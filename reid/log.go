/*
 * Copyright (c) 2017 Jon Szymaniak <jon.szymaniak@gmail.com>
 * SPDX License Identifier: GPL-3.0
 *
 * Diagnostic logging functionality
 */
package reid

import (
	"fmt"
	"os"
)

var LogLevel int = LogLevelInfo

const (
	LogLevelSilent = iota
	LogLevelError
	LogLevelInfo
	LogLevelDebug
	LogLevelVerbose
)

var prefix = [...]string{
	"",
	"[Error]   ",
	"[Info]    ",
	"[Debug]   ",
	"[Verbose] ",
}

func Errorf(format string, v ...interface{}) {
	printf(LogLevelError, format, v...)
}

func Error(v ...interface{}) {
	println(LogLevelError, v...)
}

func Infof(format string, v ...interface{}) {
	printf(LogLevelInfo, format, v...)
}

func Info(v ...interface{}) {
	println(LogLevelInfo, v...)
}

func Debugf(format string, v ...interface{}) {
	printf(LogLevelDebug, format, v...)
}

func Debug(v ...interface{}) {
	println(LogLevelDebug, v...)
}

func Verbosef(format string, v ...interface{}) {
	printf(LogLevelVerbose, format, v...)
}

func Verbose(v ...interface{}) {
	println(LogLevelVerbose, v...)
}

func println(level int, v ...interface{}) {
	if LogLevel >= level {
		fmt.Fprint(os.Stderr, prefix[level])
		fmt.Fprintln(os.Stderr, v...)
	}
}

func printf(level int, format string, v ...interface{}) {
	if LogLevel >= level {
		fmt.Fprintf(os.Stderr, prefix[level]+format, v...)
	}
}
