package logger

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"runtime"
	"strings"
	"sync/atomic"
	"time"
)

var (
	// BUILD_PWD=$(shell pwd)
	// go build -ldflags "-X github.com/hylent/sf/logger.buildPwd=${BUILD_PWD}"
	buildPwd string

	currentLevel   = LevelOfDebug
	currentPrinter = defaultLogPrinter()
)

const (
	LevelOfDebug int32 = iota
	LevelOfInfo
	LevelOfWarn
	LevelOfFatal
	LevelOfNothing
)

func getLevelStr(level int32) string {
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

func SetCurrentLevel(level int32) {
	if level >= LevelOfDebug && level <= LevelOfNothing {
		atomic.StoreInt32(&currentLevel, level)
	}
}

type LogPrinter func(data map[string]interface{})

// SetCurrentPrinter sets log printer (NTS)
func SetCurrentPrinter(printer LogPrinter) {
	currentPrinter = printer
}

func defaultLogPrinter() LogPrinter {
	printer := log.New(os.Stdout, "", 0).Print
	return func(data map[string]interface{}) {
		if ba, baErr := json.Marshal(data); baErr != nil {
			printer(fmt.Sprintf(
				"JsonEncodeFailureFallback: t=%s e=%v d=%+v",
				time.Now().Format(time.RFC3339Nano),
				baErr,
				data,
			))
		} else {
			printer(string(ba))
		}
	}
}

type LogEntryCollection interface {
	ForEachLogEntry(f func(k string, v interface{}))
}

type M map[string]interface{}

func (x M) ForEachLogEntry(f func(k string, v interface{})) {
	for k, v := range x {
		f(k, v)
	}
}

func Debug(category string, collections ...LogEntryCollection) {
	if atomic.LoadInt32(&currentLevel) <= LevelOfDebug {
		doLog(LevelOfDebug, category, collections...)
	}
}

func Info(category string, collections ...LogEntryCollection) {
	if atomic.LoadInt32(&currentLevel) <= LevelOfInfo {
		doLog(LevelOfInfo, category, collections...)
	}
}

func Warn(category string, collections ...LogEntryCollection) {
	if atomic.LoadInt32(&currentLevel) <= LevelOfWarn {
		doLog(LevelOfWarn, category, collections...)
	}
}

func Fatal(category string, collections ...LogEntryCollection) {
	if atomic.LoadInt32(&currentLevel) <= LevelOfFatal {
		doLog(LevelOfFatal, category, collections...)
	}
	os.Exit(1)
}

func doLog(level int32, category string, collections ...LogEntryCollection) {
	now := time.Now()
	var fileStr string
	if _, file, line, fileLineOk := runtime.Caller(2); fileLineOk {
		buildPwdLen := len(buildPwd)
		if buildPwdLen > 0 && strings.HasPrefix(file, buildPwd) {
			file = file[buildPwdLen:]
		}
		fileStr = fmt.Sprintf("%s:%d", file, line)
	}

	data := map[string]interface{}{
		"_level":    getLevelStr(level),
		"_category": category,
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

	currentPrinter(data)
}
