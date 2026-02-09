package errorx

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

type ErrorResponse struct {
	Message string            `json:"message"`
	Err     map[string]string `json:"errors"`
	Errors  error             `json:"-"`
}

func MapErrorToStatusCode(errType ErrorType) int {
	switch errType {
	case ErrTypeNotFound:
		return http.StatusNotFound
	case ErrTypeConflict:
		return http.StatusConflict
	case ErrTypeValidation, ErrInsufficientLimit, ErrTenorNotAvail:
		return http.StatusBadRequest
	case ErrTypeInternal:
		return http.StatusInternalServerError
	default:
		return http.StatusInternalServerError
	}
}

func SendError(c *gin.Context, log *zap.Logger, err error) {
	var appErr *AppError

	if errors.As(err, &appErr) {
		statusCode := MapErrorToStatusCode(appErr.Type)

		message := appErr.Message
		if statusCode == http.StatusInternalServerError {
			message = "internal server error, please try again later"
		}

		log.Error("error message", zap.Error(appErr.Err))

		c.JSON(statusCode, ErrorResponse{
			Message: message,
			Err:     appErr.Fields,
		})
		return
	}

	c.JSON(http.StatusInternalServerError, ErrorResponse{
		Message: "unexpected system error",
	})
}
