package logger

import (
	"encoding/json"
	"fmt"
	"os"
	"path"
	"runtime"
	"strings"
	"sync/atomic"
	"time"
)

const (
	LevelOfDebug int32 = iota
	LevelOfInfo
	LevelOfWarn
	LevelOfFatal
)

func GetLevelStr(level int32) string {
	switch level {
	case LevelOfDebug:
		return "debug"
	case LevelOfInfo:
		return "info"
	case LevelOfWarn:
		return "warn"
	case LevelOfFatal:
		return "fatal"
	}
	return ""
}

type LogEntryCollection interface {
	ForEachLogEntry(func(string, any))
}

type M map[string]any

func (x M) ForEachLogEntry(f func(string, any)) {
	for k, v := range x {
		f(k, v)
	}
}

type Logger interface {
	SetLevel(level int32)
	Debug(category string, collections ...LogEntryCollection)
	Info(category string, collections ...LogEntryCollection)
	Warn(category string, collections ...LogEntryCollection)
	Fatal(category string, collections ...LogEntryCollection)
}

type Printer func(map[string]any)

func defaultPrinter() Printer {
	writeFunc := func(ba []byte) {
		if len(ba) == 0 || ba[len(ba)-1] != '\n' {
			ba = append(ba, '\n')
		}
		_, _ = os.Stdout.Write(ba)
	}
	return func(data map[string]any) {
		if ba, baErr := json.Marshal(data); baErr != nil {
			s := fmt.Sprintf(
				"#!LogFallback: t=%s e=%v d=%+v",
				time.Now().Format(time.RFC3339Nano),
				baErr,
				data,
			)
			writeFunc([]byte(s))
		} else {
			writeFunc(ba)
		}
	}
}

func NewLogger(printer Printer, domain string, level ...int32) Logger {
	x := new(defaultLogger)
	if printer == nil {
		printer = defaultPrinter()
	}
	x.printer = printer
	x.domain = domain
	if len(level) > 0 {
		x.SetLevel(level[0])
	}
	if _, file, _, fileLineOk := runtime.Caller(1); fileLineOk {
		x.baseDir = path.Dir(file) + "/"
	}
	return x
}

type defaultLogger struct {
	printer func(map[string]any)
	domain  string
	level   int32
	baseDir string
}

func (x *defaultLogger) SetLevel(level int32) {
	if level >= LevelOfDebug && level <= LevelOfFatal {
		atomic.StoreInt32(&x.level, level)
	}
}

func (x *defaultLogger) Debug(category string, collections ...LogEntryCollection) {
	if atomic.LoadInt32(&x.level) <= LevelOfDebug {
		x.doLog(LevelOfDebug, category, collections...)
	}
}

func (x *defaultLogger) Info(category string, collections ...LogEntryCollection) {
	if atomic.LoadInt32(&x.level) <= LevelOfInfo {
		x.doLog(LevelOfInfo, category, collections...)
	}
}

func (x *defaultLogger) Warn(category string, collections ...LogEntryCollection) {
	if atomic.LoadInt32(&x.level) <= LevelOfWarn {
		x.doLog(LevelOfWarn, category, collections...)
	}
}

func (x *defaultLogger) Fatal(category string, collections ...LogEntryCollection) {
	if atomic.LoadInt32(&x.level) <= LevelOfFatal {
		x.doLog(LevelOfFatal, category, collections...)
	}
	os.Exit(1)
}

func (x *defaultLogger) doLog(level int32, category string, collections ...LogEntryCollection) {
	now := time.Now()
	var fileStr string
	if _, file, line, fileLineOk := runtime.Caller(2); fileLineOk {
		buildPwdLen := len(x.baseDir)
		if buildPwdLen > 0 && strings.HasPrefix(file, x.baseDir) {
			file = file[buildPwdLen:]
		}
		fileStr = fmt.Sprintf("%s:%d", file, line)
	}

	data := map[string]any{
		"_domain":   x.domain,
		"_category": category,
		"_level":    GetLevelStr(level),
		"_file":     fileStr,
		"_time":     now.Format(time.RFC3339Nano),
	}
	for _, collection := range collections {
		if collection != nil {
			collection.ForEachLogEntry(func(k string, v interface{}) {
				data[k] = v
			})
		}
	}

	x.printer(data)
}
