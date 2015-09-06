package drivers

import (
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
	"sync"
	"time"

	dcluster "github.com/weibocom/dockerf/cluster"
	"github.com/weibocom/dockerf/discovery"
	dutils "github.com/weibocom/dockerf/utils"
)

const (
	NGINX_DRIVER_NAME = "nginx-push"
)

type NginxServiceRegisterDriver struct {
	upstream   string
	urls       []string
	httpClient *http.Client
}

func newNginxDriver(cluster *dcluster.Cluster) (*discovery.ServiceRegisterDriver, error) {
	driverConfig, _ := discovery.GetServiceRegistryDescription(NGINX_DRIVER_NAME, cluster)
	us, ok := driverConfig["upstream"]
	if !ok {
		return nil, errors.New("Nginx driver option missed: 'upstream'")
	}
	driverName, ok := driverConfig["driver"]
	if !ok || driverName != NGINX_DRIVER_NAME {
		return nil, errors.New(fmt.Sprintf("Driver name '%s' is expected, but get '%s'", NGINX_DRIVER_NAME, driverName))
	}

	transport := &dutils.Transport{
		ConnectTimeout:        3 * time.Second,
		RequestTimeout:        10 * time.Second,
		ResponseHeaderTimeout: 5 * time.Second,
	}
	d := NginxServiceRegisterDriver{
		upstream:   us,
		httpClient: &http.Client{Transport: transport},
	}
	var pi discovery.ServiceRegisterDriver = &d
	return &pi, nil
}

func init() {
	discovery.RegDriverCreateFunction(NGINX_DRIVER_NAME, newNginxDriver)
}

func (ngxDrv *NginxServiceRegisterDriver) Registry(urls []string) {
	fmt.Printf("Nginx service register driver registry:%+v\n", urls)
	ngxDrv.urls = urls
}

func (ngxDrv *NginxServiceRegisterDriver) buildError(prefix string, errs map[string]error) error {
	errMessages := []string{}
	for url, err := range errs {
		errMessages = append(errMessages, fmt.Sprintf("%s: %s", url, err.Error()))
	}

	return errors.New(strings.Join([]string{prefix, strings.Join(errMessages, ", ")}, ", "))
}

func (ngxDrv *NginxServiceRegisterDriver) Register(host string, port int) error {
	fmt.Printf("Register service to nginx. nginx url:%+v, host:%s, port:%d\n", ngxDrv.urls, host, port)
	address := fmt.Sprintf("%s:%d", host, port)
	body := fmt.Sprintf("{\"upstream\":\"%s\",\"server\":[\"%s\"],\"method\":\"add\"}", ngxDrv.upstream, address)
	errChans := make(map[string]chan error)
	errs := make(map[string]error)
	wg := &sync.WaitGroup{}
	for _, url := range ngxDrv.urls {
		errChan := make(chan error, 1)
		errChans[url] = errChan
		wg.Add(1)
		go ngxDrv.register0(body, url, errChan, wg)
	}
	wg.Wait()

	for url, c := range errChans {
		err := <-c
		if err != nil {
			errs[url] = err
		}
	}

	if len(errs) > 0 {
		return ngxDrv.buildError("Nginx register error", errs)
	}
	fmt.Printf("Successfully Register service to nginx. nginx url:%+v, host:%s, port:%d\n", ngxDrv.urls, host, port)
	return nil
}

func (ngxDrv *NginxServiceRegisterDriver) register0(body string, url string, errChan chan error, wg *sync.WaitGroup) {
	req, _ := http.NewRequest("POST", fmt.Sprintf("http://%s/upstream_add_server", url), strings.NewReader(body))
	resp, err := ngxDrv.httpClient.Do(req)
	defer wg.Done()
	if err != nil {
		fmt.Println(fmt.Sprintf("Register error, error: %s, body: %s, url: %s", err.Error(), body, url))
		errChan <- err
		return
	}
	b, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		resp.Body.Close()
		fmt.Println(fmt.Sprintf("Read response from nginx error, error: %s, url: %s", err.Error(), url))
		errChan <- err
		return
	}
	fmt.Println(fmt.Sprintf("Register node to nginx, result.  %s", string(b)))
	resp.Body.Close()
	errChan <- nil
}

func (ngxDrv *NginxServiceRegisterDriver) UnRegister(host string, port int) error {
	fmt.Printf("Ungreister service out of nginx(%+v). ip:%s, port:%d\n", ngxDrv.urls, host, port)
	address := fmt.Sprintf("%s:%d", host, port)
	body := fmt.Sprintf("{\"upstream\":\"%s\",\"server\":[\"%s\"],\"method\":\"del\"}", ngxDrv.upstream, address)

	errChans := make(map[string]chan error)
	errs := make(map[string]error)
	wg := &sync.WaitGroup{}
	for _, url := range ngxDrv.urls {
		errChan := make(chan error, 1)
		errChans[url] = errChan
		wg.Add(1)
		go ngxDrv.unregister0(body, url, errChan, wg)
	}
	wg.Wait()

	for url, c := range errChans {
		err := <-c
		if err != nil {
			errs[url] = err
		}
	}

	if len(errs) > 0 {
		return ngxDrv.buildError("Nginx unregister error", errs)
	}
	return nil
}

func (ngxDrv *NginxServiceRegisterDriver) unregister0(body string, url string, errChan chan error, wg *sync.WaitGroup) {
	defer wg.Done()
	req, _ := http.NewRequest("POST", fmt.Sprintf("http://%s/upstream_add_server", url), strings.NewReader(body))
	resp, err := ngxDrv.httpClient.Do(req)
	if err != nil {
		fmt.Println(fmt.Sprintf("UnRegister error, error: %s, body: %s, url: %s", err.Error(), body, url))
		errChan <- err
		return
	}
	b, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		resp.Body.Close()
		fmt.Println(fmt.Sprintf("Read response from nginx error, error: %s, url: %s", err.Error(), url))
		errChan <- err
		return
	}
	fmt.Println(fmt.Sprintf("UnRegister node to nginx, result: %s", string(b)))
	resp.Body.Close()
	errChan <- nil
}

func (ngxDrv *NginxServiceRegisterDriver) Lookup() ([]discovery.Address, error) {
	return nil, nil
}

func (ngxDrv *NginxServiceRegisterDriver) Name() string {
	return NGINX_DRIVER_NAME
}
