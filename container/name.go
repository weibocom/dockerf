package container

import (
	"fmt"
	"strconv"
	"strings"
)

type ContainerName struct {
	Valid bool
	Node  string
	Group string
	Seq   int
}

func (cng *ContainerName) Parse(cName string) bool {
	if len(cName) <= 3 {
		cng.Valid = false
		return cng.Valid
	}
	name := cName
	if name[0] == '/' {
		name = name[1:]
	}

	groupIdx := strings.LastIndex(name, "/")

	node := "unkown"
	if groupIdx > 0 {
		node = name[0:groupIdx]
	}

	name = name[groupIdx+1:]
	seqIdx := strings.LastIndex(name, "-")
	if seqIdx <= 0 {
		fmt.Printf("'%s' is not a validate container name. [a-z]+(-)[0-9]+\n", name)
		return false
	}
	group := name[:seqIdx]
	seqStr := name[seqIdx+1:]
	seq, err := strconv.ParseUint(seqStr, 10, 32)
	if err != nil {
		fmt.Printf("'%s' is not a validate container name. [a-z]+(-)[0-9]+\n", name)
		return false
	}
	cng.Node = node
	cng.Group = group
	cng.Seq = int(seq)
	cng.Valid = true
	return cng.Valid
}

func (cng *ContainerName) GetName() string {
	name := fmt.Sprintf("%s-%d", cng.Group, cng.Seq)
	return name
}

func (cng *ContainerName) CompositeNameWithNode(node string, name string) string {
	return fmt.Sprintf("%s/%s", node, name)
}
