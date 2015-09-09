package drivers

import (
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/Sirupsen/logrus"
	consul "github.com/hashicorp/consul/api"
	dcluster "github.com/weibocom/dockerf/cluster"
	"github.com/weibocom/dockerf/discovery"
	dutils "github.com/weibocom/dockerf/utils"
)

type ConsulOperation string

const (
	NGINX_CONSUL_DRIVER_NAME   = "nginx-consul"
	NGINX_CONSUL_PREFIX        = "upstream"
	NGINX_CONSUL_URL_SEPARATOR = "/"
	CONSUL_REGISTER            = ConsulOperation("Register")
	CONSUL_UNREGISTER          = ConsulOperation("UnRegister")
)

type NginxConsulRegisterDriver struct {
	upstream      string
	client        *consul.Client
	consulAddress string
}

func newNginxConsulDriver(cluster *dcluster.Cluster) (*discovery.ServiceRegisterDriver, error) {
	driverConfig, _ := discovery.GetServiceRegistryDescription(NGINX_CONSUL_DRIVER_NAME, cluster)
	driverName, ok := driverConfig["driver"]
	if !ok || driverName != NGINX_CONSUL_DRIVER_NAME {
		return nil, errors.New(fmt.Sprintf("Driver name '%s' is expected, but get '%s'", NGINX_CONSUL_DRIVER_NAME, driverName))
	}
	us, ok := driverConfig["upstream"]
	if !ok {
		return nil, errors.New("Nginx driver option missed: 'upstream'")
	}

	transport := &dutils.Transport{
		ConnectTimeout:        3 * time.Second,
		RequestTimeout:        10 * time.Second,
		ResponseHeaderTimeout: 5 * time.Second,
	}
	httpClient := &http.Client{Transport: transport}

	consulAddress := cluster.ConsulCluster.Server.IPs[0]
	config := &consul.Config{
		Address:    fmt.Sprintf("%s:8500", consulAddress),
		Scheme:     "http",
		HttpClient: httpClient,
	}

	client, err := consul.NewClient(config)
	if err != nil {
		return nil, errors.New("Fail to build a consul client! ")
	}

	d := NginxConsulRegisterDriver{
		upstream:      us,
		client:        client,
		consulAddress: consulAddress,
	}
	var pi discovery.ServiceRegisterDriver = &d
	return &pi, nil
}

func init() {
	discovery.RegDriverCreateFunction(NGINX_CONSUL_DRIVER_NAME, newNginxConsulDriver)
}

func (ngxDrv *NginxConsulRegisterDriver) Registry(urls []string) {
}

func (ngxDrv *NginxConsulRegisterDriver) buildConsulUrl(upstream string, address string) string {
	return strings.Join([]string{NGINX_CONSUL_PREFIX, upstream, address}, NGINX_CONSUL_URL_SEPARATOR)
}

func (ngxDrv *NginxConsulRegisterDriver) operate(host string, port int, op ConsulOperation, callable func(kv *consul.KV, address string) error) error {
	logrus.Debug(fmt.Sprintf("%s service to nginx consul, consul address:%s, host:%s, port:%d\n", op, ngxDrv.consulAddress, host, port))
	address := fmt.Sprintf("%s:%d", host, port)
	kv := ngxDrv.client.KV()
	err := callable(kv, address)
	if err != nil {
		fmt.Println(fmt.Sprintf("Fail to %s to consul cluster, address: %s, error: %s", op, address, err.Error()))
		return err
	}

	return nil
}

func (ngxDrv *NginxConsulRegisterDriver) Register(host string, port int) error {

	registerOp := func(kv *consul.KV, address string) error {
		consulUrl := ngxDrv.buildConsulUrl(ngxDrv.upstream, address)
		p := &consul.KVPair{Key: consulUrl, Value: []byte(address)}
		_, err := kv.Put(p, nil)
		return err
	}

	return ngxDrv.operate(host, port, CONSUL_REGISTER, registerOp)
}

func (ngxDrv *NginxConsulRegisterDriver) UnRegister(host string, port int) error {

	unregisterOp := func(kv *consul.KV, address string) error {
		consulUrl := ngxDrv.buildConsulUrl(ngxDrv.upstream, address)
		_, err := kv.Delete(consulUrl, nil)
		return err
	}

	return ngxDrv.operate(host, port, CONSUL_REGISTER, unregisterOp)
}
