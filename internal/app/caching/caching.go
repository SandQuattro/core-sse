package caching

import "time"

// Cache интерфейс определяет стандартные методы для работы с кешем.
type Cache interface {
	// Set добавляет значение в кеш по ключу с опциональным временем истечения в секундах.
	Set(key string, value any, expiration time.Duration) error

	// Get получает значение из кеша по ключу.
	// Если ключ не найден, возвращает ошибку.
	Get(key string) (any, error)

	// Delete удаляет значение из кеша по ключу.
	Delete(key string) error

	// Purge очищает весь кеш.
	Purge() error
}
