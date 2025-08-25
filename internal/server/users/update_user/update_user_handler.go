package update_user

import (
	"context"
	"errors"
	"github.com/ShlykovPavel/users-microservice/internal/lib/api/body"
	"github.com/ShlykovPavel/users-microservice/internal/lib/api/models/users/create_user"
	"github.com/ShlykovPavel/users-microservice/internal/lib/api/models/users/update_user"
	resp "github.com/ShlykovPavel/users-microservice/internal/lib/api/response"
	"github.com/ShlykovPavel/users-microservice/internal/lib/services/user_service"
	"github.com/ShlykovPavel/users-microservice/internal/storage/database/repositories/users_db"
	"github.com/go-chi/chi/v5"
	"log/slog"
	"net/http"
	"strconv"
	"time"
)

func UpdateUserHandler(log *slog.Logger, userRepository users_db.UserRepository, timeout time.Duration) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "server/users.UpdateUser"
		log = log.With(slog.String("op", op))

		userID := chi.URLParam(r, "id")
		if userID == "" {
			log.Error("User ID is empty")
			resp.RenderResponse(w, r, http.StatusBadRequest, resp.Error("User ID is required"))
			return
		}
		id, err := strconv.ParseInt(userID, 10, 64)
		if err != nil {
			log.Error("User ID is invalid", "error", err)
			resp.RenderResponse(w, r, http.StatusBadRequest, resp.Error("Invalid user ID"))
			return
		}

		ctx, cancel := context.WithTimeout(r.Context(), timeout)
		defer cancel()

		var UpdateUserDto update_user.UpdateUserDto
		err = body.DecodeAndValidateJson(r, &UpdateUserDto)
		if err != nil {
			log.Error("Failed decoding body", "err", err, "body", r.Body)
			resp.RenderResponse(w, r, http.StatusInternalServerError, resp.Error("Failed reading body"))
			return
		}

		err = user_service.UpdateUser(log, userRepository, ctx, UpdateUserDto, id)
		if err != nil {
			if errors.Is(err, users_db.ErrUserNotFound) {
				resp.RenderResponse(w, r, http.StatusNotFound, resp.Error(err.Error()))
				return
			}
			log.Error("Failed to update user", "err", err)
			resp.RenderResponse(w, r, http.StatusInternalServerError, resp.Error("Failed updating user"))
			return
		}
		log.Debug("Successfully updated user", "id", id)
		resp.RenderResponse(w, r, http.StatusOK, create_user.CreateUserResponse{UserID: id, Response: resp.OK()})

	}
}
