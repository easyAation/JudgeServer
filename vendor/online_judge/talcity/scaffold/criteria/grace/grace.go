package grace

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"online_judge/talcity/scaffold/criteria/log"
	"online_judge/talcity/scaffold/criteria/merr"
)

var cleanups []func()

type panicError struct {
	panicObject interface{}
}

func (pe panicError) Error() string {
	return fmt.Sprintf("panic(%s): %v", merr.IdentifyPanic(), pe.panicObject)
}

func (pe panicError) PanicObject() interface{} {
	return pe.panicObject
}

// New build an empty Manager
func New() *Manager {
	return &Manager{}
}

// Manager for graceful shutdown
type Manager struct {
	cleanups []func()
}

// Register 注册清理函数，确保在Run执行传入的函数后（即使收到中断信号）会执行之前注册过的清理函数
func (m *Manager) Register(f func()) {
	cleanups = append(cleanups, f)
}

// Run 执行传入的函数，并运行之前注册过的清理函数
func (m *Manager) Run(f func() error) {
	signals := make(chan os.Signal, 1)
	signal.Notify(signals, syscall.SIGINT, syscall.SIGKILL, syscall.SIGTERM)

	quit := make(chan error)
	go func() {
		var err error
		defer func() {
			if e := recover(); e != nil {
				quit <- panicError{e}
			} else {
				quit <- err
			}
		}()

		err = f()
	}()

	select {
	case sig := <-signals:
		log.Infof("receive signal %s", sig)
	case err := <-quit:
		log.Errorf("exit with error %v", err)
	}

	for _, cleanup := range cleanups {
		cleanup()
	}
}
