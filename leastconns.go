package glb

import (
	"context"
	"sync"
)

type LeastConns[T any] struct {
	backends []backend[T]
	mu       sync.Mutex
}

type backend[T any] struct {
	b   T
	cnt uint64
}

var _ LoadBalancer[string] = &LeastConns[string]{}

func NewLeastConns[T any](backends []T) (*LeastConns[T], error) {
	if len(backends) == 0 {
		return nil, ErrNotEnoughBackends
	}

	bb := make([]backend[T], len(backends))
	for i, b := range backends {
		bb[i] = backend[T]{
			b: b,
		}
	}

	lc := &LeastConns[T]{
		backends: bb,
	}
	return lc, nil
}

func (lc *LeastConns[T]) Next(ctx context.Context) (T, error) {
	select {
	case <-ctx.Done():
		var zero T
		return zero, ctx.Err()
	default:
	}

	lc.mu.Lock()
	idx := lc.leastUsedBackendIdx()
	lc.backends[idx].cnt++
	lc.mu.Unlock()

	// Update the conn count once context is canceled.
	go func() {
		<-ctx.Done()
		lc.mu.Lock()
		lc.backends[idx].cnt--
		lc.mu.Unlock()
	}()

	return lc.backends[idx].b, nil
}

// leastUsedBackend returns the index of least used backend.
func (lc *LeastConns[T]) leastUsedBackendIdx() int {
	// lc.backends is guaranteed to have at least 1 element.
	idx := 0
	for i := 1; i < len(lc.backends); i++ {
		if lc.backends[i].cnt < lc.backends[idx].cnt {
			idx = i
			break
		}
	}
	return idx
}
