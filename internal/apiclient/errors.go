package apiclient

import (
	"errors"
	"fmt"
)

const (
	ExitCodeSuccess   = 0
	ExitCodeUsage     = 2
	ExitCodeConfig    = 2
	ExitCodeAuth      = 3
	ExitCodeForbidden = 4
	ExitCodeNotFound  = 5
	ExitCodeConflict  = 6
	ExitCodeAPIError  = 7
)

type APIError struct {
	StatusCode int                    `json:"statusCode"`
	Code       string                 `json:"code"`
	Message    string                 `json:"message"`
	Details    map[string]interface{} `json:"details,omitempty"`
}

type ConfigError struct{ Err error }

func (e *ConfigError) Error() string { return e.Err.Error() }
func (e *ConfigError) Unwrap() error { return e.Err }

type UsageError struct{ Err error }

func (e *UsageError) Error() string { return e.Err.Error() }
func (e *UsageError) Unwrap() error { return e.Err }

type TransportError struct{ Err error }

func (e *TransportError) Error() string { return e.Err.Error() }
func (e *TransportError) Unwrap() error { return e.Err }

func (e *APIError) Error() string {
	if e.Code != "" {
		return fmt.Sprintf("[%s] %s (HTTP %d)", e.Code, e.Message, e.StatusCode)
	}
	return fmt.Sprintf("%s (HTTP %d)", e.Message, e.StatusCode)
}

func GetExitCodeForError(err error) int {
	if err == nil {
		return ExitCodeSuccess
	}
	var configErr *ConfigError
	if errors.As(err, &configErr) {
		return ExitCodeConfig
	}
	var usageErr *UsageError
	if errors.As(err, &usageErr) {
		return ExitCodeUsage
	}
	var transportErr *TransportError
	if errors.As(err, &transportErr) {
		return ExitCodeAPIError
	}
	var apiErr *APIError
	if errors.As(err, &apiErr) {
		switch apiErr.StatusCode {
		case 401:
			return ExitCodeAuth
		case 403:
			return ExitCodeForbidden
		case 404:
			return ExitCodeNotFound
		case 409:
			return ExitCodeConflict
		case 400, 422:
			return ExitCodeUsage
		default:
			return ExitCodeAPIError
		}
	}
	return ExitCodeUsage
}
