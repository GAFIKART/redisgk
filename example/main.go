package main

import (
	"fmt"
	"log"
	"time"

	redisgklib "github.com/GAFIKART/redisgk/lib"
)

// User - example structure for demonstrating object operations
type User struct {
	ID       int    `json:"id"`
	Name     string `json:"name"`
	Email    string `json:"email"`
	Age      int    `json:"age"`
	IsActive bool   `json:"is_active"`
}

// Product - another structure for demonstration
type Product struct {
	ID          string  `json:"id"`
	Name        string  `json:"name"`
	Price       float64 `json:"price"`
	Description string  `json:"description"`
	Category    string  `json:"category"`
}

func main() {
	fmt.Println("=== RedisGK Library Usage Examples ===")

	// Redis connection configuration
	config := redisgklib.RedisConfConn{
		Host:     "localhost",
		Port:     6379,
		User:     "",
		Password: "",
		DB:       0,
		AdditionalOptions: redisgklib.RedisAdditionalOptions{
			DialTimeout:  5 * time.Second,
			ReadTimeout:  3 * time.Second,
			WriteTimeout: 3 * time.Second,
			PoolSize:     10,
			PoolTimeout:  30 * time.Second,
			BaseCtx:      10 * time.Second,
		},
	}

	// Create RedisGk instance with automatic initialization
	redisGk, err := redisgklib.NewRedisGk(config)
	if err != nil {
		log.Fatalf("Redis connection error: %v", err)
	}
	defer func() {
		if err := redisGk.Close(); err != nil {
			log.Printf("Connection close error: %v", err)
		}
	}()

	// ========================================
	// EXAMPLE 1: Getting notification channel
	// ========================================
	fmt.Println("\n=== EXAMPLE 1: Getting notification channel ===")

	// Get notification channel
	expirationChan := redisGk.ListenChannelExpirationManager()

	// Start goroutine for processing notifications
	go func() {
		for event := range expirationChan {
			fmt.Printf("📢 Key expired: %s = '%s' at %s\n",
				event.Key, event.Value, event.ExpiredAt.Format("15:04:05"))
		}
	}()

	fmt.Println("✅ Channel listener started")

	// ========================================
	// EXAMPLE 2: Data operations demonstration
	// ========================================
	fmt.Println("\n=== EXAMPLE 2: Data operations ===")

	// Example 1: Working with strings
	fmt.Println("=== 1. STRING OPERATIONS ===")
	demoStrings(redisGk)

	// Example 2: Working with objects
	fmt.Println("\n=== 2. OBJECT OPERATIONS ===")
	demoObjects(redisGk)

	// Example 3: Object search
	fmt.Println("\n=== 3. OBJECT SEARCH ===")
	demoFindObjects(redisGk)

	// Example 4: Key existence check
	fmt.Println("\n=== 4. KEY EXISTENCE CHECK ===")
	demoExists(redisGk)

	// Example 5: Key deletion
	fmt.Println("\n=== 5. KEY DELETION ===")
	demoDelete(redisGk)

	// Example 6: Getting key list
	fmt.Println("\n=== 6. GETTING KEY LIST ===")
	demoGetKeys(redisGk)

	// Example 7: List operations
	fmt.Println("\n=== 7. LIST OPERATIONS ===")
	demoListOperations(redisGk)

	// ========================================
	// EXAMPLE 3: Key expiration notification demonstration
	// ========================================
	fmt.Println("\n=== EXAMPLE 3: Notification demonstration ===")

	// Create test keys with TTL for demonstration
	testKeys := []struct {
		key   string
		value string
		ttl   time.Duration
	}{
		{"demo:expire:1", "value 1", 3 * time.Second},
		{"demo:expire:2", "value 2", 5 * time.Second},
		{"demo:expire:3", "value 3", 7 * time.Second},
		{"demo:expire:4", "value 4", 10 * time.Second},
		{"demo:expire:5", "value 5", 15 * time.Second},
	}

	fmt.Println("📝 Creating test keys with TTL...")
	for _, testKey := range testKeys {
		err := redisGk.SetString([]string{testKey.key}, testKey.value, testKey.ttl)
		if err != nil {
			log.Printf("Error creating key %s: %v", testKey.key, err)
			continue
		}
		fmt.Printf("✅ Key created: %s = '%s' (TTL: %v)\n", testKey.key, testKey.value, testKey.ttl)
	}

	fmt.Println("\n⏳ Waiting for keys to expire...")
	fmt.Println("(Press Ctrl+C to exit)")

	// Wait for program completion
	select {}
}

// demoStrings - string operations demonstration
func demoStrings(redisGk *redisgklib.RedisGk) {
	fmt.Println("📝 Saving string...")

	// Save string without TTL
	err := redisGk.SetString([]string{"user", "profile", "name"}, "John Smith")
	if err != nil {
		log.Printf("String save error: %v", err)
		return
	}
	fmt.Println("✅ String saved: user:profile:name = 'John Smith'")

	// Save string with TTL
	err = redisGk.SetString([]string{"temp", "session", "token"}, "abc123xyz", 30*time.Second)
	if err != nil {
		log.Printf("String save with TTL error: %v", err)
		return
	}
	fmt.Println("✅ String with TTL saved: temp:session:token = 'abc123xyz' (TTL: 30s)")

	// Get string
	value, err := redisGk.GetString([]string{"user", "profile", "name"})
	if err != nil {
		log.Printf("String get error: %v", err)
		return
	}
	fmt.Printf("✅ String retrieved: %s\n", value)

	// Try to get non-existent key
	_, err = redisGk.GetString([]string{"nonexistent", "key"})
	if err != nil {
		fmt.Printf("✅ Expected error for non-existent key: %v\n", err)
	}
}

// demoObjects - object operations demonstration
func demoObjects(redisGk *redisgklib.RedisGk) {
	fmt.Println("📦 Saving User object...")

	// Create User object
	user := User{
		ID:       1,
		Name:     "Anna Sidorova",
		Email:    "anna@example.com",
		Age:      28,
		IsActive: true,
	}

	// Save object
	err := redisgklib.SetObj(redisGk, []string{"users", "1"}, user)
	if err != nil {
		log.Printf("Object save error: %v", err)
		return
	}
	fmt.Println("✅ User object saved: users:1")

	// Save object with TTL
	product := Product{
		ID:          "prod_001",
		Name:        "Galaxy S21 Smartphone",
		Price:       89999.99,
		Description: "Powerful smartphone with excellent camera",
		Category:    "Electronics",
	}

	err = redisgklib.SetObj(redisGk, []string{"products", "prod_001"}, product, 1*time.Hour)
	if err != nil {
		log.Printf("Object save with TTL error: %v", err)
		return
	}
	fmt.Println("✅ Product object saved with TTL: products:prod_001 (TTL: 1h)")

	// Get User object
	retrievedUser, err := redisgklib.GetObj[User](redisGk, []string{"users", "1"})
	if err != nil {
		log.Printf("User object get error: %v", err)
		return
	}
	fmt.Printf("✅ User object retrieved: ID=%d, Name=%s, Email=%s\n",
		retrievedUser.ID, retrievedUser.Name, retrievedUser.Email)

	// Get Product object
	retrievedProduct, err := redisgklib.GetObj[Product](redisGk, []string{"products", "prod_001"})
	if err != nil {
		log.Printf("Product object get error: %v", err)
		return
	}
	fmt.Printf("✅ Product object retrieved: ID=%s, Name=%s, Price=%.2f\n",
		retrievedProduct.ID, retrievedProduct.Name, retrievedProduct.Price)
}

// demoFindObjects - object search demonstration
func demoFindObjects(redisGk *redisgklib.RedisGk) {
	fmt.Println("🔍 Searching objects...")

	// Create several users for search
	users := []User{
		{ID: 2, Name: "Peter Ivanov", Email: "petr@example.com", Age: 35, IsActive: true},
		{ID: 3, Name: "Maria Kozlova", Email: "maria@example.com", Age: 24, IsActive: false},
		{ID: 4, Name: "Sergey Volkov", Email: "sergey@example.com", Age: 42, IsActive: true},
	}

	// Save users
	for _, user := range users {
		err := redisgklib.SetObj(redisGk, []string{"users", fmt.Sprintf("%d", user.ID)}, user)
		if err != nil {
			log.Printf("Error saving user %d: %v", user.ID, err)
			continue
		}
		fmt.Printf("✅ User %d saved\n", user.ID)
	}

	// Search all users
	foundUsers, err := redisgklib.FindObj[User](redisGk, []string{"users"})
	if err != nil {
		log.Printf("User search error: %v", err)
		return
	}

	fmt.Printf("✅ Found users: %d\n", len(foundUsers))
	for key, user := range foundUsers {
		fmt.Printf("   - %s: %s (%s)\n", key, user.Name, user.Email)
	}

	// Search with result count limit
	foundUsersLimited, err := redisgklib.FindObj[User](redisGk, []string{"users"}, 2)
	if err != nil {
		log.Printf("Limited user search error: %v", err)
		return
	}

	fmt.Printf("✅ Found users (limit 2): %d\n", len(foundUsersLimited))
}

// demoExists - key existence check demonstration
func demoExists(redisGk *redisgklib.RedisGk) {
	fmt.Println("🔍 Checking key existence...")

	// Check existing key
	exists, err := redisGk.Exists([]string{"users", "1"})
	if err != nil {
		log.Printf("Key existence check error: %v", err)
		return
	}
	fmt.Printf("✅ Key 'users:1' exists: %t\n", exists)

	// Check non-existent key
	exists, err = redisGk.Exists([]string{"nonexistent", "key"})
	if err != nil {
		log.Printf("Key existence check error: %v", err)
		return
	}
	fmt.Printf("✅ Key 'nonexistent:key' exists: %t\n", exists)
}

// demoDelete - key deletion demonstration
func demoDelete(redisGk *redisgklib.RedisGk) {
	fmt.Println("🗑️ Deleting keys...")

	// Create test key for deletion
	err := redisGk.SetString([]string{"test", "delete", "key"}, "value for deletion")
	if err != nil {
		log.Printf("Test key creation error: %v", err)
		return
	}
	fmt.Println("✅ Test key created: test:delete:key")

	// Check existence before deletion
	exists, err := redisGk.Exists([]string{"test", "delete", "key"})
	if err != nil {
		log.Printf("Existence check error: %v", err)
		return
	}
	fmt.Printf("✅ Key exists before deletion: %t\n", exists)

	// Delete key
	err = redisGk.Del([]string{"test", "delete", "key"})
	if err != nil {
		log.Printf("Key deletion error: %v", err)
		return
	}
	fmt.Println("✅ Key deleted: test:delete:key")

	// Check existence after deletion
	exists, err = redisGk.Exists([]string{"test", "delete", "key"})
	if err != nil {
		log.Printf("Existence check error: %v", err)
		return
	}
	fmt.Printf("✅ Key exists after deletion: %t\n", exists)

	// Delete multiple keys
	fmt.Println("🗑️ Deleting all created keys...")

	// Delete all keys in one call
	err = redisGk.Del(
		[]string{"user", "profile", "name"},
		[]string{"temp", "session", "token"},
		[]string{"users", "1"},
		[]string{"users", "2"},
		[]string{"users", "3"},
		[]string{"users", "4"},
		[]string{"products", "prod_001"},
	)
	if err != nil {
		log.Printf("Keys deletion error: %v", err)
	} else {
		fmt.Println("✅ All keys deleted successfully")
	}
}

// demoGetKeys - getting key list demonstration
func demoGetKeys(redisGk *redisgklib.RedisGk) {
	fmt.Println("🔍 Getting key list...")

	// Get list of all keys
	keys, err := redisGk.GetKeys([]string{})
	if err != nil {
		log.Printf("Key list retrieval error: %v", err)
		return
	}

	fmt.Printf("✅ Found keys: %d\n", len(keys))
	for _, key := range keys {
		fmt.Println("   -", key)
	}

	// Get keys by pattern
	userKeys, err := redisGk.GetKeys([]string{"users"})
	if err != nil {
		log.Printf("User keys retrieval error: %v", err)
		return
	}

	fmt.Printf("✅ Found user keys: %d\n", len(userKeys))
	for _, key := range userKeys {
		fmt.Println("   -", key)
	}
}

// demoListOperations - list operations demonstration
func demoListOperations(redisGk *redisgklib.RedisGk) {
	fmt.Println("📋 List operations demonstration...")

	// Create a list
	err := redisGk.LPush([]string{"queue", "tasks"}, "task 1", "task 2", "task 3")
	if err != nil {
		log.Printf("List creation error: %v", err)
		return
	}
	fmt.Println("✅ List created with tasks")

	// Add more items to the end
	err = redisGk.RPush([]string{"queue", "tasks"}, "task 4", "task 5")
	if err != nil {
		log.Printf("List append error: %v", err)
		return
	}
	fmt.Println("✅ Added more tasks to the end")

	// Get list length
	length, err := redisGk.LLen([]string{"queue", "tasks"})
	if err != nil {
		log.Printf("List length error: %v", err)
		return
	}
	fmt.Printf("✅ List length: %d\n", length)

	// Get first item
	firstTask, err := redisGk.LPop([]string{"queue", "tasks"})
	if err != nil {
		log.Printf("List pop error: %v", err)
		return
	}
	fmt.Printf("✅ Retrieved first task: %s\n", firstTask)

	// Get last item
	lastTask, err := redisGk.RPop([]string{"queue", "tasks"})
	if err != nil {
		log.Printf("List pop error: %v", err)
		return
	}
	fmt.Printf("✅ Retrieved last task: %s\n", lastTask)

	// Get range of items
	items, err := redisGk.LRange([]string{"queue", "tasks"}, 0, -1)
	if err != nil {
		log.Printf("List range error: %v", err)
		return
	}
	fmt.Printf("✅ Remaining tasks: %v\n", items)

	// Clean up
	err = redisGk.Del([]string{"queue", "tasks"})
	if err != nil {
		log.Printf("List cleanup error: %v", err)
	} else {
		fmt.Println("✅ List cleaned up")
	}
}
