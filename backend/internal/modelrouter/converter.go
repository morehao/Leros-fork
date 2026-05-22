package modelrouter

import (
	"encoding/json"
	"errors"
	"fmt"
)

var errInvalidRequestBody = errors.New("parse request body")

func convertRequest(body []byte, entryProtocol, upstreamProtocol Protocol, upstreamModel string) ([]byte, error) {
	if entryProtocol == upstreamProtocol {
		return rewriteModelName(body, upstreamModel)
	}

	var raw map[string]interface{}
	if err := json.Unmarshal(body, &raw); err != nil {
		return nil, fmt.Errorf("%w: %w", errInvalidRequestBody, err)
	}

	decoder, ok := decoders[entryProtocol]
	if !ok {
		return nil, fmt.Errorf("no decoder for protocol %s", entryProtocol)
	}

	ir, err := decoder.DecodeRequest(raw)
	if err != nil {
		return nil, fmt.Errorf("decode %s request: %w", entryProtocol, err)
	}

	ir.Model = upstreamModel

	encoder, ok := encoders[upstreamProtocol]
	if !ok {
		return nil, fmt.Errorf("no encoder for protocol %s", upstreamProtocol)
	}

	out, err := encoder.EncodeRequest(ir)
	if err != nil {
		return nil, fmt.Errorf("encode %s request: %w", upstreamProtocol, err)
	}

	return json.Marshal(out)
}

func convertResponse(body []byte, entryProtocol, upstreamProtocol Protocol) ([]byte, error) {
	if entryProtocol == upstreamProtocol {
		return body, nil
	}

	var raw map[string]interface{}
	if err := json.Unmarshal(body, &raw); err != nil {
		return nil, fmt.Errorf("parse response body: %w", err)
	}

	decoder, ok := decoders[upstreamProtocol]
	if !ok {
		return nil, fmt.Errorf("no decoder for protocol %s", upstreamProtocol)
	}

	ir, err := decoder.DecodeResponse(raw)
	if err != nil {
		return nil, fmt.Errorf("decode %s response: %w", upstreamProtocol, err)
	}

	encoder, ok := encoders[entryProtocol]
	if !ok {
		return nil, fmt.Errorf("no encoder for protocol %s", entryProtocol)
	}

	out, err := encoder.EncodeResponse(ir)
	if err != nil {
		return nil, fmt.Errorf("encode %s response: %w", entryProtocol, err)
	}

	return json.Marshal(out)
}

func rewriteModelName(body []byte, modelName string) ([]byte, error) {
	var raw map[string]interface{}
	if err := json.Unmarshal(body, &raw); err != nil {
		return nil, fmt.Errorf("%w: %w", errInvalidRequestBody, err)
	}
	raw["model"] = modelName
	return json.Marshal(raw)
}

func convertStreamEvent(data []byte, entryProtocol, upstreamProtocol Protocol) ([][]byte, error) {
	return convertStreamEventWithState(data, entryProtocol, upstreamProtocol, nil)
}

func convertStreamEventWithState(data []byte, entryProtocol, upstreamProtocol Protocol, state *streamConversionState) ([][]byte, error) {
	if entryProtocol == upstreamProtocol {
		return [][]byte{data}, nil
	}

	var raw map[string]interface{}
	if err := json.Unmarshal(data, &raw); err != nil {
		return nil, fmt.Errorf("parse stream event: %w", err)
	}

	var irEvents []*IRStreamEvent
	switch upstreamProtocol {
	case ProtocolOpenAIChat:
		irEvents = decodeOpenAIChatStreamEvent(raw)
	case ProtocolOpenAIResponses:
		irEvents = decodeResponsesStreamEvent(raw)
	case ProtocolAnthropicMessages:
		eventType := getString(raw, "type")
		irEvents = decodeAnthropicStreamEvent(eventType, raw)
	}

	if len(irEvents) == 0 {
		return nil, nil
	}
	if state != nil {
		irEvents = state.prepareIRStreamEvents(entryProtocol, irEvents)
	}

	var result [][]byte
	for _, irEvent := range irEvents {
		var encodedEvents []map[string]interface{}

		switch entryProtocol {
		case ProtocolOpenAIChat:
			encodedEvents = encodeOpenAIChatStreamEvent(irEvent)
		case ProtocolOpenAIResponses:
			encodedEvents = encodeResponsesStreamEventWithState(irEvent, state)
		case ProtocolAnthropicMessages:
			encodedEvents = encodeAnthropicStreamEvent(irEvent)
		}

		for _, evt := range encodedEvents {
			if evt == nil {
				continue
			}
			b, err := json.Marshal(evt)
			if err != nil {
				continue
			}
			result = append(result, b)
		}
	}

	return result, nil
}

type streamConversionState struct {
	responses responsesStreamState
}

func newStreamConversionState() *streamConversionState {
	return &streamConversionState{}
}

func (s *streamConversionState) prepareIRStreamEvents(entryProtocol Protocol, events []*IRStreamEvent) []*IRStreamEvent {
	if s == nil || entryProtocol != ProtocolOpenAIResponses {
		return events
	}
	var prepared []*IRStreamEvent
	for _, event := range events {
		if event == nil {
			continue
		}
		if event.Type == IRStreamContentDelta && event.DeltaType == "text" && !s.responses.hasStartedText(event.Index) {
			prepared = append(prepared, &IRStreamEvent{Type: IRStreamContentStart, Index: event.Index})
		}
		if event.Type == IRStreamMessageDelta {
			if s.responses.hasStartedText(0) && !s.responses.hasStoppedText(0) {
				prepared = append(prepared, &IRStreamEvent{Type: IRStreamContentStop, Index: 0})
			}
			s.responses.setMessageDelta(event)
		}
		prepared = append(prepared, event)
	}
	return prepared
}
