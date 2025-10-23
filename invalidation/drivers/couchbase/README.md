# Couchbase Invalidation Driver

This driver provides distributed cache invalidation using Couchbase and the [cb-pubsub](https://github.com/halilbulentorhon/cb-pubsub) library.

## Installation

```bash
go get github.com/halilbulentorhon/invacache-go/invalidation/drivers/couchbase
```

## Usage

```go
package main

import (
    "github.com/halilbulentorhon/invacache-go"
    "github.com/halilbulentorhon/invacache-go/config"
    
    // Import the Couchbase invalidation driver
    _ "github.com/halilbulentorhon/invacache-go/invalidation/drivers/couchbase"
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
    defer cache.Close() // Always cleanup resources
    
    // Cache will now use Couchbase for distributed invalidation
}
```

## Configuration

```go
type CouchbaseInvalidationConfig struct {
    ConnectionString string `json:"connectionString"` // Couchbase connection string (e.g., "localhost:8091")
    Username         string `json:"username"`         // Couchbase username
    Password         string `json:"password"`         // Couchbase password
    BucketName       string `json:"bucketName"`       // Bucket name for invalidation messages
    CollectionName   string `json:"collectionName"`   // Collection name for invalidation messages
    ScopeName        string `json:"scopeName"`        // Optional, defaults to "_default"
    GroupName        string `json:"groupName"`        // Optional, defaults to "invacache"
}
```

## How It Works

1. **PubSub Pattern**: Uses cb-pubsub library for reliable message delivery
2. **Instance Registration**: Each cache instance registers itself in Couchbase
3. **Message Broadcasting**: When a key is invalidated, message is sent to all instances in the group
4. **Automatic Cleanup**: Inactive instances are automatically cleaned up
5. **TTL Management**: Documents have TTL to prevent accumulation

## Benefits

- **Reliable**: Built on proven Couchbase infrastructure
- **Scalable**: Supports multiple cache instances across different servers
- **Efficient**: Uses Couchbase's native pub/sub capabilities
- **Fault Tolerant**: Handles network failures and reconnections gracefully

## Dependencies

This driver requires:
- `github.com/halilbulentorhon/cb-pubsub` - The underlying PubSub library
- `github.com/couchbase/gocb/v2` - Couchbase Go SDK (via cb-pubsub)
- `github.com/google/uuid` - UUID generation (via cb-pubsub)
