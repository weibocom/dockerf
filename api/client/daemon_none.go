// +build !daemon

package client

import "github.com/docker/docker/cli"

const daemonUsage = ""

var daemonCli cli.Handler

// TODO: remove once `-d` is retired
func handleGlobalDaemonFlag() {}
