package sequence

import (
	"sync"
)

type Seq struct {
	s int
}

var sLock sync.Mutex

func (seq *Seq) Next() int {
	sLock.Lock()
	defer sLock.Unlock()
	seq.s++
	return seq.s
}

func (seq *Seq) Get() int {
	return seq.s
}

func (seq *Seq) Max(m int) {
	sLock.Lock()
	defer sLock.Unlock()
	if m > seq.s {
		seq.s = m
	}
}
