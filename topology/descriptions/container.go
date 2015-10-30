package descriptions

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
