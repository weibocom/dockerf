package topology

import (
	"fmt"
	"io/ioutil"

	"github.com/Sirupsen/logrus"
	"github.com/weibocom/dockerf/container/cluster"
	"github.com/weibocom/dockerf/machine"
	"github.com/weibocom/dockerf/options"

	"gopkg.in/yaml.v2"
)

type Topology struct {
	machineCluster   *machine.Cluster
	containerCluster *cluster.Cluster
	description      *TopologyDescription
}

func NewTopology(path string) (*Topology, error) {
	data, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}
	td := &TopologyDescription{}
	if err := yaml.Unmarshal(data, td); err != nil {
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

	m, err := getMasterMachine(driver, clusterOpts, mcluster, td)
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
	return t, nil
}

func getMasterMachine(driver string, clusterOpts *options.Options, mcluster *machine.Cluster, td *TopologyDescription) (*machine.Machine, error) {
	var m *machine.Machine
	masterName := clusterOpts.String("master-machine-name")
	if masterName != "" {
		exists := false
		m, exists = mcluster.Get(masterName)
		if !exists {
			logrus.Infof("'%s' master machine is not exists, crating a new one. name:%s", driver, masterName)
			masterGroup := clusterOpts.String("master-machine-group")
			md, err := td.GetMachineDescription(masterGroup)
			if err != nil {
				return nil, err
			}
			m, err = mcluster.Create(masterName, md.DriverName, md)
			if err != nil {
				return nil, err
			}
		} else {
			s, err := m.State()
			if err != nil {
				return nil, err
			}
			if !m.IsRunning(s) {
				return nil, fmt.Errorf("master machine '%s' exists, but not started.", masterName)
			}
		}
	} else {
		logrus.Warnf("master name not specified in container cluster.")
	}
	return m, nil
}
