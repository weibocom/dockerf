package client

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"strings"

	log "github.com/Sirupsen/logrus"
	"github.com/docker/docker/api/client"
	"github.com/docker/docker/autogen/dockerversion"
	"github.com/docker/docker/opts"
	flag "github.com/docker/docker/pkg/mflag"
	"github.com/docker/docker/pkg/term"
	dmachine "github.com/weibocom/dockerf/machine"
)

const (
	defaultTrustKeyFile = "key.json"
	defaultCaFile       = "ca.pem"
	defaultKeyFile      = "key.pem"
	defaultCertFile     = "cert.pem"
)

var (
	// all docker command except help must give '--machine' flag
	flMachine = flag.String([]string{"-machine"}, "", "Machine name where docker run. 'dockerf machine ls' to display all machines. ")
	flSwarm   = flag.Bool([]string{"-swarm"}, false, "Manage container by swarm")
)

func init() {

}

func (dcli *DockerfCli) CmdContainer(args ...string) error {

	flag.CommandLine.Parse(args)

	if *flVersion {
		showVersion()
		os.Exit(1)
	}

	if *flDaemon {
		if *flHelp {
			flag.Usage()
			os.Exit(1)
		}
		fmt.Printf("dockerf container does not support daemon.\n")
		os.Exit(1)
	}

	if *flMachine == "" {
		app := path.Base(os.Args[0])
		cmd := "container"
		fmt.Printf("%s: \"%s\" requires --machine flag for manage any container. See '%s %s --help'. \n", app, cmd, app, cmd)
		os.Exit(1)
	}

	cluster := ""
	if *flSwarm {
		cluster = "swarm"
	}
	rewriteTLSFlags(dcli, *flMachine, cluster)

	cli := newDockerClient()
	if err := cli.Cmd(flag.Args()...); err != nil {
		if sterr, ok := err.(client.StatusError); ok {
			if sterr.Status != "" {
				fmt.Fprintln(cli.Err(), sterr.Status)
				os.Exit(1)
			}
			os.Exit(sterr.StatusCode)
		}
		fmt.Fprintln(cli.Err(), err)
		os.Exit(1)
	}
	return nil
}

func rewriteTLSFlags(dcli *DockerfCli, machine string, cluster string) {
	if machine == "" {
		return
	}
	tlsFlagArgs := func(machineName string) []string {
		mp := dmachine.NewMachineProxy("container-tls-config")
		if flags, err := mp.Config(machineName, cluster); err != nil {
			fmt.Printf("Load Machine(name:%s, cluster:%s) configs error:%s\n", machine, cluster, err.Error())
			return []string{}
		} else {
			return strings.Split(flags, " ")
		}
	}(machine)

	tlsFlagset := flag.NewFlagSet("tls-docker-machine", flag.ExitOnError)

	flTls = tlsFlagset.Bool([]string{"-tls"}, false, "Use TLS; implied by --tlsverify")
	flTlsVerify = tlsFlagset.Bool([]string{"-tlsverify"}, false, "Use TLS and verify the remote")
	flCa = tlsFlagset.String([]string{"-tlscacert"}, "", "Trust certs signed only by this CA")
	flCert = tlsFlagset.String([]string{"-tlscert"}, "", "Path to TLS certificate file")
	flKey = tlsFlagset.String([]string{"-tlskey"}, "", "Path to TLS key file")

	listOpts := opts.NewListOpts(nil)
	tlsFlagset.Var(&listOpts, []string{"H", "-host"}, "Daemon socket(s) to connect to")

	tlsFlagset.Parse(tlsFlagArgs)

	flHosts = listOpts.GetAll()

}

func newDockerClient() *client.DockerCli {
	// Set terminal emulation based on platform as required.
	stdin, stdout, stderr := term.StdStreams()

	setDefaultConfFlag(flTrustKey, defaultTrustKeyFile)

	if len(flHosts) > 1 {
		log.Fatal("Please specify only one -H")
	}
	protoAddrParts := strings.SplitN(flHosts[0], "://", 2)

	var (
		cli       *client.DockerCli
		tlsConfig tls.Config
	)
	tlsConfig.InsecureSkipVerify = true

	// Regardless of whether the user sets it to true or false, if they
	// specify --tlsverify at all then we need to turn on tls
	if flag.IsSet("-tlsverify") {
		*flTls = true
	}

	// If we should verify the server, we need to load a trusted ca
	if *flTlsVerify {
		certPool := x509.NewCertPool()
		file, err := ioutil.ReadFile(*flCa)
		if err != nil {
			log.Fatalf("Couldn't read ca cert %s: %s", *flCa, err)
		}
		certPool.AppendCertsFromPEM(file)
		tlsConfig.RootCAs = certPool
		tlsConfig.InsecureSkipVerify = false
	}

	// If tls is enabled, try to load and send client certificates
	if *flTls || *flTlsVerify {
		_, errCert := os.Stat(*flCert)
		_, errKey := os.Stat(*flKey)
		if errCert == nil && errKey == nil {
			*flTls = true
			cert, err := tls.LoadX509KeyPair(*flCert, *flKey)
			if err != nil {
				log.Fatalf("Couldn't load X509 key pair: %q. Make sure the key is encrypted", err)
			}
			tlsConfig.Certificates = []tls.Certificate{cert}
		}
		// Avoid fallback to SSL protocols < TLS1.0
		tlsConfig.MinVersion = tls.VersionTLS10
	}

	cli = client.NewDockerCli(stdin, stdout, stderr, *flTrustKey, protoAddrParts[0], protoAddrParts[1], &tlsConfig)
	return cli
}

func showVersion() {
	fmt.Printf("Docker version %s, build %s\n", dockerversion.VERSION, dockerversion.GITCOMMIT)
}
