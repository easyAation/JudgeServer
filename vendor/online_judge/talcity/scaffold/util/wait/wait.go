package wait

import (
	"context"
	"sync"
)

// Group allows to start by group and waits for completion.
type Group struct {
	wg sync.WaitGroup
}

func (g *Group) Wait() {
	g.wg.Wait()
}

// Start
// start f of a new goroutine in the group.
func (g *Group) Start(f func()) {
	g.wg.Add(1)
	go func() {
		defer g.wg.Done()
		f()
	}()
}

// StartWithChannel
// start f of a new goroutine in the group.
// stopCh is an argument of f.
// f should stop when stopCh is enable.
func (g *Group) StartWithChannel(stopCh <-chan struct{}, f func(stopCh <-chan struct{})) {
	g.Start(func() {
		f(stopCh)
	})
}

// StartWithContext
// start f of a new goroutine in the group.
// ctx is an argument of f.
// f should stop when ctx.Done().
func (g *Group) StartWithContext(ctx context.Context, f func(context.Context)) {
	g.Start(func() {
		f(ctx)
	})
}
