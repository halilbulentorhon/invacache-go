package invalidation

import (
	"context"
	"fmt"
)

type PubSub interface {
	Publish(ctx context.Context, key string) error
	Subscribe(ctx context.Context, handler InvalidationHandler) error
	Close() error
}

type InvalidationHandler func(key string) error

type InvalidatorFactory func(config interface{}) (PubSub, error)

var factories = make(map[string]InvalidatorFactory)

func RegisterInvalidator(name string, factory InvalidatorFactory) {
	factories[name] = factory
}

func NewInvalidator(invalidatorType string, config interface{}) (PubSub, error) {
	factory, exists := factories[invalidatorType]
	if !exists {
		return nil, fmt.Errorf("unknown invalidator type: %s", invalidatorType)
	}
	return factory(config)
}
