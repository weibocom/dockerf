package topology

import (
	"fmt"

	"github.com/weibocom/dockerf/machine"
	"github.com/weibocom/dockerf/options"
)

type TopologyDescription struct {
	Machine   MachineTopology
	Container ContainerTopology
}

type MachineTopology struct {
	GlobalOptions map[string]string            `yaml:"global"`
	CloudOptions  map[string]map[string]string `yaml:"clouds"`
	Descriptions  []MachineDescription         `yaml:"descriptions"`
}

type MachineDescription map[string]string

type ContainerTopology struct {
	EngineOptions  map[string]string      `yaml:"engine"`
	ClusterOptions map[string]string      `yaml:"cluster"`
	Descriptions   []ContainerDescription `yaml:"descriptions"`
}

type ContainerDescription struct {
	Group           string
	Port            string
	MinNum          int `yaml:"min-num"`
	MaxNum          int `yaml:"max-num"`
	Image           string
	MachineGroup    string            `yaml:"machine-group"`
	RegisterOptions map[string]string `yaml:"register"`
}

func (t *TopologyDescription) GetMachineGlobalOptions() *options.Options {
	opts := &options.Options{
		Values: t.Machine.GlobalOptions,
	}
	for k, v := range t.Container.EngineOptions {
		opts.Apply(fmt.Sprintf("engine-%s", k), v)
	}
	return opts
}

func (t *TopologyDescription) GetClusterOptions() *options.Options {
	return &options.Options{Values: t.Container.ClusterOptions}
}

// option include:
// 1. global option
// 2. driver options
// 3. options of spcified machine group
func (t *TopologyDescription) GetMachineDescription(group string) (*machine.MachineDesc, error) {
	opts := t.GetMachineGlobalOptions()

	found := false
	for _, d := range t.Machine.Descriptions {
		if d["group"] == group {
			found = true
			for k, v := range d {
				if v != "" {
					opts.Apply(k, v)
				}
			}
			break
		}
	}
	if !found {
		return nil, fmt.Errorf("machine description of group '%s' not found.", group)
	}

	driver := opts.String("cloud-driver")
	if driver == "" {
		return nil, fmt.Errorf("cloud driver not provided for machine group '%s'.", group)
	}

	return &machine.MachineDesc{
		DriverName: driver,
		Options:    opts,
	}, nil
}
