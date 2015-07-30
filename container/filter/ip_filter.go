package filter

import (
	"fmt"
	dcontainer "github.com/weibocom/dockerf/container"
	"strings"
)

type IpFilter struct {
}

const (
	ip_separator = ","
)

func init() {
	registerFilter(&IpFilter{})
}

func (f *IpFilter) filter(filteredIp string, containers []dcontainer.ContainerInfo) ([]dcontainer.ContainerInfo, error) {
	result := []dcontainer.ContainerInfo{}

	fmt.Println(fmt.Sprintf("Apply a ip filter, filter ips: %s", filteredIp))

	ips := strings.Split(filteredIp, ip_separator)

	for _, container := range containers {
		if len(container.IpPorts) > 0 {
			for _, ip := range ips {
				if container.IpPorts[0].IP == strings.TrimSpace(ip) {
					result = append(result, container)
				}
			}
		}
	}

	return result, nil
}

func (f *IpFilter) initialize(dockerProxy *dcontainer.DockerProxy) {
}

func (f *IpFilter) name() string {
	return "ip"
}
