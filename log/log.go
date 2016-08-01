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

func writeLog(level string, v ...interface{}) {
	_, file, line, ok := runtime.Caller(2)
	if !ok {
		file = "???"
		line = 0
	}
	c := string(file + ":" + strconv.FormatInt(int64(line), 10))
	log.Printf(fmt.Sprintf("[%s] [%s] %s", level, c, fmt.Sprint(v...)))
}

func writeLogf(level, format string, v ...interface{}) {
	_, file, line, ok := runtime.Caller(2)
	if !ok {
		file = "???"
		line = 0
	}
	c := string(file + ":" + strconv.FormatInt(int64(line), 10))
	log.Printf(fmt.Sprintf("[%s] [%s] %s", level, c, format), v...)
}

func Infof(f string, v ...interface{}) {
	if logLevel > LevelInfo {
		return
	}
	writeLogf("Info", f, v...)
}

func Debugf(f string, v ...interface{}) {
	if logLevel > LevelDebug {
		return
	}
	writeLogf("Debug", f, v...)
}

func Warnf(f string, v ...interface{}) {
	if logLevel > LevelWarn {
		return
	}
	writeLogf("Warn", f, v...)
}

func Errorf(f string, v ...interface{}) {
	if logLevel > LevelError {
		return
	}
	writeLogf("Error", f, v...)
}

func Fatalf(f string, v ...interface{}) {
	if logLevel > LevelFatal {
		return
	}
	writeLogf("Fatal", f, v...)
}

func Info(v ...interface{}) {
	if logLevel > LevelInfo {
		return
	}
	writeLog("Info", v...)
}

func Debug(v ...interface{}) {
	if logLevel > LevelDebug {
		return
	}
	writeLog("Debug", v...)
}

func Warn(f string, v ...interface{}) {
	if logLevel > LevelWarn {
		return
	}
	writeLog("Warn", v...)
}

func Error(v ...interface{}) {
	if logLevel > LevelError {
		return
	}
	writeLog("Error", v...)
}

func Fatal(v ...interface{}) {
	if logLevel > LevelFatal {
		return
	}
	writeLog("Fatal", v...)
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
