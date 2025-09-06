package invacache

import (
	"github.com/halilbulentorhon/invacache-go/backend"
	"github.com/halilbulentorhon/invacache-go/backend/inmemory"
	"github.com/halilbulentorhon/invacache-go/config"
)

func NewInMemory[V any](cfg config.InvaCacheConfig) backend.Cache[V] {
	return inmemory.NewInMemoryBackend[V](cfg)
}
