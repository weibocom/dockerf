package descriptions

import (
	"io/ioutil"

	"gopkg.in/yaml.v2"

	"github.com/weibocom/dockerf/options"
)

type TopologyDescription struct {
	Machine   MachineTopology
	Container ContainerTopology
}

func NewTopologyDescription(path string) (*TopologyDescription, error) {
	data, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}
	td := &TopologyDescription{}
	if err := yaml.Unmarshal(data, td); err != nil {
		return nil, err
	}
	return td, nil
}

func (t *TopologyDescription) GetClusterOptions() *options.Options {
	return &options.Options{Values: t.Container.ClusterOptions}
}
