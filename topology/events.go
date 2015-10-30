package topology

import "github.com/Sirupsen/logrus"

func TopologyEventsHandler(id, status, from, node string, time int64, args ...interface{}) {
	logrus.Debugf("topology event received. id:%s, status:%s, from:%s, node:%s, time:%d", id, status, from, node, time)
}
