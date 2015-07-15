package client

import (
	"os"
	"path"

	dmachine "github.com/weibocom/dockerf/machine"
)

func (dcli *DockerfCli) CmdMachine(args ...string) error {
	appName := path.Base(os.Args[0]) + " machine"
	app := dmachine.NewMachineApp(appName)
	runArgs := []string{appName}
	runArgs = append(runArgs, args...)
	return app.Run(runArgs)
}
