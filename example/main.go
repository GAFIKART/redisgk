package main

import (
	"context"
	"fmt"
	"log"
	"time"

	redisgklib "github.com/GAFIKART/redisgk/lib"
)

func main() {
	// Create Redis configuration
	config := redisgklib.RedisConfConn{
		Host:     "localhost",
		Port:     6379,
		Password: "",
		DB:       0,
	}

	// Create RedisGk instance
	redisGk, err := redisgklib.NewRedisGk(config)
	if err != nil {
		log.Fatalf("Failed to create RedisGk: %v", err)
	}
	defer redisGk.Close()

	fmt.Println("RedisGk initialized successfully")
	fmt.Println("Listening for key events via keyevent channels...")
	fmt.Println("Channels: __keyevent@0__:expire, __keyevent@0__:expired, __keyevent@0__:set, __keyevent@0__:del")

	// Get channel for listening to events
	eventChan := redisGk.ListenChannelKeyEventManager()

	// Create test key with TTL
	client := redisGk.GetRedisClient()
	ctx := context.Background()

	// Set key with 5 second TTL
	err = client.Set(ctx, "test_key", "test_value", 5*time.Second).Err()
	if err != nil {
		log.Printf("Failed to set test key: %v", err)
	} else {
		fmt.Println("Test key 'test_key' set with 5 second TTL")
	}

	// Create another key without TTL to demonstrate creation events
	err = client.Set(ctx, "test_created", "created_value", 0).Err()
	if err != nil {
		log.Printf("Failed to set created key: %v", err)
	} else {
		fmt.Println("Test key 'test_created' created without TTL")
	}

	// Listen for events
	fmt.Println("\nListening for events...")
	fmt.Println("Expected sequence:")
	fmt.Println("1. TTL setting event (EventTypeExpire)")
	fmt.Println("2. Key expiration event (EventTypeExpired)")

	for {
		select {
		case event := <-eventChan:
			fmt.Printf("\n=== Event received ===\n")
			fmt.Printf("Key: %s\n", event.Key)
			fmt.Printf("Value: %s\n", event.Value)
			fmt.Printf("Event Type: %s\n", event.EventType)
			fmt.Printf("Timestamp: %s\n", event.Timestamp)
			fmt.Println("===================")

			// Filter events by type
			switch event.EventType {
			case redisgklib.EventTypeExpire:
				fmt.Printf("⏰ EXPIRE: TTL was set for key '%s'\n", event.Key)
			case redisgklib.EventTypeExpired:
				fmt.Printf("🔴 EXPIRED: Key '%s' has expired\n", event.Key)
				if event.Key == "test_key" {
					fmt.Println("Test key expired, exiting...")
					return
				}
			case redisgklib.EventTypeCreated:
				fmt.Printf("🟢 CREATED: Key '%s' was created/updated\n", event.Key)
			case redisgklib.EventTypeDeleted:
				fmt.Printf("🗑️ DELETED: Key '%s' was deleted\n", event.Key)
			default:
				fmt.Printf("❓ UNKNOWN: Key '%s' event type unknown\n", event.Key)
			}

		case <-time.After(10 * time.Second):
			fmt.Println("Timeout waiting for events")
			return
		}
	}
}
