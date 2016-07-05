package log

import (
	"errors"
	"fmt"
	stdlog "log"
	"os"
	"runtime"
	"strconv"
	"strings"
)

type LogLevel byte

const (
	_ LogLevel = iota
	LevelInfo
	LevelDebug
	LevelWarn
	LevelError
	LevelFatal
)

var (
	ErrWrongLogLevel = errors.New("log level not define")
)

var names = []string{
	LevelInfo:  "INFO",
	LevelDebug: "DEBUG",
	LevelWarn:  "WARN",
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

func writeLog(level, format string, v ...interface{}) {
	_, file, line, ok := runtime.Caller(2)
	if !ok {
		file = "???"
		line = 0
	}
	c := string(file + ":" + strconv.FormatInt(int64(line), 10))
	log.Printf(fmt.Sprintf("[%s] [%s] %s", level, c, format), v...)
}

func Info(f string, v ...interface{}) {
	if logLevel > LevelInfo {
		return
	}
	writeLog("Info", f, v...)
}

func Debug(f string, v ...interface{}) {
	if logLevel > LevelDebug {
		return
	}
	writeLog("Debug", f, v...)
}

func Warn(f string, v ...interface{}) {
	if logLevel > LevelWarn {
		return
	}
	writeLog("Warn", f, v...)
}

func Error(f string, v ...interface{}) {
	if logLevel > LevelError {
		return
	}
	writeLog("Error", f, v...)
}

func Fatal(f string, v ...interface{}) {
	if logLevel > LevelFatal {
		return
	}
	writeLog("Fatal", f, v...)
}

func SetLevel(l LogLevel) error {
	if l < LevelInfo || l > LevelFatal {
		return ErrWrongLogLevel
	}
	logLevel = l
	return nil
}

func SetLevelByName(n string) error {
	for k, v := range names {
		if v == strings.ToUpper(n) {
			logLevel = LogLevel(k)
			return nil
		}
	}
	return ErrWrongLogLevel
}

func init() {
	logLevel = LevelInfo
}
