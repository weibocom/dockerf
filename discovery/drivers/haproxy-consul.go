package drivers

import (
	"errors"
	"fmt"
	log "github.com/Sirupsen/logrus"
	consul "github.com/hashicorp/consul/api"
	dcluster "github.com/weibocom/dockerf/cluster"
	"github.com/weibocom/dockerf/discovery"
	"strconv"
	"strings"
)

const (
	HAPROXY_CONSUL_DRIVER_NAME = "haproxy-consul"
	SERVICE_NAME_KEY           = "HAPROXY_SERVICE_NAME"
	DEFAULT_DATACENTER         = "dc1"
)

type HaproxyConsulRegisterDriver struct {
	serviceName string
	catalog     *consul.Catalog
}

func newHaproxyConsulRegisterDriver(cluster *dcluster.Cluster) (*discovery.ServiceRegisterDriver, error) {
	driverConfig, _ := discovery.GetServiceRegistryDescription(HAPROXY_CONSUL_DRIVER_NAME, cluster)
	driverName, ok := driverConfig["driver"]
	if !ok || driverName != HAPROXY_CONSUL_DRIVER_NAME {
		return nil, errors.New(fmt.Sprintf("Driver name '%s' is expected, but get '%s'", HAPROXY_CONSUL_DRIVER_NAME, driverName))
	}
	serviceName, err := resolveServiceName(cluster, driverConfig)
	if err != nil {
		return nil, err
	}

	config := consul.DefaultConfig()
	consulAddress := fmt.Sprintf("%s:8500", cluster.ConsulCluster.Server.IPs[0])
	config.Address = consulAddress
	client, err := consul.NewClient(config)
	if err != nil {
		return nil, err
	}

	d := HaproxyConsulRegisterDriver{serviceName: serviceName, catalog: client.Catalog()}
	var pi discovery.ServiceRegisterDriver = &d
	return &pi, nil
}

func resolveServiceName(cluster *dcluster.Cluster, sdd dcluster.ServiceDiscoverDiscription) (string, error) {
	containerGroup, ok := sdd["container"]
	if !ok {
		return "", errors.New("Nginx driver option missed: 'container'")
	}
	description, exists := cluster.Container.Topology.GetDescription(containerGroup)
	if !exists {
		return "", errors.New(fmt.Sprintf("Fail to find container description for container group: %s...", containerGroup))
	}
	for _, env := range description.Env {
		kv := strings.Split(env, "=")
		if len(kv) < 2 {
			log.Debugf("Invalid container env %s", env)
			continue
		}
		if kv[0] == SERVICE_NAME_KEY {
			log.Debugf("Find service name for haproxy %s", env)
			return kv[1], nil
		}
	}
	return "", errors.New(fmt.Sprintf("Fail to find service name from container env, containerGroup: %s", containerGroup))
}

func init() {
	discovery.RegDriverCreateFunction(HAPROXY_CONSUL_DRIVER_NAME, newHaproxyConsulRegisterDriver)
}

func (haConsul *HaproxyConsulRegisterDriver) Registry(urls []string) {
}

func (haConsul *HaproxyConsulRegisterDriver) Register(host string, port int) error {
	service := &consul.AgentService{}
	service.Service = haConsul.serviceName
	service.Address = host
	service.Port = port
	service.ID = haConsul.buildNodeName(haConsul.serviceName, host, port)

	reg := &consul.CatalogRegistration{}
	reg.Node = haConsul.buildNodeName(haConsul.serviceName, host, port)
	reg.Datacenter = DEFAULT_DATACENTER
	reg.Address = host
	reg.Service = service

	_, err := haConsul.catalog.Register(reg, haConsul.buildWriteOptions())
	return err
}

func (haConsul *HaproxyConsulRegisterDriver) UnRegister(host string, port int) error {
	unreg := &consul.CatalogDeregistration{}
	unreg.Node = haConsul.buildNodeName(haConsul.serviceName, host, port)
	unreg.Datacenter = DEFAULT_DATACENTER
	unreg.ServiceID = haConsul.serviceName

	_, err := haConsul.catalog.Deregister(unreg, haConsul.buildWriteOptions())

	return err
}

func (haConsul *HaproxyConsulRegisterDriver) buildNodeName(serviceName string, host string, port int) string {
	return strings.Join([]string{serviceName, host, strconv.Itoa(port)}, "-")
}

func (haConsul *HaproxyConsulRegisterDriver) buildWriteOptions() *consul.WriteOptions {
	wOptions := &consul.WriteOptions{}
	wOptions.Datacenter = DEFAULT_DATACENTER
	return wOptions
}
