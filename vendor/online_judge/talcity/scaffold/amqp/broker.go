package amqp

import (
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/streadway/amqp"
)

// Broker an AMQP broker
type Broker struct {
	cnf           *AMQPConfig
	retry         bool
	retryFunc     func(chan int)
	retryStopChan chan int
	stopChan      chan int
	AMQPConnector
	processing sync.WaitGroup
}

// New creates a Broker instance.
// cnf.Broker has prefix `amqp://` or `amqps://`
func New(cnf *AMQPConfig) *Broker {
	return &Broker{cnf: cnf, retry: true, AMQPConnector: AMQPConnector{}}
}

// GetConfig
func (b *Broker) GetConfig() *AMQPConfig {
	return b.cnf
}

// GetRetry
func (b *Broker) GetRetry() bool {
	return b.retry
}

// GetRetryFunc
func (b *Broker) GetRetryFunc() func(chan int) {
	return b.retryFunc
}

// GetRetryStopChan
func (b *Broker) GetRetryStopChan() chan int {
	return b.retryStopChan
}

// GetStopChan
func (b *Broker) GetStopChan() chan int {
	return b.stopChan
}

// Publish place a new message on the queue.
func (b *Broker) Publish(routingKey string, priority uint8, body []byte, delay *time.Time) error {
	if routingKey == "" && b.GetConfig().ExchangeType == "direct" {
		// behind a direct exchange is simple.
		// a message goes to the queues whose binding key exactly matches the routing key of the message.
		routingKey = b.GetConfig().BindingKey
	}

	// priority scop 0-9
	if priority > 9 {
		priority = 9
	}

	if delay != nil {
		now := time.Now().UTC()
		if delay.After(now) {
			delayMs := int64(delay.Sub(now) / time.Millisecond)

			return b.delay(routingKey, priority, body, delayMs)
		}
	}

	conn, channel, _, confirmsChan, _, err := b.Connect(
		b.GetConfig().Broker,
		b.GetConfig().Exchange,     // exchange name
		b.GetConfig().ExchangeType, // exchange type
		"",    // queue name
		true,  // queue durable
		false, // queue delete when unused
		"",    // queue binding key
		nil,   // exchange declare args
		nil,   // queue declare args
		amqp.Table(b.GetConfig().QueueBindingArgs), // queue binding args
	)

	if err != nil {
		return err
	}

	defer b.Close("", channel, conn)

	if err := channel.Publish(
		b.GetConfig().Exchange, // publish to an exchange
		routingKey,             // routing to zero or more queues
		false,                  // mandatory
		false,                  // immediate
		amqp.Publishing{
			Headers:      amqp.Table{},
			ContentType:  "application/json",
			Body:         body,
			DeliveryMode: amqp.Persistent, // 1=transient, non-persistent, 2=persistent
			Priority:     priority,        // 0-9
		},
	); err != nil {
		return err
	}

	confirmed := <-confirmsChan

	if confirmed.Ack {
		return nil
	}

	return fmt.Errorf("Failed delivery the delivery tag: %v ", confirmed.DeliveryTag)
}

// delay a message by delayDuration miliseconds,
func (b *Broker) delay(routingKey string, priority uint8, message []byte, delayMs int64) error {
	if delayMs <= 0 {
		return errors.New("Cannot delay message by 0ms ")
	}

	// to redeclare the queue each time.
	// to zero its TTL timer.
	// format delay.[delayMs duration].[exchange name].[routingKey]
	queueName := fmt.Sprintf(
		"delay.%d.%s.%s",
		delayMs,                // delay duration in mileseconds
		b.GetConfig().Exchange, // exchange name
		routingKey,             // routing key
	)
	declareQueueArgs := amqp.Table{
		// Exchange where to send messages after TTL expiration.
		"x-dead-letter-exchange": b.GetConfig().Exchange,
		// Routing key which use when resending expired messages.
		"x-dead-letter-routing-key": routingKey,
		// Time in milliseconds
		// after that message will expire and be sent to destination.
		"x-message-ttl": delayMs,
		// Time after that the queue will be deleted.
		"x-expires": delayMs * 2,
	}
	conn, channel, _, _, _, err := b.Connect(
		b.GetConfig().Broker,
		b.GetConfig().Exchange,     // exchange name
		b.GetConfig().ExchangeType, // exchange type
		"",                                         // queue name
		true,                                       // queue durable
		false,                                      // queue delete when unused
		"",                                         // queue binding key
		nil,                                        // exchange declare args
		declareQueueArgs,                           // queue declare args
		amqp.Table(b.GetConfig().QueueBindingArgs), // queue binding args
	)

	if err != nil {
		return err
	}

	defer b.Close("", channel, conn)

	if err := channel.Publish(
		b.GetConfig().Exchange, // publish to an exchange
		queueName,              // routing key
		false,                  // mandatory
		false,                  // immediate
		amqp.Publishing{
			Headers:      amqp.Table{},
			ContentType:  "application/json",
			Body:         message,
			DeliveryMode: amqp.Persistent,
			Priority:     priority,
		},
	); err != nil {
		return err
	}

	return nil
}

// StartConsuming
func (b *Broker) StartConsuming(bindingKey, queueName, consumerTag string, concurrency int, messageProcessor MessageProcessor) (bool, error) {
	if bindingKey == "" && b.GetConfig().ExchangeType == "direct" {
		// behind a direct exchange is simple.
		// a message goes to the queues whose binding key exactly matches the routing key of the message.
		bindingKey = b.GetConfig().BindingKey
	}

	if b.retryFunc == nil {
		// use when there is a problem connecting to the broker.
		b.retryFunc = Attempt()
	}
	b.retryStopChan = make(chan int)
	b.stopChan = make(chan int)

	conn, channel, queue, _, amqpCloseChan, err := b.Connect(
		b.GetConfig().Broker,
		b.GetConfig().Exchange,     // exchange name
		b.GetConfig().ExchangeType, // exchange type
		queueName,                  // queue name
		true,                       // queue durable
		false,                      // queue delete when unused
		bindingKey,                 // queue binding key
		nil,                        // exchange declare args
		nil,                        // queue declare args
		amqp.Table(b.GetConfig().QueueBindingArgs), // queue binding args
	)

	if err != nil {
		b.GetRetryFunc()(b.GetRetryStopChan())
		return b.GetRetry(), err
	}
	defer b.Close(consumerTag, channel, conn)

	if err = channel.Qos(
		b.GetConfig().PrefetchCount, // prefetch count
		0,     // prefetch size
		false, // global
	); err != nil {
		return b.GetRetry(), fmt.Errorf("Channel qos error: %s ", err)
	}

	deliveries, err := channel.Consume(
		queue.Name,  // queue name
		consumerTag, // consumer tag
		false,       // auto ack
		false,       // exclusive
		false,       // no local
		false,       // no wait
		nil,         // arguments
	)

	if err != nil {
		return b.GetRetry(), fmt.Errorf("Queue consume error: %s ", err)
	}

	println("[*] Waiting for messages. \n To exit press CTRL+C")

	// Consume handle
	if err := b.consume(deliveries, amqpCloseChan, concurrency, messageProcessor); err != nil {
		return b.GetRetry(), err
	}

	// Waiting for any messages being processed to finish
	b.processing.Wait()

	return b.GetRetry(), nil
}

// consume handles delivered messages from the channel and managers a worker pool to process concurrently.
func (b *Broker) consume(deliveries <-chan amqp.Delivery, amqpCloseChan <-chan *amqp.Error, concurrency int, messageProcessor MessageProcessor) error {
	pool := make(chan struct{}, concurrency)

	// init worker pool
	go func() {
		for i := 0; i < concurrency; i++ {
			pool <- struct{}{}
		}
	}()

	errorChan := make(chan error)

	for {
		select {
		case amqpErr := <-amqpCloseChan:
			return amqpErr
		case err := <-errorChan:
			return err
		case d := <-deliveries:
			if concurrency > 0 {
				// get worker
				// blocks until one is available.
				<-pool
			}

			b.processing.Add(1)

			// consume inside a gotourine
			// multiple messages can be processed concurrently
			go func() {
				if err := b.consumeOne(d, messageProcessor); err != nil {
					errorChan <- err
				}

				b.processing.Done()

				if concurrency > 0 {
					// put worker to pool.
					pool <- struct{}{}
				}
			}()
		case <-b.GetStopChan():
			return nil
		}
	}
}

// consumeOne consumes a single message using MessageProcessor
func (b *Broker) consumeOne(delivery amqp.Delivery, messageProcessor MessageProcessor) error {
	var multiple, requeue = false, false
	if len(delivery.Body) == 0 {
		multiple = true
		delivery.Nack(multiple, false) // multiple, requeue
		return errors.New("Received an empty message, RabbitMQ down ? ")
	}

	if messageProcessor == nil {
		if !delivery.Redelivered {
			requeue = true
			fmt.Printf("Message processor is nil. Requeuing message: [%v-%v] %s", delivery.DeliveryMode, delivery.DeliveryTag, delivery.Body)
		}

		delivery.Nack(multiple, requeue)
		return nil
	}

	fmt.Printf("Reveived new message: delivery.mode-tag:[%v-%v] %s \n", delivery.DeliveryMode, delivery.DeliveryTag, delivery.Body)

	err := messageProcessor.Process(delivery.Body)
	delivery.Ack(multiple)
	return err
}

// StopConsuming
func (b *Broker) StopConsuming() {
	// not retry from now
	b.retry = false

	select {
	case b.retryStopChan <- 1:
		println("Stopping retry close.")
	default:
	}

	// Notify: the stop channel stops consuming messages
	b.stopChan <- 1

	// Waiting for any messages being processed to finish
	b.processing.Wait()
}
