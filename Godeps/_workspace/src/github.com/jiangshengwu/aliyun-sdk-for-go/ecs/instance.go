package ecs

import (
	"encoding/json"

	"github.com/jiangshengwu/aliyun-sdk-for-go/log"
	"github.com/jiangshengwu/aliyun-sdk-for-go/util"
)

type InstanceService interface {
	CreateInstance(params map[string]string) (CreateInstanceResponse, error)
	StartInstance(params map[string]string) (StartInstanceResponse, error)
	StopInstance(params map[string]string) (StopInstanceResponse, error)
	RebootInstance(params map[string]string) (RebootInstanceResponse, error)
	ModifyInstanceAttribute(params map[string]string) (ModifyInstanceAttributeResponse, error)
	ModifyInstanceVpcAttribute(params map[string]string) (ModifyInstanceVpcAttributeResponse, error)
	DescribeInstanceStatus(params map[string]string) (DescribeInstanceStatusResponse, error)
	DescribeInstanceAttribute(params map[string]string) (DescribeInstanceAttributeStatusResponse, error)
	DescribeInstances(params map[string]string) (DescribeInstancesResponse, error)
	DeleteInstance(params map[string]string) (DeleteInstanceResponse, error)
	JoinSecurityGroup(params map[string]string) (JoinSecurityGroupResponse, error)
	LeaveSecurityGroup(params map[string]string) (LeaveSecurityGroupResponse, error)
}

type InstanceOperator struct {
	Common *CommonParam
}

// Response struct for CreateInstance
type CreateInstanceResponse struct {
	util.ErrorResponse
	InstanceId string `json:"InstanceId"`
}

// Response struct for StartInstance
type StartInstanceResponse struct {
	util.ErrorResponse
}

// Response struct for StopInstance
type StopInstanceResponse struct {
	util.ErrorResponse
}

// Response struct for RebootInstance
type RebootInstanceResponse struct {
	util.ErrorResponse
}

// Response struct for ModifyInstanceAttribute
type ModifyInstanceAttributeResponse struct {
	util.ErrorResponse
}

// Response struct for ModifyInstanceVpcAttribute
type ModifyInstanceVpcAttributeResponse struct {
	util.ErrorResponse
}

// Response struct for DescribeInstanceStatus
type DescribeInstanceStatusResponse struct {
	util.ErrorResponse
	util.PageResponse
	AllInstanceStatuses InstanceStatusSetType `json:"InstanceStatuses"`
}

type InstanceStatusSetType struct {
	AllInstanceStatus []InstanceStatusItemType `json:"InstanceStatus"`
}

// See http://docs.aliyun.com/?spm=5176.775974174.2.4.BYfRJ2#/ecs/open-api/datatype&instancestatusitemtype
type InstanceStatusItemType struct {
	InstanceId string `json:"InstanceId"`
	Status     string `json:"Status"`
}

// See http://docs.aliyun.com/?spm=5176.775974174.2.4.BYfRJ2#/ecs/open-api/datatype&instanceattributestype
type InstanceAttributesType struct {
	InstanceId              string                  `json:"InstanceId"`
	InstanceName            string                  `json:"InstanceName"`
	Description             string                  `json:"Description"`
	ImageId                 string                  `json:"ImageId"`
	RegionId                string                  `json:"RegionId"`
	ZoneId                  string                  `json:"ZoneId"`
	ClusterId               string                  `json:"ClusterId"`
	InstanceType            string                  `json:"InstanceType"`
	HostName                string                  `json:"HostName"`
	Status                  string                  `json:"Status"`
	AllOperationLocks       OperationLocks          `json:"OperationLocks"`
	AllSecurityGroupIds     SecurityGroupIdSetType  `json:"SecurityGroupIds"`
	PublicIpAddress         IpAddressSetType        `json:"PublicIpAddress"`
	InnerIpAddress          IpAddressSetType        `json:"InnerIpAddress"`
	InternetMaxBandwidthIn  int                     `json:"InternetMaxBandwidthIn"`
	InternetMaxBandwidthOut int                     `json:"InternetMaxBandwidthOut"`
	InternetChargeType      string                  `json:"InternetChargeType"`
	CreationTime            string                  `json:"CreationTime"`
	InstanceNetworkType     string                  `json:"InstanceNetworkType"`
	VpcAttributes           VpcAttributesType       `json:"VpcAttributes"`
	EipAddress              EipAddressAssociateType `json:"EipAddress"`
}

// See http://docs.aliyun.com/?spm=5176.775974174.2.4.BYfRJ2#/ecs/open-api/datatype&vpcattributestype
type VpcAttributesType struct {
	VpcId            string           `json:"VpcId"`
	VSwitchId        string           `json:"VSwitchId"`
	PrivateIpAddress IpAddressSetType `json:"PrivateIpAddress"`
	NatIpAddress     string           `json:"NatIpAddress"`
}

// See http://docs.aliyun.com/?spm=5176.775974174.2.4.BYfRJ2#/ecs/open-api/datatype&eipaddressassociatetype
type EipAddressAssociateType struct {
	AllocationId       string `json:"AllocationId"`
	IpAddress          string `json:"IpAddress"`
	Bandwidth          string `json:"Bandwidth"`
	InternetChargeType string `json:"InternetChargeType"`
}

// Response struct for DescribeInstanceAttributeStatus
type DescribeInstanceAttributeStatusResponse struct {
	util.ErrorResponse
	InstanceAttributesType
}

// See http://docs.aliyun.com/?spm=5176.775974174.2.4.BYfRJ2#/ecs/open-api/datatype&securitygroupidsettype
type SecurityGroupIdSetType struct {
	AllSecurityGroupId []string `json:"SecurityGroupId"`
}

// See http://docs.aliyun.com/?spm=5176.775974174.2.4.BYfRJ2#/ecs/open-api/datatype&ipaddresssettype
type IpAddressSetType struct {
	AllIpAddress []string `json:"IpAddress"`
}

// Response struct for DescribeInstances
type DescribeInstancesResponse struct {
	util.ErrorResponse
	util.PageResponse
	AllInstances Instances `json:"Instances"`
}

type Instances struct {
	AllInstance []InstanceAttributesType `json:"Instance"`
}

// Response struct for DeleteInstance
type DeleteInstanceResponse struct {
	util.ErrorResponse
}

// Response struct for JoinSecurityGroup
type JoinSecurityGroupResponse struct {
	util.ErrorResponse
}

// Response struct for LeaveSecurityGroup
type LeaveSecurityGroupResponse struct {
	util.ErrorResponse
}

func (op *InstanceOperator) CreateInstance(params map[string]string) (CreateInstanceResponse, error) {
	var resp CreateInstanceResponse
	action := GetFuncName(1)
	p := op.Common.ResolveAllParams(action, params)
	result, err := RequestAPI(p)
	if err != nil {
		return CreateInstanceResponse{}, err
	}
	log.Debug(result)
	json.Unmarshal([]byte(result), &resp)
	return resp, nil
}

func (op *InstanceOperator) StartInstance(params map[string]string) (StartInstanceResponse, error) {
	var resp StartInstanceResponse
	action := GetFuncName(1)
	p := op.Common.ResolveAllParams(action, params)
	result, err := RequestAPI(p)
	if err != nil {
		return StartInstanceResponse{}, err
	}
	log.Debug(result)
	json.Unmarshal([]byte(result), &resp)
	return resp, nil
}

func (op *InstanceOperator) StopInstance(params map[string]string) (StopInstanceResponse, error) {
	var resp StopInstanceResponse
	action := GetFuncName(1)
	p := op.Common.ResolveAllParams(action, params)
	result, err := RequestAPI(p)
	if err != nil {
		return StopInstanceResponse{}, err
	}
	log.Debug(result)
	json.Unmarshal([]byte(result), &resp)
	return resp, nil
}

func (op *InstanceOperator) RebootInstance(params map[string]string) (RebootInstanceResponse, error) {
	var resp RebootInstanceResponse
	action := GetFuncName(1)
	p := op.Common.ResolveAllParams(action, params)
	result, err := RequestAPI(p)
	if err != nil {
		return RebootInstanceResponse{}, err
	}
	log.Debug(result)
	json.Unmarshal([]byte(result), &resp)
	return resp, nil
}

func (op *InstanceOperator) ModifyInstanceAttribute(params map[string]string) (ModifyInstanceAttributeResponse, error) {
	var resp ModifyInstanceAttributeResponse
	action := GetFuncName(1)
	p := op.Common.ResolveAllParams(action, params)
	result, err := RequestAPI(p)
	if err != nil {
		return ModifyInstanceAttributeResponse{}, err
	}
	log.Debug(result)
	json.Unmarshal([]byte(result), &resp)
	return resp, nil
}

func (op *InstanceOperator) ModifyInstanceVpcAttribute(params map[string]string) (ModifyInstanceVpcAttributeResponse, error) {
	var resp ModifyInstanceVpcAttributeResponse
	action := GetFuncName(1)
	p := op.Common.ResolveAllParams(action, params)
	result, err := RequestAPI(p)
	if err != nil {
		return ModifyInstanceVpcAttributeResponse{}, err
	}
	log.Debug(result)
	json.Unmarshal([]byte(result), &resp)
	return resp, nil
}

func (op *InstanceOperator) DescribeInstanceStatus(params map[string]string) (DescribeInstanceStatusResponse, error) {
	var resp DescribeInstanceStatusResponse
	action := GetFuncName(1)
	p := op.Common.ResolveAllParams(action, params)
	result, err := RequestAPI(p)
	if err != nil {
		return DescribeInstanceStatusResponse{}, err
	}
	log.Debug(result)
	json.Unmarshal([]byte(result), &resp)
	return resp, nil
}

func (op *InstanceOperator) DescribeInstanceAttribute(params map[string]string) (DescribeInstanceAttributeStatusResponse, error) {
	var resp DescribeInstanceAttributeStatusResponse
	action := GetFuncName(1)
	p := op.Common.ResolveAllParams(action, params)
	result, err := RequestAPI(p)
	if err != nil {
		return DescribeInstanceAttributeStatusResponse{}, err
	}
	log.Debug(result)
	json.Unmarshal([]byte(result), &resp)
	return resp, nil
}

func (op *InstanceOperator) DescribeInstances(params map[string]string) (DescribeInstancesResponse, error) {
	var resp DescribeInstancesResponse
	action := GetFuncName(1)
	p := op.Common.ResolveAllParams(action, params)
	result, err := RequestAPI(p)
	if err != nil {
		return DescribeInstancesResponse{}, err
	}
	log.Debug(result)
	json.Unmarshal([]byte(result), &resp)
	return resp, nil
}

func (op *InstanceOperator) DeleteInstance(params map[string]string) (DeleteInstanceResponse, error) {
	var resp DeleteInstanceResponse
	action := GetFuncName(1)
	p := op.Common.ResolveAllParams(action, params)
	result, err := RequestAPI(p)
	if err != nil {
		return DeleteInstanceResponse{}, err
	}
	log.Debug(result)
	json.Unmarshal([]byte(result), &resp)
	return resp, nil
}

func (op *InstanceOperator) JoinSecurityGroup(params map[string]string) (JoinSecurityGroupResponse, error) {
	var resp JoinSecurityGroupResponse
	action := GetFuncName(1)
	p := op.Common.ResolveAllParams(action, params)
	result, err := RequestAPI(p)
	if err != nil {
		return JoinSecurityGroupResponse{}, err
	}
	log.Debug(result)
	json.Unmarshal([]byte(result), &resp)
	return resp, nil
}

func (op *InstanceOperator) LeaveSecurityGroup(params map[string]string) (LeaveSecurityGroupResponse, error) {
	var resp LeaveSecurityGroupResponse
	action := GetFuncName(1)
	p := op.Common.ResolveAllParams(action, params)
	result, err := RequestAPI(p)
	if err != nil {
		return LeaveSecurityGroupResponse{}, err
	}
	log.Debug(result)
	json.Unmarshal([]byte(result), &resp)
	return resp, nil
}
