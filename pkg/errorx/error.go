package errorx

import "fmt"

type ErrorType string

const (
	ErrTypeNotFound      ErrorType = "resource not found"
	ErrTypeConflict      ErrorType = "resource already exists"
	ErrTypeInternal      ErrorType = "internal server error"
	ErrTypeValidation    ErrorType = "invalid validation"
	ErrInsufficientLimit ErrorType = "insufficient limit amount"
	ErrTenorNotAvail     ErrorType = "tenor option not available"
)

type AppError struct {
	Type    ErrorType         `json:"error_type"`
	Message string            `json:"message"`
	Fields  map[string]string `json:"fields,omitempty"`
	Err     error             `json:"-"`
}

func (e *AppError) Error() string {
	return fmt.Sprintf("%s: %s", e.Type, e.Message)
}

func NewValidationError(fields map[string]string) *AppError {
	return &AppError{
		Type:    ErrTypeValidation,
		Message: "invalid input parameters",
		Fields:  fields,
	}
}

func NewError(errType ErrorType, msg string, err error) *AppError {
	return &AppError{
		Type:    errType,
		Message: msg,
		Err:     err,
	}
}
