package lb

import (
	"context"
	"errors"
)

var (
	ErrNotEnoughBackends = errors.New("there should be at least one backend in args")
)

type LoadBalancer[T any] interface {
	Next(ctx context.Context) (T, error)
}

type Strategy int

const (
	RoundRobinStrat Strategy = iota
	LeastConnsStrat
)

func New[T any](strat Strategy, backends []T) (LoadBalancer[T], error) {
	if strat == RoundRobinStrat {
		return NewRoundRobin(backends)
	}
	// if strat == LeastConnsStrat {
	// 	return NewLeastConns(backends)
	// }
	panic("unexpected load balancing strategy")
}
