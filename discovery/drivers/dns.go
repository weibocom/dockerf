package drivers

// import (
// 	"fmt"
// 	consul "github.com/hashicorp/consul/api"
// 	"github.com/weibocom/dockerf/cluster"
// 	"github.com/weibocom/dockerf/discovery"
// 	dutils "github.com/weibocom/dockerf/utils"
// 	"strconv"
// 	"strings"
// )

// const (
// 	DEFAULT_DATACENTER = "dc1"
// )

// func init() {
// 	discovery.AddDriver(&DnsDiscoveryService{})
// }

// type DnsDiscoveryService struct {
// 	serviceName string
// 	catalog     *consul.Catalog
// }

// func (d *DnsDiscoveryService) Initialize(urls []string, driverConfig cluster.ServiceDiscoverDiscription) (discovery.ServiceRegisterDriver, error) {
// 	dnsDiscoveryService := &DnsDiscoveryService{}
// 	config := consul.DefaultConfig()
// 	config.Address = driverConfig["dnsserver"]
// 	client, err := consul.NewClient(config)
// 	if err != nil {
// 		dutils.Error(fmt.Sprintf("Fail to build a consul client, error: %s", err.Error()))
// 		return nil, err
// 	}

// 	dnsDiscoveryService.catalog = client.Catalog()
// 	dnsDiscoveryService.serviceName = strings.Split(driverConfig["url"], ".")[0]
// 	return dnsDiscoveryService, nil
// }

// func (d *DnsDiscoveryService) Register(host string, port uint) error {
// 	service := &consul.AgentService{}
// 	service.Service = d.serviceName

// 	reg := &consul.CatalogRegistration{}
// 	reg.Node = d.buildNodeName(d.serviceName, host, port)
// 	reg.Datacenter = DEFAULT_DATACENTER
// 	reg.Address = host
// 	reg.Service = service

// 	_, err := d.catalog.Register(reg, d.buildWriteOptions())
// 	if err != nil {
// 		dutils.Error(fmt.Sprintf("Fail to register %s:%d to service %s, err: %s", host, port, d.serviceName, err.Error()))
// 		return err
// 	}

// 	return nil
// }

// func (d *DnsDiscoveryService) UnRegister(host string, port uint) error {
// 	unreg := &consul.CatalogDeregistration{}
// 	unreg.Node = d.buildNodeName(d.serviceName, host, port)
// 	unreg.Datacenter = DEFAULT_DATACENTER
// 	unreg.ServiceID = d.serviceName

// 	_, err := d.catalog.Deregister(unreg, d.buildWriteOptions())
// 	if err != nil {
// 		dutils.Error(fmt.Sprintf("Fail to deregister %s:%d from service %s, err: %s", host, port, d.serviceName, err.Error()))
// 		return err
// 	}

// 	return nil
// }

// func (d *DnsDiscoveryService) Lookup() ([]discovery.Address, error) {
// 	qOptions := &consul.QueryOptions{}
// 	qOptions.Datacenter = DEFAULT_DATACENTER

// 	serviceNodes, _, err := d.catalog.Service(d.serviceName, "", qOptions)
// 	if err != nil {
// 		dutils.Error(fmt.Sprintf("Fail to lookup nodes from service %s, err: %s", d.serviceName, err.Error()))
// 		return nil, err
// 	}

// 	addresses := []discovery.Address{}
// 	for _, serviceNode := range serviceNodes {
// 		addr := discovery.Address{}
// 		addr.Host = serviceNode.Address
// 		addresses = append(addresses, addr)
// 	}

// 	return addresses, nil
// }

// func (d *DnsDiscoveryService) Name() string {
// 	return "dns"
// }

// func (d *DnsDiscoveryService) buildNodeName(serviceName string, host string, port uint) string {
// 	return strings.Join([]string{serviceName, host, strconv.Itoa(int(port))}, "-")
// }

// func (d *DnsDiscoveryService) buildWriteOptions() *consul.WriteOptions {
// 	wOptions := &consul.WriteOptions{}
// 	wOptions.Datacenter = DEFAULT_DATACENTER
// 	return wOptions
// }
