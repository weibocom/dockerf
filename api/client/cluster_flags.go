package client

import (
	"fmt"
	"os"

	flag "github.com/docker/docker/pkg/mflag"
)

var ClusterFlag = flag.NewFlagSet("cluster", flag.ExitOnError)

var flHelp = ClusterFlag.Bool([]string{"h", "-help"}, false, "Print usage")

func init() {

	ClusterFlag.Usage = func() {
		fmt.Fprint(os.Stdout, "Usage: dockerf cluster [COMMAND] [args]\n\nManage the whole cluster of containers.\n\nOptions:\n")

		ClusterFlag.SetOutput(os.Stdout)
		ClusterFlag.PrintDefaults()

		help := "\nCommands:\n"

		for _, command := range [][]string{
			{"deploy", "Deploy the container to the whole cluster of machines"},
			{"resize", "Create or destroy machines as needed"},
			{"start", "Start specified containers and machines"},
			{"stop", "Stop specified containers and machines"},
			{"restart", "Restart specified containers and machines"},
		} {
			help += fmt.Sprintf("    %-10.10s%s\n", command[0], command[1])
		}
		help += "\nRun 'dockerf cluster COMMAND --help' for more information on a command."
		fmt.Fprintf(os.Stdout, "%s\n", help)
	}
}
