package log

import (
	"fmt"
	stdlog "log"
	"os"
)

type LogLevel byte

const (
	_ LogLevel = iota
	LevelInfo
	LevelDebug
	LevelWard
	LevelError
	LevelFatal
)

var names = []string{
	LevelInfo:  "INFO",
	LevelDebug: "DEBUG",
	LevelWard:  "WARN",
	LevelError: "ERROR",
	LevelFatal: "FATAL",
}

func (l LogLevel) String() string {
	return names[l]
}

var (
	logLevel LogLevel                                                     // log level
	log      *stdlog.Logger = stdlog.New(os.Stdout, "", stdlog.LstdFlags) // logger
)

func Info(f string, v ...interface{}) {
	if logLevel > LevelInfo {
		return
	}
	log.Printf(fmt.Sprintf("[Info] %s", f), v...)
}

func Debug(f string, v ...interface{}) {
	if logLevel > LevelDebug {
		return
	}
	log.Printf(fmt.Sprintf("[Debug] %s", f), v...)
}

func Warn(f string, v ...interface{}) {
	if logLevel > LevelWard {
		return
	}
	log.Printf(fmt.Sprintf("[Warn] %s", f), v...)
}

func Error(f string, v ...interface{}) {
	if logLevel > LevelError {
		return
	}
	log.Printf(fmt.Sprintf("[Error] %s", f), v...)
}

func Fatal(f string, v ...interface{}) {
	if logLevel > LevelFatal {
		return
	}
	log.Printf(fmt.Sprintf("[Fatal] %s", f), v...)
}

func SetLevel(l LogLevel) {
	logLevel = l
}

func SetLevelByName(n string) {
	for k, v := range names {
		if v == n {
			logLevel = LogLevel(k)
		}
	}
}

func init() {
	logLevel = LevelInfo
}
