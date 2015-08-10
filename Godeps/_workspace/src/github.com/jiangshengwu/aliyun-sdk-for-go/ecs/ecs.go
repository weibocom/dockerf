// ECS API package
package ecs

import (
	"runtime"
	"strings"
	"time"

	"github.com/jiangshengwu/aliyun-sdk-for-go/log"
	"github.com/jiangshengwu/aliyun-sdk-for-go/util"
)

const (
	// ECS API Host
	ECSHost string = "https://ecs.aliyuncs.com/?"

	// All ECS APIs only support GET method
	ECSHttpMethod = "GET"

	// SDK only supports JSON format
	Format = "JSON"

	Version          = "2014-05-26"
	SignatureMethod  = "HMAC-SHA1"
	SignatureVersion = "1.0"
)

// struct for ECS client
type EcsClient struct {
	Common *CommonParam

	// Access to API call from this client
	Region        RegionService
	SecurityGroup SecurityGroupService
	Instance      InstanceService
	Other         OtherService
	Image         ImageService
	Snapshot      SnapshotService
	Disk          DiskService
	Network       NetworkService
	Monitor       MonitorService
}

// Initialize an ECS client
func NewClient(accessKeyId string, accessKeySecret string, resourceOwnerAccount string) *EcsClient {
	client := &EcsClient{}
	client.Common = &CommonParam{}
	client.Common.AccessKeyId = accessKeyId
	client.Common.AccessKeySecret = accessKeySecret
	client.Common.ResourceOwnerAccount = resourceOwnerAccount
	ps := map[string]string{
		"Format":           Format,
		"Version":          Version,
		"AccessKeyId":      client.Common.AccessKeyId,
		"SignatureMethod":  SignatureMethod,
		"SignatureVersion": SignatureVersion,
	}
	client.Common.attr = ps

	client.Region = &RegionOperator{client.Common}
	client.SecurityGroup = &SecurityGroupOperator{client.Common}
	client.Instance = &InstanceOperator{client.Common}
	client.Other = &OtherOperator{client.Common}
	client.Image = &ImageOperator{client.Common}
	client.Snapshot = &SnapshotOperator{client.Common}
	client.Disk = &DiskOperator{client.Common}
	client.Network = &NetworkOperator{client.Common}
	client.Monitor = &MonitorOperator{client.Common}

	return client
}

func (client *EcsClient) GetClientName() string {
	return "ECS Client"
}

func (client *EcsClient) GetVersion() string {
	return client.Common.attr["Version"]
}

func (client *EcsClient) GetSignatureMethod() string {
	return client.Common.attr["SignatureMethod"]
}

func (client *EcsClient) GetSignatureVersion() string {
	return client.Common.attr["SignatureVersion"]
}

// struct for common parameters
type CommonParam struct {
	AccessKeyId          string
	AccessKeySecret      string
	ResourceOwnerAccount string
	attr                 map[string]string
}

func RequestAPI(params map[string]string) (string, error) {
	query := util.GetQueryFromMap(params)
	req := &util.AliyunRequest{}
	req.Url = ECSHost + query
	log.Debug(req.Url)
	result, err := req.DoGetRequest()
	return result, err
}

// Get function name by skip
// which means the differs between Caller and Callers
func GetFuncName(skip int) string {
	pc, _, _, _ := runtime.Caller(skip)
	name := runtime.FuncForPC(pc).Name()
	i := strings.LastIndex(name, ".")
	if i >= 0 {
		name = name[i+1:]
	}
	return name
}

// Generate all parameters include Signature
func (c *CommonParam) ResolveAllParams(action string, params map[string]string) map[string]string {
	if params == nil {
		params = make(map[string]string, len(c.attr))
	}
	// Process parameters
	for key, value := range c.attr {
		params[key] = value
	}
	params["Action"] = action
	if c.ResourceOwnerAccount != "" {
		params["ResourceOwnerAccount"] = c.ResourceOwnerAccount
	}
	params["TimeStamp"] = time.Now().UTC().Format("2006-01-02T15:04:05Z")
	params["SignatureNonce"] = util.GetGuid()
	sign := util.MapToSign(params, c.AccessKeySecret, ECSHttpMethod)
	params["Signature"] = sign
	return params
}
