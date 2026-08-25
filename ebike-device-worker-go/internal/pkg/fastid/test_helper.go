package fastid

// NewForTest builds a generator with a fixed machine id (unit tests only).
func NewForTest(cfg Config, instanceNo int64) (*Generator, error) {
	cfg.applyDefaults()
	g := &Generator{
		cfg:             cfg,
		instanceNo:      instanceNo,
		timestampShift:  cfg.InstanceNoBits + cfg.SequenceBits,
		instanceNoShift: cfg.SequenceBits,
		timeSeq:         &from2021TimeSequence{},
		autoIncrementSeq: &defaultAutoIncrementSequence{},
	}
	g.timeSeq.init()
	g.autoIncrementSeq.init(cfg.DriftTime, cfg.SequenceBits)
	return g, nil
}
