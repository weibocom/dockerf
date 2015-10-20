package main

import (
	"fmt"
	"os"

	flag "github.com/weibocom/dockerf/dflag"
)

var (
	flVersion    = flag.DFlag.Bool([]string{"v", "-version"}, false, "Print version information and quit")
	flDebug      = flag.DFlag.Bool([]string{"D", "-debug"}, false, "Enable debug mode")
	flLogLevel   = flag.DFlag.String([]string{"l", "-log-level"}, "info", "Set the logging level")
	flHelp       = flag.DFlag.Bool([]string{"h", "-help"}, false, "Print usage")
	flCpuProfile = flag.DFlag.String([]string{"-cpu-profile"}, "", "Write cpu profile to file.")
)

func init() {
	flag.DFlag.Usage = func() {
		fmt.Fprint(os.Stdout, "Usage: dockerf [OPTIONS] COMMAND [arg...]\n\nA docker framework that makes deploying app on a cluster of containers over cloud as simple as local.\n\nOptions:\n")

		flag.DFlag.SetOutput(os.Stdout)
		flag.DFlag.PrintDefaults()

		help := "\nCommands:\n"

		for _, command := range [][]string{
			{"cluster", "Deploy and manage a cluster of containers on containers which running on machines."},
		} {
			help += fmt.Sprintf("    %-10.10s%s\n", command[0], command[1])
		}
		help += "\nRun 'dockerf COMMAND --help' for more information on a command."
		fmt.Fprintf(os.Stdout, "%s\n", help)
	}
}
