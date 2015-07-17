package client

import (
	"fmt"
	"os"
	"reflect"
	"strings"

	"github.com/docker/docker/opts"
	flag "github.com/docker/docker/pkg/mflag"
	dcluster "github.com/weibocom/dockerf/cluster"
	dcontext "github.com/weibocom/dockerf/cluster/context"
)

const DEFAULT_CLUSTER_FILE = "cluster.yml"

func GetClusterSubCmdFlags(name, signature, description string, exitOnError bool) *flag.FlagSet {
	fs := flag.NewFlagSet(name, flag.ExitOnError)

	fs.Usage = func() {
		options := ""
		if fs.FlagCountUndeprecated() > 0 {
			options = " [OPTIONS]"
			if signature != "" {
				signature = " " + signature
			}
		}
		fmt.Fprintf(os.Stdout, "Usage: dockerf cluster %s%s%s\n\n%s\n\n", name, options, signature, description)
		fs.SetOutput(os.Stdout)
		fs.PrintDefaults()
		os.Exit(0)
	}
	return fs
}

type ClusterCli struct {
	dockerfCli *DockerfCli
}

func (dcli *DockerfCli) CmdCluster(args ...string) error {
	ClusterFlag.Parse(args)
	clusterCli := newClusterCli()
	clusterCli.dockerfCli = dcli
	if err := clusterCli.Cmd(ClusterFlag.Args()...); err != nil {
		fmt.Printf("Command execute error: '%s'\n", err.Error())
		os.Exit(0)
	}
	return nil
}

func newClusterCli() *ClusterCli {
	return &ClusterCli{}
}

func (ccli *ClusterCli) Cmd(args ...string) error {
	if len(args) > 0 {
		method, exists := ccli.getMethod(args...)
		if !exists {
			fmt.Fprintf(os.Stdout, "dockerf cluster: '%s' is not a dockerf cluster command. See 'dockerf cluster --help'.\n", args[0])
			os.Exit(1)
		}
		return method(args[1:]...)
	}
	return ccli.CmdHelp()
}

func (ccli *ClusterCli) CmdDeploy(args ...string) error {
	fs := GetClusterSubCmdFlags("deploy", " PATH", "Deploy all containers on the cluster, which dedcriped by yaml file at PATH", true)
	flFile := fs.String([]string{"f", "-file"}, "", "Name of the Cluster yaml file(Default is PATH/cluster.yml")
	flCFilter := opts.NewListOpts(nil)
	fs.Var(&flCFilter, []string{"-filter"}, "Filter containers to operate, basedd on conditions provided")
	flRemove := fs.Bool([]string{"-rmc"}, true, "Remove the old container, when new container started")
	flScaleIn := fs.Bool([]string{"-scale-in"}, false, "Destroy extra num of machines, where extra-num is active machines minus necessaries in cluster.yml")
	flScaleOut := fs.Bool([]string{"-scale-out"}, false, "Create extra num of machines, where extra-num is necessary machines minus actives in cluster.yml")
	flResize := fs.Bool([]string{"-resize"}, false, "Set the num of machines exactly equal with necessaries in cluster.yml . --scale-in --scaleout")
	fs.Parse(args)

	if len(fs.Args()) != 1 {
		fmt.Printf("dockerf cluster: 'deploy' requires 1 argument. \n")
		os.Exit(1)
	}

	checkResizeFlagConfict(flScaleIn, flScaleOut, flResize, fs)

	if fs.IsSet("-resize") {
		*flScaleIn = *flResize
		*flScaleOut = *flResize
	}

	path := fs.Args()[0]
	cluster := buildCluster(*flFile, path)

	context := dcontext.NewClusterContext(*flScaleIn, *flScaleOut, *flRemove, &flCFilter, cluster, ccli.dockerfCli.CmdContainer, ccli.dockerfCli.CmdMachine)

	context.Deploy()

	return nil
}

func checkResizeFlagConfict(scaleIn, scaleOut, resize *bool, fs *flag.FlagSet) {

}

func (ccli *ClusterCli) resizeMachine(context *dcontext.ClusterContext, cluster *dcluster.Cluster, scaleIn, scaleOut bool) error {
	return nil
}

func (ccli *ClusterCli) CmdHelp(args ...string) error {
	if len(args) > 0 {
		method, exists := ccli.getMethod(args[0])
		if exists {
			method("--help")
			return nil
		} else {
			fmt.Printf("dockerf cluster: '%s' is not a dockerf cluster command. See 'dockerf cluster --help'.\n", args[0])
			return nil
		}
	}
	ClusterFlag.Usage()
	return nil
}

func (ccli *ClusterCli) getMethod(args ...string) (func(...string) error, bool) {
	if len(args) == 0 {
		return nil, false
	}
	cmd := args[0]
	methodName := "Cmd" + strings.ToUpper(cmd[:1]) + strings.ToLower(cmd[1:])
	method := reflect.ValueOf(ccli).MethodByName(methodName)
	if !method.IsValid() {
		return nil, false
	}
	return method.Interface().(func(...string) error), true
}

func buildCluster(name, path string) *dcluster.Cluster {
	fileName := name
	if fileName == "" {
		fileName = DEFAULT_CLUSTER_FILE
	}
	file := ""
	seperator := "/"
	if strings.HasSuffix(path, seperator) {
		file = path + fileName
	} else {
		file = path + seperator + fileName
	}
	cluster, err := dcluster.NewCluster(file)
	if err != nil {
		if os.IsNotExist(err) {
			fmt.Printf("Cannot locate Clusterfile: '%s'\n", file)
		} else {
			fmt.Printf("Read Clusterfile '%s' error: '%s'\n", file, err.Error())
		}
		os.Exit(1)
	}
	return cluster
}