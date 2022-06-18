package logger

import (
	"encoding/json"
	"fmt"
	"os"
	"runtime"
	"strings"
	"time"
)

const (
	TRACE = "TRACE"
	INFO  = "INFO"
	WARN  = "WARN"
	ERROR = "ERROR"
	FATAL = "FATAL"
)

var logLevel = os.Getenv("LOG_LEVEL")

func Tracef(format string, v ...interface{}) {
	if logLevel == TRACE {
		Printf(TRACE, format, v...)
	}
}
func Trace(message string) {
	if logLevel == TRACE {
		Printf(TRACE, message)
	}
}

func Infof(format string, v ...interface{}) {
	Printf(INFO, format, v...)
}
func Info(message string) {
	Printf(INFO, message)
}

func Warnf(format string, v ...interface{}) {
	Printf(WARN, format, v...)
}
func Warn(message string) {
	Printf(WARN, message)
}

func Errorf(format string, v ...interface{}) {
	Printf(ERROR, format, v...)
}
func Error(message string) {
	Printf(ERROR, message)
}

func Fatalf(format string, v ...interface{}) {
	Printf(FATAL, format, v...)

	panic(fmt.Errorf(format, v))
}

func Fatal(message string) {
	Printf(FATAL, message)

	panic(fmt.Errorf(message))
}

func Printf(level string, format string, v ...interface{}) {
	event := &LogEvent{
		Source:    getSource(2),
		Level:     level,
		Timestamp: time.Now().Format(time.RFC3339),
		Message:   fmt.Sprintf(format, v...),
		Hash:      calculateHash(format),
	}

	jsonText, err := json.Marshal(event)
	if err != nil {
		panic(err)
	}

	dest := os.Stdout

	if level == ERROR || level == FATAL {
		dest = os.Stderr
	}

	jsonText = append(jsonText, '\n')

	_, err = dest.Write(jsonText)
	if err != nil {
		panic(err)
	}

	if level == FATAL {
		os.Exit(1)
	}
}

func getSource(callDepth int) string {
	_, file, line, ok := runtime.Caller(callDepth + 1)
	if !ok {
		file = "???"
		line = 0
	}

	split := strings.Split(file, "/")
	file = split[len(split)-1]
	return fmt.Sprintf("%s:%d", file, line)
}

func calculateHash(read string) string {
	var hashedValue uint64 = 3074457345618258791
	for _, char := range read {
		hashedValue += uint64(char)
		hashedValue *= 3074457345618258799
	}

	return strings.ToUpper(fmt.Sprintf("%x", hashedValue))
}
