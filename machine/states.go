package machine

import (
	"fmt"
	"time"

	"github.com/Sirupsen/logrus"
	"github.com/docker/machine/state"
	"github.com/weibocom/dockerf/utils"
)

var (
	reqStateTimeOut = 3 * time.Second
)

// load all machines' state
func (c *Cluster) LoadStates() {
	machines := c.ListAll()
	doChan := make(chan bool)
	for _, m := range machines { //
		tm := m
		go utils.TickRun(
			fmt.Sprintf("loading machine state and ip '%s'", tm.Name()),
			doChan,
			reqStateTimeOut,
			func() {
				s := tm.LoadState()
				logrus.Debugf("loading machine state complete. %s:%s", tm.Name(), s.String())
				_, err := tm.LoadIp()
				if err != nil {
					logrus.Errorf("loading machine '%s' ip failed:%s", tm.Name(), err.Error())
				} else {
					logrus.Debugf("loading machine ip complete. '%s':%s", tm.Name(), tm.GetCachedIp())
				}
			},
			func() {
				logrus.Warnf("machine '%s' state is set be timeout", tm.Name())
				tm.setCachedState(state.Timeout)
			},
		)
	}
	for i := 0; i < len(machines); i++ {
		<-doChan
	}
	return
}

func IsRunning(s state.State) bool {
	return s == state.Running
}

func IsRunnable(s state.State) bool {
	switch s {
	case state.Stopped, state.Paused, state.Saved, state.Stopping:
		return true
	default:
		return false
	}
}
