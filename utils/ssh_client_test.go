package utils

import (
	"fmt"
	"testing"
)

func TestRunSSHCommand(t *testing.T) {

	content := "123"

	user := "root"
	key := "/Users/tangyang/.docker/machine/machines/usertag-master-1/id_rsa"
	port := 22
	host := "123.56.124.94"
	clientConfig := &SSHClientConfig{User: user, Host: host, Port: port, Key: key}

	client, err := NewSSHClient(clientConfig)
	if err != nil {
		t.Errorf("Fail to build a ssh client to host %s, error: %s", host, err.Error())
		return
	}

	echoCommand := fmt.Sprintf("echo \"%s\" > /root/aa", content)
	output, err := client.RunSSHCommand(echoCommand)
	t.Logf("Output: %s", output)
	if err != nil {
		t.Errorf("Fail to execute command %s, error: %s", echoCommand, err.Error())
	}

	catCommand := "cat /root/aa"
	output, err = client.RunSSHCommand(catCommand)
	t.Logf("Output: %s", output)
	if err != nil {
		t.Errorf("Fail to execute command %s, error: %s", catCommand, err.Error())
	}
}
