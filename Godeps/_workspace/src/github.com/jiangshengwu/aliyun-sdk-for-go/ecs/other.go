package ecs

import (
	"encoding/json"

	"github.com/jiangshengwu/aliyun-sdk-for-go/log"
	"github.com/jiangshengwu/aliyun-sdk-for-go/util"
)

type OtherService interface {
	DescribeInstanceTypes(params map[string]string) (DescribeInstanceTypesResponse, error)
}

type OtherOperator struct {
	Common *CommonParam
}

// Response struct for DescribeInstanceTypes
type DescribeInstanceTypesResponse struct {
	util.ErrorResponse
	AllInstanceTypes InstanceTypes `json:"InstanceTypes"`
}

type InstanceTypes struct {
	AllInstanceType []InstanceTypeItemType `json:"InstanceType"`
}

// See http://docs.aliyun.com/?spm=5176.775974174.2.4.BYfRJ2#/ecs/open-api/datatype&instancetypeitemtype
type InstanceTypeItemType struct {
	InstanceTypeId string  `json:"InstanceTypeId"`
	CpuCoreCount   int     `json:"CpuCoreCount"`
	MemorySize     float64 `json:"MemorySize"`
}

func (op *OtherOperator) DescribeInstanceTypes(params map[string]string) (DescribeInstanceTypesResponse, error) {
	var resp DescribeInstanceTypesResponse
	action := GetFuncName(1)
	p := op.Common.ResolveAllParams(action, params)
	result, err := RequestAPI(p)
	if err != nil {
		return DescribeInstanceTypesResponse{}, err
	}
	log.Debug(result)
	json.Unmarshal([]byte(result), &resp)
	return resp, nil
}
