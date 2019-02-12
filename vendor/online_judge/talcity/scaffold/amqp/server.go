package amqp

import "fmt"

type Server struct {
	broker *Broker
}

func NewServer(cnf *AMQPConfig) *Server {
	broker := New(cnf)

	return &Server{
		broker: broker,
	}
}

// GetBroker return broker
func (server *Server) GetBroker() *Broker {
	return server.broker
}

// Broker sets broker
func (server *Server) Broker(broker *Broker) {
	server.broker = broker
}

// GetConfig
func (server *Server) GetConfig() *AMQPConfig {
	return server.broker.cnf
}

// Config sets broker config
func (server *Server) Config(cnf *AMQPConfig) {
	server.broker.cnf = cnf
}

// SendMessage publishes a message to the queue.
func (server *Server) SendMessage(message *Message) error {
	if err := server.broker.Publish(message.RoutingKey, message.Priority, message.Body, message.Delay); err != nil {
		return fmt.Errorf("Publish message error: %s ", err)
	}

	return nil
}

// NewWorker creates worker process.
func (server *Server) NewWorker(consumerTag string, concurrency int) *Worker {
	return server.NewCustomQueueWorker(consumerTag, concurrency, "", "")
}

// NewCustomQueueWorker creates worker process with custom queue.
func (server *Server) NewCustomQueueWorker(consumerTag string, concurrency int, bindingKey, queue string) *Worker {
	return &Worker{
		server:         server,
		ConsumerTag:    consumerTag,
		Concurrency:    concurrency,
		BindingKey:     bindingKey,
		ConsumingQueue: queue,
	}
}
