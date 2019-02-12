package amqp

import (
	"fmt"
	"time"
)

// Fibonacci returns Fibonacci numbers.
// starting from 1.
func Fibonacci() func() int {
	a, b := 0, 1
	return func() int {
		a, b = b, a+b
		return a
	}
}

// FibonacciNext returns when Fibonacci number gt start.
func FibonacciNext(start int) int {
	fib := Fibonacci()
	fibNum := fib()
	for fibNum <= start {
		fibNum = fib()
	}

	return fibNum
}

// use retry func when there is a problem connecting to the broker.
// use Fibonacci space retry attempt.
var Attempt = func() func(chan int) {
	retryInside := 0
	fibonacci := Fibonacci()
	return func(stopChan chan int) {
		if retryInside > 0 {
			duration, _ := time.ParseDuration(fmt.Sprintf("%vs", retryInside))

			select {
			case <-stopChan:
				break
			case <-time.After(duration):
				break
			}
		}

		retryInside = fibonacci()
	}
}
