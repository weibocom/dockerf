package machine

import (
	"fmt"

	dcluster "github.com/weibocom/dockerf/cluster"
	dopts "github.com/weibocom/dockerf/machine/opts"
	dseq "github.com/weibocom/dockerf/sequence"
)

type MachineClusterProxy struct {
	ClusterBy     string
	Driver        string
	Discovery     string
	Master        string
	GlobalOptions []string
	DriverOptions []string
	MasterOptions []string
	SlaveOptions  []string
	NoneOptions   []string
	Proxy         *MachineProxy
	Name          string
	seqs          map[string]*dseq.Seq
}

func NewMachineClusterProxy(name, clusterBy, driver, discovery, master string, globalOptions []string, driverOptions []string) *MachineClusterProxy {
	if clusterBy != "swarm" {
		panic("'" + clusterBy + "' is not supported.")
	}
	mp := NewMachineProxy(name)
	mcp := &MachineClusterProxy{
		ClusterBy:     clusterBy,
		Driver:        driver,
		Discovery:     discovery,
		GlobalOptions: globalOptions,
		Master:        master,
		DriverOptions: driverOptions,
		Proxy:         mp,
		Name:          name,
		seqs:          map[string]*dseq.Seq{},
	}
	mcp.initSequences()
	return mcp
}

func (mp *MachineClusterProxy) getMasterOptions() []string {
	masterOptions := []string{"-d", mp.Driver, "--swarm", "--swarm-master", "--swarm-discovery", mp.Discovery, "--engine-label", "role=master"}
	if len(mp.DriverOptions) > 0 {
		masterOptions = append(masterOptions, mp.DriverOptions...)
	}
	return masterOptions
}

func (mp *MachineClusterProxy) getSlaveOptions() []string {
	slaveOptions := []string{"-d", mp.Driver, "--swarm", "--swarm-discovery", mp.Discovery, "--engine-label", "role=slave"}
	if len(mp.DriverOptions) > 0 {
		slaveOptions = append(slaveOptions, mp.DriverOptions...)
	}
	return slaveOptions
}

func (mp *MachineClusterProxy) getMachineOptions() []string {
	noneOptions := []string{"-d", mp.Driver}
	if len(mp.DriverOptions) > 0 {
		noneOptions = append(noneOptions, mp.DriverOptions...)
	}
	return noneOptions
}

func (mp *MachineClusterProxy) initSequences() {
	mifs, err := mp.List()
	if err != nil {
		fmt.Printf("Failed to init machine group sequence. err:%s\n", err.Error())

		return
	}
	for _, mi := range mifs {
		mn := MachineName{}
		if mn.Parse(mi.Name) {
			seq := mp.getSeq(mn.Prefix)
			seq.Max(mn.Seq)
		}
	}
}

func (mp *MachineClusterProxy) getSeq(group string) *dseq.Seq {
	if seq, ok := mp.seqs[group]; ok {
		return seq
	} else {
		seq := &dseq.Seq{}
		mp.seqs[group] = seq
		return seq
	}
}

func (mp *MachineClusterProxy) IP(machine string) (string, error) {
	return mp.Proxy.IP(machine)
}

func (mp *MachineClusterProxy) IPs(machines []string) ([]string, error) {
	return mp.Proxy.IPs(machines)
}

func (mp *MachineClusterProxy) CreateMaster(md dcluster.MachineDescription) error {
	opts := mp.getMasterOptions()
	extOpts, err := dopts.GetOptions(mp.Driver, md)
	if err != nil {
		return err
	}
	opts = append(opts, extOpts...)
	return mp.Proxy.Create(mp.Master, mp.GlobalOptions, opts)
}

func (mp *MachineClusterProxy) generateName(group string) string {
	seq := mp.getSeq(group)
	mn := MachineName{
		Prefix: group,
		Seq:    seq.Next(),
	}
	return mn.GetName()
}

func (mp *MachineClusterProxy) CreateSlave(group string, md dcluster.MachineDescription) (string, error) {
	nodeName := mp.generateName(group)
	opts := []string{"--engine-label", "group=" + group}
	opts = append(opts, mp.getSlaveOptions()...)
	extOpts, err := dopts.GetOptions(mp.Driver, md)
	if err != nil {
		return "", err
	}
	if len(extOpts) > 0 {
		opts = append(opts, extOpts...)
	}
	return nodeName, mp.Proxy.Create(nodeName, mp.GlobalOptions, opts)
}

func (mp *MachineClusterProxy) CreateMachine(node string, md dcluster.MachineDescription, driverOptions []string) error {
	opts := mp.getMachineOptions()
	if len(driverOptions) > 0 {
		opts = append(opts, driverOptions...)
	}
	extOpts, err := dopts.GetOptions(mp.Driver, md)
	if err != nil {
		return err
	}
	if len(extOpts) > 0 {
		opts = append(opts, extOpts...)
	}

	return mp.Proxy.Create(node, mp.GlobalOptions, opts)
}

func (mp *MachineClusterProxy) Start(names ...string) ([]string, error) {
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

func (mp *MachineClusterProxy) ListByGroup(group string) ([]MachineInfo, error) {
	clusterFilter := func(mi *MachineInfo) bool {
		if mi.Master != mp.Master {
			return false
		}
		mn := MachineName{}
		if !mn.Parse(mi.Name) {
			return false
		}
		return mn.Prefix == group
	}
	return mp.Proxy.List(clusterFilter)
}

func (mp *MachineClusterProxy) List() ([]MachineInfo, error) {
	clusterFilter := func(mi *MachineInfo) bool {
		return mi.Master == mp.Master
	}
	return mp.Proxy.List(clusterFilter)
}
