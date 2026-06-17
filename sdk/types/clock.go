package types

import "sync/atomic"

// Clock provides atomic sequence generation for event ordering
type Clock struct {
	counter uint64
}

// NewClock initializes a sequence clock
func NewClock(start uint64) *Clock {
	return &Clock{counter: start}
}

// Next returns the next sequence number atomically
func (c *Clock) Next() uint64 {
	return atomic.AddUint64(&c.counter, 1)
}

// Current returns the current sequence without incrementing
func (c *Clock) Current() uint64 {
	return atomic.LoadUint64(&c.counter)
}
