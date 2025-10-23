package main

import (
	"fmt"
	"time"

	"github.com/halilbulentorhon/invacache-go"
	"github.com/halilbulentorhon/invacache-go/backend/option"
	"github.com/halilbulentorhon/invacache-go/config"

	// Import the Redis invalidation driver
	_ "github.com/halilbulentorhon/invacache-go/invalidation/drivers/redis"
)

type User struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
	Age  int    `json:"age"`
}

func main() {
	// Create cache configuration with Redis invalidation
	cfg := config.InvaCacheConfig{
		BackendName: "in-memory",
		Backend: &config.BackendConfig{InMemory: &config.InMemoryConfig{
			ShardCount:      16,               // Number of shards for concurrent access
			Capacity:        10000,            // Maximum number of items
			SweeperInterval: 30 * time.Second, // Cleanup interval
		}},
		// Enable distributed invalidation using Redis Pub/Sub
		Invalidation: &config.InvalidationConfig{
			Type: "redis",
			DriverConfig: map[string]any{
				"Address":     "localhost:6379",
				"Password":    "",
				"DB":          0,
				"Channel":     "invacache:demo",
				"PoolSize":    10,
				"MaxRetries":  3,
				"DialTimeout": 5 * time.Second,
			},
		},
	}

	fmt.Println("=== InvaCache with Redis Invalidation Demo ===")

	// Create cache instance
	cache, err := invacache.NewCache[User](cfg)
	if err != nil {
		panic(err)
	}
	defer cache.Close() // Always cleanup resources and stop background goroutines

	// Note: In a real scenario, you would have multiple instances of this cache
	// running on different servers, all connected to the same Redis server.
	// When one instance invalidates a key, all other instances will automatically
	// receive the invalidation message via Redis Pub/Sub and remove that key
	// from their local cache.

	fmt.Println("Cache created with Redis invalidation support")

	// Set some users
	users := []User{
		{ID: 1, Name: "Alice", Age: 30},
		{ID: 2, Name: "Bob", Age: 25},
		{ID: 3, Name: "Charlie", Age: 35},
	}

	for _, user := range users {
		key := fmt.Sprintf("user:%d", user.ID)
		err = cache.Set(key, user, option.WithTTL(5*time.Minute))
		if err != nil {
			fmt.Printf("Error setting user %d: %v\n", user.ID, err)
		} else {
			fmt.Printf("Set user: %s (ID: %d)\n", user.Name, user.ID)
		}
	}

	// Get users
	fmt.Println("\n=== Getting Users ===")
	for i := 1; i <= 3; i++ {
		key := fmt.Sprintf("user:%d", i)
		user, err := cache.Get(key)
		if err != nil {
			fmt.Printf("Error getting user %d: %v\n", i, err)
		} else {
			fmt.Printf("Got user: %s (ID: %d, Age: %d)\n", user.Name, user.ID, user.Age)
		}
	}

	// Demonstrate GetOrLoad
	fmt.Println("\n=== GetOrLoad Demo ===")
	user, err := cache.GetOrLoad("user:999", func(key string) (User, time.Duration, error) {
		fmt.Printf("Loading user from database for key: %s\n", key)
		return User{
			ID:   999,
			Name: "Dynamic User",
			Age:  40,
		}, 10 * time.Minute, nil
	})
	if err != nil {
		fmt.Printf("Error in GetOrLoad: %v\n", err)
	} else {
		fmt.Printf("Loaded user: %s (ID: %d)\n", user.Name, user.ID)
	}

	// Delete a user (this will trigger invalidation across all instances)
	fmt.Println("\n=== Delete Demo (triggers Redis invalidation) ===")
	err = cache.Delete("user:2")
	if err != nil {
		fmt.Printf("Error deleting user: %v\n", err)
	} else {
		fmt.Println("Deleted user:2 - invalidation message sent via Redis Pub/Sub")
	}

	// Try to get deleted user
	_, err = cache.Get("user:2")
	if err != nil {
		fmt.Printf("Expected: user:2 not found after deletion: %v\n", err)
	}

	// Update a user (this will also trigger invalidation)
	fmt.Println("\n=== Update Demo (triggers Redis invalidation) ===")
	updatedUser := User{ID: 1, Name: "Alice Updated", Age: 31}
	err = cache.Set("user:1", updatedUser, option.WithTTL(10*time.Minute))
	if err != nil {
		fmt.Printf("Error updating user: %v\n", err)
	} else {
		fmt.Println("Updated user:1 - invalidation message sent via Redis Pub/Sub")
	}

	// Verify the update
	user, err = cache.Get("user:1")
	if err != nil {
		fmt.Printf("Error getting updated user: %v\n", err)
	} else {
		fmt.Printf("Updated user: %s (ID: %d, Age: %d)\n", user.Name, user.ID, user.Age)
	}

	fmt.Println("\n=== Demo completed! ===")
	fmt.Println("Note: In a distributed setup, you would run multiple instances")
	fmt.Println("of this application on different servers. When one instance")
	fmt.Println("deletes or updates a key, all other instances will be notified")
	fmt.Println("via Redis Pub/Sub and will remove the key from their local cache.")
	fmt.Println("\nTo test distributed invalidation:")
	fmt.Println("1. Start Redis server: docker run -d -p 6379:6379 redis:alpine")
	fmt.Println("2. Run this example in multiple terminals")
	fmt.Println("3. Observe how changes in one instance affect others")
}
