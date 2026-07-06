package idgen

import (
	"context"
	"errors"
	"sync"
	"time"
)

const (
	snowflakeEpoch int64 = 1609459200000 // 2021-01-01 UTC
	machineBits          = 10
	sequenceBits         = 12
	maxMachineID         = (1 << machineBits) - 1
	maxSequence          = (1 << sequenceBits) - 1
)

// SnowflakeGenerator produces distributed unique IDs (timestamp + machine + sequence).
type SnowflakeGenerator struct {
	mu        sync.Mutex
	machineID int64
	lastTS    int64
	sequence  int64
}

func NewSnowflakeGenerator(machineID int64) (*SnowflakeGenerator, error) {
	if machineID < 0 || machineID > maxMachineID {
		return nil, errors.New("machine id out of range")
	}
	return &SnowflakeGenerator{machineID: machineID}, nil
}

func (g *SnowflakeGenerator) Next(_ context.Context) (uint64, error) {
	g.mu.Lock()
	defer g.mu.Unlock()

	now := time.Now().UnixMilli()
	if now < g.lastTS {
		now = g.waitNextMillis(g.lastTS)
	}

	if now == g.lastTS {
		g.sequence = (g.sequence + 1) & maxSequence
		if g.sequence == 0 {
			now = g.waitNextMillis(g.lastTS)
		}
	} else {
		g.sequence = 0
	}

	g.lastTS = now
	id := ((now - snowflakeEpoch) << (machineBits + sequenceBits)) |
		(g.machineID << sequenceBits) |
		g.sequence

	return uint64(id), nil
}

func (g *SnowflakeGenerator) waitNextMillis(last int64) int64 {
	now := time.Now().UnixMilli()
	for now <= last {
		now = time.Now().UnixMilli()
	}
	return now
}
