package cluster

import "github.com/weibocom/dockerf/container"

func (c *Cluster) Run(desc *container.ContainerDesc, name string) (string, error) {
	return c.Driver.Run(desc, name)
}
