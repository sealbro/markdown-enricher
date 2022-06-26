package logger

import (
	"context"
	"fmt"
	"go.opentelemetry.io/otel/trace"
)

func Trace(ctx context.Context, format string, v ...interface{}) {
	if logLevel == TRACE {
		printfWithCtx(ctx, TRACE, format, v...)
	}
}

func Info(ctx context.Context, format string, v ...interface{}) {
	printfWithCtx(ctx, INFO, format, v...)
}

func Warn(ctx context.Context, format string, v ...interface{}) {
	printfWithCtx(ctx, WARN, format, v...)
}

func Error(ctx context.Context, format string, v ...interface{}) {
	printfWithCtx(ctx, ERROR, format, v...)
}

func Fatal(ctx context.Context, format string, v ...interface{}) {
	printfWithCtx(ctx, FATAL, format, v...)

	panic(fmt.Errorf(format, v))
}

func printfWithCtx(ctx context.Context, level string, format string, v ...interface{}) {
	event := createEvent(level, format, v)

	traceId := trace.SpanFromContext(ctx).SpanContext().TraceID()
	if traceId.IsValid() {
		event.TraceId = traceId.String()
	}

	sendToStdout(event)
}
