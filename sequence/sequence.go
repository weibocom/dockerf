package sequence

import (
	"sync"
)

type Seq struct {
	s int
	sync.Mutex
}

func (seq *Seq) Next() int {
	seq.Lock()
	defer seq.Unlock()
	seq.s++
	return seq.s
}

func (seq *Seq) Get() int {
	return seq.s
}

func (seq *Seq) Max(m int) {
	seq.Lock()
	defer seq.Unlock()
	if m > seq.s {
		seq.s = m
	}
}
