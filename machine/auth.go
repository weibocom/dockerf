package machine

import (
	"path/filepath"

	"github.com/docker/machine/utils"
	"github.com/weibocom/dockerf/options"
)

type AuthOptions struct {
	CaCertPath     string
	CaKeyPath      string
	ClientCertPath string
	ClientKeyPath  string
}

func getTLSAuthOptions(opt *options.Options) *AuthOptions {
	caCertPath := opt.String("tls-ca-cert")
	caKeyPath := opt.String("tls-ca-key")
	clientCertPath := opt.String("tls-client-cert")
	clientKeyPath := opt.String("tls-client-key")

	if caCertPath == "" {
		caCertPath = filepath.Join(utils.GetMachineCertDir(), "ca.pem")
	}

	if caKeyPath == "" {
		caKeyPath = filepath.Join(utils.GetMachineCertDir(), "ca-key.pem")
	}

	if clientCertPath == "" {
		clientCertPath = filepath.Join(utils.GetMachineCertDir(), "cert.pem")
	}

	if clientKeyPath == "" {
		clientKeyPath = filepath.Join(utils.GetMachineCertDir(), "key.pem")
	}

	return &AuthOptions{
		CaCertPath:     caCertPath,
		CaKeyPath:      caKeyPath,
		ClientCertPath: clientCertPath,
		ClientKeyPath:  clientKeyPath,
	}
}
