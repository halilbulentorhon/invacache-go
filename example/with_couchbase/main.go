package main

import (
	"fmt"
	"time"

	"github.com/halilbulentorhon/invacache-go"
	"github.com/halilbulentorhon/invacache-go/backend/option"
	"github.com/halilbulentorhon/invacache-go/config"

	// Import the Couchbase invalidation driver
	_ "github.com/halilbulentorhon/invacache-go/invalidation/drivers/couchbase"
)

type User struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
	Age  int    `json:"age"`
}

func main() {
	// Create cache configuration with Couchbase invalidation
	cfg := config.InvaCacheConfig{
		InMemory: config.InMemoryConfig{
			ShardCount:      16,               // Number of shards for concurrent access
			Capacity:        10000,            // Maximum number of items
			SweeperInterval: 30 * time.Second, // Cleanup interval
		},
		// Enable distributed invalidation using cb-pubsub
		Invalidation: &config.InvalidationConfig{
			Type: "couchbase",
			Couchbase: &config.CouchbaseInvalidationConfig{
				ConnectionString: "localhost:8091",
				Username:         "Administrator",
				Password:         "password",
				BucketName:       "PubSub",
				CollectionName:   "_default",
				ScopeName:        "_default",       // Optional, defaults to "_default"
				GroupName:        "invacache-demo", // Optional, defaults to "invacache"
			},
		},
	}

	fmt.Println("=== InvaCache with Couchbase Invalidation Demo ===")

	// Create cache instance
	cache := invacache.NewInMemory[User](cfg)
	defer cache.Close() // Always cleanup resources and stop background goroutines

	// Note: In a real scenario, you would have multiple instances of this cache
	// running on different servers, all connected to the same Couchbase cluster.
	// When one instance invalidates a key, all other instances will automatically
	// remove that key from their local cache.

	fmt.Println("Cache created with Couchbase invalidation support")

	// Set some users
	users := []User{
		{ID: 1, Name: "Alice", Age: 30},
		{ID: 2, Name: "Bob", Age: 25},
		{ID: 3, Name: "Charlie", Age: 35},
	}

	for _, user := range users {
		key := fmt.Sprintf("user:%d", user.ID)
		err := cache.Set(key, user, option.WithTTL(5*time.Minute))
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
	fmt.Println("\n=== Delete Demo (triggers invalidation) ===")
	err = cache.Delete("user:2")
	if err != nil {
		fmt.Printf("Error deleting user: %v\n", err)
	} else {
		fmt.Println("Deleted user:2 - invalidation message sent to all instances")
	}

	// Try to get deleted user
	_, err = cache.Get("user:2")
	if err != nil {
		fmt.Printf("Expected: user:2 not found after deletion: %v\n", err)
	}

	fmt.Println("\n=== Demo completed! ===")
	fmt.Println("Note: In a distributed setup, you would run multiple instances")
	fmt.Println("of this application on different servers. When one instance")
	fmt.Println("deletes or updates a key, all other instances will be notified")
	fmt.Println("via Couchbase and will remove the key from their local cache.")
}
