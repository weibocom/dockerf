package topology

import (
	"fmt"

	"github.com/Sirupsen/logrus"
	"github.com/weibocom/dockerf/container/cluster"
	"github.com/weibocom/dockerf/machine"
	"github.com/weibocom/dockerf/options"
	"github.com/weibocom/dockerf/topology/descriptions"
)

type Topology struct {
	machineCluster   *machine.Cluster
	containerCluster *cluster.Cluster
	description      *descriptions.TopologyDescription
}

func NewTopology(path string) (*Topology, error) {
	td, err := descriptions.NewTopologyDescription(path)
	if err != nil {
		return nil, err
	}
	mcluster, err := machine.NewCluster(td.GetMachineGlobalOptions())
	if err != nil {
		return nil, err
	}
	clusterOpts := td.GetClusterOptions()
	driver := clusterOpts.String("driver")
	if driver == "" {
		return nil, fmt.Errorf("container cluster driver option missed.")
	}

	m, err := getMasterMachine(clusterOpts, mcluster, td)
	if err != nil {
		return nil, err
	}

	ccluster, err := cluster.NewCluster(driver, clusterOpts, m)
	if err != nil {
		return nil, err
	}

	t := &Topology{
		machineCluster:   mcluster,
		containerCluster: ccluster,
		description:      td,
	}
	ccluster.RegisterEventHandler(TopologyEventsHandler, t)
	return t, nil
}

func getMasterMachine(clusterOpts *options.Options, mcluster *machine.Cluster, td *descriptions.TopologyDescription) (*machine.Machine, error) {
	var m *machine.Machine
	masterName := clusterOpts.String("master-machine-name")
	if masterName != "" {
		exists := false
		m, exists = mcluster.Get(masterName)
		if !exists {
			logrus.Infof("cluster master machine is not exists, crating a new one. name:%s", masterName)
			masterGroup := clusterOpts.String("master-machine-group")
			md := td.GetMachineOptionsBy(masterGroup)
			if md == nil {
				return nil, fmt.Errorf("no machine options found for master '%s'", masterGroup)
			}
			m, err := mcluster.Create(masterName, md.DriverName, md)
			if err != nil {
				return nil, err
			}
			return m, nil
		} else {
			s := m.GetCachedState()
			if !machine.IsRunning(s) {
				logrus.Infof("master machine '%s' is '%s', try to restart.", masterName, s.String())
				if err := m.Start(); err != nil {
					return nil, fmt.Errorf("master machine '%s' exists, but start failed.", masterName)
				}
			}
			return m, nil
		}
	} else {
		logrus.Warnf("master name not specified in container cluster.")
	}
	return m, nil
}
