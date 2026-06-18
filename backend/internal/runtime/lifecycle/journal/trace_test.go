package lifecyclejournal

import (
	"context"
	"strings"
	"testing"

	"github.com/insmtx/Leros/backend/internal/agent"
	"github.com/insmtx/Leros/backend/internal/runtime/events"
)

func TestRunJournalArchivesMergedMessagesAndToolEvents(t *testing.T) {
	journal := NewRunJournal(&agent.RequestContext{RunID: "run_trace"}, events.NewNoopSink())
	ctx := context.Background()

	for _, event := range []*events.Event{
		{Type: events.EventStarted, RunID: "run_trace", Seq: 1},
		events.NewMessageDelta("msg_1", "hel"),
		events.NewMessageDelta("msg_1", "lo"),
		events.NewReasoningDelta("msg_1", "think "),
		events.NewReasoningDelta("msg_1", "more"),
		events.NewToolCallStarted("call_1", "memory", map[string]any{"query": "x"}),
		events.NewToolCallCompleted("call_1", "memory", map[string]any{"ok": true}, 12),
		{Type: events.EventResult, RunID: "run_trace", Content: "hello"},
	} {
		if err := journal.Append(ctx, event); err != nil {
			t.Fatalf("emit: %v", err)
		}
	}

	payload := journal.CompletedPayload(&agent.RunResult{
		Status:  agent.RunStatusCompleted,
		Message: "hello",
		Usage:   &events.UsagePayload{TotalTokens: 7},
	})

	if payload.Usage == nil || payload.Usage.TotalTokens != 7 {
		t.Fatalf("expected usage in completed payload, got %#v", payload.Usage)
	}
	if !containsArchivedEvent(payload.Events, events.EventToolCallStarted) ||
		!containsArchivedEvent(payload.Events, events.EventToolCallCompleted) {
		t.Fatalf("expected tool events in archive timeline, got %#v", payload.Events)
	}
	if containsArchivedEvent(payload.Events, events.EventResult) {
		t.Fatalf("completed archive should not include result event: %#v", payload.Events)
	}

	message := findArchivedEvent(payload.Events, events.EventMessageDelta)
	if message == nil || contentFromEventRecord(*message) != "hello" || message.LastSeq <= message.Seq {
		t.Fatalf("expected merged message delta, got %#v", message)
	}
	reasoning := findArchivedEvent(payload.Events, events.EventReasoningDelta)
	if reasoning == nil || contentFromEventRecord(*reasoning) != "think more" || reasoning.LastSeq <= reasoning.Seq {
		t.Fatalf("expected merged reasoning delta, got %#v", reasoning)
	}
}

func TestRunJournalKeepsInterleavedMessageAndToolOrder(t *testing.T) {
	journal := NewRunJournal(&agent.RequestContext{RunID: "run_trace"}, events.NewNoopSink())
	ctx := context.Background()

	for _, event := range []*events.Event{
		{Type: events.EventStarted, RunID: "run_trace", Seq: 1},
		events.NewMessageDelta("msg_1", "first "),
		events.NewToolCallStarted("call_1", "search", map[string]any{"query": "x"}),
		events.NewToolCallCompleted("call_1", "search", map[string]any{"ok": true}, 12),
		events.NewMessageDelta("msg_1", "second"),
		{Type: events.EventResult, RunID: "run_trace", Content: "first second"},
	} {
		if err := journal.Append(ctx, event); err != nil {
			t.Fatalf("emit: %v", err)
		}
	}

	payload := journal.CompletedPayload(&agent.RunResult{
		Status:  agent.RunStatusCompleted,
		Message: "first second",
	})

	if len(payload.Events) != 4 {
		t.Fatalf("expected 4 archived events, got %#v", payload.Events)
	}
	if payload.Events[0].Type != events.EventMessageDelta || contentFromEventRecord(payload.Events[0]) != "first " {
		t.Fatalf("expected first message delta to stay before tool call, got %#v", payload.Events[0])
	}
	if payload.Events[1].Type != events.EventToolCallStarted {
		t.Fatalf("expected tool call started at index 1, got %#v", payload.Events[1])
	}
	if payload.Events[2].Type != events.EventToolCallCompleted {
		t.Fatalf("expected tool call completed at index 2, got %#v", payload.Events[2])
	}
	if payload.Events[3].Type != events.EventMessageDelta || contentFromEventRecord(payload.Events[3]) != "second" {
		t.Fatalf("expected trailing message delta to remain after tool call, got %#v", payload.Events[3])
	}
}

func containsArchivedEvent(records []events.RunEventRecord, eventType events.EventType) bool {
	return findArchivedEvent(records, eventType) != nil
}

func findArchivedEvent(records []events.RunEventRecord, eventType events.EventType) *events.RunEventRecord {
	for i := range records {
		if records[i].Type == eventType {
			return &records[i]
		}
	}
	return nil
}

func contentFromEventRecord(record events.RunEventRecord) string {
	switch record.Type {
	case events.EventMessageDelta, events.EventReasoningDelta:
		payload, err := events.DecodePayload[events.MessageDeltaPayload](&events.Event{Type: record.Type, Payload: record.Payload})
		if err == nil {
			return payload.Content
		}
	case events.EventResult:
		payload, err := events.DecodePayload[events.RunResultPayload](&events.Event{Type: record.Type, Payload: record.Payload})
		if err == nil {
			return payload.Message
		}
	}
	return strings.TrimSpace(string(record.Payload))
}
