package container

import (
	"strings"

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

func (info *ContainerInfo) IsUp() bool {
	return strings.HasPrefix(info.Status, "Up")
}

type ContainerRunConfig struct {
	Name         string
	Image        string
	PortBindings []dcluster.PortBinding
	Envs         []string
	Cmds         []string
	Hostname     string
	Bindings     []string
	DNS          []string
}
