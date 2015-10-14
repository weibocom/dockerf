package drivers

import (
	"errors"
	"fmt"
	// log "github.com/Sirupsen/logrus"
	dcluster "github.com/weibocom/dockerf/cluster"
	"github.com/weibocom/dockerf/discovery"
	"path/filepath"
	"strconv"
	"strings"

	//"github.com/aws/aws-sdk-go/aws"
	//"github.com/aws/aws-sdk-go/service/elb"
	//"github.com/docker/machine/libmachine/host"

	"github.com/docker/machine/commands/mcndirs"
	"github.com/docker/machine/drivers/amazonec2"
	"github.com/docker/machine/libmachine"
	"io/ioutil"
	"os"
)

const (
	ELB_DRIVER_NAME = "elb"
	//DEFAULT_DATACENTER = "dc1"
)

type ElbDriver struct {
	region   string
	elbName  string /* serviceName */
	protocal string
	port     int64
	//catalog     *consul.Catalog
}

func newElbDriver(cluster *dcluster.Cluster) (*discovery.ServiceRegisterDriver, error) {
	driverConfig, _ := discovery.GetServiceRegistryDescription(ELB_DRIVER_NAME, cluster)

	/* read config kv from config file */
	driverName, ok := driverConfig["driver"]
	if !ok || driverName != ELB_DRIVER_NAME {
		return nil, errors.New(fmt.Sprintf("Driver name '%s' is expected, but get '%s'", ELB_DRIVER_NAME, driverName))
	}

	elbName, ok := driverConfig["service"]
	if !ok {
		return nil, errors.New("Elb driver option missed: 'upstream', such as 'usertag'")
	}

	region, ok := driverConfig["region"]
	if !ok {
		return nil, errors.New("Elb driver option missed: 'region', such as 'cn-north-1'")
	}

	protocal, ok := driverConfig["protocal"]
	if !ok {
		return nil, errors.New("Elb driver option missed: 'protocal', such as 'http', 'tcp', etc.")
	}

	portStr, ok := driverConfig["port"]
	if !ok {
		return nil, errors.New("Elb driver option missed: 'port', such as 80, 8080, etc.")
	}
	port, err := strconv.ParseUint(portStr, 10, 64)
	if err != nil {
		return nil, errors.New("Elb driver option 'port' should be an unsigned integer")
	}

	/* create loadBalancer if not exist */
	if !ELB_Exist(region, elbName) {
		_, err := ELB_Create(region, elbName, protocal, int64(port))
		if err != nil {
			return nil, err
		}
	}

	d := &ElbDriver{
		region:   region,
		elbName:  elbName,
		protocal: protocal,
		port:     int64(port),
	}
	var pi discovery.ServiceRegisterDriver = d
	return &pi, nil
}

func init() {
	discovery.RegDriverCreateFunction(ELB_DRIVER_NAME, newElbDriver)
}

func (this *ElbDriver) Registry(urls []string) {
}

func (this *ElbDriver) Register(host string, port int) error {
	instanceId, privateIP, err := this.getInstanceIdByOuterIp(host)
	if err != nil {
		return err
	}
	fmt.Println("InstanceId:", instanceId, " privateIP:", privateIP)
	ELB_Register(this.region, this.elbName, instanceId)
	return nil
}

func (this *ElbDriver) UnRegister(host string, port int) error {
	instanceId, privateIP, err := this.getInstanceIdByOuterIp(host)
	if err != nil {
		return err
	}
	fmt.Println("InstanceId:", instanceId, " privateIP:", privateIP)
	ELB_Deregister(this.region, this.elbName, instanceId)
	return nil
}

/* Refer: Machine Filestore.List() */
func (this *ElbDriver) getInstanceIdByOuterIp(ip string) (instanceId, privateIP string, err error) {
	/* c *cli.Context
	s = &persist.Filestore{
		Path:             c.GlobalString("storage-path"),
		CaCertPath:       certInfo.CaCertPath,
		CaPrivateKeyPath: certInfo.CaPrivateKeyPath,
	}
	s := &persist.Filestore{
		Path: "/home/liubin8/.docker/machine",
		CaCertPath:       "/home/liubin8/.docker/machine/certs/ca.pem",
		CaPrivateKeyPath: "/home/liubin8/.docker/machine/certs/ca-key.pem",
	}
	*/

	// CARE: different from lastest machine code */
	//s := libmachine.NewFilestore(mcndirs.GetBaseDir(), "", "")

	//dir, err := ioutil.ReadDir(s.getMachinesDir())
	dir, err := ioutil.ReadDir(mcndirs.GetMachineDir())
	if err != nil && !os.IsNotExist(err) {
		return "", "", err
	}

	//*
	for _, file := range dir {
		if file.IsDir() && !strings.HasPrefix(file.Name(), ".") {
			// CARE: different from lastest machine code */
			hostPath := filepath.Join(mcndirs.GetMachineDir(), file.Name())
			if _, err := os.Stat(hostPath); os.IsNotExist(err) {
				return "", "", err
			}

			host := &libmachine.Host{
				Name:      file.Name(),
				StorePath: hostPath,
			}
			if err := host.LoadConfig(); err != nil {
				return "", "", err
			}

			if err != nil {
				fmt.Printf("error loading host %q: %s", file.Name(), err)
				continue
			}
			//fmt.Printf("host: %#v\n", host.Driver)
			if host.Driver.DriverName() == "amazonec2" {
				td := (host.Driver).(*amazonec2.Driver)
				if td.IPAddress == ip {
					return td.InstanceId, td.PrivateIPAddress, nil
				}
			}

		}
	}
	//*/
	return "", "", fmt.Errorf("Amazonec2 IP %s not exists", ip)
}
