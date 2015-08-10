package machine

import (
	"crypto/tls"
	"strings"
)

type MachineInfo struct {
	Name      string
	Active    string
	Driver    string
	State     string
	Master    string
	URL       string
	IP        string
	Group     string
	Seq       int
	TlsConfig tls.Config
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
