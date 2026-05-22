package modelrouter

// IRRole unified role representation.
type IRRole string

const (
	IRRoleSystem    IRRole = "system"
	IRRoleUser      IRRole = "user"
	IRRoleAssistant IRRole = "assistant"
	IRRoleTool      IRRole = "tool"
)

// IRContentBlockType content block type.
type IRContentBlockType string

const (
	IRBlockText       IRContentBlockType = "text"
	IRBlockToolUse    IRContentBlockType = "tool_use"
	IRBlockToolResult IRContentBlockType = "tool_result"
)

// IRStopReason unified stop reason.
type IRStopReason string

const (
	IRStopEndTurn      IRStopReason = "end_turn"
	IRStopMaxTokens    IRStopReason = "max_tokens"
	IRStopStopSequence IRStopReason = "stop_sequence"
	IRStopToolUse      IRStopReason = "tool_use"
	IRStopError        IRStopReason = "error"
)

// IRContentBlock unified content block.
type IRContentBlock struct {
	Type IRContentBlockType
	Text string

	ToolUseID    string
	ToolUseName  string
	ToolUseInput map[string]interface{}

	ToolResultToolUseID string
	ToolResultContent   string
}

// IRMessage unified message.
type IRMessage struct {
	Role    IRRole
	Content []IRContentBlock
}

// IRToolDecl unified tool declaration.
type IRToolDecl struct {
	Name        string
	Description string
	Parameters  map[string]interface{}
}

// IRToolChoice unified tool choice.
type IRToolChoice struct {
	Type string
	Name string
}

// IRRequest unified request.
type IRRequest struct {
	Model    string
	Messages []IRMessage
	System   string

	Temperature *float64
	TopP        *float64
	MaxTokens   int
	Stop        []string
	Stream      bool

	Tools      []IRToolDecl
	ToolChoice *IRToolChoice
	Seed       *int
	User       string

	Instructions string
	Preserved    map[string]interface{}
}

// IRResponse unified response.
type IRResponse struct {
	ID         string
	Model      string
	Created    int64
	Content    []IRContentBlock
	StopReason IRStopReason
	Usage      *IRUsage
}

// IRUsage unified token usage.
type IRUsage struct {
	InputTokens  int
	OutputTokens int
	TotalTokens  int
}

// getTextContent returns concatenated text from all text blocks.
func (m IRMessage) getTextContent() string {
	var s string
	for _, b := range m.Content {
		if b.Type == IRBlockText {
			s += b.Text
		}
	}
	return s
}

// ProtocolDecoder decodes protocol-specific request/response to IR.
type ProtocolDecoder interface {
	DecodeRequest(body map[string]interface{}) (*IRRequest, error)
	DecodeResponse(body map[string]interface{}) (*IRResponse, error)
}

// ProtocolEncoder encodes IR to protocol-specific request/response.
type ProtocolEncoder interface {
	EncodeRequest(ir *IRRequest) (map[string]interface{}, error)
	EncodeResponse(ir *IRResponse) (map[string]interface{}, error)
}

// decoder and encoder registry
var decoders = map[Protocol]ProtocolDecoder{
	ProtocolOpenAIChat:        &openAIChatDecoder{},
	ProtocolOpenAIResponses:   &openAIResponsesDecoder{},
	ProtocolAnthropicMessages: &anthropicDecoder{},
}

var encoders = map[Protocol]ProtocolEncoder{
	ProtocolOpenAIChat:        &openAIChatEncoder{},
	ProtocolOpenAIResponses:   &openAIResponsesEncoder{},
	ProtocolAnthropicMessages: &anthropicEncoder{},
}
