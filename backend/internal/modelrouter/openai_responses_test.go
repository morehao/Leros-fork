package modelrouter

import (
	"encoding/json"
	"testing"
)

func TestDecodeResponsesInputAcceptsEasyMessages(t *testing.T) {
	decoder := &openAIResponsesDecoder{}
	ir, err := decoder.DecodeRequest(map[string]interface{}{
		"model": "gpt-5",
		"input": []interface{}{
			map[string]interface{}{
				"role":    "developer",
				"content": "Talk like a pirate.",
			},
			map[string]interface{}{
				"role":    "user",
				"content": "Are semicolons optional in JavaScript?",
			},
		},
	})
	if err != nil {
		t.Fatalf("DecodeRequest() error = %v", err)
	}
	if len(ir.Messages) != 2 {
		t.Fatalf("len(Messages) = %d, want 2", len(ir.Messages))
	}
	if ir.Messages[0].Role != IRRoleSystem {
		t.Fatalf("Messages[0].Role = %q, want %q", ir.Messages[0].Role, IRRoleSystem)
	}
	if got := ir.Messages[0].getTextContent(); got != "Talk like a pirate." {
		t.Fatalf("Messages[0] text = %q", got)
	}
	if ir.Messages[1].Role != IRRoleUser {
		t.Fatalf("Messages[1].Role = %q, want %q", ir.Messages[1].Role, IRRoleUser)
	}
	if got := ir.Messages[1].getTextContent(); got != "Are semicolons optional in JavaScript?" {
		t.Fatalf("Messages[1] text = %q", got)
	}
}

func TestEncodeResponsesRequestSkipsSystemMessagesWhenInstructionsAreSet(t *testing.T) {
	input := []byte(`{
		"model": "alias",
		"messages": [
			{"role": "system", "content": "Talk like a pirate."},
			{"role": "user", "content": "Are semicolons optional in JavaScript?"}
		]
	}`)

	converted, err := convertRequest(input, ProtocolOpenAIChat, ProtocolOpenAIResponses, "gpt-5")
	if err != nil {
		t.Fatalf("convertRequest() error = %v", err)
	}

	var body map[string]interface{}
	if err := json.Unmarshal(converted, &body); err != nil {
		t.Fatalf("json.Unmarshal() error = %v", err)
	}

	if got := body["instructions"]; got != "Talk like a pirate." {
		t.Fatalf("instructions = %v, want system message text", got)
	}

	items, ok := body["input"].([]interface{})
	if !ok {
		t.Fatalf("input = %T, want []interface{}", body["input"])
	}
	if len(items) != 1 {
		t.Fatalf("len(input) = %d, want 1", len(items))
	}

	msg, ok := items[0].(map[string]interface{})
	if !ok {
		t.Fatalf("input[0] = %T, want map[string]interface{}", items[0])
	}
	if got := msg["role"]; got != "user" {
		t.Fatalf("input[0].role = %v, want user", got)
	}

	content, ok := msg["content"].([]interface{})
	if !ok || len(content) != 1 {
		t.Fatalf("input[0].content = %#v, want one content block", msg["content"])
	}
	block, ok := content[0].(map[string]interface{})
	if !ok {
		t.Fatalf("input[0].content[0] = %T, want map[string]interface{}", content[0])
	}
	if got := block["type"]; got != "input_text" {
		t.Fatalf("content block type = %v, want input_text", got)
	}
}

func TestEncodeResponsesRequestUsesSystemRoleInputText(t *testing.T) {
	encoder := &openAIResponsesEncoder{}
	body, err := encoder.EncodeRequest(&IRRequest{
		Model: "gpt-5",
		Messages: []IRMessage{
			{
				Role:    IRRoleSystem,
				Content: []IRContentBlock{{Type: IRBlockText, Text: "Talk like a pirate."}},
			},
			{
				Role:    IRRoleUser,
				Content: []IRContentBlock{{Type: IRBlockText, Text: "Hello"}},
			},
		},
	})
	if err != nil {
		t.Fatalf("EncodeRequest() error = %v", err)
	}

	items, ok := body["input"].([]map[string]interface{})
	if !ok {
		t.Fatalf("input = %T, want []map[string]interface{}", body["input"])
	}
	if len(items) != 2 {
		t.Fatalf("len(input) = %d, want 2", len(items))
	}
	if got := items[0]["role"]; got != "system" {
		t.Fatalf("input[0].role = %v, want system", got)
	}
	content, ok := items[0]["content"].([]map[string]interface{})
	if !ok || len(content) != 1 {
		t.Fatalf("input[0].content = %#v, want one content block", items[0]["content"])
	}
	if got := content[0]["type"]; got != "input_text" {
		t.Fatalf("content block type = %v, want input_text", got)
	}
}

func TestConvertChatStreamToResponsesStartsTextItemBeforeDelta(t *testing.T) {
	state := newStreamConversionState()
	start := []byte(`{"id":"chatcmpl-1","object":"chat.completion.chunk","created":1,"model":"gpt-test","choices":[{"index":0,"delta":{"role":"assistant"},"finish_reason":null}]}`)
	if _, err := convertStreamEventWithState(start, ProtocolOpenAIResponses, ProtocolOpenAIChat, state); err != nil {
		t.Fatalf("convert start event: %v", err)
	}

	delta := []byte(`{"id":"chatcmpl-1","object":"chat.completion.chunk","created":1,"model":"gpt-test","choices":[{"index":0,"delta":{"content":"hello"},"finish_reason":null}]}`)
	converted, err := convertStreamEventWithState(delta, ProtocolOpenAIResponses, ProtocolOpenAIChat, state)
	if err != nil {
		t.Fatalf("convert delta event: %v", err)
	}
	if len(converted) != 3 {
		t.Fatalf("len(converted) = %d, want output item, content part, delta", len(converted))
	}

	var types []string
	for _, data := range converted {
		var event map[string]interface{}
		if err := json.Unmarshal(data, &event); err != nil {
			t.Fatalf("unmarshal converted event: %v", err)
		}
		types = append(types, event["type"].(string))
	}
	want := []string{"response.output_item.added", "response.content_part.added", "response.output_text.delta"}
	for i := range want {
		if types[i] != want[i] {
			t.Fatalf("event type[%d] = %q, want %q (all=%v)", i, types[i], want[i], types)
		}
	}
}

func TestFormatSSEUsesConvertedResponsesEventType(t *testing.T) {
	data := []byte(`{"type":"response.output_item.added","output_index":0,"item":{"type":"message","role":"assistant","content":[]}}`)
	formatted := string(formatSSE(ProtocolOpenAIResponses, convertedEventType("ignored.upstream", data), data))
	if want := "event: response.output_item.added\n"; len(formatted) < len(want) || formatted[:len(want)] != want {
		t.Fatalf("formatted SSE = %q, want prefix %q", formatted, want)
	}
}

func TestConvertChatStreamToResponsesClosesTextItemOnFinish(t *testing.T) {
	state := newStreamConversionState()
	events := [][]byte{
		[]byte(`{"id":"chatcmpl-1","object":"chat.completion.chunk","created":1,"model":"gpt-test","choices":[{"index":0,"delta":{"role":"assistant"},"finish_reason":null}]}`),
		[]byte(`{"id":"chatcmpl-1","object":"chat.completion.chunk","created":1,"model":"gpt-test","choices":[{"index":0,"delta":{"content":"hello"},"finish_reason":null}]}`),
	}
	for _, event := range events {
		if _, err := convertStreamEventWithState(event, ProtocolOpenAIResponses, ProtocolOpenAIChat, state); err != nil {
			t.Fatalf("convert setup event: %v", err)
		}
	}

	finish := []byte(`{"id":"chatcmpl-1","object":"chat.completion.chunk","created":1,"model":"gpt-test","choices":[{"index":0,"delta":{"content":""},"finish_reason":"stop"}],"usage":{"prompt_tokens":1,"completion_tokens":1}}`)
	converted, err := convertStreamEventWithState(finish, ProtocolOpenAIResponses, ProtocolOpenAIChat, state)
	if err != nil {
		t.Fatalf("convert finish event: %v", err)
	}

	var types []string
	for _, data := range converted {
		var event map[string]interface{}
		if err := json.Unmarshal(data, &event); err != nil {
			t.Fatalf("unmarshal converted event: %v", err)
		}
		types = append(types, event["type"].(string))
	}
	want := []string{"response.output_text.done", "response.content_part.done", "response.output_item.done"}
	for i := range want {
		if types[i] != want[i] {
			t.Fatalf("event type[%d] = %q, want %q (all=%v)", i, types[i], want[i], types)
		}
	}

	done := encodeResponsesStreamEventWithState(&IRStreamEvent{Type: IRStreamDone}, state)
	if len(done) != 1 || done[0]["type"] != "response.completed" {
		t.Fatalf("done event = %#v, want response.completed", done)
	}
	resp, ok := done[0]["response"].(map[string]interface{})
	if !ok || resp["status"] != "completed" {
		t.Fatalf("response.completed payload = %#v", done[0]["response"])
	}
	if resp["id"] == "" {
		t.Fatalf("response.completed id is required, got %#v", resp)
	}
	if resp["object"] != "response" {
		t.Fatalf("response.completed object = %#v, want response", resp["object"])
	}
	usage, ok := resp["usage"].(map[string]interface{})
	if !ok || usage["input_tokens"] == nil || usage["output_tokens"] == nil || usage["total_tokens"] == nil {
		t.Fatalf("response.completed usage = %#v, want required token fields", resp["usage"])
	}
}

func TestResponsesDoneUsesResponseIDFromStreamStart(t *testing.T) {
	state := newStreamConversionState()
	start := []byte(`{"id":"chatcmpl-42","object":"chat.completion.chunk","created":1,"model":"gpt-test","choices":[{"index":0,"delta":{"role":"assistant"},"finish_reason":null}]}`)
	if _, err := convertStreamEventWithState(start, ProtocolOpenAIResponses, ProtocolOpenAIChat, state); err != nil {
		t.Fatalf("convert start event: %v", err)
	}

	done := encodeResponsesStreamEventWithState(&IRStreamEvent{Type: IRStreamDone}, state)
	resp, ok := done[0]["response"].(map[string]interface{})
	if !ok {
		t.Fatalf("response.completed payload = %#v", done[0]["response"])
	}
	if got := resp["id"]; got != "resp-chatcmpl-42" {
		t.Fatalf("response.completed id = %#v, want resp-chatcmpl-42", got)
	}
	if got := resp["model"]; got != "gpt-test" {
		t.Fatalf("response.completed model = %#v, want gpt-test", got)
	}
}
