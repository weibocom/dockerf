package utils

import (
	"time"

	"github.com/Sirupsen/logrus"
)

// it is supposed that:
// 1. 3 times is a suitable notification num during every run
// 2. notification interval should be between 3 and 60 seconds
const (
	min_interval            = 3 * time.Second
	max_interval            = 60 * time.Second
	suitable_interval_times = 3
)

func getWaitInterval(t time.Duration) time.Duration {
	interval := t / suitable_interval_times
	if interval < min_interval {
		interval = min_interval
	}
	if interval > max_interval {
		interval = max_interval
	}
	return interval
}

func TickRun(desc string, ch chan<- bool, timeout time.Duration, f func(), timeoutCallback func()) {
	execChan := make(chan bool)
	interval := getWaitInterval(timeout)
	runs := 0
	start := time.Now()

	go func(f func(), execChan chan<- bool) {
		defer func() {
			if r := recover(); r != nil {
				logrus.Errorf("panic happend:%+v", r)
			}
		}()

		f()
		execChan <- true
		logrus.Debugf("tick-tock run complete:%s", desc)
	}(f, execChan)

	complete := false
Loop:
	for {
		select {
		case <-execChan:
			complete = true
			break Loop
		case <-time.After(interval):
			runs++
			slap := time.Since(start)
			logrus.Debugf("tick-tock run interval timeout:%s. runs: %d, interval:%s, slap: %s, timeout:%s. ", desc, runs, interval, slap, timeout)
			if slap > timeout {
				if timeoutCallback != nil {
					timeoutCallback()
				}
				logrus.Warnf("tick-tock run timeout:%s. runs: %d, interval:%s, slap: %s, timeout:%s. ", desc, runs, interval, slap, timeout)
				break Loop
			}
		}
	}
	ch <- complete
	return
}

func MinInt(x, y int) int {
	if x > y {
		return y
	}
	return x
}

func MaxInt(x, y int) int {
	if x > y {
		return x
	}
	return y
}
