package config

import (
	"fmt"
	"time"

	"github.com/halilbulentorhon/invacache-go/constant"
)

type InvaCacheConfig struct {
	BackendName  string              `json:"backendName"`
	Backend      *BackendConfig      `json:"backend"`
	Invalidation *InvalidationConfig `json:"invalidation,omitempty"`
}

type BackendConfig struct {
	InMemory *InMemoryConfig `json:"inMemory"`
}

type InMemoryConfig struct {
	ShardCount      int           `json:"shardCount"`
	SweeperInterval time.Duration `json:"sweeperInterval"`
	Capacity        int           `json:"capacity"`
	Ttl             string        `json:"ttl"`
	DefaultTTL      time.Duration `json:"-"`
}

type InvalidationConfig struct {
	Type         string         `json:"type"`
	DriverConfig map[string]any `json:"driverConfig,omitempty"`
}

func (cfg *InvaCacheConfig) ApplyDefaults() {
	if cfg.Backend == nil {
		cfg.Backend = &BackendConfig{}
	}
	if cfg.Backend.InMemory == nil {
		cfg.Backend.InMemory = &InMemoryConfig{}
	}
	if cfg.Backend.InMemory.Capacity <= 0 {
		cfg.Backend.InMemory.Capacity = constant.DefaultCapacity
	}
	if cfg.Backend.InMemory.ShardCount <= 0 {
		cfg.Backend.InMemory.ShardCount = constant.DefaultShardCount
	}
	if cfg.Backend.InMemory.SweeperInterval <= 0 {
		cfg.Backend.InMemory.SweeperInterval = constant.DefaultSweeperInterval
	}
	if cfg.Backend.InMemory.Capacity <= cfg.Backend.InMemory.ShardCount {
		panic(fmt.Sprintf("shard count(%d) cannot be greater then or equal to capacity(%d)", cfg.Backend.InMemory.ShardCount, cfg.Backend.InMemory.Capacity))
	}
	if cfg.Backend.InMemory.Ttl != "" {
		duration, err := time.ParseDuration(cfg.Backend.InMemory.Ttl)
		if err != nil {
			panic(fmt.Sprintf("invalid ttl format '%s': %v", cfg.Backend.InMemory.Ttl, err))
		}
		cfg.Backend.InMemory.DefaultTTL = duration
	}
}
