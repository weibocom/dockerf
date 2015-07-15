package container

import (
	"testing"
)

func TestNameParse(t *testing.T) {
	cn := ContainerName{}
	name := "web-1"
	if !cn.Parse(name) {
		t.Errorf("'%s' is a valid name, but parse error.\n", name)
	}
}
