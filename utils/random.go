package utils

import (
	"math/rand"
	"time"
)

var uRandom = rand.New(rand.NewSource(time.Now().UnixNano()))

func RandomUInt(minimal, maximal int) int {
	min := int32(minimal)
	max := int32(maximal)
	rn := uRandom.Int31n(max)
	if rn < min {
		rn = rn%(max-min) + min
	}
	return int(rn)
}
