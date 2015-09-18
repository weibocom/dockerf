package machine

import (
	"fmt"
	"strconv"
	"strings"

	log "github.com/Sirupsen/logrus"
	dcluster "github.com/weibocom/dockerf/cluster"
	dopts "github.com/weibocom/dockerf/machine/opts"
	dseq "github.com/weibocom/dockerf/sequence"
)

type MachineClusterProxy struct {
	ClusterBy     string
	Discovery     string
	Master        string
	MasterOptions []string
	SlaveOptions  []string
	NoneOptions   []string
	Proxy         *MachineProxy
	drivers       map[string]dcluster.CloudDriverDescription
	seqs          map[string]*dseq.Seq
}

func NewMachineClusterProxy(name, clusterBy, discovery, master string, drivers dcluster.CloudDrivers) *MachineClusterProxy {
	if clusterBy != "swarm" {
		panic("'" + clusterBy + "' is not supported.")
	}
	mp := NewMachineProxy(name)
	mcp := &MachineClusterProxy{
		ClusterBy: clusterBy,
		Discovery: discovery,
		Master:    master,
		Proxy:     mp,
		seqs:      map[string]*dseq.Seq{},
		drivers:   drivers,
	}
	mcp.initSequences()
	return mcp
}

func (mp *MachineClusterProxy) getDriver(md dcluster.MachineDescription) *dcluster.CloudDriverDescription {
	name := mp.getDriverName(md)
	if driver, exists := mp.drivers[name]; !exists {
		panic(fmt.Sprintf("Cloud driver '%s' is not supported", name))
	} else {
		return &driver
	}
}

func (mp *MachineClusterProxy) getDriverName(md dcluster.MachineDescription) string {
	if md.Cloud != "" {
		return md.Cloud
	}
	for name, driver := range mp.drivers {
		if driver.Default {
			return name
		}
	}
	panic(fmt.Sprintf("No driver property or default driver provided."))
}

func (mp *MachineClusterProxy) getMasterOptions(md dcluster.MachineDescription) []string {
	masterOptions := []string{"-d", mp.getDriverName(md),
		"--swarm",
		"--swarm-master",
		"--swarm-discovery", mp.Discovery,
		"--engine-label", "role=master",
		"--swarm-opt", "filter=port",
		"--swarm-opt", "filter=affinity",
		"--swarm-opt", "filter=constraint",
	}
	masterOptions = append(masterOptions, mp.getDriver(md).GetOptions()...)
	return masterOptions
}

func (mp *MachineClusterProxy) getSlaveOptions(md dcluster.MachineDescription) []string {
	slaveOptions := []string{"-d", mp.getDriverName(md),
		"--swarm",
		"--swarm-discovery", mp.Discovery,
		"--engine-label", "role=slave",
		// "--swarm-opt", "debug",
	}
	slaveOptions = append(slaveOptions, mp.getDriver(md).GetOptions()...)
	return slaveOptions
}

func (mp *MachineClusterProxy) getMachineOptions(md dcluster.MachineDescription) []string {
	noneOptions := []string{"-d", mp.getDriverName(md)}
	noneOptions = append(noneOptions, mp.getDriver(md).GetOptions()...)
	return noneOptions
}

func (mp *MachineClusterProxy) initSequences() {
	mifs, err := mp.List()
	if err != nil {
		fmt.Printf("Failed to init machine group sequence. err:%s\n", err.Error())
	}
	for _, mi := range mifs {
		seq := mp.getSeq(mi.Group)
		seq.Max(mi.Seq)
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
	if len(md.UnmanagedIps) > 0 {
		masterAddress := md.UnmanagedIps[0]
		if len(md.UnmanagedIps) > 1 {
			log.Warnf("Multi master address is not currently supported, use first address %s instead...", masterAddress)
		}
		return mp.CreateUnmanagedMaster(md, masterAddress)
	}
	opts := mp.getMasterOptions(md)
	extOpts, err := mp.GetOptionByDescription(md)
	if err != nil {
		return err
	}
	opts = append(opts, extOpts...)
	return mp.Proxy.Create(mp.Master, mp.getDriver(md).GetGlobalOptions(), opts)
}

func (mp *MachineClusterProxy) generateName(group string) string {
	seq := mp.getSeq(group)
	return FormateMachineName(group, seq.Next())
}

func (mp *MachineClusterProxy) CreateSlave(md dcluster.MachineDescription) (string, error) {
	nodeName := mp.generateName(md.Group)
	opts := []string{"--engine-label", "group=" + md.Group}
	opts = append(opts, mp.getSlaveOptions(md)...)
	extOpts, err := mp.GetOptionByDescription(md)
	if err != nil {
		return "", err
	}
	if len(extOpts) > 0 {
		opts = append(opts, extOpts...)
	}
	return nodeName, mp.Proxy.Create(nodeName, mp.getDriver(md).GetGlobalOptions(), opts)
}

func (mp *MachineClusterProxy) CreateUnmanagedMaster(md dcluster.MachineDescription, address string) error {
	opts := mp.getMasterOptions(md)
	opts = append(opts, mp.getGenericOptions(address)...)
	return mp.Proxy.Create(mp.Master, mp.getDriver(md).GetGlobalOptions(), opts)
}

func (mp *MachineClusterProxy) CreateUnmanagedSlave(md dcluster.MachineDescription, address string) (string, error) {
	nodeName := mp.generateName(md.Group)
	opts := []string{"--engine-label", "group=" + md.Group}
	opts = append(opts, mp.getSlaveOptions(md)...)
	opts = append(opts, mp.getGenericOptions(address)...)

	return nodeName, mp.Proxy.Create(nodeName, mp.getDriver(md).GetGlobalOptions(), opts)
}

func (mp *MachineClusterProxy) getUnmanagedMachineIpPort(address string) (string, int) {
	ipPort := strings.Split(address, ":")
	ip := ipPort[0]
	port := 22
	if len(ipPort) > 1 {
		p, err := strconv.Atoi(ipPort[1])
		if err != nil {
			log.Warnf("Fail to parse unmanged slave address %s, use a default ssh port instead... ", address)
		} else {
			port = p
		}
	}
	return ip, port
}

func (mp *MachineClusterProxy) getGenericOptions(address string) []string {
	ip, port := mp.getUnmanagedMachineIpPort(address)
	options := []string{}
	options = append(options, "--generic-ip-address", ip)
	options = append(options, "--generic-ssh-port", strconv.Itoa(port))

	return options
}

func (mp *MachineClusterProxy) CreateMachine(node string, md dcluster.MachineDescription, driverOptions []string) error {
	opts := mp.getMachineOptions(md)
	if len(driverOptions) > 0 {
		opts = append(opts, driverOptions...)
	}
	extOpts, err := mp.GetOptionByDescription(md)
	if err != nil {
		return err
	}
	if len(extOpts) > 0 {
		opts = append(opts, extOpts...)
	}

	return mp.Proxy.Create(node, mp.getDriver(md).GetGlobalOptions(), opts)
}

func (mp *MachineClusterProxy) GetOptionByDescription(md dcluster.MachineDescription) ([]string, error) {
	driver := mp.getDriverName(md)
	return dopts.GetOptions(driver, md)
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
		return mi.Group == group
	}
	return mp.Proxy.List(clusterFilter)
}

func (mp *MachineClusterProxy) List() ([]MachineInfo, error) {
	clusterFilter := func(mi *MachineInfo) bool {
		return mi.Master == mp.Master
	}
	return mp.Proxy.List(clusterFilter)
}
