package context

import (
	"errors"
	"fmt"
	"math"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/docker/docker/opts"
	dcluster "github.com/weibocom/dockerf/cluster"
	dcontainer "github.com/weibocom/dockerf/container"
	"github.com/weibocom/dockerf/discovery"
	dmachine "github.com/weibocom/dockerf/machine"
	"github.com/weibocom/dockerf/sequence"
	dutils "github.com/weibocom/dockerf/utils"
)

const DEFAULT_CLUSTER_FILE = "cluster.yml"

var containerInfoLock *sync.Mutex = &sync.Mutex{}

var httpClient *http.Client

func init() {
	transport := &dutils.Transport{
		ConnectTimeout:        1 * time.Second,
		RequestTimeout:        10 * time.Second,
		ResponseHeaderTimeout: 5 * time.Second,
	}
	httpClient = &http.Client{Transport: transport}
}

type ClusterContext struct {
	create            bool
	clusterDesc       *dcluster.Cluster
	scaleIn           bool
	scaleOut          bool
	removeContainer   bool
	forceCreate       bool
	filters           *opts.ListOpts
	machinePrefix     string
	machineInfos      []dmachine.MachineInfo
	containerInfos    []dcontainer.ContainerInfo
	CmdContainer      func(args ...string) error
	CmdMachine        func(args ...string) error
	mSeq              sequence.Seq
	cSeqs             map[string]*sequence.Seq
	mProxy            *dmachine.MachineClusterProxy
	cProxy            *dcontainer.DockerProxy
	serviceRegistries map[string]*discovery.ServiceRegisterDriver
}

func NewClusterContext(scaleIn, scaleOut, rmc, forceCreate bool, cFilter *opts.ListOpts, cluster *dcluster.Cluster, machinePrefix string, cmdContainer, cmdMachine func(args ...string) error) *ClusterContext {
	if machinePrefix == "" {
		machinePrefix = "node"
	}

	clusterContext := &ClusterContext{
		create:            true,
		clusterDesc:       cluster,
		filters:           cFilter,
		scaleIn:           scaleIn,
		scaleOut:          scaleOut,
		removeContainer:   rmc,
		machinePrefix:     machinePrefix,
		CmdContainer:      cmdContainer,
		CmdMachine:        cmdMachine,
		machineInfos:      []dmachine.MachineInfo{},
		containerInfos:    []dcontainer.ContainerInfo{},
		mSeq:              sequence.Seq{},
		cSeqs:             map[string]*sequence.Seq{},
		serviceRegistries: map[string]*discovery.ServiceRegisterDriver{},
	}
	clusterContext.initContext()
	return clusterContext

}

// init the exists machine, container, and seq
func (ctx *ClusterContext) initContext() {
	fmt.Printf("Parsing port binding in cluster description.\n")
	ctx.parsePortBindings()

	fmt.Printf("Create a new machine proxy. cluster by:%s, driver:%s, discovery: %s, master: %s, driver options:%s\n", ctx.clusterDesc.ClusterBy, ctx.clusterDesc.Driver, ctx.clusterDesc.Discovery, ctx.clusterDesc.MasterNode, strings.Join(ctx.clusterDesc.DriverOptions, " "))
	machineProxy := dmachine.NewMachineClusterProxy("dockerf machine", ctx.clusterDesc.ClusterBy, ctx.clusterDesc.Driver, ctx.clusterDesc.Discovery, ctx.clusterDesc.MasterNode, ctx.clusterDesc.DriverOptions)
	ctx.mProxy = machineProxy

	fmt.Println("Loading the cluster machine info...")
	mis, err := ctx.mProxy.List()
	if err != nil {
		panic("Init cluster context error, cannot list machine infos:" + err.Error())
	}
	ctx.machineInfos = mis

	fmt.Printf("Init machine master\n")
	if err := ctx.startMaster(); err != nil {
		panic("Start cluster master failed: " + err.Error())
	}

	tlsConfig, err := machineProxy.Config()
	if err != nil {
		panic(fmt.Sprintf("Failed to load master machine node tls config. err:%s\n", err.Error()))
	}

	containerProxy, err := dcontainer.NewDockerProxy(tlsConfig)
	if err != nil {
		panic("Failed to create docker proxy:%s\n" + err.Error())
	}
	ctx.cProxy = containerProxy

	fmt.Printf("Init the consule cluster.\n")
	if err := ctx.startConsulCluster(); err != nil {
		panic("Start consul cluster failed: " + err.Error())
	}

	fmt.Println("Init the named machine sequence...")
	ctx.initMachineSequence(mis)

	cifs, err := ctx.cProxy.ListAll()
	fmt.Printf("Init the named container sequences... total containers:%d\n", len(cifs))
	if err != nil {
		panic("Init cluster context error, cannot list container infos:" + err.Error())
	}
	ctx.initContainerSequences(cifs)

	fmt.Println("cluster context inited successfully")
}

func (ctx *ClusterContext) parsePortBindings() error {
	for idx, description := range ctx.clusterDesc.Containers {
		binding := dcluster.PortBinding{}
		if err := binding.Parse(description.Port); err != nil {
			return err
		}
		ctx.clusterDesc.Containers[idx].PortBinding = binding
		fmt.Printf("Port binding parsed. group: %s, binding:%+v\n", description.Name, binding)
	}
	return nil
}

func (ctx *ClusterContext) loadContainers() error {
	infos, err := ctx.cProxy.ListAll()
	if err == nil {
		ctx.containerInfos = infos
	}
	return err
}

func (ctx *ClusterContext) getContainerDescription(group string) (dcluster.ContainerDescription, bool) {
	for _, cd := range ctx.clusterDesc.Containers {
		if cd.Name == group {
			return cd, true
		}
	}
	return dcluster.ContainerDescription{}, false
}

func (ctx *ClusterContext) initServiceDiscovery(cinfos []dcontainer.ContainerInfo) error {
	loadRegistry := func(group string, description dcluster.ContainerDescription, infos []dcontainer.ContainerInfo) []string {
		ipPorts := []string{}
		for _, c := range infos {
			if c.Group != group {
				continue
			}
			if !c.IsUp() {
				fmt.Printf("Container '%s' found for group '%s', but container is not up.\n", c.Name[0], c.Group)
			}
			for _, iPort := range c.IpPorts {
				if iPort.PrivatePort != description.PortBinding.ContainerPort {
					continue
				}
				ipPort := fmt.Sprintf("%s:%d", iPort.IP, iPort.PublicPort)
				ipPorts = append(ipPorts, ipPort)
			}
		}
		return ipPorts
	}

	for sdName, sdd := range ctx.clusterDesc.ServiceDiscovers {
		driver, err := discovery.NewRegDriver(sdd)
		if err != nil {
			return errors.New(fmt.Sprintf("Failed to create Reg Driver:'%s'. err:%s", sdName, err.Error()))
		}
		containerGroup, ok := sdd["container"]
		fmt.Printf("Load registry for '%s'\n", containerGroup)
		if ok {
			fmt.Printf("Load registry for service discover. sd name: '%s', group: '%s'.\n", sdName, containerGroup)
			description, exists := ctx.getContainerDescription(containerGroup)
			if !exists {
				panic(fmt.Sprintf("No container description found for group '%s'", containerGroup))
			}
			ipPorts := loadRegistry(containerGroup, description, cinfos)
			if len(ipPorts) == 0 {
				fmt.Printf("No available container for service discovery:%s, container group:%s\n", sdName, containerGroup)
			} else {
				fmt.Printf("Service discover registry found: %+v\n", ipPorts)
				(*driver).Registry(ipPorts)
			}
		}
		fmt.Printf("Service discover registered: name:%s. \n", sdName)
		ctx.serviceRegistries[sdName] = driver
	}
	return nil
}

func (ctx *ClusterContext) registerServiceByContainer(c *dcontainer.ContainerInfo, cd *dcluster.ContainerDescription) error {
	sd := cd.ServiceDiscover
	if sd == "" {
		fmt.Printf("Service discover missed, no need to register %s\n", c.Name[0])
		return nil
	}

	host, port, find := func(c *dcontainer.ContainerInfo) (string, int, bool) {
		for _, ipPort := range c.IpPorts {
			if ipPort.PrivatePort == cd.PortBinding.ContainerPort {
				return ipPort.IP, ipPort.PublicPort, true
			}
		}
		return "", 0, false
	}(c)

	if !find {
		return errors.New(fmt.Sprintf("Container is not expose any port as a service. (container name:%s, description:%s)", c.Name[0], cd.Name))
	}

	driver, ok := ctx.serviceRegistries[cd.ServiceDiscover]
	if !ok {
		return errors.New(fmt.Sprintf("No service register driver available for:'%s'\n", cd.ServiceDiscover))
	}
	return (*driver).Register(host, port)
}

func (ctx *ClusterContext) registerServiceByContainerId(cid string, cd *dcluster.ContainerDescription) error {
	if cd.ServiceDiscover == "" {
		fmt.Printf("Service discover missed, there is not need to register. cid:%s\n", cid)
	}
	fmt.Printf("Registering container service: cid:%s, description:%s\n, ", cid, cd.Name)

	c, exists := ctx.cProxy.GetContainerByID(cid)
	if !exists {
		return errors.New(fmt.Sprintf("Container(cid:%s, name:%s) is not exists", cid))
	}
	return ctx.registerServiceByContainer(&c, cd)
}

func (ctx *ClusterContext) unregisterService(ip string, port int, cd *dcluster.ContainerDescription) error {
	sd := cd.ServiceDiscover
	if sd == "" {
		fmt.Printf("Service discover missed, no need to unregister %s:%d\n", ip, port)
		return nil
	}
	fmt.Printf("Ungregister service ip:%s, port:%d\n", ip, port)
	driver, ok := ctx.serviceRegistries[sd]
	if !ok {
		return errors.New(fmt.Sprintf("No service register driver available for:'%s'\n", sd))
	}
	return (*driver).UnRegister(ip, port)
	fmt.Printf("Ungregistered service ip:%s, port:%d\n", ip, port)
	return nil
}

func (ctx *ClusterContext) reloadMachineInfos() error {
	mis, err := ctx.mProxy.List()
	if err == nil {
		ctx.machineInfos = mis
	}
	return err
}

func (ctx *ClusterContext) initContainerSequences(cifs []dcontainer.ContainerInfo) {

	cn := dcontainer.ContainerName{}
	for _, ci := range cifs {
		cName := ci.Name[0]
		if cn.Parse(cName) {
			group := cn.Group
			seqNum := cn.Seq
			if gSeq, ok := ctx.cSeqs[group]; ok {
				gSeq.Max(seqNum)
			} else {
				seq := &sequence.Seq{}
				seq.Max(seqNum)
				ctx.cSeqs[group] = seq
			}
		} else {
			fmt.Printf("'%s' is not a valid container name.\n", cName)
		}
	}
	fmt.Printf("Named Container info sequences inited:%+v\n", ctx.cSeqs)
}

func (ctx *ClusterContext) initMachineSequence(mifs []dmachine.MachineInfo) {
	mn := dmachine.MachineName{}
	for _, mi := range ctx.machineInfos {
		if mi.IsMaster() {
			continue
		}
		if mn.Parse(mi.Name) {
			ctx.mSeq.Max(mn.Seq)
		} else {
			fmt.Printf("'%s' is not a valid machine name.\n", mi.Name)
		}
	}
	fmt.Printf("Named Machine info sequences inited:%d\n", ctx.mSeq.Get())
}

func (ctx *ClusterContext) getMaster() (dmachine.MachineInfo, bool) {
	for _, mi := range ctx.machineInfos {
		if mi.IsMaster() {
			return mi, true
		}
	}
	return dmachine.MachineInfo{}, false
}

func (ctx *ClusterContext) Deploy() error {
	if err := ctx.ensureMachineCapacity(); err != nil {
		fmt.Printf("Ensure machine capacity error:%s\n", err.Error())
		os.Exit(1)
	}

	if err := ctx.deployContainers(); err != nil {
		fmt.Printf("Deploy container error:%s\n", err.Error())
		os.Exit(1)
	}
	fmt.Printf("Deploy successfully.\n")
	return nil
}

func (ctx *ClusterContext) getConsulPortBindings() []dcluster.PortBinding {
	return []dcluster.PortBinding{
		dcluster.PortBinding{
			Protocal:      "tcp",
			HostPort:      8300,
			ContainerPort: 8300,
		},
		dcluster.PortBinding{
			Protocal:      "tcp",
			HostPort:      8301,
			ContainerPort: 8301,
		},
		dcluster.PortBinding{
			Protocal:      "tcp",
			HostPort:      8302,
			ContainerPort: 8302,
		},
		dcluster.PortBinding{
			Protocal:      "udp",
			HostPort:      8302,
			ContainerPort: 8302,
		},
		dcluster.PortBinding{
			Protocal:      "tcp",
			HostPort:      8400,
			ContainerPort: 8400,
		},
		dcluster.PortBinding{
			Protocal:      "tcp",
			HostPort:      8500,
			ContainerPort: 8500,
		},
		dcluster.PortBinding{
			Protocal:      "udp",
			HostPort:      53,
			ContainerPort: 53,
		},
	}
}

func (ctx *ClusterContext) runConsulJoinServer(server dcluster.ConsulServer, joinNode string, joinIp string, bootstrapIp string) (string, error) {
	name := fmt.Sprintf("%s-consul-server-join", joinIp)
	portBindings := ctx.getConsulPortBindings()
	cmds := []string{
		"-server",
		"-advertise",
		joinIp,
		"-join",
		bootstrapIp,
		"-domain",
		server.Domain,
	}
	envs := []string{"constraint:node==" + joinNode}

	runConfig := dcontainer.ContainerRunConfig{
		Image:        server.Image,
		Name:         name,
		PortBindings: portBindings,
		Hostname:     name,
		Envs:         envs,
		Cmds:         cmds,
	}

	return ctx.cProxy.RunByConfig(runConfig)
}

func (ctx *ClusterContext) runConsulBootstrapServer(serverNode string, serverIP string) (string, error) {
	server := ctx.clusterDesc.ConsulCluster.Server
	name := fmt.Sprintf("%s-consulserverbootstrap", serverNode)
	portBindings := ctx.getConsulPortBindings()
	cmds := []string{
		"-server",
		"-bootstrap",
		"-domain",
		server.Domain,
		"-advertise",
		serverIP,
	}

	envs := []string{"constraint:node==" + serverNode}

	runConfig := dcontainer.ContainerRunConfig{
		Image:        server.Image,
		Name:         name,
		PortBindings: portBindings,
		Hostname:     name,
		Envs:         envs,
		Cmds:         cmds,
	}

	return ctx.cProxy.RunByConfig(runConfig)
}

func (ctx *ClusterContext) runConsulAgent(dockerProxy *dcontainer.DockerProxy, agent dcluster.ConsulAgent, agentNode string, agentIp string) (string, error) {
	name := fmt.Sprintf("%s-consul-agent", agentNode)

	serverIp := ctx.clusterDesc.ConsulCluster.Server.IPs[0]
	envs := []string{}
	cmds := []string{"-advertise", agentIp, "-join", serverIp}

	portBindings := ctx.getConsulPortBindings()

	runConfig := dcontainer.ContainerRunConfig{
		Image:        agent.Image,
		Name:         name,
		PortBindings: portBindings,
		Envs:         envs,
		Cmds:         cmds,
		Hostname:     name,
	}

	return dockerProxy.RunByConfig(runConfig)
}

func (ctx *ClusterContext) runConsulRegistrator(dockerProxy *dcontainer.DockerProxy, registrator dcluster.ConsulRegistrator, node string, ip string) (string, error) {
	name := fmt.Sprintf("%s-consul-registrator", node)
	runConfig := dcontainer.ContainerRunConfig{
		Image:    registrator.Image,
		Name:     name,
		Bindings: []string{"/var/run/docker.sock:/tmp/docker.sock"},
		Cmds: []string{
			"-ip",
			ip,
			fmt.Sprintf("consul://%s:8500", ip),
		},
	}
	return dockerProxy.RunByConfig(runConfig)
}

func (ctx *ClusterContext) initSlave(node string) error {
	fmt.Printf("Init the machine infrastructure envment: '%s'\n", node)
	// exec command on this machine
	command := strings.TrimSpace(ctx.clusterDesc.Machine.Init)
	if command != "" {
		if err := ctx.mProxy.ExecCmd(node, command); err != nil {
			fmt.Printf("Failed to exec command(%s) on '%s'\n", command, node)
		}
	}

	tlsConfig, err := ctx.mProxy.ConfigNode(node)
	if err != nil {
		return errors.New(fmt.Sprintf("Failed to load config for node '%s'", node))
	}
	ip, err := ctx.mProxy.IP(node)
	if err != nil {
		fmt.Printf("Failed to load agent ip:'%s', error:%s\n", node, err.Error())
		return err
	}
	proxy, err := dcontainer.NewDockerProxy(tlsConfig)
	if err != nil {
		fmt.Printf("Failed to new docker proxy. machine:%s, tlsConfig:%s\n", node, tlsConfig)
	}
	fmt.Printf("Run consul agent on '%s(%s)'\n", node, ip)
	if cid, err := ctx.runConsulAgent(proxy, ctx.clusterDesc.ConsulCluster.Agent, node, ip); err != nil {
		panic(fmt.Sprintf("Run consul agent on '%s' failed. err:%s\n", node, err.Error()))
	} else {
		fmt.Printf("Consul agent running successfully. id:%s\n", cid)
	}

	fmt.Printf("Run consul registrator on '%s(%s)'\n", node, ip)
	if cid, err := ctx.runConsulRegistrator(proxy, ctx.clusterDesc.ConsulCluster.Registrator, node, ip); err != nil {
		panic(fmt.Sprintf("Failed to run consul registor container on '%s'. err:%s", node, err.Error()))
	} else {
		fmt.Printf("Consul registrator running successfully. id:%s\n", cid)
	}

	fmt.Sprintf("Slave machine init successfully. node:%s\n", node)
	return nil
}

func (ctx *ClusterContext) createSlaves(num int) ([]string, []error) {
	successNodeNames := []string{}
	errs := []error{}
	var lock sync.Mutex
	var wg sync.WaitGroup
	wg.Add(int(num))
	for i := int(0); i < num; i++ {
		go func(ctx *ClusterContext) {
			defer wg.Done()
			name := ctx.genMachineName()
			fmt.Printf("Creating machine '%s'\n", name)
			err := ctx.mProxy.CreateSlave(name)
			if err != nil {
				fmt.Printf("Failed to Create machine('%s'). Error:%s\n", name, err.Error())
				errs = append(errs, err)
			} else {
				fmt.Printf("Machine(%s) created and started, begin to init slave.\n", name)
				if err := ctx.initSlave(name); err != nil {
					fmt.Printf("Failed to init slave '%s'\n", name)
				} else {
					fmt.Printf("Machine(%s) inited complete.\n", name)
					lock.Lock()
					successNodeNames = append(successNodeNames, name)
					lock.Unlock()
				}
			}
		}(ctx)
	}
	wg.Wait()
	return successNodeNames[0:], errs[0:]
}

func (ctx *ClusterContext) genMachineName() string {
	mn := dmachine.MachineName{
		Prefix: ctx.machinePrefix,
		Seq:    ctx.mSeq.Next(),
	}
	return mn.GetName()
}

func (ctx *ClusterContext) scaleMachineIn() error {
	if !ctx.scaleIn {
		fmt.Printf("No need to scale in.\n")
		return nil
	}
	runningMachines := ctx.getMachines(func(mi *dmachine.MachineInfo) bool {
		return !mi.IsMaster()
	})
	rNum := int(len(runningMachines))

	max := ctx.clusterDesc.Machine.MaxNum
	if rNum <= max {
		fmt.Printf("No extra machines in the cluster. Exists: %d. Maximal requirements: %d\n", rNum, max)
		return nil
	}
	destroyNum := rNum - max
	fmt.Printf("There are %d machines in the cluster, but maximal required num is %d. %d extra will be destroyed.\n", rNum, max, destroyNum)

	// first destroy all stopped machines
	destroyMachineNames := []string{}
	ctx.getMachines(func(mi *dmachine.MachineInfo) bool {
		if mi.IsMaster() || mi.IsRunning() {
			return false
		}
		destroyMachineNames = append(destroyMachineNames, mi.Name)
		return true
	})
	dl := int(len(destroyMachineNames))
	if dl < destroyNum {
		// need to destroy some running machines
		ctx.getMachines(func(mi *dmachine.MachineInfo) bool {
			if mi.IsMaster() || !mi.IsRunning() || dl >= destroyNum {
				return false
			}
			dl++
			destroyMachineNames = append(destroyMachineNames, mi.Name)
			return true
		})
	}
	fmt.Printf("Destroying machines: %+v\n", destroyMachineNames)
	return ctx.mProxy.Destroy(destroyMachineNames...)
}

func (ctx *ClusterContext) ensureMachineCapacity() error {
	if err := ctx.scaleMachineOut(); err != nil {
		return err
	}

	return ctx.scaleMachineIn()
}

var sequenceLock sync.Mutex

func (ctx *ClusterContext) nextContainerName(group string) string {
	gSeq := func(g string, ctx *ClusterContext) *sequence.Seq {
		sequenceLock.Lock()
		defer sequenceLock.Unlock()
		gSeq, ok := ctx.cSeqs[group]
		if ok {
			return gSeq
		}
		gSeq = &sequence.Seq{}
		gSeq.Max(0)
		ctx.cSeqs[group] = gSeq
		return gSeq
	}(group, ctx)

	seq := gSeq.Next()
	cn := dcontainer.ContainerName{
		Group: group,
		Seq:   seq,
	}
	return cn.GetName()
}

func (ctx *ClusterContext) runContainer(cd *dcluster.ContainerDescription, name string) error {
	envs := []string{"constraint:role==slave"}
	if cd.URL != "" {
		url := cd.URL
		idx := strings.IndexAny(url, ".")
		if idx > 0 {
			envs = append(envs, "SERVICE_TAGS="+url[0:idx])
			envs = append(envs, "SERVICE_NAME="+url[idx+1:])
		} else {
			envs = append(envs, "SERVICE_NAME="+url)
		}
	}
	runConfig := dcontainer.ContainerRunConfig{
		Image:        cd.Image,
		Name:         name,
		PortBindings: []dcluster.PortBinding{cd.PortBinding},
		Envs:         envs,
		DNS:          ctx.clusterDesc.ConsulCluster.Server.IPs,
	}

	cid, err := ctx.cProxy.RunByConfig(runConfig)

	if err != nil {
		return errors.New(fmt.Sprintf("Failed to run a container. name: %s, error: %s", name, err.Error()))
	}
	ctx.registerServiceByContainerId(cid, cd)
	return nil
}

func (ctx *ClusterContext) deployContainersByDescription(description *dcluster.ContainerDescription) error {
	stop := func(c *dcontainer.ContainerInfo, ctx *ClusterContext) {
		cid := c.ID
		cName := c.Name[0]
		if len(c.IpPorts) > 0 {
			ip := c.IpPorts[0].IP
			port := c.IpPorts[0].PublicPort
			fmt.Printf("Unregister service before stop a container. cid:%s, name: %s, ip:%s, port:%d\n", cid, cName, ip, port)
			if err := ctx.unregisterService(ip, port, description); err != nil {
				panic(fmt.Sprintf("Failed to unregister container service. cid:%s, name: %s, ip:%s, port:%d. Error:%s\n", cid, cName, ip, port, err.Error()))
			}
		}
		fmt.Printf("Container unregistered, begin to stop an container: container id: %s\n", cid)
		if err := ctx.cProxy.StopContainer(cid); err != nil {
			panic(fmt.Sprintf("Failed to stop container. CID:%s, name:%s, Error:%s", cid, cName, err.Error()))
		}
		fmt.Printf("Container stopped securely and successfully. name:%s, container id: %s\n", cName, cid)
	}

	remove := func(c *dcontainer.ContainerInfo, ctx *ClusterContext) {
		cid := c.ID
		name := c.Name[0]
		fmt.Printf("Remove an container: container id: %s\n", cid)
		if err := ctx.cProxy.RemoveContainer(cid); err != nil {
			fmt.Printf("Failed to remove an container. cid:%s, name:%s, Error:%s\n", cid, name, err.Error())
		}
	}

	run := func(name string, description *dcluster.ContainerDescription, ctx *ClusterContext) {
		fmt.Printf("Run a new container. name:%s, image:%s, group:%s.\n", name, description.Image, description.Name)
		err := ctx.runContainer(description, name)
		if err != nil {
			panic(fmt.Sprintf("Can not run an container. name:%s, image:%s, group:%s. err:%s", name, description.Image, description.Name, err.Error()))
		}
		fmt.Printf("A new container running. name:%s, image:%s, group:%s.\n", name, description.Image, description.Name)
	}

	restart := func(container *dcontainer.ContainerInfo, description *dcluster.ContainerDescription, ctx *ClusterContext) {
		fmt.Printf("Restarting container. cid:%s, image:%s, name:%s\n", container.ID, container.Image, container.Name[0])
		if err := ctx.cProxy.RestartContainer(container.ID); err != nil {
			panic(fmt.Sprintf("Failed to restart container.  cid:%s, image:%s, name:%s, err:%s\n", container.ID, container.Image, container.Name[0], err.Error()))
		}
		fmt.Printf("Container restarted, begin to register. cid:%s, image:%s, name:%s\n", container.ID, container.Image, container.Name[0])
		if err := ctx.registerServiceByContainer(container, description); err != nil {
			panic(fmt.Sprintf("Service Register failed. name:%s, error:%s\n", container.Name[0], err.Error()))
		}
		fmt.Printf("Container successfully. cid:%s, image:%s, name:%s\n", container.ID, container.Image, container.Name[0])
	}

	var wg sync.WaitGroup

	group := description.Name
	containers, err := ctx.cProxy.ListByGroup(group)
	if err != nil {
		panic("Cannot load containers of group '" + group + "'")
	}

	stopNum := 0
	runNum := 0
	expNum := description.Num
	// process the stopped container
	for _, c := range containers {
		if !c.IsUp() {
			stopNum++
			if ctx.removeContainer {
				wg.Add(1)
				go func(c dcontainer.ContainerInfo, ctx *ClusterContext) {
					defer wg.Done()
					remove(&c, ctx)
				}(c, ctx)
			}
		} else {
			runNum++
		}
	}
	wg.Wait()

	// process the running container
	for _, c := range containers {
		if c.IsUp() {
			if c.Image == description.Image && !description.Restart && runNum <= expNum {
				continue
			}
			runNum-- // prepare to stop
			rstart := c.Image == description.Image && runNum < expNum
			wg.Add(1)
			go func(c dcontainer.ContainerInfo, ctx *ClusterContext, rstart bool) {
				defer wg.Done()
				stop(&c, ctx)
				if rstart {
					restart(&c, description, ctx)
					runNum++
				} else {
					if ctx.removeContainer {
						remove(&c, ctx)
					}
				}
			}(c, ctx, rstart)
		}
	}
	wg.Wait()

	left := expNum - runNum
	fmt.Printf("%d Container of group '%s' is running , %d left container will be create and running.\n", runNum, group, left)

	// create new container
	if left > 0 {
		for i := 0; i < left; i++ {
			cd := description
			cName := ctx.nextContainerName(group)
			wg.Add(1)
			fmt.Printf("Deploy a new container(image:%s, name:%s).\n", cd.Image, cName)
			go func(image, name string, ctx *ClusterContext) {
				defer wg.Done()
				run(cName, cd, ctx)
				runNum++
			}(cd.Image, cName, ctx)
		}
		wg.Wait()
	}
	fmt.Printf("Container deployed for group '%s'. expect:%d started:%d.\n", group, expNum, runNum)
	return nil
}

func (ctx *ClusterContext) splitDescription() (map[string]dcluster.ContainerDescription, []string, []string) {
	gDeps := dcontainer.NewContainerGroupDeps()
	descriptions := map[string]dcluster.ContainerDescription{}
	sdGroups := map[string]bool{}

	sdNameGroupMap := map[string]string{}
	for sdName, sdd := range ctx.clusterDesc.ServiceDiscovers {
		if group, ok := sdd["container"]; ok {
			sdNameGroupMap[sdName] = group
			sdGroups[group] = true
		} else {
			fmt.Printf("Service discover container missed. discovery name:%s.\n", sdName)
		}
	}

	fmt.Printf("Service discover container group:%+v\n", sdGroups)

	for _, description := range ctx.clusterDesc.Containers {
		group := description.Name
		descriptions[group] = description
		fmt.Printf("Description: group:%s, description:%+v, total:%+v\n", group, description, descriptions)
		// add deps
		fmt.Printf("Add dependancy for group: group:%s, deps:%+v\n", group, description.Deps)

		gDeps.AddDeps(group, description.Deps)

		if description.ServiceDiscover != "" {
			if depGroup, ok := sdNameGroupMap[description.ServiceDiscover]; ok {
				fmt.Printf("Add dependancy. group:%s, deps:%+v\n", group, depGroup)
				gDeps.AddDeps(group, []string{depGroup})
			}
		}
	}

	fmt.Printf("Total description:%+v\n", descriptions)

	allGroups := gDeps.List()
	sdDGroups := []string{}
	bizDGroups := []string{}
	for _, group := range allGroups {
		if _, ok := sdGroups[group]; ok {
			sdDGroups = append(sdDGroups, group)
		} else {
			bizDGroups = append(bizDGroups, group)
		}
	}
	return descriptions, sdDGroups, bizDGroups
}

func (ctx *ClusterContext) deployContainers() error {
	deploy := func(ctx *ClusterContext, descriptions map[string]dcluster.ContainerDescription, groups []string) error {
		for _, group := range groups {
			description, ok := descriptions[group]
			if !ok {
				panic(fmt.Sprintf("No description found for group:%s", group))
			}
			fmt.Printf("Deploy container for group:%s. description:%+v\n", group, description)
			if err := ctx.deployContainersByDescription(&description); err != nil {
				panic(fmt.Sprintf("Failed to deploy container for group:%s. err:%s\n", group, err.Error()))
			}
		}
		return nil
	}

	descriptions, sdGroups, bizGroups := ctx.splitDescription()

	fmt.Printf("Deploy service discover group container:%+v\n", sdGroups)
	deploy(ctx, descriptions, sdGroups)

	if err := ctx.loadContainers(); err != nil {
		return errors.New("Failed to load containers:" + err.Error())
	}

	fmt.Printf("Init service discovery:%+v\n", ctx.clusterDesc.ServiceDiscovers)
	if err := ctx.initServiceDiscovery(ctx.containerInfos); err != nil {
		panic("Init service discovery failed:" + err.Error())
	}

	fmt.Printf("Deploy biz group container:%+v\n", bizGroups)
	deploy(ctx, descriptions, bizGroups)
	return nil
}

func (ctx *ClusterContext) scaleMachineOut() error {
	if !ctx.scaleOut {
		fmt.Printf("No need to scale machine out.\n")
		return nil
	}
	slaves := ctx.getSlaves()
	min := ctx.clusterDesc.Machine.MinNum
	stoppedMachineNames := []string{}
	runningNum := int(0)
	for _, mi := range slaves {
		if mi.IsRunning() {
			runningNum++
		} else {
			stoppedMachineNames = append(stoppedMachineNames, mi.Name)
		}
	}
	if runningNum >= min {
		fmt.Printf("Running machines num is enough(%d) for the cluster minimal need(%d)\n", runningNum, min)
		return nil
	}
	fmt.Printf("Running machines num is 'not' enough(%d) for the cluster minimal need(%d)\n", runningNum, min)

	var wg sync.WaitGroup
	sl := len(stoppedMachineNames)
	toBeStartNum := int(math.Min(float64(min-runningNum), float64(sl)))
	toBeStoppedMachineNames := stoppedMachineNames[0:toBeStartNum]
	startedNum := int(0)

	if toBeStartNum > 0 {
		fmt.Printf("Starting %d stopped machines: %+v\n", toBeStartNum, toBeStoppedMachineNames)
		wg.Add(1)
		go func(names []string, ctx *ClusterContext) {
			defer wg.Done()
			startedNames, _ := ctx.mProxy.Start(names...)
			startedNum = int(len(startedNames))
		}(toBeStoppedMachineNames, ctx)
	}
	toBeCreateNum := min - runningNum - toBeStartNum
	createdNum := int(0)
	if toBeCreateNum > 0 {
		fmt.Printf("Creating %d machines\n", toBeCreateNum)
		wg.Add(1)
		go func(num int, ctx *ClusterContext) {
			defer wg.Done()
			startedNames, _ := ctx.createSlaves(num)
			createdNum = int(len(startedNames))
		}(toBeCreateNum, ctx)
	}
	fmt.Printf("Waiting Starting(%d) and Creating(%d) machines.\n", toBeStartNum, toBeCreateNum)
	wg.Wait()
	fmt.Printf("Starting(%d) and Creating(%d) machines complete with %d started and %d created.\n", toBeStartNum, toBeCreateNum, startedNum, createdNum)
	runningNum = runningNum + startedNum + createdNum
	if runningNum < min && createdNum == toBeCreateNum && ctx.forceCreate { // just cause start stopped machine failed, and try to create new ones
		toBeforceCreateNum := min - runningNum
		fmt.Printf("%d machines started failed, try to create %d new machines", (toBeStartNum - startedNum), toBeforceCreateNum)
		fCreatedNames, _ := ctx.createSlaves(toBeforceCreateNum)
		runningNum += int(len(fCreatedNames))
	}

	if runningNum < min {
		errInfo := fmt.Sprintf("Machine num if not enough. Rumming is %d, but the minimal requirements is %d.\n", runningNum, min)
		return errors.New(errInfo)
	}
	ctx.reloadMachineInfos()
	fmt.Printf("Running machine num scale out up to %d, the minimal requirements is %d.\n", runningNum, min)
	return nil
}

func (ctx *ClusterContext) getMachines(filter func(mi *dmachine.MachineInfo) bool) []dmachine.MachineInfo {
	machines := []dmachine.MachineInfo{}

	for _, m := range ctx.machineInfos {
		if filter(&m) {
			machines = append(machines, m)
		}
	}
	return machines[0:]
}

func (ctx *ClusterContext) getSlaves() []dmachine.MachineInfo {
	return ctx.getMachines(func(mi *dmachine.MachineInfo) bool {
		return !mi.IsMaster()
	})
}

func (ctx *ClusterContext) startMaster() error {
	masterMachineInfo, exists := ctx.getMaster()
	if !exists {
		if ctx.create {
			fmt.Printf("Master is not exists on the cluster, creating an new master.\n")
			if err := ctx.mProxy.CreateMaster(); err != nil {
				return err
			}
			ctx.reloadMachineInfos()
			fmt.Printf("Master node create and running.\n")
		} else {
			return errors.New("Master Not Exists.")
		}
	} else {
		if !masterMachineInfo.IsRunning() {
			fmt.Printf("Master node(%s) is not running, try to start...\n", masterMachineInfo.Name)
			if succNames, errs := ctx.mProxy.Start(masterMachineInfo.Name); len(succNames) == 0 {
				return errs[0]
			}
			ctx.reloadMachineInfos()
			fmt.Printf("Master node(%s) started.\n", masterMachineInfo.Name)
		}
	}
	return nil
}

func (ctx *ClusterContext) startConsulCluster() error {
	server := ctx.clusterDesc.ConsulCluster.Server
	if len(server.IPs) > 0 {
		fmt.Printf("Consule server ips(%+v) provided and managed outsided of dockerf.\n", server.IPs)
		return nil
	}
	nodes := server.Nodes
	if len(nodes) == 0 {
		return errors.New("Consul server nodes missed.")
	}

	serverMachineInfos := ctx.getMachines(func(mi *dmachine.MachineInfo) bool {
		for _, node := range nodes {
			if node == mi.Name {
				return true
			}
		}
		return false
	})
	consulServerIPs := []string{}
	var lock sync.Mutex
	var wg sync.WaitGroup

	// serverMachineInfos = []dmachine.MachineInfo{}
	if len(serverMachineInfos) > 0 {
		ips := []string{}
		fmt.Printf("Consul server(num:%d) is exists, no need to create consul cluster.\n", len(serverMachineInfos))
		for _, mInfo := range serverMachineInfos {
			if !mInfo.IsRunning() {
				wg.Add(1)
				fmt.Printf("Consul server '%s' is not running, try to start it.\n")
				go func(ctx *ClusterContext, mInfo dmachine.MachineInfo) {
					defer wg.Done()
					if _, errs := ctx.mProxy.Start(mInfo.Name); len(errs) == 0 {
						lock.Lock()
						defer lock.Unlock()
						ips = append(ips, mInfo.IP)
					} else {
						fmt.Printf("Start consule server('%s') failed. err:%s\n", errs[0].Error())
					}
				}(ctx, mInfo)
			} else {
				lock.Lock()
				ips = append(ips, mInfo.IP)
				lock.Unlock()
			}
		}
		wg.Wait()
		fmt.Printf("Consul server ips is:%+v\n", ips)
		consulServerIPs = ips
	} else {
		errs := ""
		fmt.Printf("Consul server is not exists, creating a new consul cluster with machine names:%+v\n", nodes)
		for _, node := range nodes {
			wg.Add(1)
			go func(ctx *ClusterContext, n string) {
				defer wg.Done()
				opts := []string{
					"--engine-label",
					"role=consulserver",
				}
				if err := ctx.mProxy.CreateMachine(n, opts...); err != nil {
					fmt.Printf("Failed to create consul server. node:%s, err:%s\n", n, err.Error())
					errs = errs + "--" + err.Error()
				} else {
					fmt.Printf("One consul server created. name:%s\n", n)
				}
			}(ctx, node)
		}
		wg.Wait()

		if errs != "" {
			return errors.New(errs)
		}
		fmt.Printf("All Consul server created. names:%+v\n", nodes)
		serverIPs, err := ctx.mProxy.IPs(nodes)
		if err != nil {
			return errors.New("Failed to load consul server ips. err:" + err.Error())
		}
		bootstrapServerNode := nodes[0]
		bootstrapServerIp := serverIPs[0]
		fmt.Printf("Run bootstrap consul server on. node:%s, ip:%s\n", bootstrapServerNode, bootstrapServerIp)
		sleepSeconds := 3 * time.Second
		fmt.Printf("To ensure the swarm cluster to load the created consul server, sleep 60 seconds\n")
		time.Sleep(sleepSeconds)
		if cid, err := ctx.runConsulBootstrapServer(bootstrapServerNode, bootstrapServerIp); err != nil {
			fmt.Printf("Failed to run bootstrap consul server. node:%s, ip:%s. err:%s\n", bootstrapServerNode, bootstrapServerIp, err.Error())
			return err
		} else {
			fmt.Printf("Successfully run bootstrap consul server. cid:%s, node:%s, ip:%s. err:%s\n", cid, bootstrapServerNode, bootstrapServerIp)
		}
		if len(serverIPs) > 1 {
			errs := ""
			joinIps := serverIPs[1:]
			joinNodes := nodes[1:]
			for idx, ip := range joinIps {
				node := joinNodes[idx]
				fmt.Printf("Run join consul server on node:%s, ip:%s\n", node, ip)
				wg.Add(1)
				go func(ctx *ClusterContext, n string, ip string) {
					defer wg.Done()
					if cid, err := ctx.runConsulJoinServer(server, n, ip, bootstrapServerIp); err != nil {
						errs = errs + "--" + err.Error()
						fmt.Printf("Failed to run consul join server. join node:%s, join ip:%s, boot strap node:%s, boot strap ip:%s. err:%s\n", n, ip, bootstrapServerNode, bootstrapServerIp, err.Error())
					} else {
						fmt.Printf("Successfully run consul join server. cid:%s, join node:%s, join ip:%s, boot strap node:%s, boot strap ip:%s\n", cid, n, ip, bootstrapServerNode, bootstrapServerIp)
					}

				}(ctx, node, ip)
			}
			wg.Wait()
			if errs != "" {
				return errors.New(errs)
			}
		}
		consulServerIPs = serverIPs
	}
	ctx.clusterDesc.ConsulCluster.Server.IPs = consulServerIPs

	fmt.Printf("Consul server cluster start complete. ips:%+v\n", ctx.clusterDesc.ConsulCluster.Server.IPs)
	return nil
}
