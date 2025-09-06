package main

import (
	"fmt"
	"time"

	"github.com/halilbulentorhon/invacache-go"
	"github.com/halilbulentorhon/invacache-go/backend/option"
	"github.com/halilbulentorhon/invacache-go/config"
)

type User struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
	Age  int    `json:"age"`
}

func main() {
	// Create cache configuration
	cfg := config.InvaCacheConfig{
		InMemory: config.InMemoryConfig{
			ShardCount:      16,               // Number of shards for concurrent access
			Capacity:        10000,            // Maximum number of items
			SweeperInterval: 30 * time.Second, // Cleanup interval
		},
	}

	// Example 1: String cache
	fmt.Println("=== String Cache Example ===")
	stringCache := invacache.NewInMemory[string](cfg)
	defer stringCache.Close() // Always cleanup resources

	// Set a value with TTL
	err := stringCache.Set("greeting", "Hello, World!", option.WithTTL(5*time.Minute))
	if err != nil {
		panic(err)
	}

	// Get the value
	greeting, err := stringCache.Get("greeting")
	if err != nil {
		fmt.Printf("Error getting greeting: %v\n", err)
	} else {
		fmt.Printf("Greeting: %s\n", greeting)
	}

	// Example 2: Struct cache with GetOrLoad
	fmt.Println("\n=== User Cache Example ===")
	userCache := invacache.NewInMemory[User](cfg)
	defer userCache.Close() // Always cleanup resources

	// GetOrLoad pattern - loads from "database" if not in cache
	user, err := userCache.GetOrLoad("user:123", func(key string) (User, time.Duration, error) {
		fmt.Printf("Loading user from database for key: %s\n", key)
		// Simulate database lookup
		return User{
			ID:   123,
			Name: "John Doe",
			Age:  30,
		}, 10 * time.Minute, nil // Cache for 10 minutes
	})
	if err != nil {
		panic(err)
	}
	fmt.Printf("User: %+v\n", user)

	// Second call should come from cache
	user2, err := userCache.GetOrLoad("user:123", func(key string) (User, time.Duration, error) {
		fmt.Printf("This should not be called - user should be in cache\n")
		return User{}, 0, fmt.Errorf("should not reach here")
	})
	if err != nil {
		panic(err)
	}
	fmt.Printf("User from cache: %+v\n", user2)

	// Example 3: Manual set and get
	fmt.Println("\n=== Manual Set/Get Example ===")
	adminUser := User{ID: 1, Name: "Admin", Age: 25}

	// Set without expiration
	err = userCache.Set("user:admin", adminUser, option.WithNoExpiration())
	if err != nil {
		panic(err)
	}

	// Get the admin user
	admin, err := userCache.Get("user:admin")
	if err != nil {
		fmt.Printf("Error getting admin: %v\n", err)
	} else {
		fmt.Printf("Admin user: %+v\n", admin)
	}

	// Example 4: Delete operation
	fmt.Println("\n=== Delete Example ===")
	err = userCache.Delete("user:admin")
	if err != nil {
		panic(err)
	}

	// Try to get deleted user
	_, err = userCache.Get("user:admin")
	if err != nil {
		fmt.Printf("Expected error after deletion: %v\n", err)
	}

	fmt.Println("\n=== Example completed successfully! ===")
}
