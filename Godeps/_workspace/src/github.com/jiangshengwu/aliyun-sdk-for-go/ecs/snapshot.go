package ecs

import (
	"encoding/json"

	"github.com/jiangshengwu/aliyun-sdk-for-go/log"
	"github.com/jiangshengwu/aliyun-sdk-for-go/util"
)

type SnapshotService interface {
	CreateSnapshot(params map[string]string) (CreateSnapshotResponse, error)
	DeleteSnapshot(params map[string]string) (DeleteSnapshotResponse, error)
	DescribeSnapshots(params map[string]string) (DescribeSnapshotsResponse, error)
	ModifyAutoSnapshotPolicy(params map[string]string) (ModifyAutoSnapshotPolicyResponse, error)
	DescribeAutoSnapshotPolicy(params map[string]string) (DescribeAutoSnapshotPolicyResponse, error)
}

type SnapshotOperator struct {
	Common *CommonParam
}

// Response struct for CreateSnapshot
type CreateSnapshotResponse struct {
	util.ErrorResponse
	SnapshotId string `json:"SnapshotId"`
}

// Response struct for DeleteSnapshot
type DeleteSnapshotResponse struct {
	util.ErrorResponse
}

// Response struct for DescribeSnapshots
type DescribeSnapshotsResponse struct {
	util.ErrorResponse
	util.PageResponse
	AllSnapshots Snapshots `json:"Snapshots"`
}

type Snapshots struct {
	AllSnapshot SnapshotType `json:"Snapshot"`
}

// See http://docs.aliyun.com/?spm=5176.775974174.2.4.BYfRJ2#/ecs/open-api/datatype&snapshottype
type SnapshotType struct {
	SnapshotId     string `json:"SnapshotId"`
	SnapshotName   string `json:"SnapshotName"`
	Description    string `json:"Description"`
	Progress       string `json:"Progress"`
	SourceDiskId   string `json:"SourceDiskId"`
	SourceDiskSize int    `json:"SourceDiskSize"`
	SourceDiskType string `json:"SourceDiskType"`
	ProductCode    string `json:"ProductCode"`
	CreationTime   string `json:"CreationTime"`
}

// Response struct for ModifyAutoSnapshotPolicy
type ModifyAutoSnapshotPolicyResponse struct {
	util.ErrorResponse
}

// Response struct for DescribeAutoSnapshotPolicy
type DescribeAutoSnapshotPolicyResponse struct {
	util.ErrorResponse
	AutoSnapshotExecutionStatus AutoSnapshotExecutionStatusType `json:"AutoSnapshotExecutionStatus"`
	AutoSnapshotPolicy          AutoSnapshotPolicyType          `json:"AutoSnapshotPolicy"`
}

// See http://docs.aliyun.com/?spm=5176.775974174.2.4.BYfRJ2#/ecs/open-api/datatype&autosnapshotexecutionstatustype
type AutoSnapshotExecutionStatusType struct {
	SystemDiskExecutionStatus string `json:"SystemDiskExecutionStatus"`
	DataDiskExecutionStatus   string `json:"DataDiskExecutionStatus"`
}

// See http://docs.aliyun.com/?spm=5176.775974174.2.4.BYfRJ2#/ecs/open-api/datatype&autosnapshotpolicytype
type AutoSnapshotPolicyType struct {
	SystemDiskPolicyEnabled           string `json:"SystemDiskPolicyEnabled"`
	SystemDiskPolicyTimePeriod        int    `json:"SystemDiskPolicyTimePeriod"`
	SystemDiskPolicyRetentionDays     int    `json:"SystemDiskPolicyRetentionDays"`
	SystemDiskPolicyRetentionLastWeek string `json:"SystemDiskPolicyRetentionLastWeek"`
	DataDiskPolicyEnabled             string `json:"DataDiskPolicyEnabled"`
	DataDiskPolicyTimePeriod          int    `json:"DataDiskPolicyTimePeriod"`
	DataDiskPolicyRetentionDays       int    `json:"DataDiskPolicyRetentionDays"`
	DataDiskPolicyRetentionLastWeek   string `json:"DataDiskPolicyRetentionLastWeek"`
}

func (op *SnapshotOperator) CreateSnapshot(params map[string]string) (CreateSnapshotResponse, error) {
	var resp CreateSnapshotResponse
	action := GetFuncName(1)
	p := op.Common.ResolveAllParams(action, params)
	result, err := RequestAPI(p)
	if err != nil {
		return CreateSnapshotResponse{}, err
	}
	log.Debug(result)
	json.Unmarshal([]byte(result), &resp)
	return resp, nil
}

func (op *SnapshotOperator) DeleteSnapshot(params map[string]string) (DeleteSnapshotResponse, error) {
	var resp DeleteSnapshotResponse
	action := GetFuncName(1)
	p := op.Common.ResolveAllParams(action, params)
	result, err := RequestAPI(p)
	if err != nil {
		return DeleteSnapshotResponse{}, err
	}
	log.Debug(result)
	json.Unmarshal([]byte(result), &resp)
	return resp, nil
}

func (op *SnapshotOperator) DescribeSnapshots(params map[string]string) (DescribeSnapshotsResponse, error) {
	var resp DescribeSnapshotsResponse
	action := GetFuncName(1)
	p := op.Common.ResolveAllParams(action, params)
	result, err := RequestAPI(p)
	if err != nil {
		return DescribeSnapshotsResponse{}, err
	}
	log.Debug(result)
	json.Unmarshal([]byte(result), &resp)
	return resp, nil
}

func (op *SnapshotOperator) ModifyAutoSnapshotPolicy(params map[string]string) (ModifyAutoSnapshotPolicyResponse, error) {
	var resp ModifyAutoSnapshotPolicyResponse
	action := GetFuncName(1)
	p := op.Common.ResolveAllParams(action, params)
	result, err := RequestAPI(p)
	if err != nil {
		return ModifyAutoSnapshotPolicyResponse{}, err
	}
	log.Debug(result)
	json.Unmarshal([]byte(result), &resp)
	return resp, nil
}

func (op *SnapshotOperator) DescribeAutoSnapshotPolicy(params map[string]string) (DescribeAutoSnapshotPolicyResponse, error) {
	var resp DescribeAutoSnapshotPolicyResponse
	action := GetFuncName(1)
	p := op.Common.ResolveAllParams(action, params)
	result, err := RequestAPI(p)
	if err != nil {
		return DescribeAutoSnapshotPolicyResponse{}, err
	}
	log.Debug(result)
	json.Unmarshal([]byte(result), &resp)
	return resp, nil
}
