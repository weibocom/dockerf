package main

// import (
// 	"fmt"
// 	"os"
// 	"os/signal"
// 	"runtime"
// 	"runtime/pprof"
// 	"syscall"

// 	"github.com/Sirupsen/logrus"
// 	"github.com/weibocom/dockerf/api/client"
// 	_ "github.com/weibocom/dockerf/container/filter"
// 	flag "github.com/weibocom/dockerf/dflag"
// 	_ "github.com/weibocom/dockerf/discovery/drivers"
// )

// func main() {
// 	flag.DFlag.Parse(os.Args[1:])

// 	//register signal for thread dump
// 	signalChan := make(chan os.Signal, 1)
// 	signal.Notify(signalChan, syscall.SIGQUIT)

// 	go func() {
// 		stacktrace := make([]byte, 8192)
// 		for _ = range signalChan {
// 			length := runtime.Stack(stacktrace, true)
// 			fmt.Println(string(stacktrace[:length]))
// 		}
// 	}()

// 	// collect cpu profile
// 	if *flCpuProfile != "" {
// 		f, err := os.Create(*flCpuProfile)
// 		if err != nil {
// 			logrus.Fatalf("Error to create file(%s) to store cpu profile info.\n", err.Error())
// 		}
// 		pprof.StartCPUProfile(f)
// 		defer pprof.StopCPUProfile()
// 	}
// 	// end
// 	if *flVersion {
// 		showVersion()
// 		return
// 	}
// 	if *flLogLevel != "" {
// 		lvl, err := logrus.ParseLevel(*flLogLevel)
// 		if err != nil {
// 			logrus.Fatalf("Unable to parse logging level: %s", *flLogLevel)
// 			setLogLevel(lvl)
// 		}
// 	} else {
// 		setLogLevel(logrus.InfoLevel)
// 	}

// 	if *flDebug {
// 		os.Setenv("DEBUG", "1")
// 		setLogLevel(logrus.DebugLevel)
// 	}
// 	initLogging()
// 	cli := client.GetDockerfCli()
// 	if err := cli.Cmd(flag.DFlag.Args()...); err != nil {
// 		logrus.Fatal(err)
// 		os.Exit(-1)
// 	}
// }

// func showVersion() {
// 	fmt.Printf("Docker Framework version %s, build %s\n", "0.0.1", "demo")
// }
