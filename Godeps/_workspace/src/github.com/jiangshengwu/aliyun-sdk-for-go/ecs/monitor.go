package ecs

import (
	"encoding/json"

	"github.com/jiangshengwu/aliyun-sdk-for-go/log"
	"github.com/jiangshengwu/aliyun-sdk-for-go/util"
)

type MonitorService interface {
	DescribeInstanceMonitorData(params map[string]string) (DescribeInstanceMonitorDataResponse, error)
	DescribeEipMonitorData(params map[string]string) (DescribeEipMonitorDataResponse, error)
	DescribeDiskMonitorData(params map[string]string) (DescribeDiskMonitorDataResponse, error)
}

type MonitorOperator struct {
	Common *CommonParam
}

// Response struct for DescribeInstanceTypes
type DescribeInstanceMonitorDataResponse struct {
	util.ErrorResponse
	AllMonitorData InstanceMonitorData `json:"MonitorData"`
}

type InstanceMonitorData struct {
	InstanceMonitor []InstanceMonitorDataType `json:"InstanceMonitorData"`
}

// See http://docs.aliyun.com/?spm=5176.775974174.2.4.BYfRJ2#/ecs/open-api/datatype&instancemonitordatatype
type InstanceMonitorDataType struct {
	InstanceId        string `json:"InstanceId"`
	CPU               int    `json:"CPU"`
	IntranetRX        int    `json:"IntranetRX"`
	IntranetTX        int    `json:"IntranetTX"`
	IntranetBandwidth int    `json:"IntranetBandwidth"`
	InternetRX        int    `json:"InternetRX"`
	InternetTX        int    `json:"InternetTX"`
	InternetBandwidth int    `json:"InternetBandwidth"`
	IOPSRead          int    `json:"IOPSRead"`
	IOPSWrite         int    `json:"IOPSWrite"`
	BPSRead           int    `json:"BPSRead"`
	BPSWrite          int    `json:"BPSWrite"`
	TimeStamp         string `json:"TimeStamp"`
}

// Response struct for DescribeEipMonitorData
type DescribeEipMonitorDataResponse struct {
	util.ErrorResponse
	AllEipMonitorData EipMonitorData `json:"EipMonitorDatas"`
}

type EipMonitorData struct {
	EipMonitor []EipMonitorDataType `json:"EipMonitorData"`
}

// See http://docs.aliyun.com/?spm=5176.775974174.2.4.BYfRJ2#/ecs/open-api/datatype&eipmonitordatatype
type EipMonitorDataType struct {
	EipRX        int    `json:"EipRX"`
	EipTX        int    `json:"EipTX"`
	EipFlow      int    `json:"EipFlow"`
	EipBandwidth int    `json:"EipBandwidth"`
	EipPackets   int    `json:"EipPackets"`
	TimeStamp    string `json:"TimeStamp"`
}

// Response struct for DescribeDiskMonitorData
type DescribeDiskMonitorDataResponse struct {
	util.ErrorResponse
	TotalCount     int             `json:"TotalCount"`
	AllMonitorData DiskMonitorData `json:"MonitorData"`
}

type DiskMonitorData struct {
	DiskMonitor []DiskMonitorDataType `json:"DiskMonitorData"`
}

// See http://docs.aliyun.com/?spm=5176.775974174.2.4.BYfRJ2#/ecs/open-api/datatype&diskmonitordatatype
type DiskMonitorDataType struct {
	DiskId    string `json:"DiskId"`
	IOPSRead  int    `json:"IOPSRead"`
	IOPSWrite int    `json:"IOPSWrite"`
	IOPSTotal int    `json:"IOPSTotal"`
	BPSRead   int    `json:"BPSRead"`
	BPSWrite  int    `json:"BPSWrite"`
	BPSTotal  int    `json:"BPSTotal"`
	TimeStamp string `json:"TimeStamp"`
}

func (op *MonitorOperator) DescribeInstanceMonitorData(params map[string]string) (DescribeInstanceMonitorDataResponse, error) {
	var resp DescribeInstanceMonitorDataResponse
	action := GetFuncName(1)
	p := op.Common.ResolveAllParams(action, params)
	result, err := RequestAPI(p)
	if err != nil {
		return DescribeInstanceMonitorDataResponse{}, err
	}
	log.Debug(result)
	json.Unmarshal([]byte(result), &resp)
	return resp, nil
}

func (op *MonitorOperator) DescribeEipMonitorData(params map[string]string) (DescribeEipMonitorDataResponse, error) {
	var resp DescribeEipMonitorDataResponse
	action := GetFuncName(1)
	p := op.Common.ResolveAllParams(action, params)
	result, err := RequestAPI(p)
	if err != nil {
		return DescribeEipMonitorDataResponse{}, err
	}
	log.Debug(result)
	json.Unmarshal([]byte(result), &resp)
	return resp, nil
}

func (op *MonitorOperator) DescribeDiskMonitorData(params map[string]string) (DescribeDiskMonitorDataResponse, error) {
	var resp DescribeDiskMonitorDataResponse
	action := GetFuncName(1)
	p := op.Common.ResolveAllParams(action, params)
	result, err := RequestAPI(p)
	if err != nil {
		return DescribeDiskMonitorDataResponse{}, err
	}
	log.Debug(result)
	json.Unmarshal([]byte(result), &resp)
	return resp, nil
}
