package cluster

import "github.com/weibocom/dockerf/machine"

func (c *Cluster) AddMaster(m *machine.Machine) error {
	return c.Driver.AddMaster(m)
}

func (c *Cluster) AddWorker(m *machine.Machine) error {
	return c.Driver.AddWorker(m)
}
