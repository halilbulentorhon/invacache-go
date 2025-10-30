package backend

import (
	"github.com/halilbulentorhon/invacache-go/backend/option"
	"time"
)

type LoaderFunc[V any] func(key string) (V, time.Duration, error)

type Cache[V any] interface {
	Get(key string) (V, error)
	GetOrLoad(key string, loader LoaderFunc[V]) (V, error)
	Set(key string, value V, options ...option.OptFnc) error
	Delete(key string, options ...option.DelOptFnc) error
	Clear(options ...option.ClrOptFnc) error
	Close() error
}
