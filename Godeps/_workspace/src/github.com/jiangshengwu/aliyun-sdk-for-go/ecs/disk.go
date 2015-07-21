package ecs

import (
	"encoding/json"

	"github.com/jiangshengwu/aliyun-sdk-for-go/log"
	"github.com/jiangshengwu/aliyun-sdk-for-go/util"
)

type DiskService interface {
	CreateDisk(params map[string]string) (CreateDiskResponse, error)
	DescribeDisks(params map[string]string) (DescribeDisksResponse, error)
	AttachDisk(params map[string]string) (AttachDiskResponse, error)
	DetachDisk(params map[string]string) (DetachDiskResponse, error)
	ModifyDiskAttribute(params map[string]string) (ModifyDiskAttributeResponse, error)
	DeleteDisk(params map[string]string) (DeleteDiskResponse, error)
	ReInitDisk(params map[string]string) (ReInitDiskResponse, error)
	ResetDisk(params map[string]string) (ResetDiskResponse, error)
	ReplaceSystemDisk(params map[string]string) (ReplaceSystemDiskResponse, error)
	ResizeDisk(params map[string]string) (ResizeDiskResponse, error)
}

type DiskOperator struct {
	Common *CommonParam
}

// Response struct for CreateDisk
type CreateDiskResponse struct {
	util.ErrorResponse
	DiskId string `json:"DiskId"`
}

// Response struct for DescribeDisks
type DescribeDisksResponse struct {
	util.ErrorResponse
	AllDisks Disks `json:"Disks"`
}

type Disks struct {
	AllDisk []DiskItemType `json:"Disk"`
}

// See http://docs.aliyun.com/?spm=5176.775974174.2.4.BYfRJ2#/ecs/open-api/datatype&diskitemtype
type DiskItemType struct {
	DiskId             string         `json:"DiskId"`
	RegionId           string         `json:"RegionId"`
	ZoneId             string         `json:"ZoneId"`
	DiskName           string         `json:"DiskName"`
	Description        string         `json:"Description"`
	Type               string         `json:"Type"`
	Category           string         `json:"Category"`
	Size               int            `json:"Size"`
	ImageId            string         `json:"ImageId"`
	SourceSnapshotId   string         `json:"SourceSnapshotId"`
	ProductCode        string         `json:"ProductCode"`
	Portable           string         `json:"Portable"`
	Status             string         `json:"Status"`
	AllOperationLocks  OperationLocks `json:"OperationLocks"`
	InstanceId         string         `json:"InstanceId"`
	Device             string         `json:"Device"`
	DeleteWithInstance string         `json:"DeleteWithInstance"`
	DeleteAutoSnapshot string         `json:"DeleteAutoSnapshot"`
	EnableAutoSnapshot string         `json:"EnableAutoSnapshot"`
	CreationTime       string         `json:"CreationTime"`
	AttachedTime       string         `json:"AttachedTime"`
	DetachedTime       string         `json:"DetachedTime"`
}

// Response struct for AttachDisk
type AttachDiskResponse struct {
	util.ErrorResponse
}

// Response struct for DetachDisk
type DetachDiskResponse struct {
	util.ErrorResponse
}

// Response struct for ModifyDiskAttribute
type ModifyDiskAttributeResponse struct {
	util.ErrorResponse
}

// Response struct for DeleteDisk
type DeleteDiskResponse struct {
	util.ErrorResponse
}

// Response struct for ReInitDisk
type ReInitDiskResponse struct {
	util.ErrorResponse
}

// Response struct for ResetDisk
type ResetDiskResponse struct {
	util.ErrorResponse
}

// Response struct for ReplaceSystemDisk
type ReplaceSystemDiskResponse struct {
	util.ErrorResponse
	DiskId string `json:"DiskId"`
}

// Response struct for ResizeDisk
type ResizeDiskResponse struct {
	util.ErrorResponse
}

func (op *DiskOperator) CreateDisk(params map[string]string) (CreateDiskResponse, error) {
	var resp CreateDiskResponse
	action := GetFuncName(1)
	p := op.Common.ResolveAllParams(action, params)
	result, err := RequestAPI(p)
	if err != nil {
		return CreateDiskResponse{}, err
	}
	log.Debug(result)
	json.Unmarshal([]byte(result), &resp)
	return resp, nil
}

func (op *DiskOperator) DescribeDisks(params map[string]string) (DescribeDisksResponse, error) {
	var resp DescribeDisksResponse
	action := GetFuncName(1)
	p := op.Common.ResolveAllParams(action, params)
	result, err := RequestAPI(p)
	if err != nil {
		return DescribeDisksResponse{}, err
	}
	log.Debug(result)
	json.Unmarshal([]byte(result), &resp)
	return resp, nil
}

func (op *DiskOperator) AttachDisk(params map[string]string) (AttachDiskResponse, error) {
	var resp AttachDiskResponse
	action := GetFuncName(1)
	p := op.Common.ResolveAllParams(action, params)
	result, err := RequestAPI(p)
	if err != nil {
		return AttachDiskResponse{}, err
	}
	log.Debug(result)
	json.Unmarshal([]byte(result), &resp)
	return resp, nil
}

func (op *DiskOperator) DetachDisk(params map[string]string) (DetachDiskResponse, error) {
	var resp DetachDiskResponse
	action := GetFuncName(1)
	p := op.Common.ResolveAllParams(action, params)
	result, err := RequestAPI(p)
	if err != nil {
		return DetachDiskResponse{}, err
	}
	log.Debug(result)
	json.Unmarshal([]byte(result), &resp)
	return resp, nil
}

func (op *DiskOperator) ModifyDiskAttribute(params map[string]string) (ModifyDiskAttributeResponse, error) {
	var resp ModifyDiskAttributeResponse
	action := GetFuncName(1)
	p := op.Common.ResolveAllParams(action, params)
	result, err := RequestAPI(p)
	if err != nil {
		return ModifyDiskAttributeResponse{}, err
	}
	log.Debug(result)
	json.Unmarshal([]byte(result), &resp)
	return resp, nil
}

func (op *DiskOperator) DeleteDisk(params map[string]string) (DeleteDiskResponse, error) {
	var resp DeleteDiskResponse
	action := GetFuncName(1)
	p := op.Common.ResolveAllParams(action, params)
	result, err := RequestAPI(p)
	if err != nil {
		return DeleteDiskResponse{}, err
	}
	log.Debug(result)
	json.Unmarshal([]byte(result), &resp)
	return resp, nil
}

func (op *DiskOperator) ReInitDisk(params map[string]string) (ReInitDiskResponse, error) {
	var resp ReInitDiskResponse
	action := GetFuncName(1)
	p := op.Common.ResolveAllParams(action, params)
	result, err := RequestAPI(p)
	if err != nil {
		return ReInitDiskResponse{}, err
	}
	log.Debug(result)
	json.Unmarshal([]byte(result), &resp)
	return resp, nil
}

func (op *DiskOperator) ResetDisk(params map[string]string) (ResetDiskResponse, error) {
	var resp ResetDiskResponse
	action := GetFuncName(1)
	p := op.Common.ResolveAllParams(action, params)
	result, err := RequestAPI(p)
	if err != nil {
		return ResetDiskResponse{}, err
	}
	log.Debug(result)
	json.Unmarshal([]byte(result), &resp)
	return resp, nil
}

func (op *DiskOperator) ReplaceSystemDisk(params map[string]string) (ReplaceSystemDiskResponse, error) {
	var resp ReplaceSystemDiskResponse
	action := GetFuncName(1)
	p := op.Common.ResolveAllParams(action, params)
	result, err := RequestAPI(p)
	if err != nil {
		return ReplaceSystemDiskResponse{}, err
	}
	log.Debug(result)
	json.Unmarshal([]byte(result), &resp)
	return resp, nil
}

func (op *DiskOperator) ResizeDisk(params map[string]string) (ResizeDiskResponse, error) {
	var resp ResizeDiskResponse
	action := GetFuncName(1)
	p := op.Common.ResolveAllParams(action, params)
	result, err := RequestAPI(p)
	if err != nil {
		return ResizeDiskResponse{}, err
	}
	log.Debug(result)
	json.Unmarshal([]byte(result), &resp)
	return resp, nil
}
