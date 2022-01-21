package log

import (
	"flag"
	"fmt"
	"os"
	"time"
)

var (
	consoleColor bool

	FatalColorStart = ""
	ErrorColorStart = ""
	InfoColorStart  = ""
	WarnColorStart  = ""
	DebugColorStart = ""
	TraceColorStart = ""
	colorEnd        = ""

	logType string
	debug   bool
	trace   bool
)

func init() {
	flag.BoolVar(&consoleColor, "color", false, "consoleColor")
	flag.StringVar(&logType, "log", "nil", "debug or trace")
}

func Init() {
	if consoleColor {
		FatalColorStart = "\033[31m"
		ErrorColorStart = "\033[31m"
		InfoColorStart = "\033[32m"
		WarnColorStart = "\033[33m"
		DebugColorStart = "\033[34m"
		TraceColorStart = "\033[34m"
		colorEnd = "\033[0m"
	}
	switch logType {
	case "debug":
		debug = true
	case "trace":
		debug = true
		trace = true
	}
}

func getTime() string {
	return time.Now().Format("2006:01:02 15:04:05")
}

type Log struct {
	From string
}

func (l *Log) Fatal(err error) {
	fmt.Printf("%sFatal %s -> %s: %v%s\n", FatalColorStart, getTime(), l.From, err, colorEnd)
	os.Exit(1)
}

func (l *Log) Error(cid int32, sid int32, err error) {
	if err != nil {
		fmt.Printf("%sError %s -> %s(%d -> %d): %v%s\n", ErrorColorStart, getTime(), l.From, cid, sid, err, colorEnd)
	} else {
		fmt.Printf("%sError %s -> %s(%d -> %d)%s\n", ErrorColorStart, getTime(), l.From, cid, sid, colorEnd)
	}
}

func (l *Log) Info(cid int32, sid int32, message interface{}) {
	if message != nil {
		fmt.Printf("%sInfo %s -> %s(%d -> %d): %v%s\n", InfoColorStart, getTime(), l.From, cid, sid, message, colorEnd)
	} else {
		fmt.Printf("%sInfo %s -> %s(%d -> %d)%s\n", InfoColorStart, getTime(), l.From, cid, sid, colorEnd)
	}
}

func (l *Log) Warn(cid int32, sid int32, message interface{}) {
	if message != nil {
		fmt.Printf("%sWarn %s -> %s(%d -> %d): %v%s\n", WarnColorStart, getTime(), l.From, cid, sid, message, colorEnd)
	} else {
		fmt.Printf("%sWarn %s -> %s(%d -> %d)%s\n", WarnColorStart, getTime(), l.From, cid, sid, colorEnd)
	}
}

func (l *Log) Debug(cid int32, sid int32, message interface{}) {
	if debug {
		if message != nil {
			fmt.Printf("%sDebug %s -> %s(%d -> %d): %v%s\n", DebugColorStart, getTime(), l.From, cid, sid, message, colorEnd)
		} else {
			fmt.Printf("%sDebug %s -> %s(%d -> %d)%s\n", DebugColorStart, getTime(), l.From, cid, sid, colorEnd)
		}
	}
}

func (l *Log) Trace(cid int32, sid int32, message interface{}) {
	if trace {
		if message != nil {
			fmt.Printf("%sTrace %s -> %s(%d -> %d): %v%s\n", TraceColorStart, getTime(), l.From, cid, sid, message, colorEnd)
		} else {
			fmt.Printf("%sTrace %s -> %s(%d -> %d)%s\n", TraceColorStart, getTime(), l.From, cid, sid, colorEnd)
		}
	}
}
