package response

import (
	"encoding/json"
	"net/http"

	"github.com/VariableSan/go-factory-microservice/pkg/common/errors"
)

// APIResponse represents a standard API response
type APIResponse struct {
	Success bool        `json:"success"`
	Data    interface{} `json:"data,omitempty"`
	Error   *ErrorInfo  `json:"error,omitempty"`
	Message string      `json:"message,omitempty"`
}

// ErrorInfo represents error information in API response
type ErrorInfo struct {
	Code    string `json:"code"`
	Message string `json:"message"`
	Details string `json:"details,omitempty"`
}

// Success sends a successful response
func Success(w http.ResponseWriter, data interface{}) {
	response := APIResponse{
		Success: true,
		Data:    data,
	}
	sendJSON(w, http.StatusOK, response)
}

// SuccessWithMessage sends a successful response with message
func SuccessWithMessage(w http.ResponseWriter, data interface{}, message string) {
	response := APIResponse{
		Success: true,
		Data:    data,
		Message: message,
	}
	sendJSON(w, http.StatusOK, response)
}

// Created sends a created response
func Created(w http.ResponseWriter, data interface{}) {
	response := APIResponse{
		Success: true,
		Data:    data,
	}
	sendJSON(w, http.StatusCreated, response)
}

// NoContent sends a no content response
func NoContent(w http.ResponseWriter) {
	w.WriteHeader(http.StatusNoContent)
}

// Error sends an error response
func Error(w http.ResponseWriter, err error) {
	if appErr, ok := errors.IsAppError(err); ok {
		response := APIResponse{
			Success: false,
			Error: &ErrorInfo{
				Code:    string(appErr.Code),
				Message: appErr.Message,
				Details: appErr.Details,
			},
		}
		sendJSON(w, appErr.HTTPStatus, response)
		return
	}

	// Generic error
	response := APIResponse{
		Success: false,
		Error: &ErrorInfo{
			Code:    "INTERNAL_ERROR",
			Message: "Internal server error",
		},
	}
	sendJSON(w, http.StatusInternalServerError, response)
}

// BadRequest sends a bad request error
func BadRequest(w http.ResponseWriter, message string) {
	response := APIResponse{
		Success: false,
		Error: &ErrorInfo{
			Code:    "BAD_REQUEST",
			Message: message,
		},
	}
	sendJSON(w, http.StatusBadRequest, response)
}

// Unauthorized sends an unauthorized error
func Unauthorized(w http.ResponseWriter, message string) {
	response := APIResponse{
		Success: false,
		Error: &ErrorInfo{
			Code:    "UNAUTHORIZED",
			Message: message,
		},
	}
	sendJSON(w, http.StatusUnauthorized, response)
}

// Forbidden sends a forbidden error
func Forbidden(w http.ResponseWriter, message string) {
	response := APIResponse{
		Success: false,
		Error: &ErrorInfo{
			Code:    "FORBIDDEN",
			Message: message,
		},
	}
	sendJSON(w, http.StatusForbidden, response)
}

// NotFound sends a not found error
func NotFound(w http.ResponseWriter, message string) {
	response := APIResponse{
		Success: false,
		Error: &ErrorInfo{
			Code:    "NOT_FOUND",
			Message: message,
		},
	}
	sendJSON(w, http.StatusNotFound, response)
}

// InternalError sends an internal server error
func InternalError(w http.ResponseWriter, message string) {
	response := APIResponse{
		Success: false,
		Error: &ErrorInfo{
			Code:    "INTERNAL_ERROR",
			Message: message,
		},
	}
	sendJSON(w, http.StatusInternalServerError, response)
}

// sendJSON sends JSON response
func sendJSON(w http.ResponseWriter, statusCode int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(data)
}
