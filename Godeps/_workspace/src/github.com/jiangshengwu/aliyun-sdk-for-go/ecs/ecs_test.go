package ecs

import "testing"

func Test_Init1(t *testing.T) {
	cli := NewClient("testKeyId1", "testKeySecrect1", "")

	if cli.Common.ResourceOwnerAccount != "" {
		t.Error("ResourceOwnerAccount value is incorrect.")
	}
	if cli.GetSignatureMethod() != "HMAC-SHA1" {
		t.Error("SignatureMethod value is incorrect.")
	}
}

func Test_Init2(t *testing.T) {
	cli := NewClient("testKeyId2", "testKeySecrect2", "True")

	if cli.Common.ResourceOwnerAccount != "True" {
		t.Error("ResourceOwnerAccount value is incorrect.")
	}
	if cli.GetVersion() != "2014-05-26" {
		t.Error("Version value is incorrect.")
	}
}
