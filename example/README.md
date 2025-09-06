# InvaCache-Go Examples

This directory contains various examples demonstrating how to use InvaCache-Go.

## Examples

### Basic Usage
- **Directory**: `basic/`
- **Description**: Basic cache operations without invalidation
- **Features**: Set, Get, GetOrLoad, Delete operations with TTL
- **Run**: `cd basic && go run main.go`

### Couchbase Invalidation
- **Directory**: `with_couchbase/`
- **Description**: Distributed cache invalidation using Couchbase driver
- **Features**: Multi-instance cache synchronization via Couchbase
- **Requirements**: Couchbase server running on localhost:8091
- **Run**: `cd with_couchbase && go run main.go`

### Redis Invalidation
- **Directory**: `with_redis/`
- **Description**: Distributed cache invalidation using Redis Pub/Sub
- **Features**: Real-time multi-instance cache synchronization via Redis
- **Requirements**: Redis server running on localhost:6379
- **Run**: `cd with_redis && go run main.go`


## Running Examples

Each example has its own `go.mod` file to avoid package conflicts:

```bash
# Basic example (zero dependencies)
cd example/basic
go run main.go

# Couchbase example (requires Couchbase driver)
cd example/with_couchbase
go run main.go

# Redis example (requires Redis driver)
cd example/with_redis
go run main.go
```

## Creating New Examples

When creating new examples:

1. Create a new directory under `example/`
2. Add a `go.mod` file with appropriate dependencies
3. Use `main` package in your Go files
4. Add replace directives for local development

Example `go.mod`:
```go
module github.com/halilbulentorhon/invacache-go/example/my-example

go 1.21

require github.com/halilbulentorhon/invacache-go v0.0.0

replace github.com/halilbulentorhon/invacache-go => ../../
```
