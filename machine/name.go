package machine

import (
	"fmt"
	"strconv"
	"strings"
)

const (
	M_SEP = "-"
)

type MachineName struct {
	Prefix string
	Seq    uint
	Valid  bool
}

func (mn *MachineName) GetName() string {
	name := fmt.Sprintf("%s-%d", mn.Prefix, mn.Seq)
	return name
}

func (mn *MachineName) Parse(name string) bool {
	sep := M_SEP
	if idx := strings.LastIndex(name, sep); idx > 0 {
		prefix := name[0:idx]
		numStr := name[idx+len(sep):]
		seq, err := strconv.ParseInt(numStr, 10, 32)
		if err == nil {
			mn.Prefix = prefix
			mn.Seq = uint(seq)
			mn.Valid = true
			return mn.Valid
		}
	}
	mn.Valid = false
	return mn.Valid
}
