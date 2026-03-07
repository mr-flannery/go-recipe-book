package logging

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"sync"
	"time"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

type contextKey string

const wideEventKey contextKey = "wideEvent"

type WideEvent struct {
	mu     sync.Mutex
	fields map[string]any
}

func NewWideEvent() *WideEvent {
	return &WideEvent{
		fields: make(map[string]any),
	}
}

func (w *WideEvent) Set(key string, value any) {
	w.mu.Lock()
	defer w.mu.Unlock()
	w.fields[key] = value
}

func (w *WideEvent) SetMany(fields map[string]any) {
	w.mu.Lock()
	defer w.mu.Unlock()
	for k, v := range fields {
		w.fields[k] = v
	}
}

func (w *WideEvent) Get(key string) (any, bool) {
	w.mu.Lock()
	defer w.mu.Unlock()
	v, ok := w.fields[key]
	return v, ok
}

func (w *WideEvent) Fields() map[string]any {
	w.mu.Lock()
	defer w.mu.Unlock()
	copy := make(map[string]any, len(w.fields))
	for k, v := range w.fields {
		copy[k] = v
	}
	return copy
}

func ContextWithWideEvent(ctx context.Context, event *WideEvent) context.Context {
	return context.WithValue(ctx, wideEventKey, event)
}

func WideEventFromContext(ctx context.Context) *WideEvent {
	event, ok := ctx.Value(wideEventKey).(*WideEvent)
	if !ok {
		return nil
	}
	return event
}

func AddToSpan(ctx context.Context, key string, value any) {
	span := trace.SpanFromContext(ctx)
	if span == nil || !span.IsRecording() {
		return
	}

	attr := toOtelAttribute(key, value)
	if attr.Valid() {
		span.SetAttributes(attr)
	}
}

func Add(ctx context.Context, key string, value any) {
	if event := WideEventFromContext(ctx); event != nil {
		event.Set(key, value)
	}
	AddToSpan(ctx, key, value)
}

func AddMany(ctx context.Context, fields map[string]any) {
	if event := WideEventFromContext(ctx); event != nil {
		event.SetMany(fields)
	}

	span := trace.SpanFromContext(ctx)
	if span == nil || !span.IsRecording() {
		return
	}

	attrs := make([]attribute.KeyValue, 0, len(fields))
	for k, v := range fields {
		attr := toOtelAttribute(k, v)
		if attr.Valid() {
			attrs = append(attrs, attr)
		}
	}
	span.SetAttributes(attrs...)
}

func AddError(ctx context.Context, err error, message string) {
	if err == nil {
		return
	}

	fields := map[string]any{
		"error":         true,
		"error.message": message,
		"error.detail":  err.Error(),
		"error.type":    fmt.Sprintf("%T", err),
	}

	if event := WideEventFromContext(ctx); event != nil {
		event.SetMany(fields)
	}

	span := trace.SpanFromContext(ctx)
	if span != nil && span.IsRecording() {
		span.RecordError(err)
		span.SetAttributes(
			attribute.Bool("error", true),
			attribute.String("error.message", message),
			attribute.String("error.detail", err.Error()),
			attribute.String("error.type", fmt.Sprintf("%T", err)),
		)
	}
}

func Emit(ctx context.Context) {
	event := WideEventFromContext(ctx)
	if event == nil {
		return
	}

	fields := event.Fields()
	if len(fields) == 0 {
		return
	}

	fields["timestamp"] = time.Now().UTC().Format(time.RFC3339Nano)

	args := make([]any, 0, len(fields)*2)
	for k, v := range fields {
		args = append(args, k, v)
	}

	outcome, _ := fields["outcome"].(string)
	if outcome == "error" {
		slog.Error("request completed", args...)
	} else {
		slog.Info("request completed", args...)
	}
}

func toOtelAttribute(key string, value any) attribute.KeyValue {
	switch v := value.(type) {
	case string:
		return attribute.String(key, v)
	case int:
		return attribute.Int(key, v)
	case int64:
		return attribute.Int64(key, v)
	case float64:
		return attribute.Float64(key, v)
	case bool:
		return attribute.Bool(key, v)
	case []string:
		return attribute.StringSlice(key, v)
	case []int:
		return attribute.IntSlice(key, v)
	case []int64:
		return attribute.Int64Slice(key, v)
	case []float64:
		return attribute.Float64Slice(key, v)
	case []bool:
		return attribute.BoolSlice(key, v)
	case time.Duration:
		return attribute.Int64(key, v.Milliseconds())
	default:
		return attribute.KeyValue{}
	}
}

var envContext map[string]any
var envContextOnce sync.Once

// TODO: I'm not sure if this should be in the logging package, if it's only called when initiating a new wide event
func GetEnvContext() map[string]any {
	envContextOnce.Do(func() {
		envContext = map[string]any{
			"service.name":    getEnvOrDefault("OTEL_SERVICE_NAME", "recipe-book"),
			"service.version": getEnvOrDefault("RAILWAY_GIT_COMMIT_SHA", getEnvOrDefault("GIT_COMMIT", "unknown")),
			"environment":     getEnvOrDefault("RAILWAY_ENVIRONMENT", getEnvOrDefault("ENVIRONMENT", "development")),
		}

		if region := os.Getenv("RAILWAY_REGION"); region != "" {
			envContext["region"] = region
		}
		if deployID := os.Getenv("RAILWAY_DEPLOYMENT_ID"); deployID != "" {
			envContext["deployment.id"] = deployID
		}
	})
	return envContext
}

func getEnvOrDefault(key, defaultValue string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return defaultValue
}
