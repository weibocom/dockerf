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
	Group string
	Seq   int
	Valid bool
}

func (mn *MachineName) GetName() string {
	name := fmt.Sprintf("%s-%d", mn.Group, mn.Seq)
	FormateMachineName(mn.Group, mn.Seq)
	return name
}

func (mn *MachineName) Parse(name string) bool {
	sep := M_SEP
	if idx := strings.LastIndex(name, sep); idx > 0 {
		group := name[0:idx]
		numStr := name[idx+len(sep):]
		seq, err := strconv.ParseInt(numStr, 10, 32)
		if err == nil {
			mn.Group = group
			mn.Seq = int(seq)
			mn.Valid = true
			return mn.Valid
		}
	}
	mn.Valid = false
	return mn.Valid
}

func FormateMachineName(group string, seq int) string {
	name := fmt.Sprintf("%s-%d", group, seq)
	return name
}

func ParseMachineName(name string) (string, int) {
	sep := M_SEP
	if idx := strings.LastIndex(name, sep); idx > 0 {
		group := name[0:idx]
		numStr := name[idx+len(sep):]
		seq, err := strconv.ParseInt(numStr, 10, 32)
		if err == nil {
			return group, int(seq)
		}
	}
	return "", -1
}
