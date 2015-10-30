package options

import (
	"errors"
	"fmt"
	"reflect"
	"strings"

	"github.com/weibocom/dockerf/options"
)

type OptsDriver struct {
}

func RefreshOptions(driver string, opts *options.Options) error {
	od := &OptsDriver{}
	method, exists := od.getDriverFunction(driver)
	if !exists {
		return errors.New(fmt.Sprintf("Options driver '%s' is not supported.", driver))
	}
	return method(opts)
}

func (od *OptsDriver) getDriverFunction(driverName string) (func(opts *options.Options) error, bool) {
	methodName := "Refresh" + strings.ToUpper(driverName[:1]) + strings.ToLower(driverName[1:]) + "Options"
	method := reflect.ValueOf(od).MethodByName(methodName)
	if !method.IsValid() {
		return nil, false
	}
	return method.Interface().(func(opts *options.Options) error), true
}
