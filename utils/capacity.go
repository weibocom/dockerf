package utils

import (
	"strconv"
	"strings"
)

// 512m 1gb etc...
func ParseCapacity(str string) (int, error) {
	ustr := strings.ToUpper(str)
	end := len(ustr)
	if ustr[end-1] == 'B' {
		end = end - 1
	}
	base := 1
	switch ustr[end-1] {
	case 'K':
		end -= 1
		base = 1024
	case 'M':
		end -= 1
		base = 1024 * 1024
	case 'G':
		end -= 1
		base = 1024 * 1024 * 1024
	default:
	}
	ustr = ustr[0:end]

	num, err := strconv.ParseInt(ustr, 10, 64)
	if err != nil {
		return -1, err
	}
	return int(num) * int(base), nil
}
