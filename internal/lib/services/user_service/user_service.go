package user_service

import (
	"context"
	"errors"
	"github.com/ShlykovPavel/users-microservice/internal/lib/api/models/users/get_user_by_id"
	"github.com/ShlykovPavel/users-microservice/internal/storage/database/repositories/users_db"
	"log/slog"
	"strconv"
)

func GetUser(log *slog.Logger, userRepository users_db.UserRepository, userId int64, ctx context.Context) (get_user_by_id.UserInfo, error) {
	const op = "internal/lib/services/user_service/user_service.go/GetUser"
	log = log.With(slog.String("op", op),
		slog.String("UserId", strconv.FormatInt(userId, 10)))

	userInfo, err := userRepository.GetUser(ctx, userId)
	if err != nil {
		if errors.Is(err, users_db.ErrUserNotFound) {
			log.Debug("Пользователь не найден", "err", err)
			return get_user_by_id.UserInfo{}, err
		}
		log.Error("Ошибка поиска пользователя в БД", "err", err)
		return get_user_by_id.UserInfo{}, err
	}
	return get_user_by_id.UserInfo{
		Email:     userInfo.Email,
		Phone:     userInfo.Phone,
		LastName:  userInfo.LastName,
		FirstName: userInfo.FirstName,
	}, nil

}
