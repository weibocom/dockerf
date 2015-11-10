package machine

import (
	"fmt"
	"path/filepath"
	"time"

	"github.com/docker/machine/libmachine"
	"github.com/docker/machine/libmachine/auth"
	"github.com/docker/machine/libmachine/engine"
	"github.com/docker/machine/libmachine/swarm"
	"github.com/docker/machine/state"
	"github.com/docker/machine/utils"
	"github.com/weibocom/dockerf/machine/options"
)

func (c *Cluster) Create(name, driverName string, d *MachineOptions) (*Machine, error) {
	if _, ok := c.machines[name]; ok {
		return nil, fmt.Errorf("machine name '%s' exists, remove it first or provide another name.")
	}
	d.Options.Apply("engine-label", d.Options.String("engine-label")+" group="+d.Options.String("group"))

	if err := options.RefreshOptions(driverName, d.Options); err != nil {
		return nil, err
	}

	eo := getEngineOptions(d.Options)

	eopts := engine.EngineOptions(*eo)
	hostOptions := &libmachine.HostOptions{
		AuthOptions: &auth.AuthOptions{
			CaCertPath:     c.authOptions.CaCertPath,
			PrivateKeyPath: c.authOptions.CaKeyPath,
			ClientCertPath: c.authOptions.ClientCertPath,
			ClientKeyPath:  c.authOptions.ClientKeyPath,
			ServerCertPath: filepath.Join(utils.GetMachineDir(), name, "server.pem"),
			ServerKeyPath:  filepath.Join(utils.GetMachineDir(), name, "server-key.pem"),
		},
		EngineOptions: &eopts,
		SwarmOptions: &swarm.SwarmOptions{
			IsSwarm: false,
		},
	}

	host, err := c.provider.Create(name, driverName, hostOptions, d.Options)
	if err != nil {
		return nil, err
	}
	m := &Machine{
		Host:     host,
		StopTime: time.Now(),
	}
	c.Lock()
	m.setCachedState(state.Running)
	m.LoadIp()
	c.machines[name] = m
	c.Unlock()
	return m, nil
}
