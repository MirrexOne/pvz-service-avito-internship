package domain

import "errors"

var (
	ErrInternalServer    = errors.New("internal server error")
	ErrValidation        = errors.New("validation failed")
	ErrNotFound          = errors.New("entity not found")
	ErrConflict          = errors.New("resource conflict")
	ErrDatabaseError     = errors.New("database operation failed")
	ErrPVZNotFound       = errors.New("pvz not found")
	ErrReceptionNotFound = errors.New("reception not found")
	ErrInvalidRequest    = errors.New("invalid request")
	ErrPassIsRequired    = errors.New("password is required")

	ErrUnauthorized = errors.New("unauthorized")
	ErrForbidden    = errors.New("access forbidden")

	ErrPVZCityNotAllowed = errors.New("pvz creation is not allowed in this city")

	ErrReceptionInProgress = errors.New("previous reception is still in progress")
	ErrNoOpenReception     = errors.New("no open reception found for this pvz")
	ErrReceptionClosed     = errors.New("reception is already closed")

	ErrProductDeletionOrder = errors.New("products can only be deleted in LIFO order from an open reception")
	ErrNoProductsToDelete   = errors.New("no products available to delete in the current reception")
)
