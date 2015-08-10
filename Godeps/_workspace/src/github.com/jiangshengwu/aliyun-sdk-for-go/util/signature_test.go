package util

import (
	"testing"
)

func Test_MapToSign(t *testing.T) {
	params := map[string]string{
		"SignatureVersion": "1.0",
		"Action":           "DescribeInstanceAttribute",
		"Format":           "JSON",
		"TimeStamp":        "2015-05-13T15:00:42Z",
		"SignatureNonce":   "0ce21fa98484a7e855ecf84c83721452",
		"InstanceId":       "i-253op6931",
		"AccessKeyId":      "myKeyId",
		"SignatureMethod":  "HMAC-SHA1",
		"Version":          "2014-05-26",
	}
	signature := MapToSign(params, "myKeySecret", "GET")
	t.Log(signature)
	if signature != "fs0B%2B9%2BYPI0qZOZ1QXcDnlObMZM%3D" {
		t.Error("Signature is incorrect.")
	}
}
