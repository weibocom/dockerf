package container

import (
	"fmt"
	"net"
	"net/url"
	"strings"
	"time"

	"github.com/Sirupsen/logrus"
	"github.com/samalba/dockerclient"
	"github.com/weibocom/dockerf/machine"
)

type DockerClient struct {
	URL    string
	client *dockerclient.DockerClient
}

const (
	reqTimeout = 10
)

func NewDockerClientWithUrl(url string, m *machine.Machine) (*DockerClient, error) {
	ca := m.Host.HostOptions.AuthOptions.CaCertPath
	cert := m.Host.HostOptions.AuthOptions.ClientCertPath
	key := m.Host.HostOptions.AuthOptions.ClientKeyPath
	tls, err := loadTLSConfig(ca, cert, key, true)
	if err != nil {
		return nil, err
	}
	host := url
	if host == "" {
		host, err = m.Host.Driver.GetURL()
		if err != nil {
			return nil, err
		}
	}

	client, err := dockerclient.NewDockerClientTimeout(host, tls, time.Duration(reqTimeout*time.Second))
	if err != nil {
		return nil, err
	}
	return &DockerClient{
		URL:    host,
		client: client,
	}, nil
}

func NewDockerClient(m *machine.Machine) (*DockerClient, error) {
	return NewDockerClientWithUrl("", m)
}

func (d *DockerClient) List() ([]Container, error) {
	cs, err := d.client.ListContainers(true, false, "")
	if err != nil {
		return []Container{}, err
	}
	ret := make([]Container, len(cs))
	for idx, c := range cs {
		ret[idx].Container = c
		idx++
	}
	return ret, nil
}

func (d *DockerClient) Pull(image string, authConfig *dockerclient.AuthConfig) error {
	if !strings.Contains(image, ":") {
		image = image + ":latest"
	}
	if err := d.client.PullImage(image, authConfig); err != nil {
		return err
	}
	return nil
}

func (d *DockerClient) Run(desc *ContainerDesc, name string) (string, error) {
	var (
		err error
		id  string
	)
	if id, err = d.client.CreateContainer(desc.ContainerConfig, name); err != nil {
		if err != dockerclient.ErrImageNotFound {
			return "", err
		}
		logrus.Debugf("image not found, try to pull image:%s url:%s", desc.ContainerConfig.Image, d.URL)
		if err = d.Pull(desc.Image, nil); err != nil {
			return "", err
		}
		logrus.Debugf("creating container image:%s, name:%s", desc.ContainerConfig.Image, name)
		if id, err = d.client.CreateContainer(desc.ContainerConfig, name); err != nil {
			return "", err
		}
	}

	if err := d.client.StartContainer(id, &desc.HostConfig); err != nil {
		return id, err
	}
	d.client.ContainerChanges(id)

	return id, nil
}

func (d *DockerClient) GetByName(name string) (*Container, error) {
	if name == "" {
		return nil, nil
	}
	filter := fmt.Sprintf("{\"name\":[\"%s\"]}", name)
	encode := url.QueryEscape(filter)
	cs, err := d.client.ListContainers(true, false, encode)
	if err != nil {
		return nil, err
	}
	for idx, _ := range cs {
		rc := &Container{
			Container: cs[idx],
		}
		if rc.Name() == name {
			return rc, nil
		}
	}
	return nil, nil
}

func (d *DockerClient) GetById(id string) (*Container, error) {
	filter := fmt.Sprintf("{\"id\":[\"%s\"]}", id)
	encode := url.QueryEscape(filter)
	cs, err := d.client.ListContainers(true, false, encode)
	if err != nil {
		return nil, err
	}
	for idx, _ := range cs {
		rc := &Container{
			Container: cs[idx],
		}
		if rc.Id == id {
			return rc, nil
		}
	}
	return nil, nil
}

func (d *DockerClient) Rename(oldName, newName string) error {
	return d.client.RenameContainer(oldName, newName)
}

func (d *DockerClient) Stop(id string, force bool) error {
	err := d.client.StopContainer(id, reqTimeout)
	if err == nil || !force || !strings.Contains(err.Error(), "is paused") { // contains 'us paused' means containers is in 'paused' status
		return err
	}
	logrus.Warnf("stop container failed, it is supposed to be in an paused status, check and try to unpause/stop it again. id:%s, err:%s", id, err.Error())
	if err := d.client.UnpauseContainer(id); err != nil {
		return err
	}
	return d.client.StopContainer(id, reqTimeout)
}

func (d *DockerClient) StartMonitorEvents(cb dockerclient.Callback, ec chan error, args ...interface{}) {
	d.client.StartMonitorEvents(cb, ec, args...)
}

func (d *DockerClient) StopAllMonitorEvents() {
	d.client.StopAllMonitorEvents()
}

func (d *DockerClient) IsAvailable() bool {
	u, e := url.Parse(d.URL)
	if e != nil {
		return false
	}
	conn, e := net.Dial(u.Scheme, u.Host)
	if e != nil {
		return false
	}
	conn.Close()
	return true
}
