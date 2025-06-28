package main

import (
	"fmt"
	"log"
	"time"

	redisgklib "github.com/GAFIKART/redisgk/lib"
)

// User - пример структуры для демонстрации работы с объектами
type User struct {
	ID       int    `json:"id"`
	Name     string `json:"name"`
	Email    string `json:"email"`
	Age      int    `json:"age"`
	IsActive bool   `json:"is_active"`
}

// Product - еще одна структура для демонстрации
type Product struct {
	ID          string  `json:"id"`
	Name        string  `json:"name"`
	Price       float64 `json:"price"`
	Description string  `json:"description"`
	Category    string  `json:"category"`
}

func main() {
	fmt.Println("=== Примеры использования библиотеки redisgk ===")

	// Конфигурация подключения к Redis
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

	// Создание экземпляра RedisGk
	redisGk, err := redisgklib.NewRedisGk(config)
	if err != nil {
		log.Fatalf("Ошибка подключения к Redis: %v", err)
	}
	defer func() {
		if err := redisGk.Close(); err != nil {
			log.Printf("Ошибка закрытия соединения: %v", err)
		}
	}()

	fmt.Println("✅ Подключение к Redis установлено")

	// Пример 1: Работа со строками
	fmt.Println("=== 1. РАБОТА СО СТРОКАМИ ===")
	demoStrings(redisGk)

	// Пример 2: Работа с объектами
	fmt.Println("\n=== 2. РАБОТА С ОБЪЕКТАМИ ===")
	demoObjects(redisGk)

	// Пример 3: Поиск объектов
	fmt.Println("\n=== 3. ПОИСК ОБЪЕКТОВ ===")
	demoFindObjects(redisGk)

	// Пример 4: Проверка существования ключей
	fmt.Println("\n=== 4. ПРОВЕРКА СУЩЕСТВОВАНИЯ КЛЮЧЕЙ ===")
	demoExists(redisGk)

	// Пример 5: Удаление ключей
	fmt.Println("\n=== 5. УДАЛЕНИЕ КЛЮЧЕЙ ===")
	demoDelete(redisGk)

	fmt.Println("\n=== ВСЕ ПРИМЕРЫ ЗАВЕРШЕНЫ ===")
}

// demoStrings - демонстрация работы со строками
func demoStrings(redisGk *redisgklib.RedisGk) {
	fmt.Println("📝 Сохранение строки...")

	// Сохранение строки без TTL
	err := redisGk.SetString([]string{"user", "profile", "name"}, "Иван Петров")
	if err != nil {
		log.Printf("Ошибка сохранения строки: %v", err)
		return
	}
	fmt.Println("✅ Строка сохранена: user:profile:name = 'Иван Петров'")

	// Сохранение строки с TTL
	err = redisGk.SetString([]string{"temp", "session", "token"}, "abc123xyz", 30*time.Second)
	if err != nil {
		log.Printf("Ошибка сохранения строки с TTL: %v", err)
		return
	}
	fmt.Println("✅ Строка с TTL сохранена: temp:session:token = 'abc123xyz' (TTL: 30s)")

	// Получение строки
	value, err := redisGk.GetString([]string{"user", "profile", "name"})
	if err != nil {
		log.Printf("Ошибка получения строки: %v", err)
		return
	}
	fmt.Printf("✅ Строка получена: %s\n", value)

	// Попытка получить несуществующий ключ
	_, err = redisGk.GetString([]string{"nonexistent", "key"})
	if err != nil {
		fmt.Printf("✅ Ожидаемая ошибка для несуществующего ключа: %v\n", err)
	}
}

// demoObjects - демонстрация работы с объектами
func demoObjects(redisGk *redisgklib.RedisGk) {
	fmt.Println("📦 Сохранение объекта User...")

	// Создание объекта User
	user := User{
		ID:       1,
		Name:     "Анна Сидорова",
		Email:    "anna@example.com",
		Age:      28,
		IsActive: true,
	}

	// Сохранение объекта
	err := redisgklib.SetObj(redisGk, []string{"users", "1"}, user)
	if err != nil {
		log.Printf("Ошибка сохранения объекта: %v", err)
		return
	}
	fmt.Println("✅ Объект User сохранен: users:1")

	// Сохранение объекта с TTL
	product := Product{
		ID:          "prod_001",
		Name:        "Смартфон Galaxy S21",
		Price:       89999.99,
		Description: "Мощный смартфон с отличной камерой",
		Category:    "Электроника",
	}

	err = redisgklib.SetObj(redisGk, []string{"products", "prod_001"}, product, 1*time.Hour)
	if err != nil {
		log.Printf("Ошибка сохранения объекта с TTL: %v", err)
		return
	}
	fmt.Println("✅ Объект Product сохранен с TTL: products:prod_001 (TTL: 1h)")

	// Получение объекта User
	retrievedUser, err := redisgklib.GetObj[User](redisGk, []string{"users", "1"})
	if err != nil {
		log.Printf("Ошибка получения объекта User: %v", err)
		return
	}
	fmt.Printf("✅ Объект User получен: ID=%d, Name=%s, Email=%s\n",
		retrievedUser.ID, retrievedUser.Name, retrievedUser.Email)

	// Получение объекта Product
	retrievedProduct, err := redisgklib.GetObj[Product](redisGk, []string{"products", "prod_001"})
	if err != nil {
		log.Printf("Ошибка получения объекта Product: %v", err)
		return
	}
	fmt.Printf("✅ Объект Product получен: ID=%s, Name=%s, Price=%.2f\n",
		retrievedProduct.ID, retrievedProduct.Name, retrievedProduct.Price)
}

// demoFindObjects - демонстрация поиска объектов
func demoFindObjects(redisGk *redisgklib.RedisGk) {
	fmt.Println("🔍 Поиск объектов...")

	// Создание нескольких пользователей для поиска
	users := []User{
		{ID: 2, Name: "Петр Иванов", Email: "petr@example.com", Age: 35, IsActive: true},
		{ID: 3, Name: "Мария Козлова", Email: "maria@example.com", Age: 24, IsActive: false},
		{ID: 4, Name: "Сергей Волков", Email: "sergey@example.com", Age: 42, IsActive: true},
	}

	// Сохранение пользователей
	for _, user := range users {
		err := redisgklib.SetObj(redisGk, []string{"users", fmt.Sprintf("%d", user.ID)}, user)
		if err != nil {
			log.Printf("Ошибка сохранения пользователя %d: %v", user.ID, err)
			continue
		}
		fmt.Printf("✅ Пользователь %d сохранен\n", user.ID)
	}

	// Поиск всех пользователей
	foundUsers, err := redisgklib.FindObj[User](redisGk, []string{"users"})
	if err != nil {
		log.Printf("Ошибка поиска пользователей: %v", err)
		return
	}

	fmt.Printf("✅ Найдено пользователей: %d\n", len(foundUsers))
	for key, user := range foundUsers {
		fmt.Printf("   - %s: %s (%s)\n", key, user.Name, user.Email)
	}

	// Поиск с ограничением количества результатов
	foundUsersLimited, err := redisgklib.FindObj[User](redisGk, []string{"users"}, 2)
	if err != nil {
		log.Printf("Ошибка поиска пользователей с ограничением: %v", err)
		return
	}

	fmt.Printf("✅ Найдено пользователей (ограничение 2): %d\n", len(foundUsersLimited))
}

// demoExists - демонстрация проверки существования ключей
func demoExists(redisGk *redisgklib.RedisGk) {
	fmt.Println("🔍 Проверка существования ключей...")

	// Проверка существующего ключа
	exists, err := redisGk.Exists([]string{"users", "1"})
	if err != nil {
		log.Printf("Ошибка проверки существования ключа: %v", err)
		return
	}
	fmt.Printf("✅ Ключ 'users:1' существует: %t\n", exists)

	// Проверка несуществующего ключа
	exists, err = redisGk.Exists([]string{"nonexistent", "key"})
	if err != nil {
		log.Printf("Ошибка проверки существования ключа: %v", err)
		return
	}
	fmt.Printf("✅ Ключ 'nonexistent:key' существует: %t\n", exists)
}

// demoDelete - демонстрация удаления ключей
func demoDelete(redisGk *redisgklib.RedisGk) {
	fmt.Println("🗑️ Удаление ключей...")

	// Создание тестового ключа для удаления
	err := redisGk.SetString([]string{"test", "delete", "key"}, "значение для удаления")
	if err != nil {
		log.Printf("Ошибка создания тестового ключа: %v", err)
		return
	}
	fmt.Println("✅ Тестовый ключ создан: test:delete:key")

	// Проверка существования перед удалением
	exists, err := redisGk.Exists([]string{"test", "delete", "key"})
	if err != nil {
		log.Printf("Ошибка проверки существования: %v", err)
		return
	}
	fmt.Printf("✅ Ключ существует перед удалением: %t\n", exists)

	// Удаление ключа
	err = redisGk.Del([]string{"test", "delete", "key"})
	if err != nil {
		log.Printf("Ошибка удаления ключа: %v", err)
		return
	}
	fmt.Println("✅ Ключ удален: test:delete:key")

	// Проверка существования после удаления
	exists, err = redisGk.Exists([]string{"test", "delete", "key"})
	if err != nil {
		log.Printf("Ошибка проверки существования: %v", err)
		return
	}
	fmt.Printf("✅ Ключ существует после удаления: %t\n", exists)

	// Удаление нескольких ключей
	fmt.Println("🗑️ Удаление всех созданных ключей...")

	// Удаляем все ключи одним вызовом
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
		log.Printf("Ошибка удаления ключей: %v", err)
	} else {
		fmt.Println("✅ Все ключи удалены успешно")
	}
}
