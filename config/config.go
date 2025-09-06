package config

import (
	"github.com/halilbulentorhon/invacache-go/constant"
	"time"
)

type InvaCacheConfig struct {
	BackendName  string              `json:"backendName"`
	InMemory     InMemoryConfig      `json:"inMemory"`
	Invalidation *InvalidationConfig `json:"invalidation,omitempty"`
}

type InMemoryConfig struct {
	ShardCount      int           `json:"shardCount"`
	SweeperInterval time.Duration `json:"sweeperInterval"`
	Capacity        int           `json:"capacity"`
}

type InvalidationConfig struct {
	Type      string                       `json:"type"`
	Couchbase *CouchbaseInvalidationConfig `json:"couchbase,omitempty"`
	Redis     *RedisInvalidationConfig     `json:"redis,omitempty"`
}

type CouchbaseInvalidationConfig struct {
	ConnectionString string        `json:"connectionString"`
	Username         string        `json:"username"`
	Password         string        `json:"password"`
	BucketName       string        `json:"bucketName"`
	CollectionName   string        `json:"collectionName"`
	ScopeName        string        `json:"scopeName,omitempty"`
	GroupName        string        `json:"groupName,omitempty"`
	PollInterval     time.Duration `json:"pollInterval"`
}

type RedisInvalidationConfig struct {
	Address     string        `json:"address"`
	Password    string        `json:"password,omitempty"`
	DB          int           `json:"db"`
	Channel     string        `json:"channel,omitempty"`
	PoolSize    int           `json:"poolSize,omitempty"`
	MaxRetries  int           `json:"maxRetries,omitempty"`
	DialTimeout time.Duration `json:"dialTimeout,omitempty"`
}

func (cfg *InvaCacheConfig) ApplyDefaults() {
	if cfg.InMemory.Capacity <= 0 {
		cfg.InMemory.Capacity = constant.DefaultCapacity
	}
	if cfg.InMemory.ShardCount <= 0 {
		cfg.InMemory.ShardCount = constant.DefaultShardCount
	}
	if cfg.InMemory.SweeperInterval <= 0 {
		cfg.InMemory.SweeperInterval = constant.DefaultSweeperInterval
	}
}
