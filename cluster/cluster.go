package cluster

import (
	"fmt"
	"io/ioutil"
	"strconv"
	"strings"

	log "github.com/Sirupsen/logrus"

	"errors"
	"os"
	"regexp"

	dutils "github.com/weibocom/dockerf/utils"
	"gopkg.in/yaml.v2"
)

type Disk struct {
	Type     string
	Capacity string
}

type MachineDescription struct {
	MaxNum     int
	MinNum     int
	Cpu        int
	Disk       Disk
	Memory     string
	Init       string
	Region     string
	Consul     bool
	DriverOpts []string
	Cloud      string
	Group      string
}

func (md *MachineDescription) GetCpu() int {
	if md.Cpu <= 0 {
		return 1
	}
	return md.Cpu
}

func (md *MachineDescription) GetMemInBytes() int {
	if md.Memory == "" {
		return 512 * 1024 * 1024 // default is 512m
	}
	if bytes, err := dutils.ParseCapacity(md.Memory); err != nil {
		panic(fmt.Sprintf("'%s' is not a valid memory option.", md.Memory))
	} else {
		return bytes
	}
}

func (md *MachineDescription) GetDiskCapacityInBytes() int {
	if md.Disk.Capacity == "" {
		return 20 * 1024 * 1024 * 1024 // default is 10gb
	}
	if bytes, err := dutils.ParseCapacity(md.Disk.Capacity); err != nil {
		panic(fmt.Sprintf("'%s' is not a valid ssd capacity option.", md.Disk.Capacity))
	} else {
		return bytes
	}
}

type CloudDrivers map[string]CloudDriverDescription

type MachineTopology []MachineDescription

func (ct *MachineTopology) GetDescription(group string) (*MachineDescription, bool) {
	for idx, description := range *ct {
		if description.Group == group {
			return &(*ct)[idx], true
		}
	}
	return nil, false
}

type MachineCluster struct {
	OS       string
	Cloud    CloudDrivers
	Topology MachineTopology
}

func (cds *CloudDrivers) SurportedDrivers() []string {
	names := []string{}
	for name, _ := range *cds {
		names = append(names, name)
	}
	return names
}

type ContainerTopology []ContainerDescription

func (ct *ContainerTopology) GetDescription(group string) (*ContainerDescription, bool) {
	for idx, description := range *ct {
		if description.Group == group {
			return &(*ct)[idx], true
		}
	}
	return nil, false
}

type ContainerCluster struct {
	Topology ContainerTopology
}

const (
	ContainerDescription_TYPE_SD = 1
	ContainerDescription_TYPE_BZ = 2
	MultiPort_Separator          = "|"
	CONFIG_PLACEHOLDER_PATTERN   = "\\$\\{(\\w+)\\}"
)

type ContainerDescription struct {
	Num             int
	Image           string
	PreStop         string
	PostStart       string
	URL             string
	Port            string
	Deps            []string
	ServiceDiscover string
	Restart         bool
	Machine         string
	PortBinding     PortBinding
	Volums          []string
	Group           string
	Env             []string
	DepLevel        int
	Type            int
}

type SortContainerDescriptionDescByLevel []ContainerDescription

func (s SortContainerDescriptionDescByLevel) Len() int {
	return len(s)
}
func (s SortContainerDescriptionDescByLevel) Swap(i, j int) {
	s[i], s[j] = s[j], s[i]
}
func (s SortContainerDescriptionDescByLevel) Less(i, j int) bool {
	return s[i].DepLevel > s[j].DepLevel
}

type CloudDriverDescription struct {
	Options       string
	GlobalOptions string
	Default       bool
}

func (cdd *CloudDriverDescription) GetOptions() []string {
	if cdd.Options == "" {
		return []string{}
	}
	return strings.Split(cdd.Options, " ")
}

func (cdd *CloudDriverDescription) GetGlobalOptions() []string {
	if cdd.GlobalOptions == "" {
		return []string{}
	}
	return strings.Split(cdd.GlobalOptions, " ")
}

type ConsulServer struct {
	Service string
	Image   string
	Domain  string
	Nodes   []string
	IPs     []string
	Create  bool
	Machine string
}

type ConsulAgent struct {
	Image string
}

type ConsulRegistrator struct {
	Image string
}

type ConsulDescription struct {
	Server      ConsulServer
	Agent       ConsulAgent
	Registrator ConsulRegistrator
}

type ServiceDiscoverDiscription map[string]string

type Profile map[string]string

func (p *Profile) lenth() int {
	return len(map[string]string(*p))
}

type Cluster struct {
	ClusterBy   string // such as swarm
	Master      string
	MasterGroup string
	Discovery   string

	Machine   MachineCluster
	Container ContainerCluster
	// Containers []ContainerDescription

	ServiceDiscover map[string]ServiceDiscoverDiscription
	ConsulCluster   ConsulDescription
}

type ClusterProfiles struct {
	ActiveProfile string
	Profiles      map[string]Profile
}

func (p *ClusterProfiles) findProfile(profileName string) (Profile, error) {
	if profileName != "" {
		p.ActiveProfile = profileName
	}

	if profile, exist := p.Profiles[p.ActiveProfile]; !exist {
		return nil, errors.New(fmt.Sprintf("Fail to find profile %s in config files, please check again... ", p.ActiveProfile))
	} else {
		return profile, nil
	}
}

func NewCluster(configFilePos, profileName, profileFileName string) (*Cluster, error) {

	clusterProfiles, exist, err := resolveProfileFile(profileFileName)
	if err != nil {
		return nil, err
	}

	b, err := ioutil.ReadFile(configFilePos)
	if err != nil {
		return nil, err
	}
	if exist {
		profile, err := clusterProfiles.findProfile(profileName)
		if err != nil {
			return nil, err
		}
		log.Debugf("Use profile %+v", profile)
		b = []byte(applyProfile(profile, string(b)))
	}

	c := &Cluster{}
	err = yaml.Unmarshal(b, c)
	if err != nil {
		return nil, err
	}

	c.replaceClusterConfigInfo()

	return c, nil
}

func resolveProfileFile(profileFileName string) (*ClusterProfiles, bool, error) {
	clusterProfiles := &ClusterProfiles{}
	b, err := ioutil.ReadFile(profileFileName)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, false, nil
		} else {
			return nil, true, err
		}
	}
	err = yaml.Unmarshal(b, clusterProfiles)
	if err != nil {
		return nil, true, err
	}
	return clusterProfiles, true, nil
}

func applyProfile(profile Profile, configContent string) string {
	if profile.lenth() <= 0 {
		return configContent
	}

	reg := regexp.MustCompile(CONFIG_PLACEHOLDER_PATTERN)
	return reg.ReplaceAllStringFunc(configContent, func(b string) string {
		key := reg.ReplaceAllString(b, "${1}")
		if value, exist := profile[key]; exist {
			return value
		} else {
			panic(fmt.Sprintf("Fail to apply profile , because a config item named %s is not well configured..,", key))
		}
	})
}

func (cluster *Cluster) parsePortBindings() error {
	for group, description := range cluster.Container.Topology {
		binding := PortBinding{}
		if err := binding.Parse(description.Port); err != nil {
			return err
		}
		description.PortBinding = binding
		cluster.Container.Topology[group] = description
		log.Debugf("Port binding parsed. group: %s, binding:%+v\n", description.Group, binding)
	}
	return nil
}

func (cluster *Cluster) parseMultiPort(cd ContainerDescription) []ContainerDescription {
	port := cd.Port
	hostport := ContainerPort(port).GetHostPort()
	if strings.Contains(hostport, MultiPort_Separator) {
		cds := []ContainerDescription{}
		protocol := ContainerPort(port).GetContainerProtocol()
		containerPort := ContainerPort(port).GetContainerPort()
		ports := strings.Split(hostport, MultiPort_Separator)
		for _, p := range ports {
			newHostPort := ContainerPort(port).buildContainerPort(protocol, containerPort, p)
			newContainerDescription := ContainerDescription{
				Num:             cd.Num,
				Image:           cd.Image,
				PreStop:         cd.PreStop,
				PostStart:       cd.PostStart,
				URL:             cd.URL,
				Port:            newHostPort,
				Deps:            cd.Deps,
				ServiceDiscover: cd.ServiceDiscover,
				Restart:         cd.Restart,
				Machine:         cd.Machine,
				Volums:          cd.Volums,
				Group:           cd.Group,
				Env:             cd.Env,
				DepLevel:        cd.DepLevel,
				Type:            cd.Type,
			}
			cds = append(cds, newContainerDescription)
		}
		return cds
	}
	return []ContainerDescription{cd}
}

func (cluster *Cluster) replaceClusterConfigInfo() {
	cluster.replaceContainerConfigInfo()
}

func (cluster *Cluster) replaceContainerConfigInfo() {
	replacedContainerInfo := ContainerCluster{}
	replacedTopology := ContainerTopology{}
	for _, containerInfo := range cluster.Container.Topology {
		replacedTopology = append(replacedTopology, cluster.parseMultiPort(containerInfo)...)
	}
	replacedContainerInfo.Topology = replacedTopology
	cluster.Container = replacedContainerInfo

	if err := cluster.parsePortBindings(); err != nil {
		log.Fatal("Fail to parse port binding info, please check config file... ")
	}

	for index, containerInfo := range cluster.Container.Topology {
		cluster.Container.Topology[index] = cluster.replaceContainerPlaceholder(containerInfo)
	}
}

func (cluster *Cluster) replaceContainerPlaceholder(cd ContainerDescription) ContainerDescription {
	cd.URL = strings.Replace(cd.URL, "{port}", strconv.Itoa(cd.PortBinding.GetHostPort()), -1)
	cd.Group = strings.Replace(cd.Group, "{port}", strconv.Itoa(cd.PortBinding.GetHostPort()), -1)
	return cd
}
