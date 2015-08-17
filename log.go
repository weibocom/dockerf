package main

import (
	"os"

	"github.com/Sirupsen/logrus"
)

func setLogLevel(lvl logrus.Level) {
	logrus.SetLevel(lvl)
}

func initLogging() {
	logrus.SetOutput(os.Stdout)
}
