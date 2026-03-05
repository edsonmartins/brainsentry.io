package domain

import "errors"

// Sentinel domain errors for type-safe error handling.
var (
	ErrNotFound       = errors.New("resource not found")
	ErrValidation     = errors.New("validation error")
	ErrConflict       = errors.New("resource conflict")
	ErrUnauthorized   = errors.New("unauthorized")
	ErrForbidden      = errors.New("forbidden")
	ErrInternal       = errors.New("internal error")
	ErrAlreadyExists  = errors.New("resource already exists")
	ErrRateLimited    = errors.New("rate limited")
)

// DomainError wraps a sentinel error with contextual message.
type DomainError struct {
	Err     error
	Message string
	Code    string
}

func (e *DomainError) Error() string {
	if e.Message != "" {
		return e.Message
	}
	return e.Err.Error()
}

func (e *DomainError) Unwrap() error {
	return e.Err
}

// NewNotFoundError creates a not-found domain error.
func NewNotFoundError(message string) *DomainError {
	return &DomainError{Err: ErrNotFound, Message: message, Code: "not_found"}
}

// NewValidationError creates a validation domain error.
func NewValidationError(message string) *DomainError {
	return &DomainError{Err: ErrValidation, Message: message, Code: "validation"}
}

// NewConflictError creates a conflict domain error.
func NewConflictError(message string) *DomainError {
	return &DomainError{Err: ErrConflict, Message: message, Code: "conflict"}
}

// NewInternalError creates an internal domain error.
func NewInternalError(message string) *DomainError {
	return &DomainError{Err: ErrInternal, Message: message, Code: "internal"}
}
