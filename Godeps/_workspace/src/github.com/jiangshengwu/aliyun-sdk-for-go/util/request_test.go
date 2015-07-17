package util

import "testing"

func Test_DoRequest(t *testing.T) {
	req := &AliyunRequest{
		"http://weibo.com",
	}
	_, err := req.DoGetRequest()
	if err != nil {
		t.Error("Request failed.")
	}
}

func Test_GetQueryFromMap(t *testing.T) {
	params := map[string]string{
		"user": "root",
		"pass": "test",
	}
	query := GetQueryFromMap(params)
	if query != "user=root&pass=test" {
		t.Error("Query string is incorrect.")
	}
}
