package apperr

import (
	"errors"
	"net/http"
)

// AppError is a typed domain error that carries an HTTP status code.
// Handlers unwrap this to decide the response status — business logic never imports net/http.
type AppError struct {
	Code    int    // HTTP status code
	Message string // safe message returned to client
	Err     error  // internal cause — logged, never exposed
}

func (e *AppError) Error() string {
	return e.Message // always return safe client message — internal cause is for Unwrap() only
}

func (e *AppError) Unwrap() error { return e.Err }

// -- Constructors -------------------------------------------------------------

func NotFound(msg string, cause ...error) *AppError {
	return &AppError{Code: http.StatusNotFound, Message: msg, Err: wrap(cause)}
}

func Unauthorized(msg string, cause ...error) *AppError {
	return &AppError{Code: http.StatusUnauthorized, Message: msg, Err: wrap(cause)}
}

func Forbidden(msg string, cause ...error) *AppError {
	return &AppError{Code: http.StatusForbidden, Message: msg, Err: wrap(cause)}
}

func Conflict(msg string, cause ...error) *AppError {
	return &AppError{Code: http.StatusConflict, Message: msg, Err: wrap(cause)}
}

func BadRequest(msg string, cause ...error) *AppError {
	return &AppError{Code: http.StatusBadRequest, Message: msg, Err: wrap(cause)}
}

func Internal(msg string, cause ...error) *AppError {
	return &AppError{Code: http.StatusInternalServerError, Message: msg, Err: wrap(cause)}
}

// -- Handler helper -----------------------------------------------------------

// HTTPStatus extracts the HTTP status code from any error.
// Returns 500 for unknown errors so handlers always have a valid code.
func HTTPStatus(err error) (int, string) {
	var appErr *AppError
	if errors.As(err, &appErr) {
		return appErr.Code, appErr.Message
	}
	return http.StatusInternalServerError, "internal server error"
}

func wrap(causes []error) error {
	if len(causes) > 0 {
		return causes[0]
	}
	return nil
}
