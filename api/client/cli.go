package client

import (
	"fmt"
	"io"
	"os"
	"reflect"
	"strings"

	flag "github.com/docker/docker/pkg/mflag"
)

type DockerfCli struct {
	in  io.ReadCloser
	out io.Writer
	err io.Writer
}

var dockerfClient *DockerfCli

func init() {
	dockerfClient = NewDockerfCli()
}

func (cli *DockerfCli) getMethod(args ...string) (func(...string) error, bool) {
	camelArgs := make([]string, len(args))
	for i, s := range args {
		if len(s) == 0 {
			return nil, false
		}
		camelArgs[i] = strings.ToUpper(s[:1]) + strings.ToLower(s[1:])
	}
	methodName := "Cmd" + strings.Join(camelArgs, "")
	method := reflect.ValueOf(cli).MethodByName(methodName)
	if !method.IsValid() {
		return nil, false
	}
	return method.Interface().(func(...string) error), true
}

func (cli *DockerfCli) Cmd(args ...string) error {
	if len(args) > 1 {
		method, exists := cli.getMethod(args[:2]...)
		if exists {
			return method(args[2:]...)
		}
	}
	if len(args) > 0 {
		method, exists := cli.getMethod(args[0])
		if !exists {
			fmt.Fprintf(cli.err, "dockerf: '%s' is not a dockerf command. See 'dockerf --help'.\n", args[0])
			os.Exit(1)
		}
		return method(args[1:]...)
	}
	return cli.CmdHelp()
}

func (cli *DockerfCli) Subcmd(name, signature, description string, exitOnError bool) *flag.FlagSet {
	var errorHandling flag.ErrorHandling
	if exitOnError {
		errorHandling = flag.ExitOnError
	} else {
		errorHandling = flag.ContinueOnError
	}
	flags := flag.NewFlagSet(name, errorHandling)
	flags.Usage = func() {
		options := ""
		if signature != "" {
			signature = " " + signature
		}
		if flags.FlagCountUndeprecated() > 0 {
			options = " [OPTIONS]"
		}
		fmt.Fprintf(cli.out, "\nUsage: dockerf %s%s%s\n\n%s\n\n", name, options, signature, description)
		flags.SetOutput(cli.out)
		flags.PrintDefaults()
		os.Exit(0)
	}
	return flags
}

func GetDockerfCli() *DockerfCli {
	return dockerfClient
}

func NewDockerfCli() *DockerfCli {
	return &DockerfCli{
		in:  os.Stdin,
		out: os.Stdout,
		err: os.Stderr,
	}
}
