# RedisGK - Библиотека для работы с Redis

Библиотека `redisgk` предоставляет удобную обертку над официальным Redis клиентом для Go, объединяя всю необходимую логику для работы с Redis в ваших проектах.

## Особенности

- 🚀 **Простота использования** - удобный API для работы с Redis
- 🔒 **Типобезопасность** - поддержка дженериков для работы с объектами
- ⚡ **Производительность** - оптимизированные настройки подключения
- 🛡️ **Валидация** - проверка конфигурации и размеров данных
- 🔧 **Гибкость** - настраиваемые таймауты и параметры пула соединений
- 🔍 **Поиск** - поиск объектов по паттерну ключей
- 🗑️ **Массовые операции** - удаление нескольких ключей за один вызов

## Установка

```bash
go get github.com/GAFIKART/redisgk
```

## Быстрый старт

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
    // Конфигурация подключения к Redis
    config := redisgklib.RedisConfConn{
        Host:     "localhost",
        Port:     6379,
        User:     "", // Опционально
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

    // Создание клиента Redis
    redisClient, err := redisgklib.NewRedisGk(config)
    if err != nil {
        log.Fatal("Ошибка подключения к Redis:", err)
    }
    defer redisClient.Close()

    // Сохранение объекта
    user := User{ID: 1, Name: "Иван", Age: 25}
    err = redisgklib.SetObj(redisClient, []string{"users", "1"}, user, 1*time.Hour)
    if err != nil {
        log.Fatal("Ошибка сохранения:", err)
    }

    // Получение объекта
    retrievedUser, err := redisgklib.GetObj[User](redisClient, []string{"users", "1"})
    if err != nil {
        log.Fatal("Ошибка получения:", err)
    }
    log.Printf("Получен пользователь: %+v", *retrievedUser)

    // Работа со строками
    err = redisClient.SetString([]string{"greeting"}, "Привет, мир!", 30*time.Minute)
    if err != nil {
        log.Fatal("Ошибка сохранения строки:", err)
    }

    greeting, err := redisClient.GetString([]string{"greeting"})
    if err != nil {
        log.Fatal("Ошибка получения строки:", err)
    }
    log.Println("Приветствие:", greeting)

    // Поиск объектов по паттерну
    users, err := redisgklib.FindObj[User](redisClient, []string{"users"}, 100)
    if err != nil {
        log.Fatal("Ошибка поиска:", err)
    }
    log.Printf("Найдено пользователей: %d", len(users))

    // Проверка существования ключа
    exists, err := redisClient.Exists("users:1")
    if err != nil {
        log.Fatal("Ошибка проверки:", err)
    }
    log.Printf("Ключ существует: %t", exists)

    // Удаление нескольких ключей за один вызов
    err = redisClient.Del(
        []string{"users", "1"},
        []string{"greeting"},
    )
    if err != nil {
        log.Fatal("Ошибка удаления:", err)
    }
}
```

## Примеры использования

Полные примеры использования всех методов библиотеки доступны в папке [`example/`](./example/):

```bash
cd example
go run main.go
```

Примеры демонстрируют:
- Работу со строками и объектами
- Поиск объектов по паттерну
- Проверку существования ключей
- Массовое удаление ключей
- Обработку ошибок

## API

### Основные функции

#### `NewRedisGk(config RedisConfConn) (*RedisGk, error)`
Создает новый экземпляр клиента Redis.

#### `SetObj[T any](client *RedisGk, keyPath []string, value T, ttl ...time.Duration) error`
Сохраняет объект в Redis с автоматической сериализацией в JSON.

#### `GetObj[T any](client *RedisGk, keyPath []string) (*T, error)`
Получает объект из Redis с автоматической десериализацией из JSON.

#### `FindObj[T any](client *RedisGk, patternPath []string, count ...int64) (map[string]*T, error)`
Поиск объектов по паттерну ключей.

### Методы RedisGk

#### `SetString(keyPath []string, value string, ttl ...time.Duration) error`
Сохраняет строку в Redis.

#### `GetString(keyPath []string) (string, error)`
Получает строку из Redis.

#### `Del(keyPath ...[]string) error`
Удаляет один или несколько ключей из Redis. Поддерживает вариативные параметры для удаления множественных ключей за один вызов.

#### `Exists(key string) (bool, error)`
Проверяет существование ключа.

#### `Close() error`
Закрывает соединение с Redis.

## Конфигурация

### RedisConfConn
```go
type RedisConfConn struct {
    Host     string
    Port     int
    User     string        // Опционально
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

## Особенности

### Обработка ключей
- Автоматическая нормализация ключей (удаление специальных символов)
- Замена пробелов на нижнее подчеркивание
- Поддержка иерархических ключей через слайс строк
- Ограничение размера ключа до 512 МБ

### Обработка данных
- Автоматическая сериализация/десериализация объектов в JSON
- Проверка размера данных (максимум 512 МБ)
- Обработка ошибки `redis.Nil` при отсутствии ключа

### Производительность
- Настраиваемый пул соединений
- Контексты с таймаутами для всех операций
- Эффективное удаление множественных ключей

### Массовые операции
Метод `Del()` поддерживает удаление нескольких ключей за один вызов:
```go
err := redisClient.Del(
    []string{"users", "1"},
    []string{"users", "2"},
    []string{"temp", "session"},
)
```

## Требования

- Go 1.24.2+
- Redis сервер

## Лицензия

MIT License
