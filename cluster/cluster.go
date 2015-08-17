package cluster

import (
	"fmt"
	"io/ioutil"
	"strings"

	dutils "github.com/weibocom/dockerf/utils"
	"gopkg.in/yaml.v2"
)

type Disk struct {
	Type     string
	Capacity string
}

type MachineDescription struct {
	MaxNum     int
	MinNum     int
	Cpu        int
	Disk       Disk
	Memory     string
	Init       string
	Region     string
	Consul     bool
	DriverOpts []string
	Cloud      string
	Group      string
}

func (md *MachineDescription) GetCpu() int {
	if md.Cpu <= 0 {
		return 1
	}
	return md.Cpu
}

func (md *MachineDescription) GetMemInBytes() int {
	if md.Memory == "" {
		return 512 * 1024 * 1024 // default is 512m
	}
	if bytes, err := dutils.ParseCapacity(md.Memory); err != nil {
		panic(fmt.Sprintf("'%s' is not a valid memory option.", md.Memory))
	} else {
		return bytes
	}
}

func (md *MachineDescription) GetDiskCapacityInBytes() int {
	if md.Disk.Capacity == "" {
		return 20 * 1024 * 1024 * 1024 // default is 10gb
	}
	if bytes, err := dutils.ParseCapacity(md.Disk.Capacity); err != nil {
		panic(fmt.Sprintf("'%s' is not a valid ssd capacity option.", md.Disk.Capacity))
	} else {
		return bytes
	}
}

type CloudDrivers map[string]CloudDriverDescription

type MachineTopology []MachineDescription

func (ct *MachineTopology) GetDescription(group string) (*MachineDescription, bool) {
	for idx, description := range *ct {
		if description.Group == group {
			return &(*ct)[idx], true
		}
	}
	return nil, false
}

type MachineCluster struct {
	OS       string
	Cloud    CloudDrivers
	Topology MachineTopology
}

func (cds *CloudDrivers) SurportedDrivers() []string {
	names := []string{}
	for name, _ := range *cds {
		names = append(names, name)
	}
	return names
}

type ContainerTopology []ContainerDescription

func (ct *ContainerTopology) GetDescription(group string) (*ContainerDescription, bool) {
	for idx, description := range *ct {
		if description.Group == group {
			return &(*ct)[idx], true
		}
	}
	return nil, false
}

type ContainerCluster struct {
	Topology ContainerTopology
}

const (
	ContainerDescription_TYPE_SD = 1
	ContainerDescription_TYPE_BZ = 2
)

type ContainerDescription struct {
	Num             int
	Image           string
	PreStop         string
	PostStart       string
	URL             string
	Port            string
	Deps            []string
	ServiceDiscover string
	Restart         bool
	Machine         string
	PortBinding     PortBinding
	Volums          []string
	Group           string
	Env             []string
	DepLevel        int
	Type            int
}

type SortContainerDescriptionDescByLevel []ContainerDescription

func (s SortContainerDescriptionDescByLevel) Len() int {
	return len(s)
}
func (s SortContainerDescriptionDescByLevel) Swap(i, j int) {
	s[i], s[j] = s[j], s[i]
}
func (s SortContainerDescriptionDescByLevel) Less(i, j int) bool {
	return s[i].DepLevel > s[j].DepLevel
}

type CloudDriverDescription struct {
	Options       string
	GlobalOptions string
	Default       bool
}

func (cdd *CloudDriverDescription) GetOptions() []string {
	if cdd.Options == "" {
		return []string{}
	}
	return strings.Split(cdd.Options, " ")
}

func (cdd *CloudDriverDescription) GetGlobalOptions() []string {
	if cdd.GlobalOptions == "" {
		return []string{}
	}
	return strings.Split(cdd.GlobalOptions, " ")
}

type ConsulServer struct {
	Service string
	Image   string
	Domain  string
	Nodes   []string
	IPs     []string
	Create  bool
	Machine string
}

type ConsulAgent struct {
	Image string
}

type ConsulRegistrator struct {
	Image string
}

type ConsulDescription struct {
	Server      ConsulServer
	Agent       ConsulAgent
	Registrator ConsulRegistrator
}

type ServiceDiscoverDiscription map[string]string

type Cluster struct {
	ClusterBy   string // such as swarm
	Master      string
	MasterGroup string
	Discovery   string

	Machine   MachineCluster
	Container ContainerCluster
	// Containers []ContainerDescription

	ServiceDiscover map[string]ServiceDiscoverDiscription
	ConsulCluster   ConsulDescription
}

func NewCluster(configFilePos string) (*Cluster, error) {
	c := &Cluster{}
	b, err := ioutil.ReadFile(configFilePos)
	if err != nil {
		return nil, err
	}
	err = yaml.Unmarshal(b, c)
	if err != nil {
		return nil, err
	}
	return c, nil
}
