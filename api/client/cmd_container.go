package client

import (
	"fmt"
	"os"
	"path"
	"sort"
	"strings"

	log "github.com/Sirupsen/logrus"
	"github.com/docker/docker/api/client"
	"github.com/docker/docker/autogen/dockerversion"
	"github.com/docker/docker/cli"
	flag "github.com/docker/docker/pkg/mflag"
	"github.com/docker/docker/pkg/term"
	"github.com/docker/docker/utils"
	dmachine "github.com/weibocom/dockerf/machine"
)

var (
	// all docker command except help must give '--machine' flag
	flMachine = flag.String([]string{"-machine"}, "", "Machine name where docker run. 'dockerf machine ls' to display all machines. ")
	flSwarm   = flag.Bool([]string{"-swarm"}, false, "Manage container by swarm")
)

func (dcli *DockerfCli) CmdContainer(args ...string) error {
	flag.Merge(flag.CommandLine, clientFlags.FlagSet, commonFlags.FlagSet)

	flag.Usage = func() {
		fmt.Fprint(os.Stdout, "Usage: dockerf container [OPTIONS] COMMAND [arg...]\n"+daemonUsage+"       docker [ -h | --help | -v | --version ]\n\n")
		fmt.Fprint(os.Stdout, "A self-sufficient runtime for containers.\n\nOptions:\n")

		flag.CommandLine.SetOutput(os.Stdout)
		flag.PrintDefaults()

		help := "\nCommands:\n"

		// TODO(tiborvass): no need to sort if we ensure dockerCommands is sorted
		sort.Sort(byName(dockerCommands))

		for _, cmd := range dockerCommands {
			help += fmt.Sprintf("    %-10.10s%s\n", cmd.name, cmd.description)
		}

		help += "\nRun 'dockerf COMMAND --help' for more information on a command."
		fmt.Fprintf(os.Stdout, "%s\n", help)
	}

	flag.CommandLine.Parse(args)

	if *flVersion {
		showVersion()
		os.Exit(1)
	}

	if *flHelp {
		flag.Usage()
		os.Exit(1)
	}

	if *flMachine == "" {
		app := path.Base(os.Args[0])
		cmd := "container"
		log.Errorf("%s: \"%s\" requires --machine flag for manage any container. See '%s %s --help'. \n", app, cmd, app, cmd)
	}

	argsWithTls, err := mergeTlsConfig(args, *flMachine, *flSwarm)
	if err != nil {
		log.Errorf("Failed to merge tls args:%s", err.Error())
	}

	flag.CommandLine.Parse(argsWithTls)

	c := newDockerClient()
	if err := c.Run(flag.Args()...); err != nil {
		if sterr, ok := err.(cli.StatusError); ok {
			if sterr.Status != "" {
				fmt.Fprintln(os.Stderr, sterr.Status)
				os.Exit(1)
			}
			os.Exit(sterr.StatusCode)
		}
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	return nil
}

func mergeTlsConfig(args []string, machine string, swarm bool) ([]string, error) {
	// pre parse machine and swarm flag

	cluster := ""
	if swarm {
		cluster = "swarm"
	}
	mp := dmachine.NewMachineProxy("container-tls-config")
	tls, err := mp.Config(machine, cluster)
	if err != nil {
		log.Errorf("load tls config error. machine:%s, msg:%s", machine, err.Error())
	}
	argsWithTls := []string{}
	argsWithTls = append(argsWithTls, strings.Split(tls, " ")...)
	argsWithTls = append(argsWithTls, args...)
	return argsWithTls, nil
}

func newDockerClient() *cli.Cli {
	// Set terminal emulation based on platform as required.
	stdin, stdout, stderr := term.StdStreams()
	log.SetOutput(stderr)

	clientCli := client.NewDockerCli(stdin, stdout, stderr, clientFlags)

	handleGlobalDaemonFlag()

	c := cli.New(clientCli, daemonCli)

	return c
}

func showVersion() {
	if utils.ExperimentalBuild() {
		fmt.Printf("Docker version %s, build %s, experimental\n", dockerversion.VERSION, dockerversion.GITCOMMIT)
	} else {
		fmt.Printf("Docker version %s, build %s\n", dockerversion.VERSION, dockerversion.GITCOMMIT)
	}
}
