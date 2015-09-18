package utils

import (
	"fmt"
	log "github.com/Sirupsen/logrus"
	"golang.org/x/crypto/ssh"
	"io/ioutil"
	"strings"
)

type SSHClientConfig struct {
	User string
	Host string
	Port int
	Key  string
}

type NativeClient struct {
	Config   ssh.ClientConfig
	Hostname string
	Port     int
}

const (
	ErrExitCode255 = "255"
)

func newNativeConfig(config *SSHClientConfig) (ssh.ClientConfig, error) {
	var (
		authMethods []ssh.AuthMethod
	)

	key, err := ioutil.ReadFile(config.Key)
	if err != nil {
		return ssh.ClientConfig{}, err
	}

	privateKey, err := ssh.ParsePrivateKey(key)
	if err != nil {
		return ssh.ClientConfig{}, err
	}
	authMethods = append(authMethods, ssh.PublicKeys(privateKey))

	return ssh.ClientConfig{
		User: config.User,
		Auth: authMethods,
	}, nil
}

func NewSSHClient(config *SSHClientConfig) (NativeClient, error) {
	nativeConfig, err := newNativeConfig(config)
	return NativeClient{Config: nativeConfig, Hostname: config.Host, Port: config.Port}, err
}

func (client NativeClient) RunSSHCommand(command string) (string, error) {
	log.Debugf("About to run SSH command:\n%s", command)

	output, err := client.Output(command)
	log.Debugf("SSH cmd err, output: %v: %s", err, output)

	if err != nil && !isErr255Exit(err) {
		log.Error("SSH cmd error!")
		log.Errorf("command: %s", command)
		log.Errorf("err    : %v", err)
		log.Fatalf("output : %s", output)
	}

	return output, err
}

func isErr255Exit(err error) bool {
	return strings.Contains(err.Error(), ErrExitCode255)
}

func (client NativeClient) session(command string) (*ssh.Session, error) {
	conn, err := ssh.Dial("tcp", fmt.Sprintf("%s:%d", client.Hostname, client.Port), &client.Config)
	if err != nil {
		return nil, fmt.Errorf("Error dialing TCP: %s", err)
	}

	return conn.NewSession()
}

func (client NativeClient) Output(command string) (string, error) {
	session, err := client.session(command)
	if err != nil {
		return "", nil
	}

	output, err := session.CombinedOutput(command)
	defer session.Close()

	return string(output), err
}
