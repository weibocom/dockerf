package dlog

import (
	"fmt"
	"os"

	"github.com/ivpusic/golog"
	"github.com/ivpusic/golog/appenders"
)

var debugLogger *golog.Logger
var infoLogger *golog.Logger
var warnLogger *golog.Logger
var errorLogger *golog.Logger
var fatalLogger *golog.Logger

func init() {
	logPath := os.Getenv("LOG_PATH")
	if logPath == "" {
		logPath = "/tmp"
	}
	debugLogger = golog.GetLogger("debug")
	debugLogger.Enable(appenders.File(golog.Conf{
		"path": logPath + "/debug.log",
	}))
	debugLogger.Disable(golog.StdoutAppender())
	debugLogger.Level = golog.DEBUG

	infoLogger = golog.GetLogger("info")
	infoLogger.Enable(appenders.File(golog.Conf{
		"path": logPath + "/info.log",
	}))
	infoLogger.Disable(golog.StdoutAppender())
	infoLogger.Level = golog.INFO

	warnLogger = golog.GetLogger("warn")
	warnLogger.Enable(appenders.File(golog.Conf{
		"path": logPath + "/warn.log",
	}))
	warnLogger.Disable(golog.StdoutAppender())
	warnLogger.Level = golog.WARN

	errorLogger = golog.GetLogger("error")
	errorLogger.Enable(appenders.File(golog.Conf{
		"path": logPath + "/error.log",
	}))
	errorLogger.Disable(golog.StdoutAppender())
	errorLogger.Level = golog.ERROR
}

func Debug(msg interface{}, data ...interface{}) {
	debugLogger.Debug(msg, data)
}

func Debugf(msg string, data ...interface{}) {
	debugLogger.Debugf(msg, data)
}

func Info(msg interface{}, data ...interface{}) {
	infoLogger.Info(msg, data)
}

func Infof(msg string, data ...interface{}) {
	infoLogger.Info(fmt.Printf(msg, data))
}

func Warn(msg interface{}, data ...interface{}) {
	warnLogger.Warn(msg, data)
}

func Warnf(msg string, data ...interface{}) {
	warnLogger.Warn(fmt.Printf(msg, data))
}

func Error(msg interface{}, data ...interface{}) {
	errorLogger.Error(msg, data)
}

func Errorf(msg string, data ...interface{}) {
	errorLogger.Error(fmt.Printf(msg, data))
}
