package cluster

import (
	"github.com/weibocom/dockerf/container"
	"github.com/weibocom/dockerf/container/cluster/drivers"
	"github.com/weibocom/dockerf/machine"
	"github.com/weibocom/dockerf/options"
)

type Cluster struct {
	Containers []*container.Container
	Driver     drivers.Driver
}

func NewCluster(driver string, options *options.Options, master *machine.Machine) (*Cluster, error) {
	d, err := drivers.NewDriver(driver, options)
	if err != nil {
		return nil, err
	}
	if master != nil {
		if err := d.AddMaster(master); err != nil {
			return nil, err
		}
	}
	return &Cluster{
		Containers: []*container.Container{},
		Driver:     d,
	}, nil
}
