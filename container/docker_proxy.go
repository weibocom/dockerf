package container

import (
	"errors"
	"flag"
	"fmt"
	"strings"
	"time"

	"github.com/samalba/dockerclient"
)

type DockerProxy struct {
	client *dockerclient.DockerClient
}

const (
	requestTimeout = 10
)

func NewDockerProxy(tlsConfig string) (*DockerProxy, error) {
	tlsFlagset := flag.NewFlagSet("tls-docker-machine", flag.ExitOnError)
	flTlsVerify := tlsFlagset.Bool("tlsverify", false, "Use TLS and verify the remote")
	flCa := tlsFlagset.String("tlscacert", "", "Trust certs signed only by this CA")
	flCert := tlsFlagset.String("tlscert", "", "Path to TLS certificate file")
	flKey := tlsFlagset.String("tlskey", "", "Path to TLS key file")
	host := tlsFlagset.String("H", "", "docker daemon host")
	err := tlsFlagset.Parse(strings.Split(tlsConfig, " "))
	if err != nil {
		fmt.Errorf("Fail to parse machine config, data: %s, error: %s", tlsConfig, err.Error())
		return nil, err
	}
	config, err := loadTLSConfig(normalizeString(*flCa), normalizeString(*flCert), normalizeString(*flKey), *flTlsVerify)
	if err != nil {
		fmt.Errorf("Fail to load tls config, error: %s", err.Error())
		return nil, err
	}

	client, err := dockerclient.NewDockerClientTimeout(*host, config, time.Duration(requestTimeout*time.Second))
	if err != nil {
		fmt.Errorf("Fail to build a docker client to %s, error: %s", *host, err.Error())
		return nil, err
	}
	return &DockerProxy{client}, nil
}

func normalizeString(param string) string {
	return strings.TrimPrefix(strings.TrimSuffix(param, "\""), "\"")
}

func (d *DockerProxy) RunByConfig(runConfig ContainerRunConfig) (string, error) {
	portBingds := map[string][]dockerclient.PortBinding{}
	exposedPorts := map[string]struct{}{}
	for _, port := range runConfig.PortBindings {
		pb := dockerclient.PortBinding{}
		pb.HostIp = "0.0.0.0"
		pb.HostPort = fmt.Sprintf("%d", port.GetHostPort())
		key := fmt.Sprintf("%d/%s", port.ContainerPort, port.Protocal)
		portBingds[key] = []dockerclient.PortBinding{pb}

		exposedPorts[key] = struct{}{}
	}

	config := &dockerclient.ContainerConfig{}
	config.Image = runConfig.Image
	config.Env = runConfig.Envs
	config.Cmd = runConfig.Cmds
	config.Hostname = runConfig.Hostname

	config.ExposedPorts = exposedPorts

	hostConfig := &dockerclient.HostConfig{}

	hostConfig.PortBindings = portBingds
	hostConfig.Binds = runConfig.Bindings
	hostConfig.Dns = runConfig.DNS
	hostConfig.RestartPolicy = dockerclient.RestartPolicy{
		Name:              runConfig.RestartPolicy.Name,
		MaximumRetryCount: int64(runConfig.RestartPolicy.MaxTry),
	}

	config.HostConfig = *hostConfig

	cid, err := d.CreateContainer(config, runConfig.Name)
	if err != nil {
		return "", errors.New(fmt.Sprintf("Failed to create a container. name: %s, error: %s", runConfig.Name, err.Error()))
	}
	fmt.Printf("Container created. name:%s, id:%s\n", runConfig.Name, cid)

	if err := d.StartContainer(cid, hostConfig); err != nil {
		fmt.Printf("Failed to start container. name:%s, id:%s.\n", runConfig.Name, cid)
		return cid, err
	}
	fmt.Printf("Container start successfully. name:%s, id:%s.\n", runConfig.Name, cid)

	return cid, nil
}

func (d *DockerProxy) Pull(image string, authConfig *dockerclient.AuthConfig) error {
	if !strings.Contains(image, ":") {
		image = image + ":latest"
	}
	if err := d.client.PullImage(image, authConfig); err != nil {
		return err
	}
	return nil
}

func (d *DockerProxy) ListContainers(all bool, size bool, filters string) ([]ContainerInfo, error) {
	containers, err := d.client.ListContainers(all, size, filters)
	if err != nil {
		return []ContainerInfo{}, err
	}
	cifs := []ContainerInfo{}
	if len(containers) > 0 {
		for _, container := range containers {
			if !strings.HasPrefix(container.Image, "swarm:") {
				info := ContainerInfo{}
				info.ID = container.Id
				info.Image = container.Image
				info.Name = container.Names
				info.Machine = resolveMachineNameFromName(info.Name[0])
				info.Status = container.Status
				cports := container.Ports
				for _, pt := range cports {
					info.IpPorts = append(info.IpPorts, IPPort{IP: pt.IP, PublicPort: pt.PublicPort, PrivatePort: pt.PrivatePort})
				}
				cn := ContainerName{}
				if !cn.Parse(info.Name[0]) {
					// fmt.Printf("Container(id:%s, name:%s) is not valid.\n", info.ID, info.Name[0])
					continue
				}
				info.Group = cn.Group
				info.Node = cn.Node
				cifs = append(cifs, info)
			}
		}
	}
	return cifs, nil
}

func (d *DockerProxy) ListAll() ([]ContainerInfo, error) {
	return d.ListContainers(true, true, "")
}

func (d *DockerProxy) ListByGroup(group string) ([]ContainerInfo, error) {
	all, err := d.ListAll()
	if err != nil {
		return []ContainerInfo{}, errors.New(fmt.Sprintf("Load container info('%s') error:%s\n", group, err.Error()))
	}
	gcs := []ContainerInfo{}
	for _, c := range all {
		if c.Group == group {
			gcs = append(gcs, c)
		}
	}
	return gcs, nil
}

func (d *DockerProxy) GetContainerByID(id string) (ContainerInfo, bool) {
	cs, err := d.ListAll()
	if err != nil {
		fmt.Printf("Load container info('%s') error:%s\n", id, err.Error())
		return ContainerInfo{}, false
	}
	for _, c := range cs {
		if c.ID == id {
			return c, true
		}
	}
	return ContainerInfo{}, false
}

func (d *DockerProxy) CreateContainer(config *dockerclient.ContainerConfig, name string) (string, error) {
	var (
		err error
		id  string
	)
	fmt.Printf("Create a new contianer. name:%s. container config:%+v\n", name, config)
	if id, err = d.client.CreateContainer(config, name); err != nil {

		if err != dockerclient.ErrNotFound {
			return "", err
		}
		if err = d.Pull(config.Image, nil); err != nil {
			return "", err
		}
		if id, err = d.client.CreateContainer(config, name); err != nil {
			return "", err
		}
	}

	return id, nil
}

func (d *DockerProxy) StartContainer(id string, hostConfig *dockerclient.HostConfig) error {
	fmt.Printf("Start container. cid:%s, host config:%+v\n", id, hostConfig)
	return d.client.StartContainer(id, hostConfig)
}

func (d *DockerProxy) StopContainer(id string) error {
	return d.client.StopContainer(id, requestTimeout)
}

func (d *DockerProxy) RemoveContainer(id string) error {
	return d.client.RemoveContainer(id, true, true)
}

func (d *DockerProxy) RestartContainer(id string) error {
	return d.client.RestartContainer(id, requestTimeout)
}

func BuildContainerConfig(image string, entrypoint []string, cmd []string) *dockerclient.ContainerConfig {
	config := &dockerclient.ContainerConfig{}
	config.Image = image
	config.Entrypoint = entrypoint
	config.Cmd = cmd

	config.AttachStderr = true
	config.AttachStdout = true
	config.AttachStdin = false

	return config
}

func resolveMachineNameFromName(name string) string {
	parts := strings.Split(name, "/")
	if len(parts) < 2 {
		panic(fmt.Sprintf("Invalid container name, name: %s", name))
	}
	return parts[1]
}
