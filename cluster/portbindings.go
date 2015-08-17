package cluster

import (
	"errors"
	"fmt"
	"strconv"
	"strings"

	dutils "github.com/weibocom/dockerf/utils"
)

const (
	Default_Container_Protocol = "tcp"
	Port_Protocol_Separator    = "/"
	Host_Container_Separator   = ":"
)

//ContainerPort style: hostport:containerport/protocol
type ContainerPort string

func (cp ContainerPort) GetContainerProtocol() string {
	protcIdx := strings.LastIndexAny(string(cp), Port_Protocol_Separator)
	if protcIdx > 0 {
		return string(cp)[protcIdx+1:]
	} else {
		return Default_Container_Protocol
	}
}

func (cp ContainerPort) GetContainerPort() string {
	start := strings.LastIndexAny(string(cp), Host_Container_Separator) + 1
	end := strings.LastIndexAny(string(cp), Port_Protocol_Separator)
	if end == -1 {
		end = len(string(cp))
	}
	return string(cp)[start:end]
}

func (cp ContainerPort) GetHostPort() string {
	end := strings.LastIndexAny(string(cp), Host_Container_Separator)
	if end <= 0 {
		return ""
	}
	hp := string(cp)[0:end]
	return hp
}

func (cp ContainerPort) buildContainerPort(protocol string, cPort string, hostport string) string {
	hostAndContainerPort := cPort
	if hostport != "" {
		hostAndContainerPort = strings.Join([]string{hostport, cPort}, Host_Container_Separator)
	}
	if protocol == "" {
		return hostAndContainerPort
	} else {
		return strings.Join([]string{hostAndContainerPort, protocol}, Port_Protocol_Separator)
	}
}

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
	// if pb.HostPort > 0 {
	// 	return pb.HostPort
	// }
	// return pb.hostPortRange.hostPort()
	return pb.HostPort
}

func (pb *PortBinding) Parse(portStr string) error {
	pb.parstProtocal(portStr)

	if err := pb.parseContainerPort(portStr); err != nil {
		return err
	}

	return pb.parseHostPortRange(portStr)

}

func (pb *PortBinding) parstProtocal(portStr string) {
	protocol := ContainerPort(portStr).GetContainerProtocol()
	pb.Protocal = protocol
}

func (pb *PortBinding) parseContainerPort(portStr string) error {
	// start := strings.LastIndexAny(portStr, ":") + 1
	// end := strings.LastIndexAny(portStr, "/")
	// if end == -1 {
	// 	end = len(portStr)
	// }
	containerPort := ContainerPort(portStr).GetContainerPort()
	cPort, err := strconv.ParseUint(containerPort, 10, 32)
	if err != nil {
		return errors.New(fmt.Sprintf("Can not parse container port. %s", portStr))
	}
	pb.ContainerPort = int(cPort)
	return nil
}

func (pb *PortBinding) parseHostPortRange(portStr string) error {
	// end := strings.LastIndexAny(portStr, ":")
	// if end <= 0 {
	// 	pb.hostPortRange = HostPortRange{
	// 		min: pb.ContainerPort,
	// 		max: pb.ContainerPort,
	// 	}
	// 	return nil
	// }
	// hp := portStr[0:end]
	hp := ContainerPort(portStr).GetHostPort()
	if hp == "" {
		pb.hostPortRange = HostPortRange{
			min: pb.ContainerPort,
			max: pb.ContainerPort,
		}
		pb.HostPort = pb.hostPortRange.hostPort()
		return nil
	}
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
		pb.HostPort = pb.hostPortRange.hostPort()
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
	pb.HostPort = pb.hostPortRange.hostPort()
	return nil
}
