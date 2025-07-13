package users_delete

import (
	"context"
	"errors"
	resp "github.com/ShlykovPavel/users-microservice/internal/lib/api/response"
	"github.com/ShlykovPavel/users-microservice/internal/lib/services/user_service"
	"github.com/ShlykovPavel/users-microservice/internal/storage/database/repositories/users_db"
	"github.com/go-chi/chi/v5"
	"log/slog"
	"net/http"
	"strconv"
	"time"
)

func DeleteUserHandler(logger *slog.Logger, userDbRepository users_db.UserRepository, timeout time.Duration) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		log := logger.With(slog.String("op", "internal/server/users/delete/delete_user_handler.go/DeleteUserHandler"))
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

		err = user_service.DeleteUser(logger, userDbRepository, ctx, id)
		if err != nil {
			if errors.Is(err, users_db.ErrUserNotFound) {
				log.Error("User not found", "error", err)
				resp.RenderResponse(w, r, http.StatusNotFound, resp.Error("User not found"))
				return
			}
			log.Error("Error deleting user", "error", err)
			resp.RenderResponse(w, r, http.StatusInternalServerError, resp.Error("Error deleting user"))
			return
		}
		log.Info("Deleted user", "userID", userID)
		resp.RenderResponse(w, r, http.StatusNoContent, nil)
	}

}
