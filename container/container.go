package container

import (
	"strings"

	"github.com/samalba/dockerclient"
	dcluster "github.com/weibocom/dockerf/cluster"
)

type IPPort struct {
	IP          string
	PublicPort  int
	PrivatePort int
}

type ContainerInfo struct {
	ID      string
	Image   string
	Status  string
	Name    []string
	Machine string
	IpPorts []IPPort
	Group   string
	Node    string
}

type RestartPolicy struct {
	Name   string
	MaxTry int
}

func (info *ContainerInfo) IsUp() bool {
	return strings.HasPrefix(info.Status, "Up")
}

type ContainerRunConfig struct {
	Name          string
	Image         string
	PortBindings  []dcluster.PortBinding
	Envs          []string
	Cmds          []string
	Hostname      string
	Bindings      []string
	DNS           []string
	RestartPolicy RestartPolicy
}

type Container struct {
	dockerclient.Container
}

func (c *Container) Name() string {
	if len(c.Names) == 0 {
		return ""
	}
	name := c.Names[0]
	idx := strings.IndexAny(name, "/")
	if idx < 0 {
		return name
	}
	return name[idx+1:]
}

func (c *Container) IsRunning() bool {
	lower := strings.ToLower(c.Status)
	splits := strings.SplitN(lower, " ", 2)
	status := splits[0]
	switch status {
	case "running", "restarting":
		return true
	case "up":
		if strings.Contains(lower, "paused") {
			return false
		}
		return true
	default:
		return false
	}
	return false
}

func (c *Container) IsPaused() bool {
	lower := strings.ToLower(c.Status)
	return strings.Contains(lower, "paused")
}
