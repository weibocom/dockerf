package machine

import (
	"github.com/docker/machine/libmachine/engine"
	"github.com/weibocom/dockerf/options"
)

type EngineOptions engine.EngineOptions

func getEngineOptions(opt *options.Options) *EngineOptions {
	return &EngineOptions{
		ArbitraryFlags:   opt.StringSlice("engine-opt"),
		Env:              opt.StringSlice("engine-env"),
		InsecureRegistry: opt.StringSlice("engine-insecure-registry"),
		Labels:           opt.StringSlice("engine-label"),
		RegistryMirror:   opt.StringSlice("engine-registry-mirror"),
		StorageDriver:    opt.String("engine-storage-driver"),
		TlsVerify:        true,
		InstallURL:       opt.String("engine-install-url"),
	}
}
