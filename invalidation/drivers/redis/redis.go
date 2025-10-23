package redis

import (
	"context"
	"fmt"
	"time"

	"github.com/halilbulentorhon/invacache-go/backend/invalidation"
	"github.com/redis/go-redis/v9"
)

func init() {
	invalidation.RegisterInvalidator("redis", func(config any) (invalidation.PubSub, error) {
		cfgMap, ok := config.(map[string]any)
		if !ok {
			return nil, fmt.Errorf("invalid config type for redis invalidator, expected map[string]any, got %T", config)
		}

		cfg := RedisConfig{
			Address:     getString(cfgMap, "Address"),
			Password:    getString(cfgMap, "Password"),
			DB:          getInt(cfgMap, "DB"),
			Channel:     getString(cfgMap, "Channel"),
			PoolSize:    getInt(cfgMap, "PoolSize"),
			MaxRetries:  getInt(cfgMap, "MaxRetries"),
			DialTimeout: getDuration(cfgMap, "DialTimeout"),
		}

		return NewRedisInvalidator(cfg)
	})
}

func getString(m map[string]any, key string) string {
	if val, ok := m[key].(string); ok {
		return val
	}
	return ""
}

func getInt(m map[string]any, key string) int {
	if val, ok := m[key].(int); ok {
		return val
	}
	return 0
}

func getDuration(m map[string]any, key string) time.Duration {
	if val, ok := m[key].(time.Duration); ok {
		return val
	}
	return 0
}

type RedisConfig struct {
	Address     string        `json:"address"`
	Password    string        `json:"password,omitempty"`
	DB          int           `json:"db"`
	Channel     string        `json:"channel,omitempty"`
	PoolSize    int           `json:"poolSize,omitempty"`
	MaxRetries  int           `json:"maxRetries,omitempty"`
	DialTimeout time.Duration `json:"dialTimeout,omitempty"`
}

type RedisInvalidator struct {
	client  *redis.Client
	channel string
}

func NewRedisInvalidator(cfg RedisConfig) (invalidation.PubSub, error) {
	if cfg.Channel == "" {
		cfg.Channel = "invacache:invalidation"
	}
	if cfg.PoolSize <= 0 {
		cfg.PoolSize = 10
	}
	if cfg.MaxRetries <= 0 {
		cfg.MaxRetries = 3
	}
	if cfg.DialTimeout <= 0 {
		cfg.DialTimeout = 5 * time.Second
	}

	rdb := redis.NewClient(&redis.Options{
		Addr:        cfg.Address,
		Password:    cfg.Password,
		DB:          cfg.DB,
		PoolSize:    cfg.PoolSize,
		MaxRetries:  cfg.MaxRetries,
		DialTimeout: cfg.DialTimeout,
	})

	ctx, cancel := context.WithTimeout(context.Background(), cfg.DialTimeout)
	defer cancel()

	if err := rdb.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("failed to connect to Redis: %w", err)
	}

	return &RedisInvalidator{
		client:  rdb,
		channel: cfg.Channel,
	}, nil
}

func (r *RedisInvalidator) Publish(ctx context.Context, key string) error {
	err := r.client.Publish(ctx, r.channel, key).Err()
	if err != nil {
		return fmt.Errorf("failed to publish invalidation message: %w", err)
	}
	return nil
}

func (r *RedisInvalidator) Subscribe(ctx context.Context, handler invalidation.InvalidationHandler) error {
	pubsub := r.client.Subscribe(ctx, r.channel)
	defer pubsub.Close()

	ch := pubsub.Channel()

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case msg := <-ch:
			if msg == nil {
				continue
			}

			if err := handler(msg.Payload); err != nil {
				fmt.Printf("Error processing invalidation for key %s: %v\n", msg.Payload, err)
			}
		}
	}
}

func (r *RedisInvalidator) Close() error {
	return r.client.Close()
}
