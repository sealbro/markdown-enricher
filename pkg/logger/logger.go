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
		printf(TRACE, format, v...)
	}
}

func Infof(format string, v ...interface{}) {
	printf(INFO, format, v...)
}

func Warnf(format string, v ...interface{}) {
	printf(WARN, format, v...)
}

func Errorf(format string, v ...interface{}) {
	printf(ERROR, format, v...)
}

func Fatalf(format string, v ...interface{}) {
	printf(FATAL, format, v...)

	panic(fmt.Errorf(format, v))
}

func printf(level string, format string, v ...interface{}) {
	event := createEvent(level, format, v)

	sendToStdout(event)
}

func sendToStdout(event *LogEvent) {
	jsonText, err := json.Marshal(event)
	if err != nil {
		panic(err)
	}

	dest := os.Stdout

	if event.Level == ERROR || event.Level == FATAL {
		dest = os.Stderr
	}

	jsonText = append(jsonText, '\n')

	_, err = dest.Write(jsonText)
	if err != nil {
		panic(err)
	}

	if event.Level == FATAL {
		os.Exit(1)
	}
}

func createEvent(level string, format string, v []interface{}) *LogEvent {
	return &LogEvent{
		Source:    getSource(2),
		Level:     level,
		Timestamp: time.Now().Format(time.RFC3339),
		Message:   fmt.Sprintf(format, v...),
		Hash:      calculateHash(format),
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
