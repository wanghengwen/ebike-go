package fastid

import (
	"fmt"
	"math/rand"
	"sync"
)

// defaultAutoIncrementSequence mirrors DefaultAutoIncrementSequence.
type defaultAutoIncrementSequence struct {
	maxSequence int64
	wheel       [][2]int64
	index       int
	mu          sync.Mutex
}

func (s *defaultAutoIncrementSequence) init(driftTime int, sequenceBits int64) {
	s.maxSequence = -1 ^ (-1 << sequenceBits)
	s.wheel = make([][2]int64, driftTime*60)
}

func (s *defaultAutoIncrementSequence) getTimeSequenceIndex(preTimeSequence, timeSequence int64) int {
	timeSeq := s.wheel[s.index][0]
	if timeSeq == 0 {
		s.wheel[s.index][0] = timeSequence
		return s.index
	}
	if timeSeq == timeSequence {
		return s.index
	}
	if preTimeSequence > timeSeq {
		s.wheel[s.index][0] = timeSequence
		s.wheel[s.index][1] = 0
		return s.index
	}
	if timeSequence > timeSeq {
		s.index++
		if s.index >= len(s.wheel) {
			s.index = 0
		}
		return s.getTimeSequenceIndex(timeSeq, timeSequence)
	}
	idx := s.index - 1
	if timeSequence < timeSeq {
		for idx != s.index {
			if idx < 0 {
				idx = len(s.wheel) - 1
			}
			preTimeSeq := s.wheel[idx][0]
			if preTimeSeq == timeSequence {
				return idx
			}
			idx--
		}
		if idx == s.index {
			return -2
		}
	}
	return -1
}

func (s *defaultAutoIncrementSequence) sequence(timeSequence int64) int64 {
	s.mu.Lock()
	defer s.mu.Unlock()

	index := s.getTimeSequenceIndex(-1, timeSequence)
	if index < 0 {
		panic(fmt.Sprintf("time sequence wheel exception, index=%d", index))
	}
	seq := s.wheel[index][1]
	if seq == 0 {
		seq = int64(rand.Intn(8192) + 1024)
	} else {
		seq++
	}
	s.wheel[index][1] = seq
	if seq > s.maxSequence {
		panic(fmt.Sprintf("sequence out of bounds maxSequence=%d", s.maxSequence))
	}
	return seq
}
