package options

import (
	"os"
	"reflect"
	"strconv"
	"strings"

	"github.com/codegangsta/cli"
	"github.com/weibocom/dockerf/utils"
)

type Options struct {
	Values map[string]string
	Flags  []cli.Flag
}

// first: from Values
// second: from environment
// third: default value
func (m *Options) getValue(key string) interface{} {
	if m.Values != nil {
		if v, ok := m.Values[key]; ok {
			return v
		}
	}

	for _, f := range m.Flags {
		rv := reflect.ValueOf(f)
		vks := strings.Split(rv.FieldByName("Name").String(), ",")
		for _, vk := range vks {
			tk := strings.TrimSpace(vk)
			if tk == key {
				env := rv.FieldByName("EnvVar").String()
				v := os.Getenv(env)
				if v != "" {
					return v
				}
				return rv.FieldByName("Value")
			}
		}
	}
	return ""
}

func (m *Options) String(key string) string {
	v := m.getValue(key)
	if reflect.TypeOf(v).Name() == "string" {
		return v.(string)
	}
	rv := v.(reflect.Value)
	return rv.String()
}

func (m *Options) StringSlice(key string) []string {
	str := m.String(key)
	return utils.TrimSplit(str, " ")
}

func (m *Options) Int(key string) int {
	v := m.getValue(key)
	if reflect.TypeOf(v).Name() == "string" {
		str := v.(string)
		val, err := strconv.Atoi(str)
		if err != nil {
			return 0
		}
		return val
	}
	rv := v.(reflect.Value)
	return int(rv.Int())
}

func (m *Options) Bool(key string) bool {
	v := m.getValue(key)
	if reflect.TypeOf(v).Name() == "string" {
		val, err := strconv.ParseBool(v.(string))
		if err != nil {
			return false
		}
		return val
	}
	return false
}

func (o *Options) Apply(k, v string) {
	o.Values[k] = v
}
