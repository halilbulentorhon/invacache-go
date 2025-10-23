package invacache

import (
	"fmt"

	"github.com/halilbulentorhon/invacache-go/backend"
	"github.com/halilbulentorhon/invacache-go/backend/inmemory"
	"github.com/halilbulentorhon/invacache-go/config"
	"github.com/halilbulentorhon/invacache-go/constant"
)

func NewCache[V any](cfg config.InvaCacheConfig) (backend.Cache[V], error) {
	switch cfg.BackendName {
	case constant.InMemoryBackend:
		return inmemory.NewInMemoryBackend[V](cfg)
	default:
		return nil, fmt.Errorf("unknown backend name %s", cfg.BackendName)
	}
}
