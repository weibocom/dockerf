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
	machines    []*Machine
	provider    *libmachine.Provider
	authOptions *AuthOptions
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
	machines := make([]*Machine, len(hosts))
	for idx, h := range hosts {
		machines[idx] = &Machine{
			Host: h,
		}
	}

	return &Cluster{
		provider:    provider,
		machines:    machines,
		authOptions: auth,
	}, nil
}
