package swarm

import (
	"errors"
	"fmt"
	"net/url"
	"path"
	"strings"
	"sync"
	"time"

	"github.com/Sirupsen/logrus"
	"github.com/docker/machine/libmachine/auth"
	"github.com/docker/machine/libmachine/provision"
	"github.com/weibocom/dockerf/container"
	"github.com/weibocom/dockerf/container/cluster/drivers"
	"github.com/weibocom/dockerf/events"
	"github.com/weibocom/dockerf/machine"
	"github.com/weibocom/dockerf/options"
)

type Driver struct {
	options                *options.Options
	masterClients          []*container.DockerClient
	activeMasterClientIdx  int
	masterClientChangeChan chan *container.DockerClient
	startMonitorChan       chan bool // this chan can only write once, it will be close after read one ele from chan
	eventHandler           events.EventsHandler
	eventHandlerArgs       []interface{}
	sync.Mutex
}

const (
	driverName = "swarm"
)

func init() {
	drivers.Register(driverName, &drivers.RegisteredDriver{
		New: NewDriver,
	})
}

func NewDriver(options *options.Options) (drivers.Driver, error) {
	d := &Driver{
		options:                wrapOptions(options),
		masterClients:          []*container.DockerClient{},
		activeMasterClientIdx:  0,
		masterClientChangeChan: make(chan *container.DockerClient),
		startMonitorChan:       make(chan bool),
		eventHandler:           events.DefaultEventHandler,
	}
	go d.updateMasterClient()
	go d.monitorEvents()
	return d, nil
}

func (d *Driver) RegisterEventHandler(cb events.EventsHandler, args ...interface{}) {
	d.eventHandler = cb
	d.eventHandlerArgs = args
}

func (d *Driver) monitorEvents() {
	<-d.startMonitorChan // waiting master client available
	for {
		client := d.getActiveMasterClient()
		ec := make(chan error)
		dcb := wrapCallback(d.eventHandler, ec, d.eventHandlerArgs)
		client.StartMonitorEvents(dcb, ec, d.eventHandlerArgs)
		err := <-ec
		errStr := ""
		if err != nil {
			errStr = err.Error()
		}
		logrus.Errorf("monitor events failed on %s:%s", client.URL, errStr)
		client.StopAllMonitorEvents()
		close(ec)
		d.selectNextMasterClient()
		next := d.getActiveMasterClient()
		if !next.IsAvailable() {
			logrus.Warnf("docker daemon not responding, sleep 3 seconds and try to monitor events. url:%s", next.URL)
			time.Sleep(3 * time.Second)
		}
		if d.activeMasterClientIdx == 0 {
			time.Sleep(1 * time.Second) // prevent from exhuasted loop in un-expected circumstance
		}
	}
	logrus.Errorf("fatal err: events will never be monitored")
}

func (d *Driver) selectNextMasterClient() {
	d.Lock()
	defer d.Unlock()
	d.activeMasterClientIdx++
	if d.activeMasterClientIdx >= len(d.masterClients) { // loop: start from 0 idx
		d.activeMasterClientIdx = 0
	}
}

func (d *Driver) updateMasterClient() {
	for client := range d.masterClientChangeChan {
		existsIdx := -1
		for idx, m := range d.masterClients {
			if m.URL == client.URL {
				existsIdx = idx
			}
		}
		if existsIdx >= 0 {
			logrus.Warnf("the master client is already exists, replace the old one. URL:%s", client.URL)
			d.masterClients[existsIdx] = client
		} else {
			logrus.Infof("a new master client added. URL:%s", client.URL)
			d.masterClients = append(d.masterClients, client)
		}
		if len(d.masterClients) == 1 { // only one client
			d.startMonitorChan <- true // this chan will be close shortly
		}
	}
	logrus.Errorf("fatal: master client update complete")
}

// install swarm master on the 'm' machine
func (d *Driver) AddMaster(m *machine.Machine) error {
	cd := container.NewContainerDesc()
	cd.SetRestartPolicy("always", 0)

	p, err := provision.DetectProvisioner(m.Host.Driver)
	if err != nil {
		return err
	}
	swarmHost := d.options.String("host")
	u, err := url.Parse(swarmHost)
	if err != nil {
		return err
	}
	parts := strings.Split(u.Host, ":")
	port := parts[1]
	bip := parts[0]
	cd.AddPortBinding(bip, port, port, "tcp")

	dockerDir := p.GetDockerOptionsDir()
	b := fmt.Sprintf("%s:%s", dockerDir, dockerDir)
	cd.AddBind(b)

	cd.SetImage(d.options.String("image"))

	authOptions := setRemoteAuthOptions(p)

	cmds := []string{"manage",
		"--tlsverify",
		"--tlscacert", authOptions.CaCertRemotePath,
		"--tlscert", authOptions.ServerCertRemotePath,
		"--tlskey", authOptions.ServerKeyRemotePath,
		"-H", swarmHost,
		"--strategy", d.options.String("strategy"),
	}
	sopts := d.options.StringSlice("opt")
	if len(sopts) > 0 {
		cmds = append(cmds, sopts...)
	}
	cmds = append(cmds, d.options.String("discovery"))

	cd.SetCmd(cmds...)

	client, err := container.NewDockerClient(m)
	if err != nil {
		return err
	}

	name := "swarm-agent-master"
	master, err := client.GetByName(name)
	if err != nil {
		return err
	}
	id, err := checkContainerStart(name, master, cd, client)
	if err != nil {
		return err
	}
	logrus.Infof("swarm master agent is running. name:%s, id:%s", name, id)

	ip := m.GetCachedIp()
	host := fmt.Sprintf("tcp://%s:%s", ip, port)
	mc, err := container.NewDockerClientWithUrl(host, m)
	if err != nil {
		return err
	}
	d.masterClientChangeChan <- mc
	logrus.Infof("swarm master added to cluster. host:%s, name:%s", host, name)
	return nil
}

// ensure container exists and running
// if container is not running and exists, stop and rename it, then running a new one
func checkContainerStart(name string, c *container.Container, desc *container.ContainerDesc, cli *container.DockerClient) (string, error) {
	if c != nil && c.IsRunning() {
		return c.Id, nil
	}
	if c == nil {
		return cli.Run(desc, name)
	}
	// container is not nil and stopped
	if err := cli.Stop(c.Id, true); err != nil {
		return "", err
	}
	newName := fmt.Sprintf("%s.%d", name, time.Now().Unix())
	if err := cli.Rename(name, newName); err != nil {
		logrus.Warnf("failed to rename swarm container. old name:%s, new name:%s, id:%s, err:%s", name, newName, c.Id, err.Error())
	}
	return cli.Run(desc, name)
}

func setRemoteAuthOptions(p provision.Provisioner) auth.AuthOptions {
	dockerDir := p.GetDockerOptionsDir()
	authOptions := p.GetAuthOptions()

	// due to windows clients, we cannot use filepath.Join as the paths
	// will be mucked on the linux hosts
	authOptions.CaCertRemotePath = path.Join(dockerDir, "ca.pem")
	authOptions.ServerCertRemotePath = path.Join(dockerDir, "server.pem")
	authOptions.ServerKeyRemotePath = path.Join(dockerDir, "server-key.pem")

	return authOptions
}

func (d *Driver) AddWorker(m *machine.Machine) error {
	cd := container.NewContainerDesc()
	cd.SetRestartPolicy("always", 0)
	cd.SetImage(d.options.String("image"))
	ip, err := m.Host.Driver.GetIP()
	if err != nil {
		return err
	}

	cmds := []string{
		"join",
		"--addr", fmt.Sprintf("%s:%d", ip, 2376),
	}
	heartbeat := d.options.String("heartbeat")
	if heartbeat != "" {
		cmds = append(cmds, "--heartbeat", heartbeat)
	}
	cmds = append(cmds, d.options.String("discovery"))
	cd.SetCmd(cmds...)
	client, err := container.NewDockerClient(m)
	if err != nil {
		return err
	}
	name := "swarm-agent-worker"
	c, err := client.GetByName(name)
	if err != nil {
		return err
	}
	id, err := checkContainerStart(name, c, cd, client)
	if err != nil {
		return err
	}
	logrus.Infof("swarm agent added to cluster. ip:%s, name:%s, id:%s", ip, name, id)
	return nil
}

func (d *Driver) getActiveMasterClient() *container.DockerClient {
	if d.activeMasterClientIdx >= len(d.masterClients) {
		return nil
	}
	return d.masterClients[d.activeMasterClientIdx]
}

func (d *Driver) Run(cd *container.ContainerDesc, name string) (string, error) {
	client := d.getActiveMasterClient()
	if client == nil {
		return "", errors.New("swarm master not added to cluster, call AddMaster first.")
	}
	return client.Run(cd, name)
}
