package logs

import (
	"io"
	"log"
	"strings"
)

var defaultLogger *Logger

func Configure(out io.Writer, lv Level, prop int) {
	defaultLogger = &Logger{log: log.New(out, "", prop), Level: lv}
}

func GetLevel() Level {
	return defaultLogger.Level
}

func SetLevel(lv Level) {
	defaultLogger.Level = lv
}

func Printf(fmt string, v ...interface{}) {
	defaultLogger.Printf(fmt, v...)
}

func Tracef(fmt string, v ...interface{}) {
	fmt = strings.Join([]string{"[TRACE]", fmt}, " ")
	defaultLogger.writef(Trace, fmt, v...)
}

func Debugf(fmt string, v ...interface{}) {
	fmt = strings.Join([]string{"[DEBUG]", fmt}, " ")
	defaultLogger.writef(Debug, fmt, v...)
}

func Infof(fmt string, v ...interface{}) {
	fmt = strings.Join([]string{"[INFO]", fmt}, " ")
	defaultLogger.writef(Info, fmt, v...)
}

func Noticef(fmt string, v ...interface{}) {
	fmt = strings.Join([]string{"[NOTICE]", fmt}, " ")
	defaultLogger.writef(Notice, fmt, v...)
}

func Warnf(fmt string, v ...interface{}) {
	fmt = strings.Join([]string{"[WARN]", fmt}, " ")
	defaultLogger.writef(Warning, fmt, v...)
}

func Errorf(fmt string, v ...interface{}) {
	fmt = strings.Join([]string{"[ERROR]", fmt}, " ")
	defaultLogger.writef(Error, fmt, v...)
}

func Fatalf(fmt string, v ...interface{}) {
	defaultLogger.Fatalf(fmt, v...)
}
