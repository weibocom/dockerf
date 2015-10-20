package machine

import "github.com/weibocom/dockerf/options"

type MachineDesc struct {
	DriverName string
	Memory     int
	Disk       int
	Cpus       int

	Options *options.Options
}
