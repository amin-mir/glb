package lb

import (
	"context"
	"sync/atomic"
)

type RoundRobin[T any] struct {
	// We don't expect to have a huge number of backends to load balance
	// Thus, uint32 should be sufficient.
	cnt      uint32
	backends []T
}

var _ LoadBalancer[string] = &RoundRobin[string]{}

func NewRoundRobin[T any](backends []T) (*RoundRobin[T], error) {
	if len(backends) == 0 {
		return nil, ErrNotEnoughBackends
	}
	rr := &RoundRobin[T]{
		backends: backends,
	}
	return rr, nil
}

func (rr *RoundRobin[T]) Next(ctx context.Context) (T, error) {
	select {
	case <-ctx.Done():
		var zero T
		return zero, ctx.Err()
	default:
	}

	cnt := atomic.AddUint32(&rr.cnt, 1)
	idx := int(cnt-1) % len(rr.backends)
	return rr.backends[idx], nil
}
