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
	LevelDebug
	LevelInfo
	LevelWarn
	LevelError
	LevelFatal
	LevelClose
)

const (
	LogDebug = "DEBUG"
	LogInfo  = "INFO"
	LogWarn  = "WARN"
	LogError = "ERROR"
	LogFatal = "FATAL"
)

var (
	ErrWrongLogLevel = errors.New("log level not define")
)

var names = []string{
	LevelDebug: "DEBUG",
	LevelInfo:  "INFO",
	LevelWarn:  "WARN",
	LevelError: "ERROR",
	LevelFatal: "FATAL",
}

func (l LogLevel) String() string {
	return names[l]
}

var (
	logLevel LogLevel                                                     // log level
	logger   *stdlog.Logger = stdlog.New(os.Stdout, "", stdlog.LstdFlags) // logger
)

func logSite() string {
	_, file, line, ok := runtime.Caller(3)
	if !ok {
		file = "???"
		line = 0
	}
	c := string(file + ":" + strconv.FormatInt(int64(line), 10))
	return c
}

func writeLog(level string, v ...interface{}) {
	logger.Printf(fmt.Sprintf("[%s] [%s] %s", level, logSite(), fmt.Sprint(v...)))
}

func writeLogf(level, format string, v ...interface{}) {
	logger.Printf(fmt.Sprintf("[%s] [%s] %s", level, logSite(), format), v...)
}

func Tracef(f string, v ...interface{}) {
	if logLevel > LevelFatal {
		return
	}
	buf := make([]byte, 10000)
	n := runtime.Stack(buf, false)
	buf = buf[:n]
	v = append(v, string(buf))
	writeLogf("Trace", f+"\n%s", v...)
}

func Debugf(f string, v ...interface{}) {
	if logLevel > LevelDebug {
		return
	}
	writeLogf(LogDebug, f, v...)
}

func Infof(f string, v ...interface{}) {
	if logLevel > LevelInfo {
		return
	}
	writeLogf(LogInfo, f, v...)
}

func Warnf(f string, v ...interface{}) {
	if logLevel > LevelWarn {
		return
	}
	writeLogf(LogWarn, f, v...)
}

func Errorf(f string, v ...interface{}) {
	if logLevel > LevelError {
		return
	}
	writeLogf(LogError, f, v...)
}

func Fatalf(f string, v ...interface{}) {
	if logLevel > LevelFatal {
		return
	}
	writeLogf(LogFatal, f, v...)
	os.Exit(-1)
}

func Trace(v ...interface{}) {
	if logLevel > LevelFatal {
		return
	}
	buf := make([]byte, 10000)
	n := runtime.Stack(buf, false)
	buf = buf[:n]
	v = append(v, string(buf))
	writeLogf("Trace", "%s\n%s", v...)
}

func Debug(v ...interface{}) {
	if logLevel > LevelDebug {
		return
	}
	writeLog(LogDebug, v...)
}

func Info(v ...interface{}) {
	if logLevel > LevelInfo {
		return
	}
	writeLog(LogInfo, v...)
}

func Warn(f string, v ...interface{}) {
	if logLevel > LevelWarn {
		return
	}
	writeLog(LogWarn, v...)
}

func Error(v ...interface{}) {
	if logLevel > LevelError {
		return
	}
	writeLog(LogError, v...)
}

func Fatal(v ...interface{}) {
	if logLevel > LevelFatal {
		return
	}
	writeLog(LogFatal, v...)
	os.Exit(-1)
}

func SetLevel(l LogLevel) error {
	if l < LevelDebug || l > LevelClose {
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
