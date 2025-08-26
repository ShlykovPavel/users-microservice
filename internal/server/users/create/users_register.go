package users

import (
	"context"
	"errors"
	"github.com/ShlykovPavel/users-microservice/internal/lib/api/body"
	resp "github.com/ShlykovPavel/users-microservice/internal/lib/api/response"
	users "github.com/ShlykovPavel/users-microservice/internal/server/users"
	"github.com/ShlykovPavel/users-microservice/internal/storage/database/repositories/users_db"
	"github.com/ShlykovPavel/users-microservice/models/users/create_user"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-playground/validator"
	"log/slog"
	"net/http"
	"time"
)

// CreateUser godoc
// @Summary Создать пользователя
// @Description Регистрирует пользователя в системе
// @Tags Users
// @Param input body create_user.UserCreate true "Данные пользователя"
// @Success 201 {object} create_user.CreateUserResponse
// @Router /register [post]
func CreateUser(log *slog.Logger, userRepository users_db.UserRepository, timeout time.Duration) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "server/users.CreateUser"
		log = log.With(
			slog.String("operation", op),
			slog.String("request_id", middleware.GetReqID(r.Context())),
			slog.String("url", r.URL.Path))
		//Создаём контекст для управления временем обработки запроса
		ctx, cancel := context.WithTimeout(r.Context(), timeout)
		defer cancel()

		var user create_user.UserCreate
		err := body.DecodeAndValidateJson(r, &user)
		if err != nil {
			log.Error("Error while decoding request body", "err", err)
			if validationErrors, ok := err.(validator.ValidationErrors); ok {
				resp.RenderResponse(w, r, http.StatusBadRequest, resp.ValidationError(validationErrors))
				return
			}
			resp.RenderResponse(w, r, http.StatusBadRequest, resp.Error(err.Error()))
			return
		}

		//	Хешируем пароль
		passwordHash, err := users.HashUserPassword(user.Password, log)
		if err != nil {
			log.Error("Error while hashing password", "err", err)
			resp.RenderResponse(w, r, http.StatusInternalServerError, resp.Error(err.Error()))
			return
		}

		user.Password = passwordHash
		//Записываем в бд
		userId, err := userRepository.CreateUser(ctx, &user)
		if err != nil {
			log.Error("Error while creating user", "err", err)
			if errors.Is(err, users_db.ErrEmailAlreadyExists) {
				resp.RenderResponse(w, r, http.StatusBadRequest, resp.Error(
					err.Error()))
				return
			}
			if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
				log.Warn("Request canceled or timed out", slog.Any("err", err))
				resp.RenderResponse(w, r, http.StatusGatewayTimeout, resp.Error("Request timed out or canceled"))
				return
			}
			resp.RenderResponse(w, r, http.StatusInternalServerError, resp.Error(err.Error()))
			return
		}

		log.Info("Created user", "user id", userId)
		resp.RenderResponse(w, r, http.StatusCreated, create_user.CreateUserResponse{
			resp.OK(),
			userId,
		})
	}
}
