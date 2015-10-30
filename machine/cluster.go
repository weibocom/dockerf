package machine

import (
	"os"
	"sync"

	"github.com/docker/machine/libmachine"
	"github.com/docker/machine/ssh"
	"github.com/weibocom/dockerf/options"
)

type Cluster struct {
	sync.Mutex
	machines      map[string]*Machine
	provider      *libmachine.Provider
	authOptions   *AuthOptions
	globleOptions *options.Options
}

func NewCluster(gopt *options.Options) (*Cluster, error) {
	rootPath := gopt.String("storage-path")
	os.Setenv("MACHINE_STORAGE_PATH", rootPath)
	if gopt.Bool("native-ssh") {
		ssh.SetDefaultClient(ssh.Native)
	}

	auth := getTLSAuthOptions(gopt)
	store := libmachine.NewFilestore(rootPath, auth.CaCertPath, auth.CaKeyPath)
	provider, _ := libmachine.New(store)
	hosts, err := provider.List()
	if err != nil {
		return nil, err
	}
	machines := make(map[string]*Machine, len(hosts))
	for _, h := range hosts {
		machines[h.Name] = &Machine{
			Host: h,
		}
	}

	c := &Cluster{
		provider:    provider,
		machines:    machines,
		authOptions: auth,
	}
	c.LoadStates()
	return c, nil
}
