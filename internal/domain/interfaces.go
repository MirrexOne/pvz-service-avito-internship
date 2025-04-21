package domain

import (
	"context"
	"time"

	"github.com/google/uuid"
)

// PVZRepository определяет методы для работы с сущностями ПВЗ.
type PVZRepository interface {
	// Create сохраняет новый ПВЗ в хранилище.
	Create(ctx context.Context, pvz *PVZ) error
	// GetByID находит ПВЗ по его уникальному идентификатору.
	// Возвращает ErrNotFound, если ПВЗ не найден.
	GetByID(ctx context.Context, id uuid.UUID) (*PVZ, error)

	// ListIDsAndTotal возвращает слайс ID ПВЗ для текущей страницы с учетом фильтров
	// и общее количество ПВЗ, удовлетворяющих фильтрам (без пагинации).
	// Заменил старый метод List.
	ListIDsAndTotal(ctx context.Context, limit, offset int, startDate, endDate *time.Time) ([]uuid.UUID, int, error)

	// GetByIDs находит и возвращает список ПВЗ по списку их ID.
	// Добавлен для получения данных после ListIDsAndTotal.
	GetByIDs(ctx context.Context, ids []uuid.UUID) ([]PVZ, error)

	// ListAll возвращает список всех ПВЗ без фильтрации и пагинации (для gRPC).
	ListAll(ctx context.Context) ([]PVZ, error)
}

// ReceptionRepository определяет методы для работы с сущностями Приемок.
type ReceptionRepository interface {
	// Create сохраняет новую приемку в хранилище.
	Create(ctx context.Context, reception *Reception) error
	// GetByID находит приемку по ее уникальному идентификатору.
	// Возвращает ErrNotFound, если приемка не найдена.
	GetByID(ctx context.Context, id uuid.UUID) (*Reception, error)
	// FindOpenByPVZID находит последнюю незавершенную приемку ('in_progress') для указанного ПВЗ.
	// Возвращает ErrNoOpenReception, если открытых приемок нет.
	FindOpenByPVZID(ctx context.Context, pvzID uuid.UUID) (*Reception, error)
	// UpdateStatus обновляет статус приемки по ее ID.
	// Возвращает ErrNotFound, если приемка не найдена.
	UpdateStatus(ctx context.Context, id uuid.UUID, status ReceptionStatus) error
	// ListByPVZIDsAndDate возвращает мапу приемок с товарами для указанных ПВЗ и диапазона дат.
	// Ключ мапы - PVZ ID. Используется для обогащения данных в PVZService.ListPVZs.
	ListByPVZIDsAndDate(ctx context.Context, pvzIDs []uuid.UUID, startDate, endDate *time.Time) (map[uuid.UUID][]ReceptionWithProducts, error)
}

// ProductRepository определяет методы для работы с сущностями Товаров.
type ProductRepository interface {
	// Create сохраняет новый товар в хранилище, привязывая его к приемке.
	Create(ctx context.Context, product *Product) error
	// FindLastByReceptionID находит самый последний добавленный товар для указанной приемки (по времени добавления).
	// Возвращает ErrNoProductsToDelete, если в приемке нет товаров.
	FindLastByReceptionID(ctx context.Context, receptionID uuid.UUID) (*Product, error)
	// DeleteByID удаляет товар по его уникальному идентификатору.
	// Возвращает ErrNotFound, если товар не найден.
	DeleteByID(ctx context.Context, id uuid.UUID) error
	// ListByReceptionIDs возвращает мапу товаров для указанных ID приемок.
	// Ключ мапы - Reception ID. Используется для обогащения данных в ReceptionRepository.
	ListByReceptionIDs(ctx context.Context, receptionIDs []uuid.UUID) (map[uuid.UUID][]Product, error)
}

// UserRepository определяет методы для работы с сущностями Пользователей.
type UserRepository interface {
	// Create сохраняет нового пользователя в хранилище.
	// Возвращает ErrConflict, если пользователь с таким email уже существует.
	Create(ctx context.Context, user *User) error
	// GetByEmail находит пользователя по его email.
	// Возвращает ErrNotFound, если пользователь не найден.
	GetByEmail(ctx context.Context, email string) (*User, error)
	// GetByID находит пользователя по его ID.
	// Возвращает ErrNotFound, если пользователь не найден.
	GetByID(ctx context.Context, id uuid.UUID) (*User, error)
}

// --- Интерфейсы Сервисов ---
// Определяют методы бизнес-логики (use cases).

// AuthService определяет методы, связанные с аутентификацией и регистрацией.
type AuthService interface {
	// DummyLogin генерирует тестовый JWT токен для указанной роли.
	DummyLogin(ctx context.Context, role UserRole) (string, error)
	// Register регистрирует нового пользователя.
	Register(ctx context.Context, email, password string, role UserRole) (*User, error)
	// Login аутентифицирует пользователя и возвращает JWT токен.
	Login(ctx context.Context, email, password string) (string, error)
}

// PVZService определяет методы бизнес-логики для работы с ПВЗ.
type PVZService interface {
	// CreatePVZ создает новый ПВЗ с учетом бизнес-правил (допустимые города).
	CreatePVZ(ctx context.Context, city City) (*PVZ, error)
	// ListPVZs возвращает список ПВЗ с деталями и пагинацией.
	ListPVZs(ctx context.Context, limit, page int, startDate, endDate *time.Time) ([]PVZWithDetails, int, error)
}

// ReceptionService определяет методы бизнес-логики для работы с приемками.
type ReceptionService interface {
	// CreateReception инициирует новую приемку для ПВЗ.
	CreateReception(ctx context.Context, pvzID uuid.UUID) (*Reception, error)
	// CloseReception закрывает последнюю активную приемку для ПВЗ.
	CloseReception(ctx context.Context, pvzID uuid.UUID) (*Reception, error)
}

// ProductService определяет методы бизнес-логики для работы с товарами.
type ProductService interface {
	// AddProduct добавляет товар в текущую открытую приемку ПВЗ.
	AddProduct(ctx context.Context, pvzID uuid.UUID, productType ProductType) (*Product, error)
	// DeleteLastProduct удаляет последний добавленный товар из открытой приемки (LIFO).
	DeleteLastProduct(ctx context.Context, pvzID uuid.UUID) error
}

// --- Вспомогательные Интерфейсы ---

// PasswordHasher определяет контракт для хеширования и сравнения паролей.
type PasswordHasher interface {
	// Hash генерирует хеш для заданного пароля.
	Hash(password string) (string, error)
	// Compare сравнивает хеш с паролем. Возвращает nil при совпадении.
	Compare(hashedPassword, password string) error
}

// MetricsCollector определяет контракт для сбора метрик Prometheus.
type MetricsCollector interface {
	IncRequestsTotal(method, path, statusCode string)
	ObserveRequestDuration(method, path string, duration float64)
	IncPVZCreated()
	IncReceptionsCreated()
	IncProductsAdded()
}
