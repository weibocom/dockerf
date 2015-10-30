package topology

import (
	"fmt"
	"time"

	"github.com/weibocom/dockerf/sequence"
)

const (
	GROUP_SEP = '-'
	SEQ_SEP   = '.'
)

var (
	bootUnixTime = time.Now().Unix()
	seq          = &sequence.Seq{}
)

func GenerateName(group string) string {
	sn := seq.Next()
	name := fmt.Sprintf("%s%c%d%c%d", group, GROUP_SEP, bootUnixTime, SEQ_SEP, sn)
	return name
}

func ParseGroup(name string) string {
	ignore := false // ignore the '_' only once
	for idx := len(name) - 1; idx > 0; idx-- {
		c := name[idx]
		if c >= '0' && c <= '9' {
			continue
		}
		if c == SEQ_SEP && !ignore {
			ignore = true
			continue
		}
		if c == GROUP_SEP {
			return name[0:idx]
		}
		break
	}
	return name
}
