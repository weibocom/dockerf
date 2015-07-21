package util

import "testing"

func Test_MapToSign(t *testing.T) {
	params := map[string]string{
		"SignatureVersion": "1",
		"Action":           "DescribeInstanceAttribute",
		"Format":           "JSON",
		"TimeStamp":        "2015-05-13T15:00:42",
		"SignatureNonce":   "0ce21fa98484a7e855ecf84c83721452",
		"InstanceId":       "i-253op6931",
		"AccessKeyId":      "myKeyId",
		"SignatureMethod":  "HMAC-SHA1",
		"Version":          "2014-05-26",
	}
	signature := MapToSign(params, "myKeySecrect", "GET")
	if signature != "MuppLklVpLWM%2BpLCt7uSDvyAs60%3D" {
		t.Error("Signature is incorrect.")
	}
}
