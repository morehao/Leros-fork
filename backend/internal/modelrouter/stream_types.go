package modelrouter

// IRStreamEventType stream event type.
type IRStreamEventType string

const (
	IRStreamMessageStart IRStreamEventType = "message_start"
	IRStreamContentStart IRStreamEventType = "content_block_start"
	IRStreamContentDelta IRStreamEventType = "content_block_delta"
	IRStreamContentStop  IRStreamEventType = "content_block_stop"
	IRStreamMessageDelta IRStreamEventType = "message_delta"
	IRStreamDone         IRStreamEventType = "done"
	IRStreamError        IRStreamEventType = "error"
)

// IRStreamEvent unified stream event.
type IRStreamEvent struct {
	Type IRStreamEventType

	Index int

	ResponseID    string
	ResponseModel string
	ContentBlock  *IRContentBlock

	DeltaType string
	DeltaText string
	DeltaJSON string

	StopReason IRStopReason
	Usage      *IRUsage

	ErrorMessage string
}
