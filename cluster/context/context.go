package context

import (
	"errors"
	"fmt"
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
	mScaleIn          bool
	mScaleOut         bool
	cScaleIn          bool
	cScaleOut         bool
	rmc               bool
	forceCreate       bool
	cStepPercent      int
	filters           *opts.ListOpts
	machineInfos      []dmachine.MachineInfo
	containerInfos    []dcontainer.ContainerInfo
	mSeq              sequence.Seq
	cSeqs             map[string]*sequence.Seq
	mProxy            *dmachine.MachineClusterProxy
	cProxy            *dcontainer.DockerProxy
	serviceRegistries map[string]*discovery.ServiceRegisterDriver
}

func NewClusterContext(mScaleIn, mScaleOut, cScaleIn, cScaleout, rmc bool, cFilter *opts.ListOpts, cStepPercent int, cluster *dcluster.Cluster) *ClusterContext {
	clusterContext := &ClusterContext{
		create:            true,
		clusterDesc:       cluster,
		filters:           cFilter,
		mScaleIn:          mScaleIn,
		mScaleOut:         mScaleOut,
		cScaleIn:          cScaleIn,
		cScaleOut:         cScaleout,
		rmc:               rmc,
		cStepPercent:      cStepPercent,
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

	fmt.Printf("Starting machine master\n")
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

	fmt.Printf("Loading all filtered container infos")
	if err := ctx.loadContainers(); err != nil {
		panic("Init cluster context error, cannot list container infos:" + err.Error())
	}

	fmt.Printf("Init container sequences.\n")
	if err := ctx.initContainerSequences(); err != nil {
		panic("Init cluster context error, cannot init container seqs:" + err.Error())
	}

	fmt.Println("cluster context inited successfully")
}

func (ctx *ClusterContext) parsePortBindings() error {
	for group, description := range ctx.clusterDesc.Container.Topology {
		binding := dcluster.PortBinding{}
		if err := binding.Parse(description.Port); err != nil {
			return err
		}
		description.PortBinding = binding
		ctx.clusterDesc.Container.Topology[group] = description
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

func (ctx *ClusterContext) loadAllContainers() ([]dcontainer.ContainerInfo, error) {
	return ctx.cProxy.ListAll()
}

func (ctx *ClusterContext) getContainerByGroup(group string) []dcontainer.ContainerInfo {
	infos := []dcontainer.ContainerInfo{}
	for _, info := range ctx.containerInfos {
		if info.Group == group {
			infos = append(infos, info)
		}
	}
	return infos
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

	for sdName, sdd := range ctx.clusterDesc.ServiceDiscover {
		driver, err := discovery.NewRegDriver(sdd)
		if err != nil {
			return errors.New(fmt.Sprintf("Failed to create Reg Driver:'%s'. err:%s", sdName, err.Error()))
		}
		containerGroup, ok := sdd["container"]
		fmt.Printf("Load registry for '%s'\n", containerGroup)
		if ok {
			fmt.Printf("Load registry for service discover. sd name: '%s', group: '%s'.\n", sdName, containerGroup)
			description, exists := ctx.clusterDesc.Container.Topology[containerGroup]
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
	fmt.Printf("Registering container service: cid:%s, description:%s\n, ", cid, cd)

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

func (ctx *ClusterContext) initContainerSequences() error {
	cifs := ctx.containerInfos
	if len(ctx.filters.GetAll()) > 0 {
		aCifs, err := ctx.loadAllContainers()
		if err != nil {
			return err
		}
		cifs = aCifs
	}
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
	return nil
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

func (ctx *ClusterContext) initSlave(node string, md dcluster.MachineDescription) error {
	fmt.Printf("Init the machine infrastructure envment: '%s'\n", node)
	// exec command on this machine
	command := strings.TrimSpace(md.Init)
	if command != "" {
		if err := ctx.mProxy.ExecCmd(node, command); err != nil {
			fmt.Printf("Failed to exec command(%s) on '%s'\n", command, node)
		}
	}

	if md.Consul {
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
	}

	fmt.Sprintf("Slave machine init successfully. node:%s\n", node)
	return nil
}

func (ctx *ClusterContext) createSlaves(group string, md dcluster.MachineDescription, num int) ([]string, error) {
	successNodeNames := []string{}
	errs := []string{}
	var lock sync.Mutex
	var wg sync.WaitGroup
	wg.Add(int(num))
	for i := 0; i < num; i++ {
		go func(ctx *ClusterContext) {
			defer wg.Done()
			name, err := ctx.mProxy.CreateSlave(group, md)
			if err != nil {
				fmt.Printf("Failed to Create machine of group '%s'. Error:%s\n", group, err.Error())
				errs = append(errs, err.Error())
			} else {
				fmt.Printf("Machine(%s) created and started, begin to init slave.\n", name)
				if err := ctx.initSlave(name, md); err != nil {
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
	var err error = nil
	if len(errs) > 0 {
		err = errors.New(strings.Join(errs, "---"))
	}
	return successNodeNames[0:], err
}

func (ctx *ClusterContext) scaleMachineIn() error {
	if !ctx.mScaleIn {
		fmt.Printf("No need to scale in.\n")
		return nil
	}
	runningMachines := ctx.getMachines(func(mi *dmachine.MachineInfo) bool {
		return !mi.IsMaster()
	})
	rNum := int(len(runningMachines))

	// max := ctx.clusterDesc.Machine.MaxNum
	max := 5
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

func (ctx *ClusterContext) runContainer(cd *dcluster.ContainerDescription, group string) {
	name := ctx.nextContainerName(group)
	fmt.Printf("Run a new container. name:%s, image:%s, group:%s.\n", name, cd.Image, group)
	envs := []string{"constraint:role==slave", "constraint:group==" + cd.Machine}
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
		Bindings:     cd.Volums,
	}

	cid, err := ctx.cProxy.RunByConfig(runConfig)

	if err != nil {
		panic(fmt.Sprintf("Failed to run a container. name: %s, error: %s", name, err.Error()))
	}
	err = ctx.registerServiceByContainerId(cid, cd)
	if err != nil {
		panic(fmt.Sprintf("Failed to register service of container. id:%s, description:%+v err:%s", cid, *cd, err.Error()))
	}
}

func (ctx *ClusterContext) removeStoppedContainerByGroup(group string) {
	if !ctx.rmc {
		fmt.Printf("No need to remove stopped container of group '%s'\n", group)
		return
	}
	containers := ctx.getContainerByGroup(group)
	var wg sync.WaitGroup
	for _, c := range containers {
		if c.IsUp() {
			continue
		}
		wg.Add(1)
		fmt.Printf("Remove an container: container id: %s name:%s\n", c.ID, c.Name[0])

		go func(c *dcontainer.ContainerInfo) {
			defer wg.Done()
			if err := ctx.cProxy.RemoveContainer(c.ID); err != nil {
				fmt.Printf("Failed to remove an container. cid:%s, name:%s, Error:%s\n", c.ID, c.Name[0], err.Error())
			} else {
				fmt.Printf("Successfully to remove an container. cid:%s, name:%s. %s\n", c.ID, c.Name[0])
			}
		}(&c)
	}
	wg.Wait()
}

func (ctx *ClusterContext) stopContainer(c *dcontainer.ContainerInfo, description *dcluster.ContainerDescription) {
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

func (ctx *ClusterContext) startContainer(container *dcontainer.ContainerInfo, description *dcluster.ContainerDescription) {
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

func (ctx *ClusterContext) deployRunningContainersByDescription(group string, description *dcluster.ContainerDescription) error {
	runningContainers := func() []dcontainer.ContainerInfo {
		containers := ctx.getContainerByGroup(group)
		running := []dcontainer.ContainerInfo{}
		for _, c := range containers {
			if c.IsUp() {
				running = append(running, c)
			}
		}
		return running
	}()
	var lock sync.Mutex
	total := len(runningContainers)
	sim := int(float64(total) * float64(ctx.cStepPercent) / float64(100))
	if sim <= 0 {
		fmt.Printf("The simutaneous container is less than 0(%d). 1 will be set.", sim)
		sim = 1
	}
	fmt.Printf("Simutaneous of group '%s' is %d\n", group, sim)
	blocking := make(chan int, sim)
	done := make(chan bool)
	doneNum := 0

	for idx, c := range runningContainers {
		if c.Image == description.Image && !description.Restart {
			continue
		}
		blocking <- idx
		go func(c dcontainer.ContainerInfo, ctx *ClusterContext) {
			defer func() {
				lock.Lock()
				defer lock.Unlock()
				<-blocking
				doneNum++
				if doneNum >= total {
					done <- true
				}
			}()
			ctx.stopContainer(&c, description)
			if c.Image == description.Image {
				ctx.startContainer(&c, description) // just restart
			} else {
				ctx.runContainer(description, group)
			}

		}(c, ctx)
	}
	<-done
	return nil
}

func (ctx *ClusterContext) scaleOutContainersByDescription(group string, description *dcluster.ContainerDescription) error {
	if !ctx.cScaleOut {
		fmt.Printf("--scale-out is set to false(not set), not need to scale container out for group '%s'.\n", group)
	}
	containers := ctx.getContainerByGroup(group)
	running := 0
	for _, c := range containers {
		if c.IsUp() {
			running++
		}
	}
	createNum := description.Num - running
	if createNum <= 0 {
		fmt.Printf("No need to scale out container for group '%s'. running:%d, need:%d\n", group, running, description.Num)
		return nil
	}
	fmt.Printf("%d container will be created and run of group '%s'.\n", createNum, group)

	var wg sync.WaitGroup
	for i := 0; i < createNum; i++ {
		wg.Add(1)
		go func(group string, cd *dcluster.ContainerDescription) {
			defer wg.Done()
			ctx.runContainer(cd, group)
		}(group, description)
	}
	wg.Wait()
	return nil
}

func (ctx *ClusterContext) scaleInContainersByDescription(group string, description *dcluster.ContainerDescription) error {
	if !ctx.cScaleIn {
		fmt.Printf("--scale-in is set to false(not set), not need to scale container out for group '%s'.\n", group)
	}
	containers := ctx.getContainerByGroup(group)
	running := 0
	for _, c := range containers {
		if c.IsUp() {
			running++
		}
	}
	needStopped := running - description.Num
	if needStopped <= 0 {
		fmt.Printf("No need to scale in container for group '%s'. running:%d, need:%d\n", group, running, description.Num)
		return nil
	}
	fmt.Printf("%d container will be stopped of group '%s'.\n", needStopped, group)

	var wg sync.WaitGroup
	stopped := 0
	for _, c := range containers {
		if !c.IsUp() {
			continue
		}
		if stopped >= needStopped {
			break
		}
		stopped++
		wg.Add(1)
		go func(group string, cd *dcluster.ContainerDescription) {
			defer wg.Done()
			ctx.stopContainer(&c, description)
		}(group, description)
	}
	wg.Wait()
	return nil
}

func (ctx *ClusterContext) deployContainersByDescription(group string, description *dcluster.ContainerDescription) error {
	ctx.deployRunningContainersByDescription(group, description)

	ctx.scaleOutContainersByDescription(group, description)

	ctx.scaleInContainersByDescription(group, description)

	ctx.removeStoppedContainerByGroup(group)

	return nil
}

func (ctx *ClusterContext) splitDescription() (map[string]dcluster.ContainerDescription, []string, []string) {
	gDeps := dcontainer.NewContainerGroupDeps()
	descriptions := map[string]dcluster.ContainerDescription{}
	sdGroups := map[string]bool{}

	sdNameGroupMap := map[string]string{}
	for sdName, sdd := range ctx.clusterDesc.ServiceDiscover {
		if group, ok := sdd["container"]; ok {
			sdNameGroupMap[sdName] = group
			sdGroups[group] = true
		} else {
			fmt.Printf("Service discover container missed. discovery name:%s.\n", sdName)
		}
	}

	fmt.Printf("Service discover container group:%+v\n", sdGroups)

	for group, description := range ctx.clusterDesc.Container.Topology {
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
			if err := ctx.deployContainersByDescription(group, &description); err != nil {
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

	fmt.Printf("Init service discovery:%+v\n", ctx.clusterDesc.ServiceDiscover)
	if err := ctx.initServiceDiscovery(ctx.containerInfos); err != nil {
		panic("Init service discovery failed:" + err.Error())
	}

	fmt.Printf("Deploy biz group container:%+v\n", bizGroups)
	deploy(ctx, descriptions, bizGroups)
	return nil
}

// return: running, restart, error
func (ctx *ClusterContext) startMachines(group string, machines []dmachine.MachineInfo, md dcluster.MachineDescription) (int, int, error) {
	running := 0
	stoppedNodes := func(ms []dmachine.MachineInfo) []string {
		stopped := []string{}
		for _, m := range ms {
			if !m.IsRunning() {
				stopped = append(stopped, m.Name)
			} else {
				running++
			}
		}
		return stopped
	}(machines)

	orgRunNum := running

	max := md.MinNum
	if running >= max {
		fmt.Printf("Enough node running, no need to start extra machines. group:%s, running:%d, max:%d\n", group, running, max)
		return running, 0, nil
	}

	errs := []string{}

	start := 0
	end := -1
	for {
		start = end + 1
		need := max - running
		end = start + need
		if end > len(stoppedNodes) {
			end = len(stoppedNodes)
		}
		if start >= end {
			break
		}
		fmt.Printf("nodes:%+v, start:%d, end:%d\n", stoppedNodes, start, end)
		toBeStart := stoppedNodes[start:end]

		if len(toBeStart) == 0 || running >= max {
			fmt.Printf("No stopped machines exists or running machine is enough. stopped:%d, running:%d, max nedd:%d\n", len(toBeStart), running, max)
			break
		}
		succesNames, err := ctx.mProxy.Start(toBeStart...)
		running = running + len(succesNames)
		if err != nil {
			fmt.Printf("Start stopped machine failed. names:%+v, err:%s\n", toBeStart, err.Error())
			errs = append(errs, err.Error())
		}
	}

	var err error = nil
	if len(errs) > 0 {
		err = errors.New(strings.Join(errs, "--"))
	}

	return running, running - orgRunNum, err
}

func (ctx *ClusterContext) scaleMachineOutByGroup(group string, md dcluster.MachineDescription) error {
	min := md.MinNum
	machines, err := ctx.mProxy.ListByGroup(group)
	if err != nil {
		fmt.Printf("Failed to load machine for group:%s\n", err.Error())
		return err
	}

	fmt.Printf("Start stopped machines for group:%s\n", group)
	runningNum, restarted, err := ctx.startMachines(group, machines, md)
	if err != nil {
		fmt.Printf("Error happend when Start machines for group:%s. total running:%d, need: %d, err:%s\n", group, runningNum, min, err.Error())
	} else {
		fmt.Printf("Start machines complete for group:%s. running:%d, need: %d\n", group, runningNum, min)
	}

	if runningNum >= min {
		fmt.Printf("Running machines num is enough(%d) for the cluster minimal need(%d)\n", runningNum, min)
		return nil
	}
	fmt.Printf("Running machines num is 'not' enough(%d) for the cluster minimal need(%d)\n", runningNum, min)

	toBeCreateNum := min - runningNum
	fmt.Printf("Creating %d machines of group '%s'\n", toBeCreateNum, group)
	startedNames, err := ctx.createSlaves(group, md, toBeCreateNum)
	createdNum := len(startedNames)
	runningNum += createdNum
	if err != nil {
		fmt.Printf("Error happened when create machines of group '%s'. err:%s\n", group, err.Error())
	}

	fmt.Printf("%d machines running, of which %d restarted and %d created.group: %s.\n", runningNum, restarted, createdNum, group)

	if runningNum < min {
		errInfo := fmt.Sprintf("Machine num if not enough. Rumming is %d, but the minimal requirements is %d.\n", runningNum, min)
		return errors.New(errInfo)
	}

	fmt.Printf("Running machine num scale out up to %d, the minimal requirements is %d.\n", runningNum, min)
	return nil
}

func (ctx *ClusterContext) scaleMachineOut() error {
	if !ctx.mScaleOut {
		fmt.Printf("No need to scale machine out.\n")
		return nil
	}
	for group, md := range ctx.clusterDesc.Machine.Topology {
		if err := ctx.scaleMachineOutByGroup(group, md); err != nil {
			return err
		}
	}
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
			if succNames, err := ctx.mProxy.Start(masterMachineInfo.Name); len(succNames) == 0 {
				return err
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
					if _, err := ctx.mProxy.Start(mInfo.Name); err == nil {
						lock.Lock()
						defer lock.Unlock()
						ips = append(ips, mInfo.IP)
					} else {
						fmt.Printf("Start consule server('%s') failed. err:%s\n", err.Error())
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
