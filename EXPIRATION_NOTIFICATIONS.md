# Key Expiration Notifications

This document describes the key expiration notification system implemented in the RedisGK library.

## Overview

The RedisGK library provides automatic key expiration notifications that allow you to receive real-time events when keys expire in Redis. This feature is useful for:

- Cleaning up related resources when keys expire
- Logging expiration events for monitoring
- Implementing custom expiration logic
- Building event-driven applications

## How It Works

### 1. Automatic Setup

When you create a new RedisGk instance, the library automatically:

1. **Checks Redis Configuration**: Verifies if `notify-keyspace-events` is configured to include expiration events (`E`)
2. **Configures Redis**: If not configured, automatically sets `notify-keyspace-events` to `Ex`
3. **Creates Subscription**: Subscribes to the `__keyevent@0__:expired` channel
4. **Starts Listener**: Begins listening for expiration events in a background goroutine

### 2. Event Processing

The expiration manager:

- **Captures Events**: Listens for expiration notifications from Redis
- **Retrieves Values**: Attempts to get the key's value before it expires (with 50ms timeout)
- **Creates Events**: Packages the event with key, value, and expiration timestamp
- **Delivers Events**: Sends events through an unbuffered channel to ensure delivery

### 3. Thread Safety

The system uses:

- **Mutex Protection**: Thread-safe operations for starting/stopping the listener
- **WaitGroup**: Proper goroutine cleanup
- **Context Cancellation**: Graceful shutdown when closing the connection

## Usage

### Basic Example

```go
package main

import (
    "fmt"
    "log"
    "time"
    
    "github.com/GAFIKART/redisgk/lib"
)

func main() {
    // Create Redis client
    config := redisgklib.RedisConfConn{
        Host:     "localhost",
        Port:     6379,
        Password: "your_password",
        DB:       0,
    }
    
    redisClient, err := redisgklib.NewRedisGk(config)
    if err != nil {
        log.Fatal("Redis connection error:", err)
    }
    defer redisClient.Close()

    // Get expiration notification channel
    expirationChan := redisClient.ListenChannelExpirationManager()

    // Start listening for expiration events
    go func() {
        for event := range expirationChan {
            fmt.Printf("Key expired: %s = '%s' at %s\n",
                event.Key, event.Value, event.ExpiredAt.Format("2006-01-02 15:04:05"))
        }
    }()

    // Create some keys with TTL for testing
    err = redisClient.SetString([]string{"test:key1"}, "value1", 5*time.Second)
    if err != nil {
        log.Printf("Error setting key: %v", err)
    }

    err = redisClient.SetString([]string{"test:key2"}, "value2", 10*time.Second)
    if err != nil {
        log.Printf("Error setting key: %v", err)
    }

    // Wait for keys to expire
    time.Sleep(15 * time.Second)
}
```

### Advanced Example

```go
package main

import (
    "fmt"
    "log"
    "sync"
    "time"
    
    "github.com/GAFIKART/redisgk/lib"
)

type ExpirationHandler struct {
    mu sync.RWMutex
    handlers map[string]func(string, string)
}

func NewExpirationHandler() *ExpirationHandler {
    return &ExpirationHandler{
        handlers: make(map[string]func(string, string)),
    }
}

func (eh *ExpirationHandler) RegisterHandler(pattern string, handler func(string, string)) {
    eh.mu.Lock()
    defer eh.mu.Unlock()
    eh.handlers[pattern] = handler
}

func (eh *ExpirationHandler) HandleExpiration(event redisgklib.KeyExpirationEvent) {
    eh.mu.RLock()
    defer eh.mu.RUnlock()
    
    for pattern, handler := range eh.handlers {
        if strings.HasPrefix(event.Key, pattern) {
            handler(event.Key, event.Value)
        }
    }
}

func main() {
    // Create Redis client
    config := redisgklib.RedisConfConn{
        Host:     "localhost",
        Port:     6379,
        Password: "your_password",
        DB:       0,
    }
    
    redisClient, err := redisgklib.NewRedisGk(config)
    if err != nil {
        log.Fatal("Redis connection error:", err)
    }
    defer redisClient.Close()

    // Create expiration handler
    handler := NewExpirationHandler()
    
    // Register handlers for different key patterns
    handler.RegisterHandler("session:", func(key, value string) {
        fmt.Printf("Session expired: %s\n", key)
        // Clean up session-related resources
    })
    
    handler.RegisterHandler("cache:", func(key, value string) {
        fmt.Printf("Cache entry expired: %s\n", key)
        // Update cache statistics
    })
    
    handler.RegisterHandler("temp:", func(key, value string) {
        fmt.Printf("Temporary data expired: %s\n", key)
        // Log temporary data expiration
    })

    // Get expiration notification channel
    expirationChan := redisClient.ListenChannelExpirationManager()

    // Start listening for expiration events
    go func() {
        for event := range expirationChan {
            handler.HandleExpiration(event)
        }
    }()

    // Create test keys with different patterns
    testKeys := []struct {
        key   string
        value string
        ttl   time.Duration
    }{
        {"session:user123", "user_data", 3 * time.Second},
        {"cache:product456", "product_info", 5 * time.Second},
        {"temp:upload789", "file_data", 7 * time.Second},
    }

    for _, testKey := range testKeys {
        err := redisClient.SetString([]string{testKey.key}, testKey.value, testKey.ttl)
        if err != nil {
            log.Printf("Error setting key %s: %v", testKey.key, err)
        }
    }

    // Wait for all keys to expire
    time.Sleep(10 * time.Second)
}
```

## Event Structure

```go
type KeyExpirationEvent struct {
    Key       string    `json:"key"`        // Key name
    Value     string    `json:"value"`      // Record body (value)
    ExpiredAt time.Time `json:"expired_at"` // Expiration time
}
```

## Configuration

### Redis Server Configuration

The library automatically configures Redis to enable expiration notifications by setting:

```
notify-keyspace-events Ex
```

Where:
- `E` - enables keyspace events
- `x` - enables expired events

### Client Configuration

You can configure the expiration notification system through the `RedisAdditionalOptions`:

```go
config := redisgklib.RedisConfConn{
    Host:     "localhost",
    Port:     6379,
    Password: "your_password",
    DB:       0,
    AdditionalOptions: redisgklib.RedisAdditionalOptions{
        BaseCtx:      10 * time.Second,  // Context timeout for operations
        DialTimeout:  10 * time.Second,  // Connection timeout
        ReadTimeout:  30 * time.Second,  // Read timeout
        WriteTimeout: 30 * time.Second,  // Write timeout
        PoolSize:     20,                // Connection pool size
        PoolTimeout:  30 * time.Second,  // Pool timeout
    },
}
```

## Performance Considerations

### Memory Usage

- **Unbuffered Channel**: Events are delivered synchronously to prevent memory buildup
- **Value Retrieval**: Attempts to get key values with a 50ms timeout to avoid blocking
- **Goroutine Management**: Proper cleanup prevents goroutine leaks

### Network Usage

- **Single Subscription**: One subscription per RedisGk instance
- **Efficient Events**: Only expiration events are processed
- **Automatic Cleanup**: Subscription is closed when the client is closed

### Error Handling

- **Graceful Degradation**: If value retrieval fails, empty string is used
- **Connection Errors**: Proper error handling for network issues
- **Redis Errors**: Handles Redis-specific errors appropriately

## Best Practices

### 1. Channel Reading

Always read from the expiration channel in a goroutine to prevent blocking:

```go
// Good
go func() {
    for event := range expirationChan {
        // Handle event
    }
}()

// Bad - will block
for event := range expirationChan {
    // Handle event
}
```

### 2. Error Handling

Handle potential errors in your expiration handlers:

```go
go func() {
    for event := range expirationChan {
        defer func() {
            if r := recover(); r != nil {
                log.Printf("Panic in expiration handler: %v", r)
            }
        }()
        
        // Handle event
    }
}()
```

### 3. Resource Cleanup

Ensure proper cleanup when closing the client:

```go
defer func() {
    if err := redisClient.Close(); err != nil {
        log.Printf("Error closing Redis client: %v", err)
    }
}()
```

### 4. Pattern Matching

Use efficient pattern matching for different key types:

```go
switch {
case strings.HasPrefix(event.Key, "session:"):
    handleSessionExpiration(event)
case strings.HasPrefix(event.Key, "cache:"):
    handleCacheExpiration(event)
case strings.HasPrefix(event.Key, "temp:"):
    handleTempExpiration(event)
default:
    log.Printf("Unknown key pattern: %s", event.Key)
}
```

## Troubleshooting

### Common Issues

1. **No Expiration Events Received**
   - Check if Redis server supports keyspace notifications (Redis 2.8.0+)
   - Verify `notify-keyspace-events` configuration
   - Ensure keys actually have TTL set

2. **High Memory Usage**
   - Ensure you're reading from the expiration channel
   - Check for goroutine leaks in your handlers
   - Monitor channel buffer usage

3. **Missing Values**
   - Values may be empty if retrieval times out
   - Consider the 50ms timeout for value retrieval
   - Implement fallback logic for missing values

### Debugging

Enable debug logging to troubleshoot issues:

```go
// Add debug logging to your expiration handler
go func() {
    for event := range expirationChan {
        log.Printf("DEBUG: Received expiration event: %+v", event)
        // Handle event
    }
}()
```

## Limitations

1. **Value Retrieval**: Values are retrieved with a 50ms timeout, which may fail for large values
2. **Single Database**: Notifications are only received for the configured database
3. **Network Dependency**: Requires stable network connection to Redis
4. **Memory Usage**: Large numbers of concurrent expirations may impact performance

## Future Enhancements

Planned improvements include:

- Configurable value retrieval timeout
- Support for multiple database notifications
- Batch processing of expiration events
- Metrics and monitoring integration
- Custom event filtering
