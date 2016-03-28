package starx

import (
	"fmt"
	"log"
)

var Log *log.Logger // logger

// Error logs a message at error level.
func Error(f string, v ...interface{}) {
	Log.Printf(fmt.Sprintf("[Error] %s", f), v...)
}

// compatibility alias for Warning()
func Info(f string, v ...interface{}) {
	Log.Printf(fmt.Sprintf("[Info] %s", f), v...)
}

func Warning(f string, v ...interface{}) {
	Log.Printf(fmt.Sprintf("[Warning] %s", f), v...)
}

func Debug(f string, v ...interface{}) {
	Log.Printf(fmt.Sprintf("[Debug] %s", f), v...)
}
