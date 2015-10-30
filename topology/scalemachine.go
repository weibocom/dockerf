package topology

import (
	"fmt"
	"sync"
	"time"

	"github.com/Sirupsen/logrus"
	"github.com/weibocom/dockerf/machine"
	"github.com/weibocom/dockerf/utils"
)

var (
	defaultOperationTimeout = 3 * time.Minute
)

// group with value of "all" means scale all machine group
func (t *Topology) ScaleMachine(group string) error {
	mds := []*machine.MachineOptions{}
	if group == "all" {
		mds = t.description.GetAllMachineOptions()
	} else {
		md := t.description.GetMachineOptionsBy(group)
		if md != nil {
			mds = append(mds, md)
		}
	}

	if len(mds) == 0 {
		logrus.Warnf("no machine description found for group '%s'", group)
		return nil
	}
	var wg sync.WaitGroup
	for _, md := range mds {
		wg.Add(1)
		tmd := md
		go func() {
			defer func() {
				wg.Done()
				if r := recover(); r != nil {
					logrus.Errorf("panic happend:%+v", r)
				}
			}()
			t.scaleMachineBy(tmd)
		}()
	}
	wg.Wait()
	logrus.Infof("machine '%s' scaled complete.", group)
	return nil
}

func (t *Topology) scaleMachineBy(md *machine.MachineOptions) {
	group := md.Options.String("group")
	max := getMaxNum(md.Options)
	min := getMinNum(md.Options)
	if min < 0 || max < 0 {
		logrus.Debugf("'min-num' or 'max-num' options is not set for group '%s', it is supposed not to be scalable", group)
		return
	}

	if md.Options.Bool("create") {
		t.CreateMachineByGrup(group, md)
	} else {
		logrus.Debugf("'create' option is not set or set to be false for group '%s', it is supposed not to be creating new machines", group)
	}

	if md.Options.Bool("restart") {
		t.RestartMachineByGrup(group, md)
	} else {
		logrus.Debugf("'restart' option is not set or set to be false for group '%s', it is supposed not to be restarting machines", group)
	}

	if md.Options.Bool("remove") {
		t.RemoveMachineByGrup(group, md)
	} else {
		logrus.Debugf("'remove' option is not set or set to be false for group '%s', it is supposed not to be restarting machines", group)
	}
}

// machine removing policy
// 1. first, remove none-running and then running machine
// 2. reduce the number to (min+max) / 2
func (t *Topology) RemoveMachineByGrup(group string, md *machine.MachineOptions) {
	runnings, nonRunnings := []*machine.Machine{}, []*machine.Machine{}
	total := t.machineCluster.List(
		func(m *machine.Machine) bool {
			pg := ParseGroup(m.Name())
			if group != pg {
				return false
			}
			if machine.IsRunning(m.GetCachedState()) {
				runnings = append(runnings, m)
			} else {
				nonRunnings = append(nonRunnings, m)
			}
			return true
		},
	)

	max := getMaxNum(md.Options)
	min := getMinNum(md.Options)
	totalNum := len(total)
	if min < 0 {
		logrus.Debugf("'min' option not provided, it is supposed not to be remove machine '%s'", group)
		return
	}
	if min > max {
		logrus.Debugf("machine of '%s' mininal number is greater than maximal number. %d > %d ", group, min, max)
	}

	toBeRemovedNum := totalNum - (max+min)/2
	if toBeRemovedNum <= 0 {
		logrus.Debugf("no need to remove machine '%s'. total:%d, to be removed:%d", group, totalNum, toBeRemovedNum)
		return
	}
	removing := append(nonRunnings, runnings...)
	if toBeRemovedNum <= len(removing) {
		removing = removing[0:toBeRemovedNum]
	}

	leftNum := totalNum - len(removing)
	if leftNum < min {
		logrus.Warnf("removing machine process interupted, for the left num of machine is less than minimal num. group:%s, total:%d, min:%d, to be removed:%d", group, totalNum, min, toBeRemovedNum)
		return
	}

	doneChan := make(chan bool)
	timeout := parseDuration(md.Options.String("remove-timeout"), defaultOperationTimeout)

	logrus.Warnf("%d machines '%s' will be removed", len(removing), group)

	for _, m := range removing {
		tm := m
		go utils.TickRun(fmt.Sprintf("removing machine '%s' cached state:%s", tm.Name(), tm.GetCachedState().String()),
			doneChan,
			timeout,
			func() {
				err := t.RemoveMachine(tm)
				if err != nil {
					logrus.Errorf("remove machine '%s' failed:%s", tm.Name(), err.Error())
				} else {
					logrus.Infof("remove machine '%s' successfully.", tm.Name())
				}
			},
			nil)
	}

	for i := 0; i < len(removing); i++ {
		<-doneChan
	}

	return
}

// try start all not-running-state machines
func (t *Topology) RestartMachineByGrup(group string, md *machine.MachineOptions) {
	nonRunnings := t.machineCluster.List(
		func(m *machine.Machine) bool {
			pg := ParseGroup(m.Name())
			if group != pg {
				return false
			}
			return !machine.IsRunning(m.GetCachedState())
		},
	)
	if len(nonRunnings) == 0 {
		logrus.Debugf("all machine '%s' are in runnint state.", group)
		return
	}
	timeout := parseDuration(md.Options.String("restart-timeout"), defaultOperationTimeout)

	logrus.Infof("%d-non-running machines will be started '%s'", len(nonRunnings), group)
	doneChan := make(chan bool)
	for _, m := range nonRunnings {
		tm := m
		go utils.TickRun(fmt.Sprintf("restarting machine '%s'", tm.Name()),
			doneChan,
			timeout,
			func() {
				err := t.RestartMachine(tm)
				if err != nil {
					logrus.Errorf("failed to start machine '%s':", tm.Name(), err.Error())
				} else {
					logrus.Debugf("start machine '%s' successfully.", tm.Name())
				}
			},
			nil)
	}
	for i := 0; i < len(nonRunnings); i++ {
		<-doneChan
	}
	return
}

func (t *Topology) CreateMachineByGrup(group string, md *machine.MachineOptions) {
	tl := 0
	runnings := t.machineCluster.List(
		func(m *machine.Machine) bool {
			pg := ParseGroup(m.Name())
			if group != pg {
				return false
			}
			tl++
			return machine.IsRunning(m.GetCachedState())
		},
	)

	max := getMaxNum(md.Options)
	min := getMinNum(md.Options)

	if len(runnings) >= min {
		logrus.Debugf("the number of running machines is enough to meet the minimal requirements. group:%s, running:%d, mininal requirements:%d", group, len(runnings), min)
		return
	}
	if tl >= max {
		logrus.Warnf("the total number of machine '%s' reach its limits. total:%d, limits:%d", group, tl, max)
		return
	}
	createNum := min - len(runnings)
	if createNum+tl >= max {
		createNum = max - tl
	}

	logrus.Infof("%d machines will be created of group '%s'", createNum, group)

	driverName := md.Options.String("cloud-driver")
	timeout := parseDuration(md.Options.String("create-timeout"), defaultOperationTimeout)
	doneChan := make(chan bool)
	for c := 0; c < createNum; c++ {
		name := GenerateName(group)
		go utils.TickRun(
			fmt.Sprintf("creating machine name:%s driver:%s", name, driverName),
			doneChan,
			timeout,
			func() {
				_, err := t.CreateMachine(name, driverName, md)
				if err != nil {
					logrus.Errorf("machine '%s' create failed: %s", name, err.Error())
					return
				}
				logrus.Infof("machine '%s' created.", name)
			},
			nil,
		)
	}
	for c := 0; c < createNum; c++ {
		<-doneChan
	}
	logrus.Infof("creating machine process complete. group: %s", group)
	return
}

func parseDuration(str string, def time.Duration) time.Duration {
	if str == "" {
		return def
	}
	d, err := time.ParseDuration(str)
	if err != nil {
		logrus.Infof("parse time duration err:%s", err.Error())
		return def
	}
	return d
}
