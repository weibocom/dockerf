package main

import (
	"fmt"
	"os"
	"time"

	"github.com/Sirupsen/logrus"
	"github.com/weibocom/dockerf/container/cluster"
	"github.com/weibocom/dockerf/machine"
	"github.com/weibocom/dockerf/options"
	"github.com/weibocom/dockerf/topology"
)

const (
	driver = "virtualbox"
)

func main() {
	logrus.SetLevel(logrus.DebugLevel)
	path := "/Users/crystal/go/src/github.com/weibocom/dockerf/cluster-default.yml"
	t, err := topology.NewTopology(path)
	if err != nil {
		fmt.Printf("Error:%s\n", err.Error())
		return
	}
	t.ScaleMachine("all")

	<-time.After(3 * time.Minute)
	// // os.Setenv("debug", "true")
	// // os.Setenv("DEBUG", "true")
	// mc := getMachineCluster()
	// mname := "master"
	// master, exists := mc.Get(mname)
	// if !exists {
	// 	master = createHost(mc, mname)
	// }
	// cc := getContainerCluster(master)

	// // var wg sync.WaitGroup

	// // wg.Add(1)
	// // go func() {
	// // 	defer wg.Done()
	// // 	m := createHost(mc, "master")
	// // 	addMaster(cc, m)
	// // }()

	// // workers := 1
	// // wg.Add(workers)
	// // for n := 0; n < workers; n++ {
	// // 	go func() {
	// // 		defer wg.Done()
	// // 		m := createHost(mc, fmt.Sprintf("worker-%d", n))
	// // 		addSlave(cc, m)
	// // 	}()
	// // }

	// // wg.Wait()

	// desc := container.NewContainerDesc()
	// desc.SetCmd("sleep", "30")
	// desc.SetImage("busybox:latest")
	// id, err := cc.Run(desc, fmt.Sprintf("test-%d", time.Now().Unix()))
	// if err != nil {
	// 	logrus.Fatalf("err:%+s", err.Error())
	// }
	// logrus.Infof("id:%s", id)
	// cc.Run(desc, "test2")
}

func addSlave(clu *cluster.Cluster, m *machine.Machine) {
	err := clu.AddWorker(m)
	if err != nil {
		logrus.Fatalf("failed to add worker '%s' to container cluster. err:%s", m, err.Error())
	}
	logrus.Infof("add to slave")
}

func listMachines(c *machine.Cluster) {
	ms := c.ListAll()
	for _, m := range ms {
		s := m.GetCachedState()
		fmt.Printf("name:%s\tstate:%t\n", m.Name(), s.String())
	}
}

func getContainerCluster(master *machine.Machine) *cluster.Cluster {
	m := make(map[string]string)
	// m["swarm-host"] = ":3376"
	m["swarm-image"] = "registry.intra.weibo.com/icycrystal4/swarm:latest"
	m["swam-discover"] = "consul://101.200.173.242:8500/orginal"
	options := options.Options{
		Values: m,
	}
	cluster, err := cluster.NewCluster("swarm", &options, master)
	if err != nil {
		fmt.Printf("new cluster error:%s\n", err.Error())
		os.Exit(0)
	}
	return cluster
}

func getMachineCluster() *machine.Cluster {
	opts := make(map[string]string)
	opts["debug"] = "true"
	opts["engine-insecure-registry"] = "registry.intra.weibo.com"
	opts["engine-storage-driver"] = "aufs"
	// options["native-ssh"] = "true"
	cluster, err := machine.NewCluster(&options.Options{Values: opts})
	if err != nil {
		fmt.Printf("Failed to create machine cluster:%s\n", err.Error())
		os.Exit(0)
	}
	return cluster
}

func createHost(cluster *machine.Cluster, name string) *machine.Machine {
	opts := make(map[string]string)
	opts["virtualbox-memory"] = "512"
	opts["engine-label"] = "role=worker group=tomcat"
	os.Setenv("VIRTUALBOX_DISK_SIZE", "5000")
	md := &machine.MachineOptions{
		Options: &options.Options{Values: opts},
	}
	m, err := cluster.Create(name, "virtualbox", md)
	if err != nil {
		fmt.Printf("%s\n", err.Error())
	}
	return m
}
