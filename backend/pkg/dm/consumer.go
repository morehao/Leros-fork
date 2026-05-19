package dm

import "fmt"

// OrchestratorConsumer 构造 orchestrator 的持久化消费者名称。
func OrchestratorConsumer(topic string) string {
	return fmt.Sprintf("orchestrator-%s", topic)
}
