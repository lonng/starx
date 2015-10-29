package starx

import (
	"fmt"
	"log"
)

var Log *log.Logger // logger

// Error logs a message at error level.
func Error(info string) {
	Log.Panicln(fmt.Sprintf("[Panic] %s", info))
}

// compatibility alias for Warning()
func Info(info string) {
	Log.Println(fmt.Sprintf("[Info] %s", info))
}

func Warning(info string) {
	Log.Println(fmt.Sprintf("[Warning] %s", info))
}

func Debug(info string) {
	Log.Println(fmt.Sprintf("[Debug] %s", info))
}
