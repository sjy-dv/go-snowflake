package snowflake

import (
	"fmt"
	"math/rand"
	"sync"
	"sync/atomic"
	"time"
)

// Snowflake struct represents a snowflake ID generator
type Snowflake struct {
	mu       sync.Mutex
	sequence uint32
	lastMs   int64
	node     uint32
	ready    chan struct{}
}

// Constructor for Snowflake
func NewSnowflake() *Snowflake {
	return &Snowflake{
		ready: make(chan struct{}),
	}
}

// Returns the current time in milliseconds
func (s *Snowflake) timeMs() int64 {
	return time.Now().UTC().UnixNano() / 1e6
}

// Wait until the next millisecond
func (s *Snowflake) waitUntilNextMs() {
	for atomic.LoadInt64(&s.lastMs) == s.timeMs() {
		time.Sleep(time.Microsecond)
	}
}

// Generates a new ID and increments the sequence number
func (s *Snowflake) generateID() uint64 {
	ms := s.timeMs()
	// If time is behind the current time, wait until the next millisecond
	if ms < atomic.LoadInt64(&s.lastMs) {
		s.waitUntilNextMs()
		ms = s.timeMs()
	}
	// If time is the same as the current time, increment the sequence number
	if ms == atomic.LoadInt64(&s.lastMs) {
		atomic.AddUint32(&s.sequence, 1)
		if s.sequence >= 1<<12 {
			// If sequence number exceeds the maximum, wait until the next millisecond
			s.waitUntilNextMs()
			atomic.StoreUint32(&s.sequence, 0)
			ms = s.timeMs()
		}
	} else {
		// Reset the sequence number if the time is ahead of the current time
		atomic.StoreUint32(&s.sequence, 0)
	}
	// Store the current time and sequence number
	atomic.StoreInt64(&s.lastMs, ms)

	// Generate the ID
	s.node = rand.Uint32()
	id := uint64(ms)<<22 | uint64(s.node)<<10 | uint64(s.sequence)
	return id
}

// Generates a new ID and waits until it is time to generate a new ID
func (s *Snowflake) GeneratorID() string {
	// Lock the mutex while generating the ID
	s.mu.Lock()
	defer s.mu.Unlock()

	id := s.generateID()

	// Wait until it is time to generate a new ID
	select {
	case <-s.ready:
	default:
		time.Sleep(time.Millisecond)
	}

	return fmt.Sprintf("%d", id)
}

// Notify when it is time to generate a new ID
func (s *Snowflake) NotifyReady() {
	close(s.ready)
	s.ready = make(chan struct{})
}
