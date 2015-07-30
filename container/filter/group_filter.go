package filter

import (
	"encoding/json"
	"fmt"
	dcontainer "github.com/weibocom/dockerf/container"
)

const (
	GroupFilterName = "group"
)

type GroupFilter struct {
	dockerProxy *dcontainer.DockerProxy
}

func init() {
	registerFilter(&GroupFilter{})
}

func (g *GroupFilter) filter(groupName string, containers []dcontainer.ContainerInfo) ([]dcontainer.ContainerInfo, error) {
	m := map[string][]string{"name": []string{g.buildGroupFilterKey(groupName)}}
	filter, err := json.Marshal(m)
	fmt.Println(fmt.Sprintf("Apply a group filter, groupName: %s, filter: %s", groupName, string(filter)))
	if err != nil {
		fmt.Println(fmt.Sprintf("Fail to build a group filter key, groupName: %s, error: %s", groupName, err.Error()))
		return g.dockerProxy.ListContainers(true, true, "")
	}

	return g.dockerProxy.ListContainers(true, true, string(filter))
}

func (g *GroupFilter) buildGroupFilterKey(originalKey string) string {
	return "^" + originalKey + "-" + "[0-9]*$"
}

func (g *GroupFilter) initialize(dockerProxy *dcontainer.DockerProxy) {
	g.dockerProxy = dockerProxy
}

func (g *GroupFilter) name() string {
	return GroupFilterName
}
