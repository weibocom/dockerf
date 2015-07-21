package ecs

import (
	"encoding/json"

	"github.com/jiangshengwu/aliyun-sdk-for-go/log"
	"github.com/jiangshengwu/aliyun-sdk-for-go/util"
)

type SecurityGroupService interface {
	CreateSecurityGroup(params map[string]string) (CreateSecurityGroupResponse, error)
	AuthorizeSecurityGroup(params map[string]string) (AuthorizeSecurityGroupResponse, error)
	DescribeSecurityGroupAttribute(params map[string]string) (DescribeSecurityGroupAttributeResponse, error)
	DescribeSecurityGroups(params map[string]string) (DescribeSecurityGroupsResponse, error)
	RevokeSecurityGroup(params map[string]string) (RevokeSecurityGroupResponse, error)
	DeleteSecurityGroup(params map[string]string) (DeleteSecurityGroupResponse, error)
	ModifySecurityGroupAttribute(params map[string]string) (ModifySecurityGroupAttributeResponse, error)
	AuthorizeSecurityGroupEgress(params map[string]string) (AuthorizeSecurityGroupEgressResponse, error)
	RevokeSecurityGroupEgress(params map[string]string) (RevokeSecurityGroupEgressResponse, error)
}

type SecurityGroupOperator struct {
	Common *CommonParam
}

// Response struct for CreateSecurityGroup
type CreateSecurityGroupResponse struct {
	util.ErrorResponse
	SecurityGroupId string `json:"SecurityGroupId"`
}

// Response struct for AuthorizeSecurityGroup
type AuthorizeSecurityGroupResponse struct {
	util.ErrorResponse
}

// Response struct for DescribeSecurityGroupAttribute
type DescribeSecurityGroupAttributeResponse struct {
	util.ErrorResponse
	RegionId        string      `json:"RegionId"`
	SecurityGroupId string      `json:"SecurityGroupId"`
	Description     string      `json:"Description"`
	AllPermissions  Permissions `json:"Permissions"`
	VpcId           string      `json:"VpcId"`
}

type Permissions struct {
	AllPermission []PermissionType `json:"Permission"`
}

// See http://docs.aliyun.com/?spm=5176.775974174.2.4.BYfRJ2#/ecs/open-api/datatype&permissiontype
type PermissionType struct {
	IpProtocol              string `json:"IpProtocol"`
	PortRange               string `json:"PortRange"`
	SourceCidrIp            string `json:"SourceCidrIp"`
	SourceGroupId           string `json:"SourceGroupId"`
	SourceGroupOwnerAccount string `json:"SourceGroupOwnerAccount"`
	DestCidrIp              string `json:"DestCidrIp"`
	DestGroupId             string `json:"DestGroupId"`
	DestGroupOwnerAccount   string `json:"DestGroupOwnerAccount"`
	Policy                  string `json:"Policy"`
	NicType                 string `json:"NicType"`
	Priority                string `json:"Priority"`
}

// Response struct for DescribeSecurityGroupsResponse
type DescribeSecurityGroupsResponse struct {
	util.ErrorResponse
	util.PageResponse
	RegionId  string `json:"RegionId"`
	AllGroups Groups `json:"SecurityGroups"`
}

type Groups struct {
	AllGroup []SecurityGroupItemType `json:"SecurityGroup"`
}

// See http://docs.aliyun.com/?spm=5176.775974174.2.4.BYfRJ2#/ecs/open-api/datatype&securitygroupitemtype
type SecurityGroupItemType struct {
	SecurityGroupId   string `json:"SecurityGroupId"`
	SecurityGroupName string `json:"SecurityGroupName"`
	Description       string `json:"Description"`
	VpcId             string `json:"VpcId"`
	CreationTime      string `json:"CreationTime"`
}

// Response struct for RevokeSecurity
type RevokeSecurityGroupResponse struct {
	util.ErrorResponse
}

// Response struct for DeleteSecurityGroup
type DeleteSecurityGroupResponse struct {
	util.ErrorResponse
}

// Response struct for ModifySecurityGroupAttribute
type ModifySecurityGroupAttributeResponse struct {
	util.ErrorResponse
}

// Response struct for AuthorizeSecurityGroupEgress
type AuthorizeSecurityGroupEgressResponse struct {
	util.ErrorResponse
}

// Response struct for RevokeSecurityGroupEgress
type RevokeSecurityGroupEgressResponse struct {
	util.ErrorResponse
}

func (op *SecurityGroupOperator) CreateSecurityGroup(params map[string]string) (CreateSecurityGroupResponse, error) {
	var resp CreateSecurityGroupResponse
	action := GetFuncName(1)
	p := op.Common.ResolveAllParams(action, params)
	result, err := RequestAPI(p)
	if err != nil {
		return CreateSecurityGroupResponse{}, err
	}
	log.Debug(result)
	json.Unmarshal([]byte(result), &resp)
	return resp, nil
}

func (op *SecurityGroupOperator) AuthorizeSecurityGroup(params map[string]string) (AuthorizeSecurityGroupResponse, error) {
	var resp AuthorizeSecurityGroupResponse
	action := GetFuncName(1)
	p := op.Common.ResolveAllParams(action, params)
	result, err := RequestAPI(p)
	if err != nil {
		return AuthorizeSecurityGroupResponse{}, err
	}
	log.Debug(result)
	json.Unmarshal([]byte(result), &resp)
	return resp, nil
}

func (op *SecurityGroupOperator) DescribeSecurityGroupAttribute(params map[string]string) (DescribeSecurityGroupAttributeResponse, error) {
	var resp DescribeSecurityGroupAttributeResponse
	action := GetFuncName(1)
	p := op.Common.ResolveAllParams(action, params)
	result, err := RequestAPI(p)
	if err != nil {
		return DescribeSecurityGroupAttributeResponse{}, err
	}
	log.Debug(result)
	json.Unmarshal([]byte(result), &resp)
	return resp, nil
}

func (op *SecurityGroupOperator) DescribeSecurityGroups(params map[string]string) (DescribeSecurityGroupsResponse, error) {
	var resp DescribeSecurityGroupsResponse
	action := GetFuncName(1)
	p := op.Common.ResolveAllParams(action, params)
	result, err := RequestAPI(p)
	if err != nil {
		return DescribeSecurityGroupsResponse{}, err
	}
	log.Debug(result)
	json.Unmarshal([]byte(result), &resp)
	return resp, nil
}

func (op *SecurityGroupOperator) RevokeSecurityGroup(params map[string]string) (RevokeSecurityGroupResponse, error) {
	var resp RevokeSecurityGroupResponse
	action := GetFuncName(1)
	p := op.Common.ResolveAllParams(action, params)
	result, err := RequestAPI(p)
	if err != nil {
		return RevokeSecurityGroupResponse{}, err
	}
	log.Debug(result)
	json.Unmarshal([]byte(result), &resp)
	return resp, nil
}

func (op *SecurityGroupOperator) DeleteSecurityGroup(params map[string]string) (DeleteSecurityGroupResponse, error) {
	var resp DeleteSecurityGroupResponse
	action := GetFuncName(1)
	p := op.Common.ResolveAllParams(action, params)
	result, err := RequestAPI(p)
	if err != nil {
		return DeleteSecurityGroupResponse{}, err
	}
	log.Debug(result)
	json.Unmarshal([]byte(result), &resp)
	return resp, nil
}

func (op *SecurityGroupOperator) ModifySecurityGroupAttribute(params map[string]string) (ModifySecurityGroupAttributeResponse, error) {
	var resp ModifySecurityGroupAttributeResponse
	action := GetFuncName(1)
	p := op.Common.ResolveAllParams(action, params)
	result, err := RequestAPI(p)
	if err != nil {
		return ModifySecurityGroupAttributeResponse{}, err
	}
	log.Debug(result)
	json.Unmarshal([]byte(result), &resp)
	return resp, nil
}

func (op *SecurityGroupOperator) AuthorizeSecurityGroupEgress(params map[string]string) (AuthorizeSecurityGroupEgressResponse, error) {
	var resp AuthorizeSecurityGroupEgressResponse
	action := GetFuncName(1)
	p := op.Common.ResolveAllParams(action, params)
	result, err := RequestAPI(p)
	if err != nil {
		return AuthorizeSecurityGroupEgressResponse{}, err
	}
	log.Debug(result)
	json.Unmarshal([]byte(result), &resp)
	return resp, nil
}

func (op *SecurityGroupOperator) RevokeSecurityGroupEgress(params map[string]string) (RevokeSecurityGroupEgressResponse, error) {
	var resp RevokeSecurityGroupEgressResponse
	action := GetFuncName(1)
	p := op.Common.ResolveAllParams(action, params)
	result, err := RequestAPI(p)
	if err != nil {
		return RevokeSecurityGroupEgressResponse{}, err
	}
	log.Debug(result)
	json.Unmarshal([]byte(result), &resp)
	return resp, nil
}
