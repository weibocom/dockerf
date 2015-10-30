package topology

import (
	"github.com/Sirupsen/logrus"
	"github.com/weibocom/dockerf/machine"
)

func (t *Topology) CreateMachine(name, driverName string, d *machine.MachineOptions) (*machine.Machine, error) {
	m, err := t.machineCluster.Create(name, driverName, d)
	if err == nil {
		if d.Options.Bool("master") {
			logrus.Debugf("adding machine '%s' to container cluster as a master", name)
			if err := t.containerCluster.AddMaster(m); err != nil {
				logrus.Warnf("failed to add machine '%s' to container cluster as a master:%s", name, err.Error())
			}
		} else {
			logrus.Debugf("adding machine '%s' to container cluster as a worker", name)
			if err := t.containerCluster.AddWorker(m); err != nil {
				logrus.Warnf("failed to add machine '%s' to container cluster as a worker:%s", name, err.Error())
			}
		}
	}
	return m, err
}

func (t *Topology) RestartMachine(m *machine.Machine) error {
	return t.machineCluster.Start(m)
}

func (t *Topology) RemoveMachine(m *machine.Machine) error {
	return t.machineCluster.Remove(m)
}
