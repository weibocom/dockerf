package machine

import (
	"crypto/tls"
	"strings"

	"github.com/docker/machine/libmachine"
	"github.com/docker/machine/state"
)

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

type State state.State

type Machine struct {
	Host *libmachine.Host
}

func (m *Machine) Name() string {
	return m.Host.Name
}

func (m *Machine) Id() string {
	return ""
}

func (m *Machine) GetIP() (string, error) {
	return m.Host.Driver.GetIP()
}

func (m *Machine) State() (State, error) {
	s, err := m.Host.Driver.GetState()
	return State(s), err
}

func (mi *Machine) IsRunning(s State) bool {
	return s == State(state.Running)
}

func (m *Machine) Start() error {
	s, err := m.State()
	if err != nil {
		return err
	}
	if m.IsRunning(s) {
		return nil
	}
	return m.Host.Start()
}

func (m *Machine) Stop() error {
	s, err := m.State()
	if err != nil {
		return err
	}
	if !m.IsRunning(s) {
		return nil
	}
	return m.Host.Stop()
}

func (m *Machine) Remove() error {
	return m.Host.Remove(false)
}
