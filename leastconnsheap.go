package glb

import (
	"context"
	"sync"
)

type LeastConnsHeap[T any] struct {
	backends []*backend[T]
	mu       sync.Mutex
}

var _ LoadBalancer[string] = &LeastConnsHeap[string]{}

func NewLeastConnsHeap[T any](backends []T) (*LeastConnsHeap[T], error) {
	if len(backends) == 0 {
		return nil, ErrNotEnoughBackends
	}

	bb := make([]*backend[T], len(backends))
	for i, b := range backends {
		bb[i] = &backend[T]{
			b: b,
		}
	}

	lc := &LeastConnsHeap[T]{
		backends: bb,
	}
	return lc, nil
}

func (lc *LeastConnsHeap[T]) Next(ctx context.Context) (T, error) {
	select {
	case <-ctx.Done():
		var zero T
		return zero, ctx.Err()
	default:
	}

	lc.mu.Lock()
	leastUsed := lc.leastUsedBackend()
	leastUsed.cnt++
	lc.mu.Unlock()

	// Update the conn count once context is canceled.
	go func() {
		<-ctx.Done()
		lc.mu.Lock()
		leastUsed.cnt--
		lc.mu.Unlock()
	}()

	return leastUsed.b, nil
}

// leastUsedBackend returns the *backend with the least
// number of connections.
func (lc *LeastConnsHeap[T]) leastUsedBackend() *backend[T] {
	// lc.backends is guaranteed to have at least 1 element.
	leastUsed := lc.backends[0]
	for _, b := range lc.backends[1:] {
		if b.cnt < leastUsed.cnt {
			leastUsed = b
			break
		}
	}
	return leastUsed
}
