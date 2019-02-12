/*
用于记录必要的操作信息, 建议保存在context中, 用法见测试
*/
package trace

import (
	"context"
	"fmt"
	"sync"
)

const (
	// DefaultKey the default key for context's trace value
	DefaultKey Key = "default_trace_key"
	// StagingKey 用于记录不同阶段的trace, 可选
	StagingKey = "stages"
)

// Key context key
type Key string

// T 用于记录trace
type T struct {
	mu     sync.RWMutex
	keys   []string
	traces map[string]interface{}
}

// New buil new T
func New() *T {
	return &T{
		traces: make(map[string]interface{}),
	}
}

// Keys get keys
func (t *T) Keys() []string {
	t.mu.RLock()
	result := make([]string, len(t.keys))
	copy(result, t.keys)
	t.mu.RUnlock()
	return result
}

// Traces 返回keys, traces
func (t *T) Traces() ([]string, map[string]interface{}) {
	t.mu.RLock()

	keys := make([]string, len(t.keys))
	copy(keys, t.keys)
	traces := make(map[string]interface{})
	for k, v := range t.traces {
		traces[k] = lightCopy(v)
	}

	t.mu.RUnlock()
	return keys, traces
}

// Value get trace by key
func (t *T) Value(key string) interface{} {
	t.mu.RLock()
	val := t.traces[key]
	t.mu.RUnlock()

	return lightCopy(val)
}

// Stage 提供默认key 用于记录不同阶段的trace
func (t *T) Stage(msgs ...interface{}) {
	t.Add("stages", msgs...)
}

// Add build msgs to a singal trace, and store key and trace to t
func (t *T) Add(key string, msgs ...interface{}) {
	if len(msgs) == 0 {
		return
	}
	var trace interface{}
	if len(msgs) > 1 {
		trace = fmt.Sprintf(msgs[0].(string), msgs[1:]...)
	} else {
		trace = msgs[0]
	}

	t.mu.Lock()
	val, ok := t.traces[key]

	if ok {
		arr, y := val.([]interface{})
		if y {
			val = append(arr, trace)
		} else {
			val = []interface{}{val, trace}
		}
	} else {
		val = trace
		t.keys = append(t.keys, key)
	}

	t.traces[key] = val
	t.mu.Unlock()
}

// Add add trace to context, trace must already set in context
// otherwise panic
func Add(ctx context.Context, key string, msgs ...interface{}) {
	t := Get(ctx)
	t.Add(key, msgs...)
}

// Stage 提供默认key 用于记录不同阶段的trace
func Stage(ctx context.Context, msgs ...interface{}) {
	t := Get(ctx)
	t.Add("stages", msgs...)
}

// ToCtx set trace to context
func ToCtx(ctx context.Context, t *T) context.Context {
	return context.WithValue(ctx, DefaultKey, t)
}

// Get Get get trace by key
func Get(ctx context.Context) *T {
	return FromCtx(ctx, DefaultKey)
}

// FromCtx Get get trace by key, if value not set, will panic
func FromCtx(ctx context.Context, key Key) *T {
	return ctx.Value(key).(*T)
}

// lightCopy 浅copy, 如果是slice则copy, 否则返回原值
func lightCopy(src interface{}) interface{} {
	arr, ok := src.([]interface{})
	if ok {
		result := make([]interface{}, len(arr))
		copy(result, arr)
		return result
	}
	return src
}
