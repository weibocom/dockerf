package sequence

import (
	"sync"
)

type Seq struct {
	s uint
}

var sLock sync.Mutex

func (seq *Seq) Next() uint {
	sLock.Lock()
	defer sLock.Unlock()
	seq.s++
	return seq.s
}

func (seq *Seq) Get() uint {
	return seq.s
}

func (seq *Seq) Max(m uint) {
	sLock.Lock()
	defer sLock.Unlock()
	if m > seq.s {
		seq.s = m
	}
}
