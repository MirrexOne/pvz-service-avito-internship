package domain

import "errors"

// Стандартные доменные ошибки, используемые в приложении.
// Они помогают абстрагироваться от конкретных ошибок нижележащих слоев (БД, внешние сервисы).
var (
	// Общие ошибки
	ErrInternalServer    = errors.New("internal server error")     // Общая ошибка сервера (500)
	ErrValidation        = errors.New("validation failed")         // Ошибка валидации данных (400)
	ErrNotFound          = errors.New("entity not found")          // Сущность не найдена (404)
	ErrConflict          = errors.New("resource conflict")         // Конфликт ресурсов (например, email занят) (409)
	ErrDatabaseError     = errors.New("database operation failed") // Ошибка при операции с БД (маскирует детали)
	ErrPVZNotFound       = errors.New("pvz not found")             // ПВЗ не найден (404)
	ErrReceptionNotFound = errors.New("reception not found")       // Приемка не найдена (404)

	// Ошибки авторизации/аутентификации
	ErrUnauthorized = errors.New("unauthorized")     // Неверные учетные данные или отсутствует токен (401)
	ErrForbidden    = errors.New("access forbidden") // Недостаточно прав для выполнения операции (403)

	// Ошибки бизнес-логики ПВЗ
	ErrPVZCityNotAllowed = errors.New("pvz creation is not allowed in this city") // Попытка создать ПВЗ в неразрешенном городе (400)

	// Ошибки бизнес-логики Приемок
	ErrReceptionInProgress = errors.New("previous reception is still in progress") // Попытка создать новую приемку при наличии незавершенной (400)
	ErrNoOpenReception     = errors.New("no open reception found for this pvz")    // Попытка добавить/удалить товар или закрыть приемку, когда нет активной (400)
	ErrReceptionClosed     = errors.New("reception is already closed")             // Попытка закрыть уже закрытую приемку или удалить товар из закрытой (400)

	// Ошибки бизнес-логики Товаров
	ErrProductDeletionOrder = errors.New("products can only be deleted in LIFO order from an open reception") // Нарушение порядка LIFO при удалении (400)
	ErrNoProductsToDelete   = errors.New("no products available to delete in the current reception")          // Попытка удалить товар из пустой приемки (400)
)
