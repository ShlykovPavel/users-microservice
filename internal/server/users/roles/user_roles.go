package roles

import (
	"context"
	"errors"
	"fmt"
	resp "github.com/ShlykovPavel/users-microservice/internal/lib/api/response"
	"github.com/ShlykovPavel/users-microservice/internal/server/users"
	"github.com/ShlykovPavel/users-microservice/internal/storage/database/repositories/users_db"
	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"log/slog"
	"net/http"
	"strconv"
)

func CheckAdminInDB(poll *pgxpool.Pool, log *slog.Logger) error {
	userRepository := users_db.NewUsersDB(poll, log)

	user, err := userRepository.CheckAdminInDB(context.Background())
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			log.Error("no admin role found", "error", err)
			//Создаём захэшированный пароль
			passwordHash, err := users.HashUserPassword("password", log)
			if err != nil {
				log.Error("Error while hashing password", "err", err)
				return fmt.Errorf("failed to hash password: %w", err)
			}
			err = userRepository.AddFirstAdmin(context.Background(), passwordHash)
			if err != nil {
				log.Error("error adding admin role", "error", err)
				return err
			}
		}
		log.Error("error checking admin role", "error", err)
		return err
	}
	log.Info("admin role check ok. no need to create admin role. Found admin:", "user", user)
	return nil
}

func SetAdminRole(poll *pgxpool.Pool, log *slog.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
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
		userRepository := users_db.NewUsersDB(poll, log)
		err = userRepository.SetAdminRole(context.Background(), id)
		if err != nil {
			if errors.Is(err, users_db.ErrUserNotFound) {
				log.Debug("user not found", "error", err)
				resp.RenderResponse(w, r, http.StatusBadRequest, resp.Error("User not found"))
				return
			}

			log.Error("error setting admin role", "error", err)
			resp.RenderResponse(w, r, http.StatusInternalServerError, resp.Error("Failed to set admin role"))
			return
		}
	}
}
