package cluster

import (
	"io/ioutil"

	"gopkg.in/yaml.v2"
)

type OS struct {
	Provider string
	Bits     int
	Version  string
}

type Disk struct {
	Ssd  string
	Sata string
	Sas  string
}

type MachineDescription struct {
	MaxNum  int
	MinNum  int
	Os      OS
	Cpu     int
	Disk    Disk
	Memory  string
	Init    string
	Regions map[string]int
}

type ContainerDescription struct {
	Name            string
	Num             int
	Image           string
	PreStop         string
	PostStart       string
	URL             string
	Port            string
	Deps            []string
	ServiceDiscover string
	Restart         bool
	PortBinding     PortBinding
}

type ConsulServer struct {
	Service string
	Image   string
	Domain  string
	Nodes   []string
	IPs     []string
	Create  bool
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
	ClusterBy     string // such as swarm
	MasterNode    string
	Driver        string // aliyun do virtualbox
	DriverOptions []string
	Discovery     string

	Machine    MachineDescription
	Containers []ContainerDescription

	ServiceDiscovers map[string]ServiceDiscoverDiscription

	ConsulCluster ConsulDescription
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
