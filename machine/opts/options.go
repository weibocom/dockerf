package opts

import (
	"errors"
	"fmt"
	"reflect"
	"strings"

	dcluster "github.com/weibocom/dockerf/cluster"
)

type OptsDriver struct {
}

func GetOptions(driver string, md dcluster.MachineDescription) ([]string, error) {
	od := &OptsDriver{}
	method, exists := od.getDriverFunction(driver)
	if !exists {
		return []string{}, errors.New(fmt.Sprintf("Options driver '%s' is not supported.", driver))
	}
	return method(md)
}

func (od *OptsDriver) getDriverFunction(driverName string) (func(dcluster.MachineDescription) ([]string, error), bool) {
	methodName := "Get" + strings.ToUpper(driverName[:1]) + strings.ToLower(driverName[1:]) + "Options"
	method := reflect.ValueOf(od).MethodByName(methodName)
	if !method.IsValid() {
		return nil, false
	}
	return method.Interface().(func(dcluster.MachineDescription) ([]string, error)), true
}
