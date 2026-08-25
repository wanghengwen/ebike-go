package fastid

import (
	"fmt"
	"sync"
)

// Generator mirrors com.xyy.fastid.core.FastId snowflake layout.
type Generator struct {
	cfg              Config
	instanceNo       int64
	timestampShift   int64
	instanceNoShift  int64
	timeSeq          *from2021TimeSequence
	autoIncrementSeq *defaultAutoIncrementSequence
}

var (
	global     *Generator
	globalOnce sync.Once
	globalErr  error
)

// Init configures the process-wide FastId generator (call once at startup).
func Init(cfg Config) error {
	globalOnce.Do(func() {
		global, globalErr = New(cfg)
	})
	return globalErr
}

// MustInit panics when FastId cannot be initialized.
func MustInit(cfg Config) {
	if err := Init(cfg); err != nil {
		panic(err)
	}
}

// Next returns the next distributed id.
func Next() int64 {
	if global == nil {
		panic("fastid not initialized")
	}
	return global.Next()
}

// Ready reports whether FastId has been initialized.
func Ready() bool {
	return global != nil
}

// New creates a FastId generator after fetching machine id from register center.
func New(cfg Config) (*Generator, error) {
	cfg.applyDefaults()
	g := &Generator{
		cfg:              cfg,
		timestampShift:   cfg.InstanceNoBits + cfg.SequenceBits,
		instanceNoShift:  cfg.SequenceBits,
		timeSeq:          &from2021TimeSequence{},
		autoIncrementSeq: &defaultAutoIncrementSequence{},
	}
	instanceNo, err := resolveInstanceNo(cfg)
	if err != nil {
		return nil, err
	}
	maxInstanceNo := int64(-1 ^ (-1 << cfg.InstanceNoBits))
	if instanceNo > maxInstanceNo {
		return nil, fmt.Errorf("instance no %d exceeds max %d", instanceNo, maxInstanceNo)
	}
	g.instanceNo = instanceNo
	g.timeSeq.init()
	g.autoIncrementSeq.init(cfg.DriftTime, cfg.SequenceBits)
	return g, nil
}

func (g *Generator) Next() int64 {
	timeSequence := g.timeSeq.sequence()
	autoIncrement := g.autoIncrementSeq.sequence(timeSequence)
	return (timeSequence << g.timestampShift) | (g.instanceNo << g.instanceNoShift) | autoIncrement
}
