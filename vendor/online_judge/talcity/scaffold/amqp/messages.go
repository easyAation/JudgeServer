package amqp

import "time"

// MessageProcessor can process a delivered message.
// probable always be a worker instance.
type MessageProcessor interface {
	Process([]byte) error
	CustomQueue() string
}

// Message a single message.
type Message struct {
	RoutingKey string
	Priority   uint8
	Body       []byte
	Delay      *time.Time
}
