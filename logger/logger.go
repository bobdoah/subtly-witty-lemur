package logger

import (
	"log"
	"os"
	"sync"
)

// Logger is a logging interface
type Logger interface {
	Printf(format string, v ...interface{})
}

type discardLog struct{}

func (*discardLog) Printf(format string, v ...interface{}) {
}

var logger Logger = &discardLog{}

// Enabled is if the debug logger is enabled
var Enabled bool
var once sync.Once

// GetLogger provides the instance of the debug logger
func GetLogger() Logger {
	once.Do(func() {
		if Enabled {
			logger = log.New(os.Stderr, "", log.Ldate|log.Ltime|log.Lmicroseconds|log.Lshortfile)
		}
	})
	return logger
}
