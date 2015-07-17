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
	MasterOptions []string
	SlaveOptions  []string
	NoneOptions   []string
	Proxy         *MachineProxy
	Name          string
	seqs          map[string]*dseq.Seq
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
		seqs:          map[string]*dseq.Seq{},
	}
	mcp.initSequences()
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

func (mp *MachineClusterProxy) CreateMaster() error {
	return mp.Proxy.Create(mp.Master, mp.MasterOptions...)
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
	opts = append(opts, mp.SlaveOptions...)
	extOpts, err := dopts.GetOptions(mp.Driver, md)
	if err != nil {
		return "", err
	}
	if len(extOpts) > 0 {
		opts = append(opts, extOpts...)
	}
	return nodeName, mp.Proxy.Create(nodeName, opts...)
}

func (mp *MachineClusterProxy) CreateMachine(group string, opts ...string) error {
	cOptions := []string{"--engine-label", "group=" + group}
	cOptions = append(cOptions, mp.NoneOptions...)
	if len(opts) > 0 {
		cOptions = append(cOptions, opts...)
	}
	nodeName := mp.generateName(group)
	return mp.Proxy.Create(nodeName, cOptions...)
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
