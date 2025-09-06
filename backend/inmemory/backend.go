package inmemory

import (
	"context"
	"fmt"
	"hash/fnv"
	"time"

	"github.com/halilbulentorhon/invacache-go/backend"
	"github.com/halilbulentorhon/invacache-go/backend/invalidation"
	"github.com/halilbulentorhon/invacache-go/backend/option"
	"github.com/halilbulentorhon/invacache-go/config"
	"github.com/halilbulentorhon/invacache-go/constant"
	"github.com/halilbulentorhon/invacache-go/pkg/logger"
)

type inMemoryBackend[V any] struct {
	ctx          context.Context
	invalidator  invalidation.PubSub
	logger       logger.Logger
	cancel       context.CancelFunc
	singleFlight singleFlight[V]
	shards       []inMemoryShard[V]
}

func (i *inMemoryBackend[V]) Get(key string) (V, error) {
	shard := i.getShard(key)
	shard.mu.Lock()
	defer shard.mu.Unlock()

	return shard.get(key)
}

func (i *inMemoryBackend[V]) GetOrLoad(key string, loader backend.LoaderFunc[V]) (V, error) {
	shard := i.getShard(key)

	shard.mu.Lock()
	if value, err := shard.get(key); err == nil {
		shard.mu.Unlock()
		return value, nil
	}
	shard.mu.Unlock()

	value, ttl, err := i.singleFlight.Do(key, func() (V, time.Duration, error) {
		return loader(key)
	})
	if err != nil {
		var zero V
		return zero, err
	}

	shard.mu.Lock()
	defer shard.mu.Unlock()

	if existing, err := shard.get(key); err == nil {
		return existing, nil
	}

	if setErr := shard.set(key, value, option.WithTTL(ttl)); setErr != nil {
		var zero V
		return zero, setErr
	}
	return value, nil
}

func (i *inMemoryBackend[V]) Set(key string, value V, options ...option.OptFnc) error {
	shard := i.getShard(key)
	shard.mu.Lock()
	err := shard.set(key, value, options...)
	shard.mu.Unlock()

	if err == nil {
		if pubErr := i.publishInvalidation(key); pubErr != nil {
			i.logger.Warn("failed to publish invalidation", "key", key, "error", pubErr)
		}
	}

	return err
}

func (i *inMemoryBackend[V]) publishInvalidation(key string) error {
	if i.invalidator != nil {
		err := i.invalidator.Publish(i.ctx, key)
		if err != nil {
			return fmt.Errorf("failed to publish invalidation for key %s: %w", key, err)
		}
	}
	return nil
}

func (i *inMemoryBackend[V]) Delete(key string) error {
	shard := i.getShard(key)
	shard.mu.Lock()
	err := shard.delete(key)
	shard.mu.Unlock()

	if err == nil {
		if pubErr := i.publishInvalidation(key); pubErr != nil {
			i.logger.Warn("failed to publish invalidation", "key", key, "error", pubErr)
		}
	}

	return err
}

func (i *inMemoryBackend[V]) runSweeper(shard *inMemoryShard[V], interval time.Duration) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		select {
		case <-i.ctx.Done():
			return
		case <-ticker.C:
			shard.mu.Lock()
			shard.sweepExpired()
			shard.mu.Unlock()
		}
	}
}

func (i *inMemoryBackend[V]) Close() error {
	i.logger.Info("closing inmemory cache")
	if i.invalidator != nil {
		err := i.invalidator.Close()
		if err != nil {
			i.logger.Error("failed to close invalidator", "error", err)
			return err
		}
	}
	if i.cancel != nil {
		i.cancel()
	}
	i.logger.Info("inmemory cache closed successfully")
	return nil
}

func (i *inMemoryBackend[V]) getShard(key string) *inMemoryShard[V] {
	h := fnv.New32a()
	h.Write([]byte(key))
	return &i.shards[h.Sum32()%uint32(len(i.shards))]
}

func getInvalidatorConfig(cfg *config.InvalidationConfig) interface{} {
	switch cfg.Type {
	case constant.CouchbaseInvalidationConfigType:
		return cfg.Couchbase
	case constant.RedisInvalidationConfigType:
		return cfg.Redis
	default:
		return nil
	}
}

func NewInMemoryBackend[V any](cfg config.InvaCacheConfig) backend.Cache[V] {
	log := logger.NewLogger("inmemory-cache")
	cfg.ApplyDefaults()

	log.Info("initializing inmemory cache",
		"shard_count", cfg.InMemory.ShardCount,
		"capacity", cfg.InMemory.Capacity,
		"sweeper_interval", cfg.InMemory.SweeperInterval)

	shards := make([]inMemoryShard[V], cfg.InMemory.ShardCount)
	baseCapacity := cfg.InMemory.Capacity / cfg.InMemory.ShardCount
	remainder := cfg.InMemory.Capacity % cfg.InMemory.ShardCount
	for i := range shards {
		capacity := baseCapacity
		if i == len(shards)-1 && remainder > 0 {
			capacity += remainder
		}
		shards[i] = newInMemoryShard[V](capacity)
	}

	ctx, cancel := context.WithCancel(context.Background())
	be := &inMemoryBackend[V]{
		shards:       shards,
		singleFlight: singleFlight[V]{},
		ctx:          ctx,
		cancel:       cancel,
		logger:       log,
	}

	for i := range shards {
		go be.runSweeper(&shards[i], cfg.InMemory.SweeperInterval)
	}

	if cfg.Invalidation != nil {
		log.Info("initializing invalidation", "type", cfg.Invalidation.Type)
		invalidatorConfig := getInvalidatorConfig(cfg.Invalidation)
		if inv, err := invalidation.NewInvalidator(cfg.Invalidation.Type, invalidatorConfig); err == nil {
			be.invalidator = inv
			log.Info("invalidation initialized successfully", "type", cfg.Invalidation.Type)
			go func() {
				log.Debug("starting invalidation subscription")
				err = inv.Subscribe(ctx, func(key string) error {
					log.Debug("received invalidation", "key", key)
					return be.Delete(key)
				})
				if err != nil {
					log.Info("invalidation subscription ended", "error", err)
				}
			}()
		} else {
			log.Error("failed to create invalidator", "type", cfg.Invalidation.Type, "error", err)
			panic(fmt.Sprintf("failed to create invalidator: %v\n", err))
		}
	}

	return be
}
