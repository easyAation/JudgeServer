package amqp

import (
	"fmt"
	"math/rand"
	"os"
	"testing"
	"time"
)

var (
	defaultAMQPURI = "amqp://admin:hwl123456@221.122.65.176:5672/"
	workCount      = 1
	concurrency    = 2
	messageCount   = 2
	interval       = 10 // interval at which send messages in ms
	errorHandler   = func(err error) {
		fmt.Printf("worker error: %s \n", err)
	}
)

func TestWorkerCustomQueue(t *testing.T) {
	rand.Seed(time.Now().UTC().UnixNano())
	amqpURI := os.Getenv("AMQP_URL")
	amqpURI = defaultAMQPURI
	if amqpURI == "" {
		t.Log("AMQP_URL is not defined")
		t.Logf("Default AMQP_URL: %s", defaultAMQPURI)
		amqpURI = defaultAMQPURI
	}

	t.Logf("AMQP_URL: %s", amqpURI)

	cnf := defalutCnf
	(*cnf).Broker = amqpURI
	(*cnf).Exchange = "logs"

	t.Logf("AMQP Config: %#v", cnf)

	server := NewServer(cnf)
	logRouting := []string{"info", "debug"}
	infoWorker := server.NewCustomQueueWorker("worker.log.info", 2, cnf.Exchange+"."+logRouting[0], logRouting[0])
	debugWorker := server.NewCustomQueueWorker("worker.log.debug", 2, cnf.Exchange+"."+logRouting[1], logRouting[1])

	for i := 0; i < workCount; i++ {
		go func() {
			infoWorker.errorHandler = errorHandler
			infoWorker.Launch()
		}()

		go func() {
			debugWorker.errorHandler = errorHandler
			debugWorker.Launch()
		}()
	}

	defer func() {
		go func() {
			infoWorker.Quit()
		}()

		go func() {
			debugWorker.Quit()
		}()
	}()

	done := make(chan error)

	doing := 0
	published := false
	fmt.Printf("publish message count: %d \n", concurrency*messageCount)
	for {
		select {
		case err := <-done:
			if err != nil {
				fmt.Printf("publish error: %+v \n", err)
			}

			doing++
			fmt.Printf("doing: %d \n", doing)
		case <-time.After(5 * time.Millisecond):
			if published {
				goto END
			}

			if !published {
				go func() {
					publishMessage(server, cnf, logRouting, done)
				}()
			}

			published = true
		}

	END:
		if doing == concurrency*messageCount {
			fmt.Printf("published: done count: %d, message count: %d \n", doing, concurrency*messageCount)
			time.Sleep(5 * time.Second)
			break
		}
	}

	return
}

func publishMessage(server *Server, cnf *AMQPConfig, routingKey []string, done chan error) {
	message := &Message{
		RoutingKey: "",
		Priority:   0,
		Body:       []byte("ok"),
	}
	for i := 0; i < concurrency; i++ {
		go func(i int) {
			doneCount := 1
			for {
				message.RoutingKey = cnf.Exchange + "." + routingKey[rand.Intn(len(routingKey))]
				fmt.Printf("Send Message: %#v \n", message)
				err := server.SendMessage(message)

				if err != nil {
					fmt.Printf("Send Message Error: %s \n", err)
					done <- err
					break
				}

				done <- nil

				doneCount++
				fmt.Printf("concurrency %d, message count: %d \n", i, doneCount-1)
				if messageCount != 0 && doneCount > messageCount {
					break
				}

				time.Sleep(time.Duration(interval) * time.Millisecond)
			}
		}(i)
	}
}
