package watcher

import (
	logger "log"
	"os"
)

type Log interface {
	Printf(string, ...interface{})
	Fatalf(string, ...interface{})
}

// The logger used package wide.
var log Log = logger.New(os.Stderr, "", logger.LstdFlags)

// Set the logger used by the package.
// This isn't thread safe. Only call it before starting any watcher.
func SetLog(l Log) {
	log = l
}
