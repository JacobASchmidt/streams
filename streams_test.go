package streams_test

import (
	"testing"

	"github.com/JacobAlbertSchmidt/streams"
)

func TestStreams(t *testing.T) {
	const n = 1000
	range_ := streams.Range(0, n)
	sum := streams.Reduce(range_, 0, func(init int, val int) int {
		return init + val
	})
	expected := (n * (n - 1)) / 2
	if sum != expected {
		t.Fatalf("expected %v, got %v", expected, sum)
	}

}
