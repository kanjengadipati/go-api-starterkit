package httpx

import (
	"errors"
	"net/http"

	"pleco-api/internal/domain"

	"github.com/gin-gonic/gin"
)

func HandleError(c *gin.Context, err error) {
	if err == nil {
		return
	}

	var apiErr *domain.APIError
	if errors.As(err, &apiErr) {
		status := apiErr.HTTPStatus
		if status == 0 {
			status = http.StatusInternalServerError
		}
		ErrorWithCode(c, status, string(apiErr.Code), apiErr.Message)
		return
	}

	ErrorWithCode(c, http.StatusInternalServerError, ErrCodeInternalError, "internal server error")
}
