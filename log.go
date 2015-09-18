package main

import (
	// "fmt"
	"github.com/Sirupsen/logrus"
	"os"
)

func setLogLevel(lvl logrus.Level) {
	logrus.SetLevel(lvl)
}

func initLogging() {
	// logrus.SetOutput(os.Stdout)
	fileName := "./console.log"
	console, err := os.Create(fileName)
	if err != nil {
		panic("Fail to log console output, error: " + err.Error())
	}
	logrus.SetOutput(console)
}
