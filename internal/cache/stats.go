package cache

import "sync/atomic"

type Stats struct {
	Hits       atomic.Int64
	Misses     atomic.Int64
	BytesSaved atomic.Int64
	Entries    atomic.Int64
}

func (s *Stats) RecordHit(bodySize int64) {
	s.Hits.Add(1)
	s.BytesSaved.Add(bodySize)
}

func (s *Stats) RecordMiss() {
	s.Misses.Add(1)
}

func (s *Stats) HitRatio() float64 {
	hits := s.Hits.Load()
	total := hits + s.Misses.Load()
	if total == 0 {
		return 0
	}
	return float64(hits) / float64(total)
}
