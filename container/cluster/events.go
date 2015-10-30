package cluster

import "github.com/weibocom/dockerf/events"

func (c *Cluster) RegisterEventHandler(cb events.EventsHandler, args ...interface{}) {
	c.Driver.RegisterEventHandler(cb, args)
}
