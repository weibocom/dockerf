package options

import (
	"fmt"

	"github.com/weibocom/dockerf/options"
	"github.com/weibocom/dockerf/utils"
)

func (od *OptsDriver) RefreshVirtualboxOptions(opts *options.Options) error {
	cpu := opts.Int("cpu")
	if cpu <= 0 {
		cpu = 1
	}
	opts.Apply("virtualbox-cpu-count", fmt.Sprintf("%d", cpu))

	memory := opts.String("memory")
	if memory == "" {
		memory = "1g"
	}
	membytes, merr := utils.ParseCapacity(memory)
	if merr != nil {
		return fmt.Errorf("'%s' is not a valid memory option.", memory)
	}
	opts.Apply("virtualbox-memory", fmt.Sprintf("%d", membytes/1024/1024))

	disk := opts.String("disk")
	if disk == "" {
		disk = "10g"
	}
	diskbytes, derr := utils.ParseCapacity(disk)
	if derr != nil {
		return fmt.Errorf("'%s' is not a valid disk option.", disk)
	}
	opts.Apply("virtualbox-disk-size", fmt.Sprintf("%d", diskbytes/1024/1024))
	return nil
}
