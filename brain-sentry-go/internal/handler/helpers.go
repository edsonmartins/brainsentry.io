package handler

import (
	"encoding/json"
	"errors"
	"net/http"
	"time"

	"github.com/integraltech/brainsentry/internal/domain"
	"github.com/integraltech/brainsentry/internal/dto"
)

// writeJSON writes a JSON response.
func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(v)
}

// writeError writes a JSON error response with structured fields.
func writeError(w http.ResponseWriter, status int, message string) {
	category := errorCategoryFromStatus(status)
	writeJSON(w, status, dto.ErrorResponse{
		Error:         http.StatusText(status),
		Message:       message,
		Status:        status,
		ErrorCode:     errorCodeFromStatus(status),
		ErrorCategory: category,
		Timestamp:     time.Now().UTC().Format(time.RFC3339),
	})
}

// writeDomainError maps a domain error to the appropriate HTTP status and writes a structured response.
func writeDomainError(w http.ResponseWriter, err error) {
	var domErr *domain.DomainError
	if errors.As(err, &domErr) {
		status := httpStatusFromDomainError(domErr)
		writeJSON(w, status, dto.ErrorResponse{
			Error:         http.StatusText(status),
			Message:       domErr.Message,
			Status:        status,
			ErrorCode:     domErr.Code,
			ErrorCategory: errorCategoryFromDomainError(domErr),
			Timestamp:     time.Now().UTC().Format(time.RFC3339),
		})
		return
	}
	writeError(w, http.StatusInternalServerError, err.Error())
}

func httpStatusFromDomainError(err *domain.DomainError) int {
	switch {
	case errors.Is(err, domain.ErrNotFound):
		return http.StatusNotFound
	case errors.Is(err, domain.ErrValidation):
		return http.StatusBadRequest
	case errors.Is(err, domain.ErrConflict), errors.Is(err, domain.ErrAlreadyExists):
		return http.StatusConflict
	case errors.Is(err, domain.ErrUnauthorized):
		return http.StatusUnauthorized
	case errors.Is(err, domain.ErrForbidden):
		return http.StatusForbidden
	case errors.Is(err, domain.ErrRateLimited):
		return http.StatusTooManyRequests
	default:
		return http.StatusInternalServerError
	}
}

func errorCategoryFromDomainError(err *domain.DomainError) string {
	switch {
	case errors.Is(err, domain.ErrNotFound):
		return "NOT_FOUND"
	case errors.Is(err, domain.ErrValidation):
		return "VALIDATION"
	case errors.Is(err, domain.ErrConflict), errors.Is(err, domain.ErrAlreadyExists):
		return "CONFLICT"
	case errors.Is(err, domain.ErrUnauthorized), errors.Is(err, domain.ErrForbidden):
		return "AUTH"
	default:
		return "INTERNAL"
	}
}

func errorCategoryFromStatus(status int) string {
	switch {
	case status == http.StatusNotFound:
		return "NOT_FOUND"
	case status == http.StatusBadRequest:
		return "VALIDATION"
	case status == http.StatusConflict:
		return "CONFLICT"
	case status == http.StatusUnauthorized || status == http.StatusForbidden:
		return "AUTH"
	default:
		return "INTERNAL"
	}
}

func errorCodeFromStatus(status int) string {
	switch status {
	case http.StatusNotFound:
		return "not_found"
	case http.StatusBadRequest:
		return "validation"
	case http.StatusConflict:
		return "conflict"
	case http.StatusUnauthorized:
		return "unauthorized"
	case http.StatusForbidden:
		return "forbidden"
	case http.StatusTooManyRequests:
		return "rate_limited"
	default:
		return "internal"
	}
}
