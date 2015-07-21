package util

import "fmt"

type SdkError struct {
	Resp ErrorResponse
	Url  string
}

func (err *SdkError) Error() string {
	return fmt.Sprintf("Aliyun SDK Request Error:\nURL: %s\nCode: %s\nMessage: %s", err.Url, err.Resp.Code, err.Resp.Message)
}
