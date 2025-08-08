# RedisGK - Redis Library

The `redisgk` library provides a convenient wrapper over the official Redis client for Go, combining all the necessary logic for working with Redis in your projects with enhanced security and performance features.

## Features

- 🚀 **Easy to use** - convenient API for working with Redis
- 🔒 **Type safety** - support for generics for working with objects
- ⚡ **Performance** - optimized connection settings and goroutine pool
- 🛡️ **Security** - comprehensive input validation and nil checks
- 🔧 **Flexibility** - configurable timeouts and connection pool parameters
- 🔍 **Search** - search objects by key pattern with optimized processing
- 🗑️ **Bulk operations** - delete multiple keys in one call
- 🔔 **Notifications** - automatic key expiration notifications
- 📋 **Lists** - support for Redis list operations
- 🛡️ **Resource safety** - proper cleanup and goroutine management
- 🔍 **Error handling** - detailed error messages and validation

## Installation

```bash
go get github.com/GAFIKART/redisgk
```

## Quick Start

```go
package main

import (
    "log"
    "time"
    
    "github.com/GAFIKART/redisgk/lib"
)

type User struct {
    ID   int    `json:"id"`
    Name string `json:"name"`
    Age  int    `json:"age"`
}

func main() {
    // Redis connection configuration
    config := redisgklib.RedisConfConn{
        Host:     "localhost",
        Port:     6379,
        User:     "", // Optional
        Password: "your_password",
        DB:       0,
        AdditionalOptions: redisgklib.RedisAdditionalOptions{
            BaseCtx:      10 * time.Second,
            DialTimeout:  10 * time.Second,
            ReadTimeout:  30 * time.Second,
            WriteTimeout: 30 * time.Second,
            PoolSize:     20,
            PoolTimeout:  30 * time.Second,
        },
    }

    // Create Redis client
    redisClient, err := redisgklib.NewRedisGk(config)
    if err != nil {
        log.Fatal("Redis connection error:", err)
    }
    defer redisClient.Close()

    // Get key expiration notification channel
    expirationChan := redisClient.ListenChannelExpirationManager()
    go func() {
        for event := range expirationChan {
            log.Printf("Key expired: %s = '%s'", event.Key, event.Value)
        }
    }()

    // Save object
    user := User{ID: 1, Name: "John", Age: 25}
    err = redisgklib.SetObj(redisClient, []string{"users", "1"}, user, 1*time.Hour)
    if err != nil {
        log.Fatal("Save error:", err)
    }

    // Get object
    retrievedUser, err := redisgklib.GetObj[User](redisClient, []string{"users", "1"})
    if err != nil {
        log.Fatal("Get error:", err)
    }
    log.Printf("Retrieved user: %+v", *retrievedUser)

    // Work with strings
    err = redisClient.SetString([]string{"greeting"}, "Hello, world!", 30*time.Minute)
    if err != nil {
        log.Fatal("String save error:", err)
    }

    greeting, err := redisClient.GetString([]string{"greeting"})
    if err != nil {
        log.Fatal("String get error:", err)
    }
    log.Println("Greeting:", greeting)

    // Work with lists
    err = redisClient.LPush([]string{"queue", "tasks"}, "task 1", "task 2")
    if err != nil {
        log.Fatal("List add error:", err)
    }

    task, err := redisClient.LPop([]string{"queue", "tasks"})
    if err != nil {
        log.Fatal("List get error:", err)
    }
    log.Println("Retrieved task:", task)

    // Search objects by pattern
    users, err := redisgklib.FindObj[User](redisClient, []string{"users"}, 100)
    if err != nil {
        log.Fatal("Search error:", err)
    }
    log.Printf("Found users: %d", len(users))

    // Check key existence
    exists, err := redisClient.Exists([]string{"users", "1"})
    if err != nil {
        log.Fatal("Check error:", err)
    }
    log.Printf("Key exists: %t", exists)

    // Get list of keys
    keys, err := redisClient.GetKeys([]string{"users"})
    if err != nil {
        log.Fatal("Get keys error:", err)
    }
    log.Printf("Found keys: %v", keys)

    // Delete multiple keys in one call
    err = redisClient.Del(
        []string{"users", "1"},
        []string{"greeting"},
    )
    if err != nil {
        log.Fatal("Delete error:", err)
    }
}
```

## Usage Examples

Complete examples of using all library methods are available in the [`example/`](./example/) folder:

```bash
cd example
go run main.go
```

Examples demonstrate:
- Working with strings and objects
- Key expiration notifications
- Working with lists
- Searching objects by pattern
- Checking key existence
- Getting list of keys
- Bulk key deletion
- Error handling and validation

## API

### Main Functions

#### `NewRedisGk(config RedisConfConn) (*RedisGk, error)`
Creates a new Redis client instance with automatic key expiration notification setup. Includes comprehensive validation and security checks.

#### `SetObj[T any](client *RedisGk, keyPath []string, value T, ttl ...time.Duration) error`
Saves an object to Redis with automatic JSON serialization. Includes data size validation and nil checks.

#### `GetObj[T any](client *RedisGk, keyPath []string) (*T, error)`
Gets an object from Redis with automatic JSON deserialization. Handles missing keys gracefully.

#### `FindObj[T any](client *RedisGk, patternPath []string, count ...int64) (map[string]*T, error)`
Search objects by key pattern with optimized processing and goroutine safety.

### RedisGk Methods

#### Strings
- `SetString(keyPath []string, value string, ttl ...time.Duration) error`
- `GetString(keyPath []string) (string, error)`

#### Lists
- `LPush(keyPath []string, values ...string) error` - add to beginning of list
- `RPush(keyPath []string, values ...string) error` - add to end of list
- `LPop(keyPath []string) (string, error)` - get first element
- `RPop(keyPath []string) (string, error)` - get last element
- `LRange(keyPath []string, start, stop int64) ([]string, error)` - get range
- `LLen(keyPath []string) (int64, error)` - get list length

#### Key Management
- `Del(keyPath ...[]string) error` - delete one or multiple keys
- `Exists(key []string) (bool, error)` - check key existence
- `GetKeys(patternPath []string) ([]string, error)` - get list of keys

#### Expiration Notifications
- `ListenChannelExpirationManager() <-chan KeyExpirationEvent` - get notification channel

#### Connection Management
- `Close() error` - close Redis connection with proper cleanup

## Configuration

### RedisConfConn
```go
type RedisConfConn struct {
    Host     string
    Port     int
    User     string        // Optional
    Password string
    DB       int
    AdditionalOptions RedisAdditionalOptions
}
```

### RedisAdditionalOptions
```go
type RedisAdditionalOptions struct {
    DialTimeout  time.Duration
    ReadTimeout  time.Duration
    WriteTimeout time.Duration
    PoolSize     int
    PoolTimeout  time.Duration
    BaseCtx      time.Duration
}
```

## Security Features

### Input Validation
- **Nil checks** - All methods validate input parameters
- **Configuration validation** - Comprehensive Redis connection validation
- **Data size limits** - Maximum 512 MB for keys and values
- **Domain validation** - Proper hostname and IP address validation
- **Key normalization** - Automatic key sanitization and normalization

### Resource Safety
- **Goroutine management** - Proper cleanup of background goroutines
- **Channel safety** - Safe channel operations with nil checks
- **Context handling** - Proper context cancellation and timeout management
- **Connection cleanup** - Graceful shutdown of Redis connections

### Error Handling
- **Detailed error messages** - Comprehensive error information
- **Graceful degradation** - Proper handling of missing keys and network issues
- **Validation errors** - Clear feedback for invalid inputs

## Features

### Key Processing
- Automatic key normalization (removing special characters)
- Replacing spaces with underscores
- Support for hierarchical keys via string slice
- Key size limit of 512 MB
- Input validation and sanitization

### Data Processing
- Automatic object serialization/deserialization to JSON
- Data size validation (maximum 512 MB)
- Handling `redis.Nil` error when key is missing
- Comprehensive error handling

### Performance
- Configurable connection pool
- Contexts with timeouts for all operations
- Efficient multiple key deletion
- Goroutine pool for key expiration notification processing
- Optimized object search processing with proper cleanup

### Key Expiration Notifications
- Automatic Redis configuration for notifications
- Unbuffered channel for synchronous event transmission
- Goroutine pool for event processing
- Guaranteed delivery of all notifications
- Graceful shutdown when closing connection
- Thread-safe operations with mutex protection

### Bulk Operations
The `Del()` method supports deleting multiple keys in one call:
```go
err := redisClient.Del(
    []string{"users", "1"},
    []string{"users", "2"},
    []string{"temp", "session"},
)
```

### List Operations
Enhanced list operations with validation:
```go
// Add elements to list
err := redisClient.LPush([]string{"queue"}, "task1", "task2")

// Get elements from list
task, err := redisClient.LPop([]string{"queue"})

// Get list range
items, err := redisClient.LRange([]string{"queue"}, 0, -1)
```

## Requirements

- Go 1.24.0+
- Redis server version 2.8.0+

## License

MIT License
