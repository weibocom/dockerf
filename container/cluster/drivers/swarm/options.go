package swarm

import (
	"github.com/codegangsta/cli"
	"github.com/weibocom/dockerf/options"
)

func wrapOptions(opts *options.Options) *options.Options {
	extFlags := []cli.Flag{
		cli.StringFlag{
			Name:   "swarm-image",
			Usage:  "Specify Docker image to use for Swarm",
			Value:  "swarm:latest",
			EnvVar: "MACHINE_SWARM_IMAGE",
		},
		cli.StringFlag{
			Name:   "engine-install-url",
			Usage:  "Custom URL to use for engine installation",
			Value:  "https://get.docker.com",
			EnvVar: "MACHINE_DOCKER_INSTALL_URL",
		},
		cli.StringFlag{
			Name:  "swarm-strategy",
			Usage: "Define a default scheduling strategy for Swarm",
			Value: "spread",
		},
		cli.StringFlag{
			Name:  "swarm-host",
			Usage: "ip/socket to listen on for Swarm master",
			Value: "tcp://0.0.0.0:3376",
		},
	}
	opts.Flags = append(opts.Flags, extFlags...)
	return opts
}
