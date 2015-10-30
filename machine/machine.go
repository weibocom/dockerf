package machine

import (
	"crypto/tls"
	"strings"
	"time"

	"github.com/Sirupsen/logrus"
	"github.com/docker/machine/libmachine"
	"github.com/docker/machine/state"
)

import "github.com/weibocom/dockerf/options"

type MachineOptions struct {
	DriverName string
	Memory     int
	Disk       int
	Cpus       int

	Options *options.Options
}

type MachineInfo struct {
	Name      string
	Driver    string
	Active    string
	State     string
	Master    string
	URL       string
	IP        string
	Group     string
	Seq       int
	TlsConfig tls.Config
	Host      *libmachine.Host
}

func (mi *MachineInfo) IsMaster() bool {
	return mi.Name == mi.Master
}

func (mi *MachineInfo) IsRunning() bool {
	return strings.ToUpper(mi.State) == "RUNNING"
}

func (mi *MachineInfo) parseName() {
	group, seq := ParseMachineName(mi.Name)
	mi.Group = group
	mi.Seq = seq
}

type Machine struct {
	Host        *libmachine.Host
	CachedState state.State
	CachedIp    string
	StopTime    time.Time
}

func (m *Machine) Name() string {
	return m.Host.Name
}

func (m *Machine) Id() string {
	return ""
}

func (m *Machine) GetCachedIp() string {
	return m.CachedIp
}

func (m *Machine) LoadIp() (string, error) {
	ip, err := m.Host.Driver.GetIP()
	if err == nil {
		m.CachedIp = ip
	}
	return ip, err
}

func (m *Machine) setCachedState(s state.State) {
	if m.CachedState == state.Running && s != state.Running {
		m.StopTime = time.Now()
	}
	m.CachedState = s
}

func (m *Machine) GetCachedState() state.State {
	return m.CachedState
}

// this Method must be timeout-able
func (m *Machine) LoadState() state.State {
	s, err := m.Host.Driver.GetState()
	if err != nil {
		logrus.Errorf("loading machine '%s' state failed:%s", m.Name(), err.Error())
	}
	m.setCachedState(s)
	return s
}

// 1. start machine
// 2. load ip and cache it
// 3. load state and cache it
func (m *Machine) Start() error {
	err := m.Host.Start()
	if err == nil {
		m.setCachedState(state.Running)
		if _, e := m.LoadIp(); e != nil {
			logrus.Warnf("machine '%s' started, but ip loading failed: ", m.Name(), err.Error())
		}
	} else {
		m.setCachedState(state.Error)
	}
	return err
}

func (m *Machine) Stop() error {
	err := m.Host.Stop()
	if err == nil {
		m.setCachedState(state.Stopped)
	} else {
		m.setCachedState(state.Error)
	}
	return err
}

func (m *Machine) Remove() error {
	err := m.Host.Remove(false)
	if err == nil {
		m.setCachedState(state.None)
	} else {
		m.setCachedState(state.Error)
	}
	return err
}
