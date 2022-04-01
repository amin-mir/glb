package glb

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestNewRoundRobin_NotEnoughBackends(t *testing.T) {
	var backends []string
	rr, err := NewRoundRobin(backends)
	require.Nil(t, rr)
	require.ErrorIs(t, err, ErrNotEnoughBackends)
}

func TestRoundRobin_Next(t *testing.T) {
	t.Run("should return immediately with context error", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		cancel()

		rr, err := NewRoundRobin([]string{"backend-0"})
		require.NoError(t, err)

		res, err := rr.Next(ctx)
		require.ErrorIs(t, err, context.Canceled)
		require.Equal(t, "", res)
	})

	t.Run("should return the same backend if length=1", func(t *testing.T) {
		ctx := context.Background()

		rr, err := NewRoundRobin([]string{"backend-0"})
		require.NoError(t, err)

		for i := 0; i < 10; i++ {
			res, err := rr.Next(ctx)
			require.NoError(t, err)
			require.Equal(t, "backend-0", res)
		}
	})

	t.Run("should choose backends based on cnt", func(t *testing.T) {
		ctx := context.Background()

		rr, err := NewRoundRobin([]string{"backend-0", "backend-1", "backend-2"})
		require.NoError(t, err)

		rr.cnt = 1
		res, err := rr.Next(ctx)
		require.NoError(t, err)
		require.Equal(t, "backend-1", res)

		res, err = rr.Next(ctx)
		require.NoError(t, err)
		require.Equal(t, "backend-2", res)

		res, err = rr.Next(ctx)
		require.NoError(t, err)
		require.Equal(t, "backend-0", res)
	})
}

func TestRoundRobin_Next_ConcurrentUsage(t *testing.T) {
	ctx := context.Background()

	backends := []string{"backend-0", "backend-1", "backend-2"}

	rr, err := NewRoundRobin(backends)
	require.NoError(t, err)

	resCh := make(chan BackendCounts[string])

	// Call `RoundRobin.Next` 10 times and return the `BackendCounts`.
	f := func() {
		bc := NewBackendCounts[string]()

		for i := 0; i < 10; i++ {
			res, err := rr.Next(ctx)
			require.NoError(t, err)
			bc.Inc(res)
		}

		resCh <- bc
	}

	go f()
	go f()
	go f()

	bc := <-resCh
	bc.Merge(<-resCh)
	bc.Merge(<-resCh)

	require.Equal(t, 10, bc.Get("backend-0"))
	require.Equal(t, 10, bc.Get("backend-1"))
	require.Equal(t, 10, bc.Get("backend-2"))
}
