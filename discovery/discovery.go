package discovery

import (
	"errors"
	"fmt"
	"github.com/weibocom/dockerf/cluster"
	"sync"
)

type IpPort struct {
	host string
	port int
}

type ServiceRegisterDriver interface {
	// Initialize(urls []string, driverConfig cluster.ServiceDiscoverDiscription) (ServiceRegisterDriver, error)
	Registry(urls []string)
	Register(host string, port int) error
	UnRegister(host string, port int) error
	// Lookup() ([]Address, error)
	// Name() string
}

var (
	driverLock       sync.Mutex
	ServRegDrivers   map[string]ServiceRegisterDriver
	ErrNotSupported  = errors.New("driver not supported")
	dcLock           sync.Mutex
	driverCreateFuns map[string]func(driverConfig cluster.ServiceDiscoverDiscription) (*ServiceRegisterDriver, error)
)

func init() {
	ServRegDrivers = make(map[string]ServiceRegisterDriver)
	driverCreateFuns = map[string]func(driverConfig cluster.ServiceDiscoverDiscription) (*ServiceRegisterDriver, error){}
}

func AddDriver(driver ServiceRegisterDriver) error {
	driverLock.Lock()
	defer driverLock.Unlock()
	// dName := driver.Name()
	// if _, ok := ServRegDrivers[dName]; ok {
	// 	return errors.New(fmt.Sprintf("Driver('%s') is already registered.", dName))
	// }
	// ServRegDrivers[dName] = driver
	return nil
}

func RegDriverCreateFunction(driverName string, creatorFun func(driverConfig cluster.ServiceDiscoverDiscription) (*ServiceRegisterDriver, error)) error {
	dcLock.Lock()
	defer dcLock.Unlock()
	if f, ok := driverCreateFuns[driverName]; ok {
		return errors.New(fmt.Sprintf("Service register driver('%s') function is already registered. func:%+v\n", driverName, f))
	} else {
		driverCreateFuns[driverName] = creatorFun
		return nil
	}
}

func NewRegDriver(driverConfig cluster.ServiceDiscoverDiscription) (*ServiceRegisterDriver, error) {
	driverName, ok := driverConfig["driver"]
	if !ok {
		return nil, errors.New("New register driver failed: 'drover' option missed.\n")
	}
	driverFunc, ok := driverCreateFuns[driverName]
	if !ok {
		return nil, errors.New("Not a validate driver registered for '" + driverName + "'")
	}
	return driverFunc(driverConfig)
}

// func New(driverName string, urls []string, driverConfig cluster.ServiceDiscoverDiscription) (ServiceRegisterDriver, error) {

// 	if driver, exists := ServRegDrivers[driverName]; exists {
// 		d, err := driver.Initialize(urls, driverConfig)
// 		return d, err
// 	}

// 	return nil, ErrNotSupported
// }

type Address struct {
	Host string
	port int
}
