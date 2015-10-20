package container

import (
	"fmt"

	"github.com/samalba/dockerclient"
)

type ContainerDesc struct {
	*dockerclient.ContainerConfig
}

func NewContainerDesc() *ContainerDesc {
	config := &dockerclient.ContainerConfig{}
	config.ExposedPorts = make(map[string]struct{})
	config.HostConfig.PortBindings = make(map[string][]dockerclient.PortBinding)
	config.HostConfig.Binds = []string{}
	return &ContainerDesc{
		config,
	}
}

func (c *ContainerDesc) SetRestartPolicy(pn string, count int) {
	c.HostConfig.RestartPolicy = dockerclient.RestartPolicy{
		Name:              pn,
		MaximumRetryCount: int64(count),
	}
}

func (c *ContainerDesc) AddPortBinding(ip, cPort, hPort, protoc string) {
	hostIp := ip
	if ip == "" {
		ip = "0.0.0.0"
	}
	prot := "tcp"
	if protoc != "" {
		prot = protoc
	}
	key := fmt.Sprintf("%s/%s", cPort, prot)
	c.ExposedPorts[key] = struct{}{}

	pb := dockerclient.PortBinding{
		HostIp:   hostIp,
		HostPort: hPort,
	}
	c.HostConfig.PortBindings[key] = append(c.HostConfig.PortBindings[key], pb)

}

func (c *ContainerDesc) AddBind(v string) {
	c.HostConfig.Binds = append(c.HostConfig.Binds, v)
}

func (c *ContainerDesc) SetImage(img string) {
	c.Image = img
}

func (c *ContainerDesc) SetCmd(cmds ...string) {
	c.Cmd = cmds[0:]
}
