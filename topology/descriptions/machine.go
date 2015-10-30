package descriptions

import (
	"fmt"

	"github.com/weibocom/dockerf/machine"
	"github.com/weibocom/dockerf/options"
)

type MachineTopology struct {
	GlobalOptions map[string]string            `yaml:"global"`
	CloudOptions  map[string]map[string]string `yaml:"clouds"`
	Descriptions  []MachineDescription         `yaml:"descriptions"`
}

type MachineDescription map[string]string

func (t *TopologyDescription) GetMachineGlobalOptions() *options.Options {
	vals := make(map[string]string, len(t.Machine.GlobalOptions))
	for k, v := range t.Machine.GlobalOptions {
		vals[k] = v
	}

	for k, v := range t.Container.EngineOptions {
		vals[fmt.Sprintf("engine-%s", k)] = v
	}
	return &options.Options{
		Values: vals,
	}
}

func (t *TopologyDescription) GetAllMachineOptions() []*machine.MachineOptions {
	mds := make([]*machine.MachineOptions, len(t.Machine.Descriptions))
	cnt := 0
	for _, d := range t.Machine.Descriptions {
		opts := t.GetMachineGlobalOptions()
		for k, v := range d {
			if v != "" {
				opts.Apply(k, v)
			}
		}
		mds[cnt] = &machine.MachineOptions{
			DriverName: opts.String("cloud-driver"),
			Options:    opts,
		}
		cnt++
	}
	return mds[0:cnt]
}

func (t *TopologyDescription) GetMachineOptionsBy(group string) *machine.MachineOptions {
	mds := t.GetAllMachineOptions()

	for _, md := range mds {
		if md.Options.String("group") == group {
			return md
		}
	}
	return nil
}
