package machine

import (
	"path/filepath"

	"github.com/docker/machine/libmachine"
	"github.com/docker/machine/libmachine/auth"
	"github.com/docker/machine/libmachine/engine"
	"github.com/docker/machine/libmachine/swarm"
	"github.com/docker/machine/utils"
)

func (c *Cluster) Create(name, driverName string, d *MachineDesc) (*Machine, error) {
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
		Host: host,
	}
	c.Lock()
	c.machines = append(c.machines, m)
	c.Unlock()
	return m, nil
}
