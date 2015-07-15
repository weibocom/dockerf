package cluster

import (
	"errors"
	"fmt"
	"strconv"
	"strings"

	dutils "github.com/weibocom/dockerf/utils"
)

type HostPortRange struct {
	min int
	max int
}

func (hp *HostPortRange) hostPort() int {
	if hp.min == hp.max {
		return hp.min
	}
	return dutils.RandomUInt(hp.min, hp.max)
}

type PortBinding struct {
	Protocal      string
	hostPortRange HostPortRange
	HostPort      int
	ContainerPort int
}

func (pb *PortBinding) GetHostPort() int {
	if pb.HostPort > 0 {
		return pb.HostPort
	}
	return pb.hostPortRange.hostPort()
}

func (pb *PortBinding) Parse(portStr string) error {
	pb.parstProtocal(portStr)

	if err := pb.parseContainerPort(portStr); err != nil {
		return err
	}

	return pb.parseHostPortRange(portStr)

}

func (pb *PortBinding) parstProtocal(portStr string) {
	protcIdx := strings.LastIndexAny(portStr, "/")
	if protcIdx > 0 {
		pb.Protocal = portStr[protcIdx+1:]
	} else {
		pb.Protocal = "tcp"
	}
}

func (pb *PortBinding) parseContainerPort(portStr string) error {
	start := strings.LastIndexAny(portStr, ":") + 1
	end := strings.LastIndexAny(portStr, "/")
	if end == -1 {
		end = len(portStr)
	}
	cPort, err := strconv.ParseUint(portStr[start:end], 10, 32)
	if err != nil {
		return errors.New(fmt.Sprintf("Can not parse container port. %s", portStr))
	}
	pb.ContainerPort = int(cPort)
	return nil
}

func (pb *PortBinding) parseHostPortRange(portStr string) error {
	end := strings.LastIndexAny(portStr, ":")
	if end <= 0 {
		pb.hostPortRange = HostPortRange{
			min: pb.ContainerPort,
			max: pb.ContainerPort,
		}
		return nil
	}
	hp := portStr[0:end]
	tildeIdx := strings.LastIndexAny(portStr, "~")

	if tildeIdx == -1 {
		hPort, err := strconv.ParseUint(hp, 10, 32)
		if err != nil {
			return errors.New(fmt.Sprintf("Failed to parse host port. %s, err:%s", hp, err.Error()))
		}
		pb.hostPortRange = HostPortRange{
			min: int(hPort),
			max: int(hPort),
		}
		return nil
	}

	minPortStr, maxPortStr := "1025", "65535"
	if tildeIdx > 0 {
		minPortStr = hp[0:tildeIdx]
		if tildeIdx+1 < len(hp) {
			maxPortStr = hp[tildeIdx+1:]
		}
	}
	minPort, minErr := strconv.ParseInt(minPortStr, 10, 32)
	maxPort, maxErr := strconv.ParseInt(maxPortStr, 10, 32)
	if minPort > maxPort || minErr != nil || maxErr != nil {
		return errors.New(fmt.Sprintf("Parse host port(%s) range error:%s", portStr))
	}
	pb.hostPortRange = HostPortRange{
		min: int(minPort),
		max: int(maxPort),
	}
	return nil
}
