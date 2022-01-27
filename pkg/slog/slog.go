// very simple logger

package slog

import (
	"fmt"
	"os"
	"sync"
	"time"
)

var lck sync.Mutex
var ident string
var debug int
var tformat string

func init() {
	ident = ""
	debug = 0
	tformat = "2006-01-02 15:04:05"
}

// init logger
func Init(identString string, debugLevel int, timeFormat string) {
	ident = identString
	tformat = timeFormat

	if debugLevel < 0 {
		debug = 0
	} else if debugLevel > 9 {
		debug = 9
	} else {
		debug = debugLevel
	}
}

// log error message
func genLog(level string, format string, args ...interface{}) {
	lck.Lock()
	ts := time.Now().Format(tformat)
	fmt.Fprintf(os.Stderr, ts+"|"+ident+"["+level+"] "+format+"\n", args...)
	lck.Unlock()
}

// Err logs ERROR messages to stderr
func Err(args ...interface{}) {
	genLog("ERR", args[0].(string), args[1:]...)
}

// Info logs INFO messages to stderr
func Info(args ...interface{}) {
	genLog("INFO", args[0].(string), args[1:]...)
}

// Warn logs WARNING messages to stderr
func Warn(args ...interface{}) {
	genLog("WARN", args[0].(string), args[1:]...)
}

// Fatal logs FATAL messages to stderr and calls os.Exit()
func Fatal(args ...interface{}) {
	genLog("FATL", args[0].(string), args[1:]...)
	os.Exit(1)
}

// Debug logs DEBUG messages at specified debug level
func Debug(lvl int, args ...interface{}) {
	if lvl >= debug {
		genLog(fmt.Sprintf("DBG%d", lvl), args[0].(string), args[1:]...)
	}
}
