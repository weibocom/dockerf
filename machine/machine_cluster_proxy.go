package machine

import (
	"fmt"
	"net/url"
	"strings"

	dutils "github.com/weibocom/dockerf/utils"
)

type MachineClusterProxy struct {
	ClusterBy     string
	Driver        string
	Discovery     string
	Master        string
	MasterOptions []string
	SlaveOptions  []string
	NoneOptions   []string
	Proxy         *MachineProxy
	Name          string
}

func NewMachineClusterProxy(name, clusterBy, driver, discovery, master string, driverOptions []string) *MachineClusterProxy {
	mp := NewMachineProxy(name)
	mOpts, sOpts, nOpts := GetOptions(clusterBy, driver, discovery, driverOptions)

	mcp := &MachineClusterProxy{
		ClusterBy:     clusterBy,
		Driver:        driver,
		Discovery:     discovery,
		Master:        master,
		MasterOptions: mOpts,
		SlaveOptions:  sOpts,
		NoneOptions:   nOpts,
		Proxy:         mp,
		Name:          name,
	}
	return mcp
}

func GetOptions(clusterBy, driver, discovery string, driverOptions []string) ([]string, []string, []string) {
	if clusterBy != "swarm" {
		panic("'" + clusterBy + "' is not supported.")
	}
	masterOptions := []string{"-d", driver, "--swarm", "--swarm-master", "--swarm-discovery", discovery, "--engine-label", "role=master"}
	slaveOptions := []string{"-d", driver, "--swarm", "--swarm-discovery", discovery, "--engine-label", "role=slave"}

	noneOptions := []string{"-d", driver, "--swarm", "--swarm-discovery", discovery}

	if len(driverOptions) > 0 {
		masterOptions = append(masterOptions, driverOptions...)
		slaveOptions = append(slaveOptions, driverOptions...)
		noneOptions = append(noneOptions, driverOptions...)
	}
	return masterOptions, slaveOptions, noneOptions
}

func (mp *MachineClusterProxy) IP(machine string) (string, error) {
	return mp.Proxy.IP(machine)
}

func (mp *MachineClusterProxy) IPs(machines []string) ([]string, error) {
	return mp.Proxy.IPs(machines)
}

func (mp *MachineClusterProxy) CreateMaster() error {
	return mp.Proxy.Create(mp.Master, mp.MasterOptions...)
}

func (mp *MachineClusterProxy) CreateSlave(nodeName string) error {
	return mp.Proxy.Create(nodeName, mp.SlaveOptions...)
}

func (mp *MachineClusterProxy) CreateMachine(nodeName string, opts ...string) error {
	cOptions := []string{}
	cOptions = append(cOptions, mp.NoneOptions...)
	if len(opts) > 0 {
		cOptions = append(cOptions, opts...)
	}
	return mp.Proxy.Create(nodeName, cOptions...)
}

func (mp *MachineClusterProxy) Start(names ...string) ([]string, []error) {
	return mp.Proxy.Start(names...)
}

func (mp *MachineClusterProxy) ExecCmd(machine, command string) error {
	return mp.Proxy.ExecCmd(machine, command)
}

func (mp *MachineClusterProxy) Config() (string, error) {
	return mp.Proxy.Config(mp.Master, mp.ClusterBy)
}

func (mp *MachineClusterProxy) ConfigNode(node string) (string, error) {
	return mp.Proxy.Config(node, "")
}

func (mp *MachineClusterProxy) Destroy(names ...string) error {
	return mp.Proxy.Destroy(names...)
}

func (mp *MachineClusterProxy) List() ([]MachineInfo, error) {
	clusterFilter := func(mi *MachineInfo) bool {
		return mi.Master == mp.Master
	}
	return mp.Proxy.List(clusterFilter)
}

//解析machine ip，machine url格式：tcp://192.168.99.100:2376
func parseMachineIpFromUrl(machineUrl string) (string, error) {
	u, err := url.Parse(machineUrl)
	if err != nil {
		dutils.Error(fmt.Sprintf("Invalid machine url, url: %s, error: %s", machineUrl, err.Error()))
		return "", err
	}
	return strings.Split(u.Host, ":")[0], nil
}
