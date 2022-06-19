package trace

import (
	"context"
	"go.opentelemetry.io/otel/trace"
)

var stubEmptyTracer = &emptyTracer{}
var stubEmptySpan = &emptySpan{}

type emptySpan struct {
	trace.Span
}

func (*emptySpan) End(...trace.SpanEndOption) {
}

type emptyTracer struct {
}

func (*emptyTracer) Start(ctx context.Context, _ string, _ ...trace.SpanStartOption) (context.Context, trace.Span) {
	return ctx, stubEmptySpan
}
