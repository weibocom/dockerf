package topology

import (
	"strconv"

	"github.com/Sirupsen/logrus"
	"github.com/weibocom/dockerf/options"
)

func getNum(key string, errVal int, opts *options.Options) int {
	str := opts.String(key)
	if str == "" {
		logrus.Warnf("'%s' options is not found, %d is the default", key, errVal)
		return errVal
	}
	val, err := strconv.Atoi(str)
	if err != nil {
		logrus.Warnf("'%s' is not a valid number, %d is the default. err:%s", str, errVal, err.Error())
		return errVal
	}
	return int(val)
}

func getMaxNum(opts *options.Options) int {
	return getNum("max-num", -1, opts)
}

func getMinNum(opts *options.Options) int {
	return getNum("min-num", -1, opts)
}
