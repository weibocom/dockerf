package main

import (
	"fmt"
	"os"
	"runtime/pprof"

	"github.com/Sirupsen/logrus"
	"github.com/weibocom/dockerf/api/client"
	flag "github.com/weibocom/dockerf/dflag"
	_ "github.com/weibocom/dockerf/discovery/drivers"
)

func main() {
	flag.DFlag.Parse(os.Args[1:])
	// collect cpu profile
	if *flCpuProfile != "" {
		f, err := os.Create(*flCpuProfile)
		if err != nil {
			logrus.Fatalf("Error to create file(%s) to store cpu profile info.\n", err.Error())
		}
		pprof.StartCPUProfile(f)
		defer pprof.StopCPUProfile()
	}
	// end
	if *flVersion {
		showVersion()
		return
	}
	if *flLogLevel != "" {
		lvl, err := logrus.ParseLevel(*flLogLevel)
		if err != nil {
			logrus.Fatalf("Unable to parse logging level: %s", *flLogLevel)
			setLogLevel(lvl)
		}
	} else {
		setLogLevel(logrus.InfoLevel)
	}

	if *flDebug {
		os.Setenv("DEBUG", "1")
		setLogLevel(logrus.DebugLevel)
	}
	cli := client.GetDockerfCli()
	if err := cli.Cmd(flag.DFlag.Args()...); err != nil {
		logrus.Fatal(err)
		os.Exit(-1)
	}
}

func showVersion() {
	fmt.Printf("Docker Framework version %s, build %s\n", "0.0.1", "demo")
}
