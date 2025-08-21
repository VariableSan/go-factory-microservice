package errors

import (
	"fmt"
	"net/http"
)

// ErrorCode represents application error codes
type ErrorCode string

const (
	// Authentication errors
	ErrUnauthorized    ErrorCode = "UNAUTHORIZED"
	ErrForbidden       ErrorCode = "FORBIDDEN"
	ErrInvalidToken    ErrorCode = "INVALID_TOKEN"
	ErrExpiredToken    ErrorCode = "EXPIRED_TOKEN"
	
	// Validation errors
	ErrValidation      ErrorCode = "VALIDATION_ERROR"
	ErrInvalidInput    ErrorCode = "INVALID_INPUT"
	ErrMissingField    ErrorCode = "MISSING_FIELD"
	
	// Resource errors
	ErrNotFound        ErrorCode = "NOT_FOUND"
	ErrAlreadyExists   ErrorCode = "ALREADY_EXISTS"
	ErrConflict        ErrorCode = "CONFLICT"
	
	// Server errors
	ErrInternal        ErrorCode = "INTERNAL_ERROR"
	ErrServiceUnavailable ErrorCode = "SERVICE_UNAVAILABLE"
	ErrDatabaseError   ErrorCode = "DATABASE_ERROR"
	ErrExternalService ErrorCode = "EXTERNAL_SERVICE_ERROR"
)

// AppError represents application error with code and message
type AppError struct {
	Code       ErrorCode `json:"code"`
	Message    string    `json:"message"`
	Details    string    `json:"details,omitempty"`
	HTTPStatus int       `json:"-"`
}

// Error implements error interface
func (e *AppError) Error() string {
	if e.Details != "" {
		return fmt.Sprintf("[%s] %s: %s", e.Code, e.Message, e.Details)
	}
	return fmt.Sprintf("[%s] %s", e.Code, e.Message)
}

// NewAppError creates a new application error
func NewAppError(code ErrorCode, message string) *AppError {
	return &AppError{
		Code:       code,
		Message:    message,
		HTTPStatus: getHTTPStatusForCode(code),
	}
}

// NewAppErrorWithDetails creates a new application error with details
func NewAppErrorWithDetails(code ErrorCode, message, details string) *AppError {
	return &AppError{
		Code:       code,
		Message:    message,
		Details:    details,
		HTTPStatus: getHTTPStatusForCode(code),
	}
}

// getHTTPStatusForCode returns HTTP status code for error code
func getHTTPStatusForCode(code ErrorCode) int {
	switch code {
	case ErrUnauthorized, ErrInvalidToken, ErrExpiredToken:
		return http.StatusUnauthorized
	case ErrForbidden:
		return http.StatusForbidden
	case ErrNotFound:
		return http.StatusNotFound
	case ErrValidation, ErrInvalidInput, ErrMissingField:
		return http.StatusBadRequest
	case ErrAlreadyExists, ErrConflict:
		return http.StatusConflict
	case ErrServiceUnavailable:
		return http.StatusServiceUnavailable
	case ErrInternal, ErrDatabaseError, ErrExternalService:
		return http.StatusInternalServerError
	default:
		return http.StatusInternalServerError
	}
}

// IsAppError checks if error is AppError
func IsAppError(err error) (*AppError, bool) {
	if appErr, ok := err.(*AppError); ok {
		return appErr, true
	}
	return nil, false
}

// Wrap wraps an error with AppError
func Wrap(err error, code ErrorCode, message string) *AppError {
	details := ""
	if err != nil {
		details = err.Error()
	}
	return NewAppErrorWithDetails(code, message, details)
}

// Predefined errors
var (
	ErrUserNotFound     = NewAppError(ErrNotFound, "User not found")
	ErrInvalidCredentials = NewAppError(ErrUnauthorized, "Invalid credentials")
	ErrTokenExpired     = NewAppError(ErrExpiredToken, "Token has expired")
	ErrAccessDenied     = NewAppError(ErrForbidden, "Access denied")
)
