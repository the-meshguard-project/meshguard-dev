package types

import (
	"sync"
	"testing"
)

func TestClockNext(t *testing.T) {
	clock := NewClock(0)

	if got := clock.Next(); got != 1 {
		t.Errorf("First Next() = %d, want 1", got)
	}

	if got := clock.Next(); got != 2 {
		t.Errorf("Second Next() = %d, want 2", got)
	}

	if got := clock.Next(); got != 3 {
		t.Errorf("Third Next() = %d, want 3", got)
	}
}

func TestClockCurrent(t *testing.T) {
	clock := NewClock(100)

	if got := clock.Current(); got != 100 {
		t.Errorf("Initial Current() = %d, want 100", got)
	}

	clock.Next()
	if got := clock.Current(); got != 101 {
		t.Errorf("After Next(), Current() = %d, want 101", got)
	}

	// Current should not increment
	if got := clock.Current(); got != 101 {
		t.Errorf("Second Current() = %d, want 101", got)
	}
}

func TestClockConcurrency(t *testing.T) {
	clock := NewClock(0)
	iterations := 1000
	goroutines := 10

	var wg sync.WaitGroup
	sequences := make([][]uint64, goroutines)

	for i := 0; i < goroutines; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			sequences[idx] = make([]uint64, iterations)
			for j := 0; j < iterations; j++ {
				sequences[idx][j] = clock.Next()
			}
		}(i)
	}

	wg.Wait()

	// Verify all sequences are unique
	seen := make(map[uint64]bool)
	for _, seq := range sequences {
		for _, num := range seq {
			if seen[num] {
				t.Errorf("Duplicate sequence number: %d", num)
			}
			seen[num] = true
		}
	}

	expectedTotal := goroutines * iterations
	if len(seen) != expectedTotal {
		t.Errorf("Expected %d unique sequences, got %d", expectedTotal, len(seen))
	}

	if got := clock.Current(); got != uint64(expectedTotal) {
		t.Errorf("Final counter = %d, want %d", got, expectedTotal)
	}
}
