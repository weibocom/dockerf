package opts

import (
	"fmt"

	dcluster "github.com/weibocom/dockerf/cluster"
)

func (od *OptsDriver) GetVirtualboxOptions(md dcluster.MachineDescription) ([]string, error) {
	options := []string{
		"--virtualbox-cpu-count",
		fmt.Sprintf("%d", md.GetCpu()),
		"--virtualbox-memory",
		fmt.Sprintf("%d", md.GetMemInBytes()/1024/1024),
		"--virtualbox-disk-size",
		fmt.Sprintf("%d", md.GetDiskCapacityInBytes()/1024/1024),
	}
	return options, nil
}
