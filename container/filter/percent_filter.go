package filter

import (
	"fmt"
	dcontainer "github.com/weibocom/dockerf/container"
	"strconv"
)

type PercentFilter struct {
}

func init() {
	registerFilter(&PercentFilter{})
}

func (p *PercentFilter) filter(percentStr string, containers []dcontainer.ContainerInfo) ([]dcontainer.ContainerInfo, error) {
	total := len(containers)
	percent, err := strconv.ParseFloat(percentStr, 64)
	if err != nil {
		panic(fmt.Sprintf("Invalid percent num specified, percent: %s", percentStr))
	}
	num := int(float64(total) * percent / float64(100))

	fmt.Println(fmt.Sprintf("Apply a percent filter, percent: %s, num: %d", percentStr, num))
	return containers[0:num], nil
}

func (p *PercentFilter) initialize(dockerProxy *dcontainer.DockerProxy) {
}

func (p *PercentFilter) name() string {
	return "percent"
}
