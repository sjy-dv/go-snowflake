package snowflake

import (
	"fmt"
	"sync"
	"time"
)

type Snowflake struct {
	mu       sync.Mutex
	lastMs   int64
	sequence uint16
	Node     uint8
}

func Init(s *Snowflake) *Snowflake {
	return s
}

func (s *Snowflake) timeMs() int64 {
	return time.Now().UnixNano() / 1e6
}

func (s *Snowflake) GeneratorID() string {
	s.mu.Lock()
	defer s.mu.Unlock()
	ms := s.timeMs()
	//search last generate id
	if s.lastMs == ms {
		s.sequence++
		if s.sequence >= 1<<12 {
			//if sequence maximum, waiting for next ms
			for ms <= s.lastMs {
				ms = s.timeMs()
			}
		}
	} else {
		s.sequence = 0
	}
	s.lastMs = ms
	id := uint64(ms<<2 | int64(s.Node)<<16 | int64(s.sequence))
	return fmt.Sprintf("%d", id)
}
