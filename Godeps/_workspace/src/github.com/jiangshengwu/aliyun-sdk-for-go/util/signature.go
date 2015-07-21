package util

import (
	"crypto/hmac"
	"crypto/md5"
	"crypto/rand"
	"crypto/sha1"
	"encoding/base64"
	"encoding/hex"
	"io"
	"net/url"
	"sort"
	"strings"
)

// Get signature from params in map
func MapToSign(params map[string]string, keySecret string, httpMethod string) string {
	key := []byte(keySecret + "&")
	h := hmac.New(sha1.New, key)
	query := canonicalizedFromMap(params)
	h.Write([]byte(httpMethod + "&%2F&" + query))
	sign := PercentEncode(base64.StdEncoding.EncodeToString(h.Sum(nil)))
	return sign
}

// Get canonicalized query string from params in map
func canonicalizedFromMap(params map[string]string) string {
	keys := make([]string, len(params))
	i := 0
	for k := range params {
		keys[i] = k
		i++
	}
	sort.Strings(keys)
	sign := ""
	for _, v := range keys {
		sign += "&" + PercentEncode(v) + "=" + PercentEncode(params[v])
	}
	sign = PercentEncode(sign[1:])
	return sign
}

// URL encode
func PercentEncode(str string) string {
	s := url.QueryEscape(str)
	s = strings.Replace(s, "+", "%20", -1)
	s = strings.Replace(s, "*", "%2A", -1)
	s = strings.Replace(s, "%7E", "~", -1)
	return s
}

// MD5 hash
func getMd5String(s string) string {
	h := md5.New()
	h.Write([]byte(s))
	return hex.EncodeToString(h.Sum(nil))
}

// Generate random guid
func GetGuid() string {
	b := make([]byte, 48)
	if _, err := io.ReadFull(rand.Reader, b); err != nil {
		return ""
	}
	return getMd5String(base64.URLEncoding.EncodeToString(b))
}
