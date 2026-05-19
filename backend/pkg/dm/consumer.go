package dm

import "fmt"

// WorkerTaskConsumer 构造 worker 任务消费者的持久化消费者名称。
func WorkerTaskConsumer(orgID, workerID uint) string {
	return fmt.Sprintf("worker-task-%d-%d", orgID, workerID)
}

// SessionTitleConsumer 构造会话标题处理器的持久化消费者名称。
func SessionTitleConsumer() string {
	return "session-title-handler"
}

// SessionCompletedConsumer 构造会话完成处理器的持久化消费者名称。
func SessionCompletedConsumer() string {
	return "session-completed-handler"
}
