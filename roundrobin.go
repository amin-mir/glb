package glb

import (
	"context"
	"sync/atomic"
)

type RoundRobin[T any] struct {
	cnt      uint64
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

	cnt := atomic.AddUint64(&rr.cnt, 1)
	idx := (cnt - 1) % uint64(len(rr.backends))
	return rr.backends[int(idx)], nil
}
