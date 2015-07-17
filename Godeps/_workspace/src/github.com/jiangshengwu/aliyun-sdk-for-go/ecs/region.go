package ecs

import (
	"encoding/json"

	"github.com/jiangshengwu/aliyun-sdk-for-go/log"
	"github.com/jiangshengwu/aliyun-sdk-for-go/util"
)

type RegionService interface {
	DescribeRegions(params map[string]string) (DescribeRegionsResponse, error)
	DescribeZones(params map[string]string) (DescribeZonesResponse, error)
}

type RegionOperator struct {
	Common *CommonParam
}

// Response struct for DescribeRegions
type DescribeRegionsResponse struct {
	util.ErrorResponse
	AllRegions Regions `json:"Regions"`
}

type Regions struct {
	AllRegion []RegionType `json:"Region"`
}

// See http://docs.aliyun.com/?spm=5176.775974174.2.4.BYfRJ2#/ecs/open-api/datatype&regiontype
type RegionType struct {
	RegionId   string `json:"RegionId"`
	RegionName string `json:"RegionName"`
}

// Response struct for DescribeZones
type DescribeZonesResponse struct {
	util.ErrorResponse
	AllZones Zones `json:"Zones"`
}

type Zones struct {
	AllZone []Zone `json:"Zone"`
}

// See http://docs.aliyun.com/?spm=5176.775974174.2.4.BYfRJ2#/ecs/open-api/datatype&zonetype
type Zone struct {
	ZoneId                    string                        `json:"ZoneId"`
	LocalName                 string                        `json:"LocalName"`
	AvailableResourceCreation AvailableResourceCreationType `json:"AvailableResourceCreation"`
	AvailableDiskCategories   AvailableDiskCategoriesType   `json:"AvailableDiskCategories"`
}

// See http://docs.aliyun.com/?spm=5176.775974174.2.4.BYfRJ2#/ecs/open-api/datatype&availableresourcecreationtype
type AvailableResourceCreationType struct {
	ResourceTypes []string `json:"ResourceTypes"`
}

// See http://docs.aliyun.com/?spm=5176.775974174.2.4.BYfRJ2#/ecs/open-api/datatype&availablediskcategoriestype
type AvailableDiskCategoriesType struct {
	DiskCategories []string `json:"DiskCategories"`
}

func (op *RegionOperator) DescribeRegions(params map[string]string) (DescribeRegionsResponse, error) {
	var resp DescribeRegionsResponse
	action := GetFuncName(1)
	p := op.Common.ResolveAllParams(action, params)
	result, err := RequestAPI(p)
	if err != nil {
		return DescribeRegionsResponse{}, err
	}
	log.Debug(result)
	json.Unmarshal([]byte(result), &resp)
	return resp, nil
}

func (op *RegionOperator) DescribeZones(params map[string]string) (DescribeZonesResponse, error) {
	var resp DescribeZonesResponse
	action := GetFuncName(1)
	p := op.Common.ResolveAllParams(action, params)
	result, err := RequestAPI(p)
	if err != nil {
		return DescribeZonesResponse{}, err
	}
	log.Debug(result)
	json.Unmarshal([]byte(result), &resp)
	return resp, nil
}
