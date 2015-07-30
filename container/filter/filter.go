package filter

import (
	// "encoding/json"
	"fmt"
	dcontainer "github.com/weibocom/dockerf/container"
)

type Filter interface {
	filter(pattern string, containers []dcontainer.ContainerInfo) ([]dcontainer.ContainerInfo, error)
	name() string
	initialize(dockerProxy *dcontainer.DockerProxy)
}

type FilterChain struct {
	filters map[string]Filter
}

func NewFilterChain(dockerProxy *dcontainer.DockerProxy) *FilterChain {
	filters := map[string]Filter{}
	for k, v := range filterMap {
		v.initialize(dockerProxy)
		filters[k] = v
	}
	return &FilterChain{filters: filters}
}

func (f *FilterChain) Filter(filterKVs map[string]string) ([]dcontainer.ContainerInfo, error) {
	groupFilterKey, exist := filterKVs[GroupFilterName]
	if !exist {
		panic("You must specify a group filter... ")
	}
	containers, err := f.filters[GroupFilterName].filter(groupFilterKey, nil)
	if err != nil {
		fmt.Println(fmt.Sprintf("Fail to execute group filter, error: %s, group: %s", err.Error(), groupFilterKey))
		return nil, err
	}

	for key, value := range filterKVs {
		if key == GroupFilterName {
			continue
		}
		filter, _ := f.filters[key]
		containers, err = filter.filter(value, containers)
		if err != nil {
			fmt.Println(fmt.Sprintf("Fail to execute filter, name: %s, filterKey: %s, error: %s", key, value, err.Error()))
			return nil, err
		}
	}
	return containers, nil
}

var filterMap map[string]Filter

func init() {
	filterMap = make(map[string]Filter)
}

func registerFilter(filter Filter) {
	if _, exist := filterMap[filter.name()]; exist {
		fmt.Println(fmt.Sprintf("Filter is already registered, name: %s", filter.name()))
		return
	}
	filterMap[filter.name()] = filter
}
