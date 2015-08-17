package drivers

// import (
// 	"dockerf/dlog"
// 	"encoding/json"
// 	"github.com/coreos/go-etcd/etcd"
// 	"path"
// )

// type EtcdDiscovery struct {
// 	url    *DiscoveryUrl
// 	client *etcd.Client
// 	ttl    int64
// }

// const (
// 	name        = "etcd"
// 	DefaultTTL  = 0
// 	DockerfPath = "/dockerf"
// )

// func init() {
// 	registerDiscovery(name, &EtcdDiscovery{})
// }

// func (e *EtcdDiscovery) GetDiscoveryServiceName() string {
// 	return name
// }

// func (e *EtcdDiscovery) Initialize(url *DiscoveryUrl) error {

// 	e.url = url

// 	var entries []string
// 	for _, ip := range e.url.ips {
// 		entries = append(entries, "http://"+ip)
// 	}

// 	e.client = etcd.NewClient(entries)
// 	if _, err := e.client.CreateDir(DockerfPath, DefaultTTL); err != nil {
// 		if etcdError, ok := err.(*etcd.EtcdError); ok {
// 			if etcdError.ErrorCode != 105 { // skip key already exists
// 				return err
// 			}
// 		} else {
// 			return err
// 		}
// 	}
// 	return nil
// }

// func (e *EtcdDiscovery) Register(group string, key string, value string) error {
// 	_, err := e.client.Set(e.buildPath(e.buildDir(group), key), value, DefaultTTL)
// 	return err
// }

// func (e *EtcdDiscovery) Lookup(group string) (DiscoveryServiceNodes, error) {
// 	response, err := e.client.Get(e.buildDir(group), true, true)
// 	if err != nil {
// 		return nil, err
// 	}

// 	keys := []string{}
// 	values := []map[string]string{}

// 	for _, n := range response.Node.Nodes {
// 		_, k := path.Split(n.Key)
// 		keys = append(keys, k)
// 		v := make(map[string]string)
// 		err = json.Unmarshal([]byte(n.Value), v)
// 		if err != nil {
// 			utils.Errorf("Fail to parse value from etcd, path: %s, value: %s", n.Key, n.Value)
// 			continue
// 		}
// 		values = append(values, v)
// 	}

// 	return toDiscoveryServiceNodes(keys, values), nil
// }

// func (e *EtcdDiscovery) Watch(callback DiscoveryCallback) {

// }

// func (e *EtcdDiscovery) buildDir(group string) string {
// 	return path.Join(DockerfPath, group)
// }

// func (e *EtcdDiscovery) buildPath(dir string, key string) string {
// 	return path.Join(dir, key)
// }
