package glb

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestNewLeastConns_NotEnoughBackends(t *testing.T) {
	var backends []string
	rr, err := NewLeastConns(backends)
	require.Nil(t, rr)
	require.ErrorIs(t, err, ErrNotEnoughBackends)
}

func TestLeastConns_Next(t *testing.T) {
	t.Run("should return immediately with context error", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		cancel()

		lc, err := NewLeastConns([]string{"backend-0"})
		require.NoError(t, err)

		res, err := lc.Next(ctx)
		require.ErrorIs(t, err, context.Canceled)
		require.Equal(t, "", res)
	})

	t.Run("should return the same backend if length=1", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())

		lc, err := NewLeastConns([]string{"backend-0"})
		require.NoError(t, err)

		for i := 0; i < 10; i++ {
			res, err := lc.Next(ctx)
			require.NoError(t, err)
			require.Equal(t, "backend-0", res)
		}

		// Make sure goroutines quit.
		cancel()
		time.Sleep(10 * time.Millisecond)
		require.Equal(t, uint64(0), lc.backends[0].cnt)
	})

	t.Run("should choose backends based on cnt", func(t *testing.T) {
		lc, err := NewLeastConns([]string{"backend-1", "backend-2", "backend-3"})
		require.NoError(t, err)

		ctx1, cancel1 := context.WithCancel(context.Background())
		res, err := lc.Next(ctx1)
		require.NoError(t, err)
		require.Equal(t, "backend-1", res)

		ctx2, cancel2 := context.WithCancel(context.Background())
		res, err = lc.Next(ctx2)
		require.NoError(t, err)
		require.Equal(t, "backend-2", res)

		ctx3, cancel3 := context.WithCancel(context.Background())
		res, err = lc.Next(ctx3)
		require.NoError(t, err)
		require.Equal(t, "backend-3", res)

		// backend-1 becomes the least used.
		cancel1()
		ctx4, cancel4 := context.WithCancel(context.Background())
		res, err = lc.Next(ctx4)
		require.NoError(t, err)
		require.Equal(t, "backend-1", res)

		// Make sure goroutines quit.
		cancel2()
		cancel3()
		cancel4()
		time.Sleep(50 * time.Millisecond)
		for _, b := range lc.backends {
			require.Equal(t, uint64(0), b.cnt)
		}
	})
}

func TestLeastConns_Next_ConcurrentUsage(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	backends := []string{"backend-1", "backend-2", "backend-3"}

	lc, err := NewLeastConns(backends)
	require.NoError(t, err)

	resCh := make(chan BackendCounts[string])

	// Call `RoundRobin.Next` 10 times and return the `BackendCounts`.
	f := func() {
		bc := NewBackendCounts[string]()

		// Divisible by 3.
		for i := 0; i < 10; i++ {
			res, err := lc.Next(ctx)
			require.NoError(t, err)
			bc.Inc(res)
		}

		resCh <- bc
	}

	go f()
	go f()
	go f()

	// Allow goroutines to finish.
	time.Sleep(50 * time.Millisecond)

	bc := <-resCh
	bc.Merge(<-resCh)
	bc.Merge(<-resCh)

	// Same behavior as RoudnRobin if context is not canceled.
	require.Equal(t, 10, bc.Get("backend-1"))
	require.Equal(t, 10, bc.Get("backend-2"))
	require.Equal(t, 10, bc.Get("backend-3"))

	// Make sure goroutines quit.
	cancel()
	time.Sleep(50 * time.Millisecond)
	for _, b := range lc.backends {
		require.Equal(t, uint64(0), b.cnt)
	}
}

func TestLeastConns_leastUsedBackendIdx(t *testing.T) {
	tests := []struct {
		name        string
		lc          *LeastConns[string]
		expectedIdx int
	}{
		{
			name: "Should work when only 1 backend",
			lc: &LeastConns[string]{
				backends: []backend[string]{
					{b: "backend-0"},
				},
			},
			expectedIdx: 0,
		},
		{
			name: "Should return 1",
			lc: &LeastConns[string]{
				backends: []backend[string]{
					{b: "backend-0", cnt: 3},
					{b: "backend-1", cnt: 1},
					{b: "backend-2", cnt: 2},
				},
			},
			expectedIdx: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			actual := tt.lc.leastUsedBackendIdx()
			require.Equal(t, tt.expectedIdx, actual)
		})
	}
}
