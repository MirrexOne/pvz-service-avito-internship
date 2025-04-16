package hash

import (
	"fmt"
	"golang.org/x/crypto/bcrypt"

	// Импортируем интерфейс из домена
	"pvz-service-avito-internship/internal/domain"
)

// BcryptHasher реализует интерфейс domain.PasswordHasher с использованием bcrypt.
type BcryptHasher struct {
	cost int // Стоимость хеширования bcrypt
}

// NewBcryptHasher создает новый экземпляр BcryptHasher.
// Принимает cost - сложность хеширования. Если cost равен 0, используется bcrypt.DefaultCost.
func NewBcryptHasher(cost int) *BcryptHasher {
	if cost == 0 {
		cost = bcrypt.DefaultCost // Значение по умолчанию из библиотеки bcrypt
	}
	return &BcryptHasher{cost: cost}
}

// Hash генерирует bcrypt хеш для заданного пароля.
func (h *BcryptHasher) Hash(password string) (string, error) {
	// Используем стандартную функцию bcrypt для генерации хеша
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), h.cost)
	if err != nil {
		// Оборачиваем ошибку для добавления контекста
		return "", fmt.Errorf("failed to generate bcrypt hash: %w", err)
	}
	return string(bytes), nil // Возвращаем хеш как строку
}

// Compare сравнивает предоставленный хеш пароля с чистым паролем.
// Возвращает nil, если пароль совпадает с хешем.
// Возвращает ошибку (обычно bcrypt.ErrMismatchedHashAndPassword), если пароли не совпадают.
func (h *BcryptHasher) Compare(hashedPassword, password string) error {
	// Используем стандартную функцию bcrypt для сравнения
	err := bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password))
	if err != nil {
		// Если есть ошибка сравнения (например, несовпадение), возвращаем ее как есть.
		// Это позволяет вышестоящему коду (AuthService) корректно определить причину ошибки.
		return err
	}
	return nil // Пароли совпали
}

// Проверка соответствия интерфейсу во время компиляции
var _ domain.PasswordHasher = (*BcryptHasher)(nil)
