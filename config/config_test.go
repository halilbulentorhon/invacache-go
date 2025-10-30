package config

import (
	"testing"
	"time"

	"github.com/halilbulentorhon/invacache-go/constant"
)

func TestApplyDefaultsBasic(t *testing.T) {
	cfg := InvaCacheConfig{}
	cfg.ApplyDefaults()

	if cfg.Backend == nil {
		t.Fatal("Backend should not be nil")
	}
	if cfg.Backend.InMemory == nil {
		t.Fatal("InMemory config should not be nil")
	}
	if cfg.Backend.InMemory.Capacity != constant.DefaultCapacity {
		t.Errorf("expected capacity %d, got %d", constant.DefaultCapacity, cfg.Backend.InMemory.Capacity)
	}
	if cfg.Backend.InMemory.ShardCount != constant.DefaultShardCount {
		t.Errorf("expected shard count %d, got %d", constant.DefaultShardCount, cfg.Backend.InMemory.ShardCount)
	}
	if cfg.Backend.InMemory.SweeperInterval != constant.DefaultSweeperInterval {
		t.Errorf("expected sweeper interval %v, got %v", constant.DefaultSweeperInterval, cfg.Backend.InMemory.SweeperInterval)
	}
}

func TestApplyDefaultsWithCustomValues(t *testing.T) {
	cfg := InvaCacheConfig{
		Backend: &BackendConfig{
			InMemory: &InMemoryConfig{
				Capacity:        500,
				ShardCount:      8,
				SweeperInterval: 2 * time.Minute,
			},
		},
	}
	cfg.ApplyDefaults()

	if cfg.Backend.InMemory.Capacity != 500 {
		t.Errorf("expected capacity 500, got %d", cfg.Backend.InMemory.Capacity)
	}
	if cfg.Backend.InMemory.ShardCount != 8 {
		t.Errorf("expected shard count 8, got %d", cfg.Backend.InMemory.ShardCount)
	}
	if cfg.Backend.InMemory.SweeperInterval != 2*time.Minute {
		t.Errorf("expected sweeper interval 2m, got %v", cfg.Backend.InMemory.SweeperInterval)
	}
}

func TestApplyDefaultsWithTTL(t *testing.T) {
	cfg := InvaCacheConfig{
		Backend: &BackendConfig{
			InMemory: &InMemoryConfig{
				Capacity:   100,
				ShardCount: 4,
				Ttl:        "10m",
			},
		},
	}
	cfg.ApplyDefaults()

	if cfg.Backend.InMemory.DefaultTTL != 10*time.Minute {
		t.Errorf("expected default TTL 10m, got %v", cfg.Backend.InMemory.DefaultTTL)
	}
}

func TestApplyDefaultsWithInvalidTTL(t *testing.T) {
	cfg := InvaCacheConfig{
		Backend: &BackendConfig{
			InMemory: &InMemoryConfig{
				Capacity:   100,
				ShardCount: 4,
				Ttl:        "invalid",
			},
		},
	}

	defer func() {
		if r := recover(); r == nil {
			t.Error("expected panic for invalid TTL format")
		}
	}()

	cfg.ApplyDefaults()
}

func TestApplyDefaultsWithEmptyTTL(t *testing.T) {
	cfg := InvaCacheConfig{
		Backend: &BackendConfig{
			InMemory: &InMemoryConfig{
				Capacity:   100,
				ShardCount: 4,
				Ttl:        "",
			},
		},
	}
	cfg.ApplyDefaults()

	if cfg.Backend.InMemory.DefaultTTL != 0 {
		t.Errorf("expected default TTL 0, got %v", cfg.Backend.InMemory.DefaultTTL)
	}
}

func TestApplyDefaultsCapacityLessThanShardCount(t *testing.T) {
	cfg := InvaCacheConfig{
		Backend: &BackendConfig{
			InMemory: &InMemoryConfig{
				Capacity:   10,
				ShardCount: 10,
			},
		},
	}

	defer func() {
		if r := recover(); r == nil {
			t.Error("expected panic when capacity equals shard count")
		}
	}()

	cfg.ApplyDefaults()
}

func TestApplyDefaultsCapacityLessThanShardCountNegative(t *testing.T) {
	cfg := InvaCacheConfig{
		Backend: &BackendConfig{
			InMemory: &InMemoryConfig{
				Capacity:   5,
				ShardCount: 10,
			},
		},
	}

	defer func() {
		if r := recover(); r == nil {
			t.Error("expected panic when capacity is less than shard count")
		}
	}()

	cfg.ApplyDefaults()
}

func TestApplyDefaultsVariousTTLFormats(t *testing.T) {
	tests := []struct {
		name     string
		ttl      string
		expected time.Duration
	}{
		{"seconds", "30s", 30 * time.Second},
		{"minutes", "5m", 5 * time.Minute},
		{"hours", "2h", 2 * time.Hour},
		{"combined", "1h30m", 90 * time.Minute},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := InvaCacheConfig{
				Backend: &BackendConfig{
					InMemory: &InMemoryConfig{
						Capacity:   100,
						ShardCount: 4,
						Ttl:        tt.ttl,
					},
				},
			}
			cfg.ApplyDefaults()

			if cfg.Backend.InMemory.DefaultTTL != tt.expected {
				t.Errorf("expected TTL %v, got %v", tt.expected, cfg.Backend.InMemory.DefaultTTL)
			}
		})
	}
}
