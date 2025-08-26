package body

import (
	"errors"
	validators "github.com/ShlykovPavel/users-microservice/internal/lib/api/validator"
	"github.com/go-chi/render"
	"github.com/go-playground/validator"
	"log/slog"
	"net/http"
)

// Ошибки
var ErrDecodeJSON = errors.New("failed to decode JSON")

// DecodeAndValidateJson декодирует JSON и валидирует структуру
func DecodeAndValidateJson(r *http.Request, v interface{}) error {
	// Декодируем JSON
	if err := render.DecodeJSON(r.Body, v); err != nil {
		slog.Default().Error("DecodeAndValidateJson: error decoding body or validating", "error", err)
		return ErrDecodeJSON
	}

	// Валидируем структуру
	if err := validators.GetValidator().Struct(v); err != nil {
		validateErr := err.(validator.ValidationErrors)
		return validateErr
	}

	return nil
}
