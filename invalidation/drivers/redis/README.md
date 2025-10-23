# Redis Invalidation Driver

This driver provides distributed cache invalidation using Redis Pub/Sub for InvaCache-Go.

## Installation

```bash
go get github.com/halilbulentorhon/invacache-go/invalidation/drivers/redis
```

## Usage

```go
package main

import (
    "time"
    
    "github.com/halilbulentorhon/invacache-go"
    "github.com/halilbulentorhon/invacache-go/config"
    
    // Import the Redis invalidation driver
    _ "github.com/halilbulentorhon/invacache-go/invalidation/drivers/redis"
)

func main() {
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
    
    // Cache will now use Redis for distributed invalidation
}
```

## Configuration

```go
type RedisInvalidationConfig struct {
    Address     string        `json:"address"`     // Redis server address (e.g., "localhost:6379")
    Password    string        `json:"password"`    // Redis password (optional)
    DB          int           `json:"db"`          // Redis database number
    Channel     string        `json:"channel"`     // Pub/Sub channel name (optional, defaults to "invacache:invalidation")
    PoolSize    int           `json:"poolSize"`    // Connection pool size (optional, defaults to 10)
    MaxRetries  int           `json:"maxRetries"`  // Maximum retry attempts (optional, defaults to 3)
    DialTimeout time.Duration `json:"dialTimeout"` // Connection timeout (optional, defaults to 5s)
}
```

## How It Works

1. **Redis Pub/Sub**: Uses Redis's native pub/sub mechanism for message broadcasting
2. **Channel-based**: All cache instances subscribe to the same Redis channel
3. **Message Broadcasting**: When a key is invalidated, message is published to the channel
4. **Real-time**: Immediate invalidation across all connected instances
5. **Lightweight**: Minimal overhead with Redis's efficient pub/sub implementation

## Benefits

- **High Performance**: Redis's optimized pub/sub implementation
- **Scalable**: Supports unlimited cache instances
- **Real-time**: Immediate invalidation propagation
- **Reliable**: Built on Redis's proven infrastructure
- **Simple**: Minimal configuration required
- **Fault Tolerant**: Automatic reconnection and retry logic

## Redis Setup

### Using Docker

```bash
# Start Redis server
docker run -d --name redis -p 6379:6379 redis:alpine

# With password
docker run -d --name redis -p 6379:6379 redis:alpine redis-server --requirepass mypassword
```

### Using Redis CLI

```bash
# Install Redis
brew install redis  # macOS
sudo apt install redis-server  # Ubuntu

# Start Redis
redis-server

# Test connection
redis-cli ping
```

## Configuration Examples

### Basic Configuration

```go
Redis: &config.RedisInvalidationConfig{
    Address: "localhost:6379",
    DB:      0,
}
```

### Production Configuration

```go
Redis: &config.RedisInvalidationConfig{
    Address:     "redis.example.com:6379",
    Password:    "secure-password",
    DB:          1,
    Channel:     "myapp:cache:invalidation",
    PoolSize:    20,
    MaxRetries:  5,
    DialTimeout: 10 * time.Second,
}
```

### Redis Cluster Configuration

```go
Redis: &config.RedisInvalidationConfig{
    Address:     "redis-cluster.example.com:6379",
    Password:    "cluster-password",
    DB:          0,
    PoolSize:    50,
    MaxRetries:  3,
    DialTimeout: 5 * time.Second,
}
```

## Monitoring

You can monitor Redis pub/sub activity using Redis CLI:

```bash
# Monitor all pub/sub activity
redis-cli monitor

# Subscribe to invalidation channel
redis-cli subscribe invacache:invalidation

# Check active subscriptions
redis-cli pubsub channels
```

## Dependencies

This driver requires:
- `github.com/redis/go-redis/v9` - Redis Go client library

## Performance Considerations

- **Connection Pooling**: Configure `PoolSize` based on your load
- **Network Latency**: Consider Redis server location for minimal latency
- **Channel Names**: Use specific channel names to avoid cross-talk
- **Monitoring**: Monitor Redis memory usage and connection count

## Troubleshooting

### Connection Issues

```go
// Increase timeout for slow networks
DialTimeout: 30 * time.Second,

// Increase retries for unstable connections
MaxRetries: 10,
```

### High Load Scenarios

```go
// Increase pool size for high concurrency
PoolSize: 100,

// Use dedicated Redis instance for cache invalidation
DB: 15, // Use a separate database
```
