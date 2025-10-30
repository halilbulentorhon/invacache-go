# InvaCache-Go

A high-performance, thread-safe, generic in-memory cache library for Go with TTL support, automatic cleanup, and
distributed invalidation capabilities.

## Features

- 🚀 **High Performance**: Shard-based architecture for concurrent access
- 🔧 **Generic Types**: Full support for Go generics (Go 1.21+)
- ⏰ **TTL Support**: Automatic expiration with configurable cleanup intervals
- 🔄 **SingleFlight**: Prevents duplicate requests for the same key
- 🌐 **Distributed Invalidation**: Optional cache invalidation with pluggable drivers (zero deps core)
- 🧵 **Thread-Safe**: Concurrent read/write operations with minimal locking
- 📊 **Configurable**: Customizable shard count, capacity, and sweep intervals
- 📝 **Structured Logging**: Built-in structured logging with configurable levels and formats

## Installation

### Core Library (Zero Dependencies)

```bash
go get github.com/halilbulentorhon/invacache-go
```

This installs only the core cache functionality with no external dependencies.

### Optional Invalidation Drivers

InvaCache-Go uses a **pluggable driver architecture** to keep the core library dependency-free. Install only the drivers
you need:

#### Couchbase Driver

```bash
go get github.com/halilbulentorhon/invacache-go/invalidation/drivers/couchbase
```

#### Redis Driver

```bash
go get github.com/halilbulentorhon/invacache-go/invalidation/drivers/redis
```

### Dependency Isolation

| Usage Scenario     | Dependencies                     |
|--------------------|----------------------------------|
| **In-Memory Only** | ✅ Zero dependencies              |
| **+ Couchbase**    | ✅ Only Couchbase SDK & cb-pubsub |
| **+ Redis**        | ✅ Only Redis client              |
| **+ Both Drivers** | ✅ Both driver dependencies       |

Each driver is a separate module, so you only pull in what you actually use.

## Quick Start

### Basic Usage (No External Dependencies)

```go
package main

import (
	"fmt"
	"time"

	"github.com/halilbulentorhon/invacache-go"
	"github.com/halilbulentorhon/invacache-go/config"
	"github.com/halilbulentorhon/invacache-go/backend/option"
)

func main() {
	// Create cache configuration
	cfg := config.InvaCacheConfig{
		BackendName: "in-memory",
		Backend: &config.BackendConfig{
			InMemory: &config.InMemoryConfig{
				ShardCount:      16,               // Number of shards for concurrent access
				Capacity:        10000,            // Maximum number of items
				SweeperInterval: 30 * time.Second, // Cleanup interval
				Ttl:             "10m",            // Default TTL for all items (optional)
			},
		},
	}

	// Create a new cache instance for string values
	cache, err := invacache.NewCache[string](cfg)
	if err != nil {
		panic(err)
	}
	defer cache.Close() // Always cleanup resources

	// Set a value with TTL
	err = cache.Set("user:123", "John Doe", option.WithTTL(5*time.Minute))
	if err != nil {
		panic(err)
	}

	// Get a value
	value, err := cache.Get("user:123")
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}
	fmt.Printf("User: %s\n", value)

	// Get or load pattern
	user, err := cache.GetOrLoad("user:456", func(key string) (string, time.Duration, error) {
		// This function is called only if the key is not in cache
		// Simulate database lookup
		return "Jane Doe", 10 * time.Minute, nil
	})
	if err != nil {
		panic(err)
	}
	fmt.Printf("Loaded user: %s\n", user)
}
```

## API Reference

### Cache Interface

```go
type Cache[V any] interface {
    Get(key string) (V, error)
    GetOrLoad(key string, loader LoaderFunc[V]) (V, error)
    Set(key string, value V, options ...option.OptFnc) error
    Delete(key string, options ...option.DelOptFnc) error
    Clear(options ...option.ClrOptFnc) error
    Close() error
}
```

### Configuration

```go
type InvaCacheConfig struct {
    BackendName  string              `json:"backendName"`
    Backend      *BackendConfig      `json:"backend"`
    Invalidation *InvalidationConfig `json:"invalidation,omitempty"`
}

type BackendConfig struct {
    InMemory *InMemoryConfig `json:"inMemory"`
}

type InMemoryConfig struct {
    ShardCount      int           `json:"shardCount"`      // Default: 8
    SweeperInterval time.Duration `json:"sweeperInterval"` // Default: 10 minutes
    Capacity        int           `json:"capacity"`        // Default: 1000
    Ttl             string        `json:"ttl"`             // Default TTL for all items (e.g., "10m", "1h")
}

### Options

**Set Options:**
- `option.WithTTL(duration)` - Set expiration time for specific key
- `option.WithNoExpiration()` - Set item to never expire
- `option.WithInvalidation()` - Trigger distributed invalidation on Set

**Delete Options:**
- `option.WithDeleteInvalidation()` - Trigger distributed invalidation on Delete

**Clear Options:**
- `option.WithClearInvalidation()` - Trigger distributed invalidation on Clear

## Advanced Usage

### Custom Types

```go
type User struct {
    ID   int    `json:"id"`
    Name string `json:"name"`
}

cache, err := invacache.NewCache[User](cfg)
if err != nil {
	panic(err)
}
defer cache.Close() // Always cleanup resources

user := User{ID: 1, Name: "John"}
cache.Set("user:1", user, option.WithTTL(time.Hour))

// Clear all cache
err = cache.Clear()
if err != nil {
	panic(err)
}

// Clear with distributed invalidation
err = cache.Clear(option.WithClearInvalidation())
if err != nil {
	panic(err)
}
```

### Structured Logging

InvaCache-Go includes built-in structured logging to help you monitor cache operations and troubleshoot issues.

#### Production Logging (JSON format, Info level)

```go
cache, err := invacache.NewCache[string](cfg)
if err != nil {
	panic(err)
}
```

#### Development Logging (Text format, Debug level)

```go
// Note: NewCache function handles backend selection based on BackendName
// Use NewCache with custom logger configuration for development
cache, err := invacache.NewCache[string](cfg)
if err != nil {
	panic(err)
}
```

#### Log Levels

- **debug**: Detailed operation logs (cache hits/misses, individual operations)
- **info**: General information (cache initialization, invalidation events)
- **warn**: Warning messages (invalidation failures)
- **error**: Error conditions (cache operation failures)

#### Sample Production Output (JSON)

```json
{"time":"2025-09-06T13:07:07.846723+03:00","level":"INFO","msg":"initializing inmemory cache","component":"inmemory-cache","shard_count":16,"capacity":10000}
{"time":"2025-09-06T13:07:07.847048+03:00","level":"INFO","msg":"invalidation initialized successfully","component":"inmemory-cache","type":"couchbase"}
```

#### Sample Development Output (Text, Debug)

```
time=2025-09-06T13:11:32.172+03:00 level=DEBUG msg="cache set operation" component=inmemory-cache key=user:123
time=2025-09-06T13:11:32.172+03:00 level=DEBUG msg="cache hit" component=inmemory-cache key=user:123
time=2025-09-06T13:11:32.172+03:00 level=DEBUG msg="cache delete successful" component=inmemory-cache key=user:123
```

### Distributed Invalidation

InvaCache-Go supports **pluggable invalidation drivers** to keep your core library dependency-free. Each driver is a
separate module that you install only when needed.

**Available Drivers:**

#### Couchbase Invalidation Driver (using cb-pubsub)

#### Redis Invalidation Driver (using Redis Pub/Sub)

Install the Couchbase invalidation driver separately:

```bash
go get github.com/halilbulentorhon/invacache-go/invalidation/drivers/couchbase
```

Then use it in your application:

```go
import (
    "github.com/halilbulentorhon/invacache-go"
    "github.com/halilbulentorhon/invacache-go/config"
    
    // Import the Couchbase invalidation driver
    _ "github.com/halilbulentorhon/invacache-go/invalidation/drivers/couchbase"
)

cfg := config.InvaCacheConfig{
    BackendName: "in-memory",
    Backend: &config.BackendConfig{
        InMemory: &config.InMemoryConfig{
            ShardCount: 16,
            Capacity:   10000,
        },
    },
    Invalidation: &config.InvalidationConfig{
        Type: "couchbase",
        DriverConfig: map[string]any{
            "ConnectionString": "localhost:8091",
            "Username":         "admin",
            "Password":         "password",
            "BucketName":       "default",
            "CollectionName":   "_default",
            "ScopeName":        "_default",
            "GroupName":        "my-cache-group",
        },
    },
}

cache, err := invacache.NewCache[string](cfg)
if err != nil {
	panic(err)
}

// Important: Always close the cache to properly cleanup resources
defer cache.Close() // This will stop invalidation goroutines and cleanup connections
```

The Couchbase driver uses the [cb-pubsub](https://github.com/halilbulentorhon/cb-pubsub) library for reliable
distributed messaging.

**Important Notes:**

- Invalidation subscription runs in a separate goroutine (non-blocking)
- Always call `cache.Close()` to properly cleanup resources and stop background goroutines
- **Resilient Operation**: Cache operations never fail due to invalidation issues
- **Driver Responsibility**: Retry logic and reconnection are handled by the invalidation driver (cb-pubsub)
- **Graceful Degradation**: Cache works locally when invalidation is unavailable

Install the Redis invalidation driver separately:

```bash
go get github.com/halilbulentorhon/invacache-go/invalidation/drivers/redis
```

Then use it in your application:

```go
import (
    "time"
    
    "github.com/halilbulentorhon/invacache-go"
    "github.com/halilbulentorhon/invacache-go/config"
    
    // Import the Redis invalidation driver
    _ "github.com/halilbulentorhon/invacache-go/invalidation/drivers/redis"
)

cfg := config.InvaCacheConfig{
    BackendName: "in-memory",
    Backend: &config.BackendConfig{
        InMemory: &config.InMemoryConfig{
            ShardCount: 16,
            Capacity:   10000,
        },
    },
    Invalidation: &config.InvalidationConfig{
        Type: "redis",
        DriverConfig: map[string]any{
            "Address":     "localhost:6379",
            "Password":    "",
            "DB":          0,
            "Channel":     "invacache:invalidation",
            "PoolSize":    10,
            "MaxRetries":  3,
            "DialTimeout": 5 * time.Second,
        },
    },
}

cache, err := invacache.NewCache[string](cfg)
if err != nil {
	panic(err)
}
defer cache.Close() // Always cleanup resources
```

The Redis driver uses Redis's native Pub/Sub mechanism for real-time distributed messaging.

**Redis Driver Benefits:**

- **High Performance**: Redis's optimized pub/sub implementation
- **Real-time**: Immediate invalidation propagation
- **Scalable**: Supports unlimited cache instances
- **Simple Setup**: Minimal Redis configuration required
- **Reliable**: Built on Redis's proven infrastructure

## Performance

InvaCache-Go is designed for high-performance scenarios:

- **Sharded Architecture**: Reduces lock contention by distributing keys across multiple shards
- **Minimal Allocations**: Efficient memory usage with pre-allocated structures
- **SingleFlight**: Prevents thundering herd problems for expensive operations
- **Concurrent Sweeping**: Background cleanup doesn't block cache operations

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.
