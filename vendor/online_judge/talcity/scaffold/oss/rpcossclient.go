package oss

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/vitessio/vitess/go/pools"
	"google.golang.org/grpc"
)

var (
	poolNoInit        = errors.New("oss client pool without init")
	defaultDialOption = grpc.WithInsecure()
)

// OOSClient resource.
type OOSClientResource struct {
	OOSClient
	closed bool
}

func (ocr OOSClientResource) Close() {
	if !ocr.closed {
		ocr.OOSClient.(*oOSClient).cc.Close()
	}
}

// oOSClientFactory create a resource of OOSClient.
func oOSClientFactory(address string) pools.Factory {
	return func() (pools.Resource, error) {
		// default a DialOption which disables transport security for this ClientConn.
		maxMsgLimitCallOption := grpc.WithDefaultCallOptions(grpc.MaxCallSendMsgSize(100 * 1000 * 1000))
		conn, err := grpc.Dial(address, defaultDialOption, maxMsgLimitCallOption)
		// conn, err := grpc.Dial(address, defaultDialOption)

		if err != nil {
			return nil, err
		}

		return &OOSClientResource{NewOOSClient(conn), false}, nil
	}
}

// OOSClientPool oss client pool
type OOSClientPool struct {
	*pools.ResourcePool
}

// NewOOSClientPool create oss client pool.
// address rpc server address.
// cap is the number of possible resources in the pool
// maxCap specifies the extent to which the pool can be resized
// in the future through the SetCapacity function.
// cannot resize the pool beyond maxCap.
// If a resource is unused beyond idleTimeout, it's discarded.
// An idleTimeout of 0 means that there is no timeout.
func NewOOSClientPool(address string, cap, maxCap int, idleTimeout time.Duration) (*OOSClientPool, error) {
	if cap <= 0 || maxCap <= 0 || cap > maxCap {
		return nil, fmt.Errorf("capacity %d is out of range", cap)
	}

	return CustomOOSClientPool(oOSClientFactory(address), cap, maxCap, idleTimeout), nil
}

func CustomOOSClientPool(factory pools.Factory, cap, maxCap int, idleTimeout time.Duration) *OOSClientPool {
	return &OOSClientPool{
		pools.NewResourcePool(factory, cap, maxCap, idleTimeout),
	}
}

// SetCapacity changes the capacity of the pool.
func (op *OOSClientPool) SetOOSClientPoolCap(cap int) error {
	return op.SetCapacity(cap)
}

// GetOOSClient get the oos client from pool.
func (op *OOSClientPool) GetOOSClient(ctx context.Context) (*OOSClientResource, error) {
	if op == nil {
		return nil, poolNoInit
	}

	conn, err := op.Get(ctx)
	if err != nil {
		return nil, err
	}

	osc, ok := conn.(*OOSClientResource)
	if !ok {
		op.Put(conn)
		return nil, errors.New("pool error: oss client resource no exist")
	}

	return osc, nil
}
