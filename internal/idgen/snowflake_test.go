package idgen

import (
	"context"
	"testing"
)

func TestSnowflakeGeneratorUniqueIDs(t *testing.T) {
	gen, err := NewSnowflakeGenerator(1)
	if err != nil {
		t.Fatalf("NewSnowflakeGenerator() error = %v", err)
	}

	seen := make(map[uint64]struct{})
	for i := 0; i < 1000; i++ {
		id, err := gen.Next(context.Background())
		if err != nil {
			t.Fatalf("Next() error = %v", err)
		}
		if _, ok := seen[id]; ok {
			t.Fatalf("duplicate id %d", id)
		}
		seen[id] = struct{}{}
	}
}

func TestSnowflakeGeneratorInvalidMachineID(t *testing.T) {
	if _, err := NewSnowflakeGenerator(2000); err == nil {
		t.Fatal("expected error for invalid machine id")
	}
}
