package runconfig

import (
	"strings"

	"github.com/Sirupsen/logrus"
	"github.com/docker/docker/nat"
)

func Merge(userConf, imageConf *Config) error {
	if userConf.User == "" {
		userConf.User = imageConf.User
	}
	if len(userConf.ExposedPorts) == 0 {
		userConf.ExposedPorts = imageConf.ExposedPorts
	} else if imageConf.ExposedPorts != nil {
		if userConf.ExposedPorts == nil {
			userConf.ExposedPorts = make(nat.PortSet)
		}
		for port := range imageConf.ExposedPorts {
			if _, exists := userConf.ExposedPorts[port]; !exists {
				userConf.ExposedPorts[port] = struct{}{}
			}
		}
	}

	if len(userConf.PortSpecs) > 0 {
		if userConf.ExposedPorts == nil {
			userConf.ExposedPorts = make(nat.PortSet)
		}
		ports, _, err := nat.ParsePortSpecs(userConf.PortSpecs)
		if err != nil {
			return err
		}
		for port := range ports {
			if _, exists := userConf.ExposedPorts[port]; !exists {
				userConf.ExposedPorts[port] = struct{}{}
			}
		}
		userConf.PortSpecs = nil
	}
	if len(imageConf.PortSpecs) > 0 {
		// FIXME: I think we can safely remove this. Leaving it for now for the sake of reverse-compat paranoia.
		logrus.Debugf("Migrating image port specs to container: %s", strings.Join(imageConf.PortSpecs, ", "))
		if userConf.ExposedPorts == nil {
			userConf.ExposedPorts = make(nat.PortSet)
		}

		ports, _, err := nat.ParsePortSpecs(imageConf.PortSpecs)
		if err != nil {
			return err
		}
		for port := range ports {
			if _, exists := userConf.ExposedPorts[port]; !exists {
				userConf.ExposedPorts[port] = struct{}{}
			}
		}
	}

	if len(userConf.Env) == 0 {
		userConf.Env = imageConf.Env
	} else {
		for _, imageEnv := range imageConf.Env {
			found := false
			imageEnvKey := strings.Split(imageEnv, "=")[0]
			for _, userEnv := range userConf.Env {
				userEnvKey := strings.Split(userEnv, "=")[0]
				if imageEnvKey == userEnvKey {
					found = true
				}
			}
			if !found {
				userConf.Env = append(userConf.Env, imageEnv)
			}
		}
	}

	if userConf.Labels == nil {
		userConf.Labels = map[string]string{}
	}
	if imageConf.Labels != nil {
		for l := range userConf.Labels {
			imageConf.Labels[l] = userConf.Labels[l]
		}
		userConf.Labels = imageConf.Labels
	}

	if userConf.Entrypoint.Len() == 0 {
		if userConf.Cmd.Len() == 0 {
			userConf.Cmd = imageConf.Cmd
		}

		if userConf.Entrypoint == nil {
			userConf.Entrypoint = imageConf.Entrypoint
		}
	}
	if userConf.WorkingDir == "" {
		userConf.WorkingDir = imageConf.WorkingDir
	}
	if len(userConf.Volumes) == 0 {
		userConf.Volumes = imageConf.Volumes
	} else {
		for k, v := range imageConf.Volumes {
			userConf.Volumes[k] = v
		}
	}
	return nil
}
