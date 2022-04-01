// This file contains BackendCounts which is a helper for tests.
// We don't need it in the application, so by putting it here
// it won't be part of the compiled binary.
//
// Another option is to put this file under `/internal/backendcounts.go`
// but for a simple project such as this, this approach works just fine.
//
// We may want to export the functionality defined here at some point
// if it turns out to be useful to consumers of the package.
package glb

import (
	"testing"

	"github.com/stretchr/testify/require"
)

type BackendCounts[T comparable] map[T]int

func NewBackendCounts[T comparable]() BackendCounts[T] {
	return make(BackendCounts[T])
}

func (bc BackendCounts[T]) Inc(backend T) {
	bc[backend]++
}

func (bc BackendCounts[T]) Get(backend T) int {
	return bc[backend]
}

func (bc BackendCounts[T]) Merge(other BackendCounts[T]) {
	for backend, count := range other {
		bc[backend] += count
	}
}

func TestBackendCounts_Inc(t *testing.T) {
	bc := NewBackendCounts[string]()

	bc.Inc("b1")
	bc.Inc("b1")
	bc.Inc("b2")

	expected := BackendCounts[string](map[string]int{"b1": 2, "b2": 1})
	require.Equal(t, expected, bc)
}

func TestBackendCounts_Merge(t *testing.T) {
	bc1 := BackendCounts[string](map[string]int{"b1": 1})
	bc2 := BackendCounts[string](map[string]int{"b2": 2, "b3": 3})

	bc1.Merge(bc2)

	expected := BackendCounts[string](map[string]int{"b1": 1, "b2": 2, "b3": 3})
	require.Equal(t, expected, bc1)
}
