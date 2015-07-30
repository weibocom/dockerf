package filter

import (
	"fmt"
	dcontainer "github.com/weibocom/dockerf/container"
	"strings"
)

type MachineNameFilter struct {
}

const (
	machine_name_separator = ","
)

func init() {
	registerFilter(&MachineNameFilter{})
}

func (f *MachineNameFilter) filter(filteredMachineNames string, containers []dcontainer.ContainerInfo) ([]dcontainer.ContainerInfo, error) {
	result := []dcontainer.ContainerInfo{}

	fmt.Println(fmt.Sprintf("Apply a machine filter, filter machine names: %s", filteredMachineNames))

	names := strings.Split(filteredMachineNames, machine_name_separator)

	for _, container := range containers {
		if len(container.IpPorts) > 0 {
			for _, name := range names {
				if container.Machine == strings.TrimSpace(name) {
					result = append(result, container)
				}
			}
		}
	}

	return result, nil
}

func (m *MachineNameFilter) initialize(dockerProxy *dcontainer.DockerProxy) {
}

func (m *MachineNameFilter) name() string {
	return "machine"
}
