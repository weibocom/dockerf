package drivers

import (
	"fmt"

	"github.com/weibocom/dockerf/container"
	"github.com/weibocom/dockerf/events"
	"github.com/weibocom/dockerf/machine"
	"github.com/weibocom/dockerf/options"
)

type RegisteredDriver struct {
	New func(options *options.Options) (Driver, error)
}

var (
	drivers map[string]*RegisteredDriver
)

func init() {
	drivers = make(map[string]*RegisteredDriver)
}

type Driver interface {
	AddMaster(m *machine.Machine) error
	AddWorker(m *machine.Machine) error
	RegisterEventHandler(cb events.EventsHandler, args ...interface{})
	Run(d *container.ContainerDesc, name string) (string, error)
}

func NewDriver(name string, options *options.Options) (Driver, error) {
	d, exists := drivers[name]
	if !exists {
		return nil, fmt.Errorf("hosts: Unknown driver %q", name)
	}
	return d.New(options)
}

// Register a driver
func Register(name string, r *RegisteredDriver) error {
	if _, exists := drivers[name]; exists {
		return fmt.Errorf("Name already registered %s", name)
	}
	drivers[name] = r
	return nil
}
