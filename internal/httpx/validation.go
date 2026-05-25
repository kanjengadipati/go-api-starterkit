package httpx

import (
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
)

type FieldError struct {
	Field   string `json:"field"`
	Message string `json:"message"`
}

func FormatValidationError(err error) []FieldError {
	if err == nil {
		return nil
	}

	var errors []FieldError
	validationErrors, ok := err.(validator.ValidationErrors)
	if !ok {
		return []FieldError{
			{
				Field:   "request",
				Message: err.Error(),
			},
		}
	}

	for _, e := range validationErrors {
		errors = append(errors, FieldError{
			Field:   e.Field(),
			Message: e.Tag(),
		})
	}

	return errors
}

func ValidationError(c *gin.Context, errors interface{}) {
	if message, ok := errors.(string); ok && strings.Contains(message, "request body too large") {
		ErrorWithCode(c, 413, "REQUEST_BODY_TOO_LARGE", "Request body is too large")
		return
	}
	if fields, ok := errors.([]FieldError); ok {
		for _, field := range fields {
			if strings.Contains(field.Message, "request body too large") {
				ErrorWithCode(c, 413, "REQUEST_BODY_TOO_LARGE", "Request body is too large")
				return
			}
		}
	}
	c.JSON(400, Envelope{
		Status:  "error",
		Message: "Validation failed",
		Errors:  errors,
	})
}
