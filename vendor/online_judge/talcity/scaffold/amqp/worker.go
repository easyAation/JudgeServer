package amqp

import (
	"errors"
	"fmt"
	"os"
	"os/signal"
	"syscall"
)

// Worker a single worker process.
type Worker struct {
	server         *Server
	ConsumerTag    string
	Concurrency    int
	ConsumingQueue string
	BindingKey     string
	errorHandler   func(err error)
	processor      MessageProcessor
}

// CustomQueue returns consuming queue of the running worker process.
func (worker *Worker) CustomQueue() string {
	return worker.ConsumingQueue
}

// Quit down the running worker process.
func (worker *Worker) Quit() {
	worker.server.GetBroker().StopConsuming()
}

// Process worker handle message
func (worker *Worker) Process(message []byte) error {
	fmt.Printf("Got %dB delivery message: %s \n", len(message), message)
	return nil
}

// GetServer returns used server.
func (worker *Worker) GetServer() *Server {
	return worker.server
}

// SetErrorHandler sets a custom error handler for message handle errors.
func (worker *Worker) SetErrorHandler(handler func(err error)) {
	worker.errorHandler = handler
}

// SetProcessor sets a custom message processor.
func (worker *Worker) SetProcessor(processor MessageProcessor) {
	worker.processor = processor
}

// Launch starts a new worker process.
// the worker binding custom queue and handling incomming message.
func (worker *Worker) Launch() error {
	errorChan := make(chan error)
	worker.LaunchAsync(errorChan)
	return <-errorChan
}

// LaunchAsync a non blocking Launch.
func (worker *Worker) LaunchAsync(errorChan chan<- error) {
	cnf := worker.server.GetConfig()
	broker := worker.server.GetBroker()

	// print some useful information about configuration
	println("Launching a worker with the following settings: ")
	fmt.Printf("- Broker: %s \n", cnf.Broker)
	fmt.Printf("- CustomingQueue: %s \n", worker.ConsumingQueue)
	fmt.Printf("  - Exchange: %s \n", cnf.Exchange)
	fmt.Printf("  - ExchangeType: %s \n", cnf.ExchangeType)
	fmt.Printf("  - BindingKey: %s \n", cnf.BindingKey)
	fmt.Printf("  - PrefetchCount: %d \n", cnf.PrefetchCount)
	fmt.Println("- Worker: ")
	fmt.Printf("  - ConsumerTag: %s \n", worker.ConsumerTag)
	fmt.Printf("  - BindingKey: %s \n", worker.BindingKey)
	fmt.Printf("  - ConsumingQueue: %s \n", worker.ConsumingQueue)

	if worker.processor == nil {
		worker.processor = worker
	}
	// start broker to consume
	// attempt when broker connection dies
	go func() {
		for {
			retry, err := broker.StartConsuming(worker.BindingKey, worker.ConsumingQueue, worker.ConsumerTag, worker.Concurrency, worker.processor)

			fmt.Printf("Broker start consume %s, retry: %t \n", worker.ConsumerTag, retry)
			if retry {
				if err != nil && worker.errorHandler != nil {
					worker.errorHandler(err)
				}
			} else {
				errorChan <- err
				return
			}
		}
	}()

	// Handle SIGINT and SIGTERM signals
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, os.Interrupt, syscall.SIGTERM)
	var signalReceived uint

	go func() {
		for {
			select {
			case s := <-sig:
				fmt.Printf("Signal received: %v", s)
				signalReceived++

				if signalReceived < 2 {
					// After first `Ctrl+C`, the worker gracefully quitting.
					go func() {
						worker.Quit()
						errorChan <- fmt.Errorf("%s Worker quit gracefully ", worker.ConsumerTag)
					}()
				} else {
					// Abort the program when hitting `Ctrl+C` second.
					errorChan <- errors.New("Worker quit , when hitting `Ctrl+C` second time. ")
				}
			}
		}
	}()
}
