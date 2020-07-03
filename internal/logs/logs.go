// Package logs provides primitive logger with filtering level
package logs

import (
	"io"
	"log"
	"strings"
)

type Logging interface {
	Printf(string, ...interface{})
	Tracef(string, ...interface{})
	Debugf(string, ...interface{})
	Infof(string, ...interface{})
	Noticef(string, ...interface{})
	Warnf(string, ...interface{})
	Errorf(string, ...interface{})
}

type Logger struct {
	log   *log.Logger
	Level Level
}

func New(out io.Writer, lv Level, prop int) *Logger {
	return &Logger{log: log.New(out, "", prop), Level: lv}
}

func (self *Logger) Printf(fmt string, v ...interface{}) {
	self.log.Printf(fmt, v...)
}

func (self *Logger) writef(lv Level, fmt string, v ...interface{}) {
	if lv >= self.Level {
		self.log.Printf(fmt, v...)
	}
}

func (self *Logger) Tracef(fmt string, v ...interface{}) {
	fmt = strings.Join([]string{"[TRACE]", fmt}, " ")
	self.writef(Trace, fmt, v...)
}

func (self *Logger) Debugf(fmt string, v ...interface{}) {
	fmt = strings.Join([]string{"[DEBUG]", fmt}, " ")
	self.writef(Debug, fmt, v...)
}

func (self *Logger) Infof(fmt string, v ...interface{}) {
	fmt = strings.Join([]string{"[INFO]", fmt}, " ")
	self.writef(Info, fmt, v...)
}

func (self *Logger) Noticef(fmt string, v ...interface{}) {
	fmt = strings.Join([]string{"[NOTICE]", fmt}, " ")
	self.writef(Notice, fmt, v...)
}

func (self *Logger) Warnf(fmt string, v ...interface{}) {
	fmt = strings.Join([]string{"[WARN]", fmt}, " ")
	self.writef(Warning, fmt, v...)
}

func (self *Logger) Errorf(fmt string, v ...interface{}) {
	fmt = strings.Join([]string{"[ERROR]", fmt}, " ")
	self.writef(Error, fmt, v...)
}

func (self *Logger) Fatalf(fmt string, v ...interface{}) {
	self.log.Fatalf(fmt, v...)
}
