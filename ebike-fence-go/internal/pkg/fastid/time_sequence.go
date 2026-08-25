package fastid

import "time"

type from2021TimeSequence struct {
	epochMs int64
}

func (s *from2021TimeSequence) init() {
	epoch, err := time.ParseInLocation("20060102150405", "20210101000000", time.Local)
	if err != nil {
		panic(err)
	}
	s.epochMs = epoch.UnixMilli()
}

func (s *from2021TimeSequence) sequence() int64 {
	return (time.Now().UnixMilli() - s.epochMs) / 1000
}
