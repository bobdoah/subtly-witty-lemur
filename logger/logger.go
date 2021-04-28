package logger

import (
	"log"
	"os"
)

// Logger is a logging interface
type Logger interface {
	Printf(format string, v ...interface{})
}

type discardLog struct{}

func (*discardLog) Printf(format string, v ...interface{}) {
}

// DebugLogger is a logger for print debug statements
var DebugLogger Logger = &discardLog{}

func debugLogger(enable bool) {
	if enable {
		DebugLogger = log.New(os.Stderr, "", log.Ldate|log.Ltime|log.Lmicroseconds|log.Lshortfile)
	}
}
