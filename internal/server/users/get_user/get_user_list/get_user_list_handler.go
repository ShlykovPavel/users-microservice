package get_user_list

import (
	"context"
	"github.com/ShlykovPavel/users-microservice/internal/lib/api/query_params"
	resp "github.com/ShlykovPavel/users-microservice/internal/lib/api/response"
	"github.com/ShlykovPavel/users-microservice/internal/lib/services/user_service"
	"github.com/ShlykovPavel/users-microservice/internal/storage/database/repositories/users_db"
	"log/slog"
	"net/http"
	"time"
)

func GetUserList(logger *slog.Logger, userDbRepository users_db.UserRepository, timeout time.Duration) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "internal/server/users/get_user/get_user_list/get_user_list_handler.go/get_user_list"
		log := logger.With(slog.String("op", op))

		ctx, cancel := context.WithTimeout(r.Context(), timeout)
		defer cancel()
		requestQuery := r.URL.Query()
		parsedQuery, err := query_params.ParseStandardQueryParams(requestQuery, log)
		if err != nil {
			log.Error("Ошибка парсинга параметров", "error", err, "request", requestQuery)
			resp.RenderResponse(w, r, http.StatusBadRequest, resp.Error("Ошибка параметров запроса"))
			return
		}

		userList, err := user_service.GetUserList(log, userDbRepository, ctx, parsedQuery)
		if err != nil {
			log.Error("Error while getting user list", "error", err)
			resp.RenderResponse(w, r, http.StatusInternalServerError, resp.Error("Something went wrong, while getting user list"))
			return
		}
		log.Debug("Successful get users list")
		resp.RenderResponse(w, r, http.StatusOK, userList)
		return

	}
}
