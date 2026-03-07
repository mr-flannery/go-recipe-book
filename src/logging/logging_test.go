package logging

import (
	"context"
	"fmt"
	"testing"
)

func TestWideEvent_SetAndGet(t *testing.T) {
	event := NewWideEvent()

	event.Set("key1", "value1")
	event.Set("key2", 42)

	val, ok := event.Get("key1")
	if !ok {
		t.Error("expected key1 to exist")
	}
	if val != "value1" {
		t.Errorf("expected value1, got %v", val)
	}

	val, ok = event.Get("key2")
	if !ok {
		t.Error("expected key2 to exist")
	}
	if val != 42 {
		t.Errorf("expected 42, got %v", val)
	}

	_, ok = event.Get("nonexistent")
	if ok {
		t.Error("expected nonexistent key to not exist")
	}
}

func TestWideEvent_SetMany(t *testing.T) {
	event := NewWideEvent()

	event.SetMany(map[string]any{
		"a": 1,
		"b": "two",
		"c": true,
	})

	fields := event.Fields()
	if len(fields) != 3 {
		t.Errorf("expected 3 fields, got %d", len(fields))
	}
	if fields["a"] != 1 || fields["b"] != "two" || fields["c"] != true {
		t.Error("fields not set correctly")
	}
}

func TestContextWithWideEvent(t *testing.T) {
	event := NewWideEvent()
	event.Set("test", "value")

	ctx := ContextWithWideEvent(context.Background(), event)
	retrieved := WideEventFromContext(ctx)

	if retrieved == nil {
		t.Fatal("expected to retrieve wide event from context")
	}

	val, ok := retrieved.Get("test")
	if !ok || val != "value" {
		t.Error("expected to retrieve correct value from context event")
	}
}

func TestWideEventFromContext_ReturnsNilWhenMissing(t *testing.T) {
	ctx := context.Background()
	event := WideEventFromContext(ctx)

	if event != nil {
		t.Error("expected nil when no wide event in context")
	}
}

func TestAdd_UpdatesBothEventAndSpan(t *testing.T) {
	event := NewWideEvent()
	ctx := ContextWithWideEvent(context.Background(), event)

	Add(ctx, "test.key", "test.value")

	val, ok := event.Get("test.key")
	if !ok || val != "test.value" {
		t.Error("expected Add to update wide event")
	}
}

func TestAddMany_UpdatesBothEventAndSpan(t *testing.T) {
	event := NewWideEvent()
	ctx := ContextWithWideEvent(context.Background(), event)

	AddMany(ctx, map[string]any{
		"key1": "value1",
		"key2": 123,
	})

	fields := event.Fields()
	if fields["key1"] != "value1" || fields["key2"] != 123 {
		t.Error("expected AddMany to update wide event")
	}
}

func TestFields_ReturnsCopy(t *testing.T) {
	event := NewWideEvent()
	event.Set("original", "value")

	fields := event.Fields()
	fields["original"] = "modified"

	val, _ := event.Get("original")
	if val != "value" {
		t.Error("Fields should return a copy, not modify original")
	}
}

func TestAddError_AddsErrorFieldsToWideEvent(t *testing.T) {
	event := NewWideEvent()
	ctx := ContextWithWideEvent(context.Background(), event)

	testErr := fmt.Errorf("something went wrong")
	AddError(ctx, testErr, "Failed to save recipe")

	fields := event.Fields()

	if fields["error"] != true {
		t.Error("expected error=true")
	}
	if fields["error.message"] != "Failed to save recipe" {
		t.Errorf("expected error.message='Failed to save recipe', got %v", fields["error.message"])
	}
	if fields["error.detail"] != "something went wrong" {
		t.Errorf("expected error.detail='something went wrong', got %v", fields["error.detail"])
	}
	if fields["error.type"] != "*errors.errorString" {
		t.Errorf("expected error.type='*errors.errorString', got %v", fields["error.type"])
	}
}

func TestAddError_DoesNothingWhenErrIsNil(t *testing.T) {
	event := NewWideEvent()
	ctx := ContextWithWideEvent(context.Background(), event)

	AddError(ctx, nil, "This should not appear")

	fields := event.Fields()
	if len(fields) != 0 {
		t.Error("expected no fields when err is nil")
	}
}
