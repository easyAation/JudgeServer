package amqp

import (
	"fmt"

	"github.com/streadway/amqp"
)

var (
	defalutCnf = &AMQPConfig{
		Broker:        "amqp://guest:guest@localhost:5672/",
		Exchange:      "tal_exchange",
		ExchangeType:  "direct",
		BindingKey:    "tal_msg",
		PrefetchCount: 3,
	}
)

// QueueBindingArgs optional arguments which are used when binding to the exchange.
type QueueBindingArgs map[string]interface{}

// AMQPConfig wraps RabbitMQ related configurations.
type AMQPConfig struct {
	Broker           string           `json:"broker" toml:"broker" yaml:"broker"`
	Exchange         string           `json:"exchange" toml:"exchange" yaml:"exchange"`
	ExchangeType     string           `json:"exchange_type" toml:"exchangeType" yaml:"exchangeType"`
	QueueBindingArgs QueueBindingArgs `json:"queue_binding_args" toml:"queueBindingArgs" yaml:"queueBindingArgs"`
	BindingKey       string           `json:"binding_key" toml:"bindingKey" yaml:"bindingKey"`
	PrefetchCount    int              `json:"prefetch_count" toml:"prefetchCount" yaml:"prefetchCount"`
}

// AMQPConnector
type AMQPConnector struct{}

// Connect opens a connection to RabbitMQ.
// declares an exchange, opens a channel, declares and binds the queue
// enables publish notifications.
func (ac *AMQPConnector) Connect(amqpURI string, exchange, exchangeType, queueName string, queueDurable, queueDelete bool, queueBindingKey string,
	exchangeDeclareArgs, queueDeclareArgs, queueBindingArgs amqp.Table) (*amqp.Connection, *amqp.Channel, amqp.Queue,
	<-chan amqp.Confirmation, <-chan *amqp.Error, error) {
	// connect to server
	conn, channel, err := ac.Open(amqpURI)
	if err != nil {
		return nil, nil, amqp.Queue{}, nil, nil, err
	}

	if exchange != "" {
		// Declare an exchange
		if err := channel.ExchangeDeclare(
			exchange,            // name of the exchange
			exchangeType,        // type
			true,                // durable
			false,               // delete when complete
			false,               // internal
			false,               // noWait
			exchangeDeclareArgs, // arguments
		); err != nil {
			return conn, channel, amqp.Queue{}, nil, nil, fmt.Errorf("Exchange declare error: %s ", err)
		}
	}

	var queue amqp.Queue
	if queueName != "" {
		// Declare a queue
		queue, err = channel.QueueDeclare(
			queueName,        // name of the queue
			queueDurable,     // durable
			queueDelete,      // delete when unused
			false,            // exclusive
			false,            // noWait
			queueDeclareArgs, // arguments
		)

		if err != nil {
			return conn, channel, amqp.Queue{}, nil, nil, fmt.Errorf("Queue declare error: %s ", err)
		}

		// Bind the queue to exchange
		if err = channel.QueueBind(
			queue.Name,       // name of the queue
			queueBindingKey,  // binding key(routing key)
			exchange,         // source exchange
			false,            // noWait
			queueBindingArgs, // arguments
		); err != nil {
			return conn, channel, queue, nil, nil, fmt.Errorf("Queue bind error: %s ", err)
		}
	}

	// Enable publishings confirm mode
	// confirm mode: the client can ensure all publishings have successfully been received by the server.
	if err = channel.Confirm(false); err != nil {
		return conn, channel, queue, nil, nil, fmt.Errorf("Channel could not be putted into confirm mode, error: %s ", err)
	}

	return conn, channel, queue, channel.NotifyPublish(make(chan amqp.Confirmation, 1)), conn.NotifyClose(make(chan *amqp.Error, 1)), nil
}

// DeleteQueue delete a queue by name.
func (ac *AMQPConnector) DeleteQueue(channel *amqp.Channel, queueName string) error {
	_, err := channel.QueueDelete(
		queueName, // name
		false,     // ifUnused
		false,     // ifEmpty
		false,     // noWait
	)

	return err
}

// InspectQueue
// inspect the current message count and consumer count by queue name.
func (ac *AMQPConnector) InspectQueue(channel *amqp.Channel, queueName string) (*amqp.Queue, error) {
	if channel != nil {
		queueState, err := channel.QueueInspect(queueName)
		if err != nil {
			return nil, fmt.Errorf("Queue inspect error: %s ", err)
		}

		return &queueState, nil
	}

	return nil, fmt.Errorf("amqp.Channel is nil")
}

// Open new RabbitMQ connection.
func (ac *AMQPConnector) Open(amqpURI string) (*amqp.Connection, *amqp.Channel, error) {
	// It is equivalent to calling DialTLS(amqp, nil) when it encounters an amqps:// scheme.
	conn, err := amqp.Dial(amqpURI)
	if err != nil {
		return nil, nil, fmt.Errorf("Dial error: %s ", err)
	}

	// opens a unique, concurrent server channel
	channel, err := conn.Channel()
	if err != nil {
		return nil, nil, fmt.Errorf("Open channel error: %s ", err)
	}

	return conn, channel, nil
}

// Close
// close connection and the deliveries channel.
func (ac *AMQPConnector) Close(consumerTag string, channel *amqp.Channel, conn *amqp.Connection) error {
	if consumerTag != "" && channel != nil {
		if err := channel.Cancel(consumerTag, true); err != nil {
			return fmt.Errorf("Consumer %s cancel failed: %s ", consumerTag, err)
		}

		return nil
	}

	if channel != nil {
		if err := channel.Close(); err != nil {
			return fmt.Errorf("Close channel error: %s ", err)
		}
	}

	if conn != nil {
		if err := conn.Close(); err != nil {
			return fmt.Errorf("Close connection error: %s ", err)
		}
	}

	return nil
}
