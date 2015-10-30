package swarm

import (
	"strings"

	"github.com/samalba/dockerclient"
	"github.com/weibocom/dockerf/events"
)

func wrapCallback(eh events.EventsHandler, ec chan error, args ...interface{}) dockerclient.Callback {
	return func(e *dockerclient.Event, ec chan error, args ...interface{}) {
		id := e.Id
		splits := strings.SplitN(e.From, " node:", 2)
		from := splits[0]
		node := ""
		if len(splits) >= 2 {
			node = splits[1]
		}
		status := e.Status
		time := e.Time
		eh(id, status, from, node, time, args)
	}
}
