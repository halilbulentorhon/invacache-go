package couchbase

import (
	"context"
	"fmt"

	"github.com/halilbulentorhon/cb-pubsub/config"
	"github.com/halilbulentorhon/cb-pubsub/pubsub"
	"github.com/halilbulentorhon/invacache-go/backend/invalidation"
)

func init() {
	invalidation.RegisterInvalidator("couchbase", func(config any) (invalidation.PubSub, error) {
		cfgMap, ok := config.(map[string]any)
		if !ok {
			return nil, fmt.Errorf("invalid config type for couchbase invalidator, expected map[string]any, got %T", config)
		}

		cfg := CouchbaseConfig{
			ConnectionString: getString(cfgMap, "ConnectionString"),
			Username:         getString(cfgMap, "Username"),
			Password:         getString(cfgMap, "Password"),
			BucketName:       getString(cfgMap, "BucketName"),
			CollectionName:   getString(cfgMap, "CollectionName"),
			ScopeName:        getString(cfgMap, "ScopeName"),
			GroupName:        getString(cfgMap, "GroupName"),
		}

		return NewCouchbaseInvalidator(cfg)
	})
}

func getString(m map[string]any, key string) string {
	if val, ok := m[key].(string); ok {
		return val
	}
	return ""
}

type CouchbaseConfig struct {
	ConnectionString string `json:"connectionString"`
	Username         string `json:"username"`
	Password         string `json:"password"`
	BucketName       string `json:"bucketName"`
	CollectionName   string `json:"collectionName"`
	ScopeName        string `json:"scopeName,omitempty"`
	GroupName        string `json:"groupName,omitempty"`
}

type CouchbaseInvalidator struct {
	pubsub pubsub.PubSub[string]
}

func NewCouchbaseInvalidator(cfg CouchbaseConfig) (invalidation.PubSub, error) {
	if cfg.ScopeName == "" {
		cfg.ScopeName = "_default"
	}
	if cfg.GroupName == "" {
		cfg.GroupName = "invacache"
	}

	pubsubConfig := config.PubSubConfig{
		CouchbaseConfig: config.CouchbaseConfig{
			Host:                cfg.ConnectionString,
			Username:            cfg.Username,
			Password:            cfg.Password,
			BucketName:          cfg.BucketName,
			ScopeName:           cfg.ScopeName,
			CollectionName:      cfg.CollectionName,
			ConnectTimeoutSec:   10,
			OperationTimeoutSec: 5,
		},
		PollIntervalSeconds:    1,
		CleanupIntervalSeconds: 15,
		SubscribeRetryAttempts: 10,
		CleanupRetryAttempts:   5,
	}

	ps, err := pubsub.NewCbPubSub[string](cfg.GroupName, pubsubConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to create cb-pubsub instance: %w", err)
	}

	return &CouchbaseInvalidator{
		pubsub: ps,
	}, nil
}

func (c *CouchbaseInvalidator) Publish(ctx context.Context, key string) error {
	return c.pubsub.Publish(ctx, key)
}

func (c *CouchbaseInvalidator) Subscribe(ctx context.Context, handler invalidation.InvalidationHandler) error {
	return c.pubsub.Subscribe(ctx, func(messages []string) error {
		for _, key := range messages {
			if err := handler(key); err != nil {
				fmt.Printf("Error processing invalidation for key %s: %v\n", key, err)
			}
		}
		return nil
	})
}

func (c *CouchbaseInvalidator) Close() error {
	return c.pubsub.Close()
}
