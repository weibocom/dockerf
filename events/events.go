package events

import "github.com/Sirupsen/logrus"

type EventsHandler func(id, status, from, node string, time int64, args ...interface{})

func DefaultEventHandler(id, status, from, node string, time int64, args ...interface{}) {
	logrus.Debugf("event received. id:%s, status:%s, from:%s, node:%s, time:%d", id, status, from, node, time)
}
