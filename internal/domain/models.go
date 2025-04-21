package domain

import (
	"time"

	"github.com/google/uuid"
)

// --- User ---

// UserRole представляет роль пользователя в системе.
type UserRole string

// Константы для ролей пользователя.
const (
	RoleEmployee  UserRole = "employee"  // Сотрудник ПВЗ
	RoleModerator UserRole = "moderator" // Модератор системы
)

// IsValid проверяет, является ли строка допустимой ролью пользователя.
func (r UserRole) IsValid() bool {
	switch r {
	case RoleEmployee, RoleModerator:
		return true
	default:
		return false
	}
}

// User представляет пользователя системы.
type User struct {
	ID           uuid.UUID `json:"id"`    // Уникальный идентификатор
	Email        string    `json:"email"` // Email (логин)
	PasswordHash string    `json:"-"`     // Хеш пароля (не экспортируется в JSON)
	Role         UserRole  `json:"role"`  // Роль пользователя
}

// --- PVZ (Пункт Выдачи Заказов) ---

// City представляет город, в котором может находиться ПВЗ.
type City string

// Константы для разрешенных городов.
const (
	Moscow          City = "Москва"
	SaintPetersburg City = "Санкт-Петербург"
	Kazan           City = "Казань"
)

// allowedCities содержит список разрешенных городов для быстрого поиска.
var allowedCities = map[City]bool{
	Moscow:          true,
	SaintPetersburg: true,
	Kazan:           true,
}

// IsValid проверяет, разрешено ли создание ПВЗ в данном городе.
func (c City) IsValid() bool {
	_, ok := allowedCities[c]
	return ok
}

// PVZ представляет Пункт Выдачи Заказов.
type PVZ struct {
	ID               uuid.UUID `json:"id"`               // Уникальный идентификатор
	RegistrationDate time.Time `json:"registrationDate"` // Дата и время регистрации в системе
	City             City      `json:"city"`             // Город расположения
}

// --- Reception (Приёмка Товаров) ---

// ReceptionStatus представляет статус приемки товаров.
type ReceptionStatus string

// Константы для статусов приемки.
const (
	StatusInProgress ReceptionStatus = "in_progress" // Приемка активна
	StatusClosed     ReceptionStatus = "close"       // Приемка завершена
)

// IsValid проверяет, является ли строка допустимым статусом приемки.
func (s ReceptionStatus) IsValid() bool {
	switch s {
	case StatusInProgress, StatusClosed:
		return true
	default:
		return false
	}
}

// Reception представляет процесс приемки товаров на конкретном ПВЗ.
type Reception struct {
	ID       uuid.UUID       `json:"id"`       // Уникальный идентификатор приемки
	DateTime time.Time       `json:"dateTime"` // Дата и время начала приемки
	PVZID    uuid.UUID       `json:"pvzId"`    // Идентификатор ПВЗ, где проходит приемка
	Status   ReceptionStatus `json:"status"`   // Текущий статус приемки
}

// --- Product (Товар) ---

// ProductType представляет тип принимаемого товара.
type ProductType string

// Константы для типов товаров.
const (
	TypeElectronics ProductType = "электроника"
	TypeClothing    ProductType = "одежда"
	TypeShoes       ProductType = "обувь"
)

// allowedProductTypes содержит список разрешенных типов товаров.
var allowedProductTypes = map[ProductType]bool{
	TypeElectronics: true,
	TypeClothing:    true,
	TypeShoes:       true,
}

// IsValid проверяет, является ли строка допустимым типом товара.
func (pt ProductType) IsValid() bool {
	_, ok := allowedProductTypes[pt]
	return ok
}

// Product представляет конкретный товар, принятый в рамках приемки.
type Product struct {
	ID          uuid.UUID   `json:"id"`          // Уникальный идентификатор товара
	DateTime    time.Time   `json:"dateTime"`    // Дата и время добавления товара в систему (в рамках приемки)
	Type        ProductType `json:"type"`        // Тип товара
	ReceptionID uuid.UUID   `json:"receptionId"` // Идентификатор приемки, к которой относится товар
}

// --- Вспомогательные структуры для комплексных запросов/ответов ---

// ReceptionWithProducts используется для представления приемки вместе со списком ее товаров.
type ReceptionWithProducts struct {
	Reception Reception `json:"reception"` // Данные о приемке
	Products  []Product `json:"products"`  // Список товаров в этой приемке
}

// PVZWithDetails используется для представления ПВЗ вместе со списком его приемок (и их товаров).
type PVZWithDetails struct {
	PVZ        PVZ                     `json:"pvz"`        // Данные о ПВЗ
	Receptions []ReceptionWithProducts `json:"receptions"` // Список приемок в этом ПВЗ
}
