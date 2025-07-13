package user_service

import (
	"context"
	"errors"
	"github.com/ShlykovPavel/users-microservice/internal/lib/api/models/users/get_user_by_id"
	"github.com/ShlykovPavel/users-microservice/internal/lib/api/models/users/get_users_list"
	"github.com/ShlykovPavel/users-microservice/internal/lib/api/query_params"
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

// GetUserList retrieves a list of users from the repository and converts them to DTOs.
// It takes a logger, user repository, and context as input.
// Returns a slice of UserInfoList DTOs or an error if the operation fails.
func GetUserList(log *slog.Logger, userRepository users_db.UserRepository, ctx context.Context, queryParams query_params.ListUsersParams) (get_users_list.UsersList, error) {
	const op = "internal/lib/services/user_service/user_service.go/GetUserList"
	log = log.With(slog.String("op", op))

	result, err := userRepository.GetUserList(ctx, queryParams.Search, queryParams.Limit, queryParams.Offset, queryParams.Sort)
	if err != nil {
		log.Error("Failed to get users list", "err", err)
		return get_users_list.UsersList{}, err
	}
	userList := make([]get_users_list.UserInfoList, 0, len(result.Users))
	for _, user := range result.Users {
		userInfo := get_users_list.UserInfoList{
			Id:        user.ID,
			Email:     user.Email,
			Phone:     user.Phone,
			LastName:  user.LastName,
			FirstName: user.FirstName,
			Role:      user.Role,
		}
		userList = append(userList, userInfo)
	}
	metaData := get_users_list.UsersListMetaData{
		Page:   queryParams.Page,
		Total:  result.Total,
		Limit:  queryParams.Limit,
		Offset: queryParams.Offset,
	}
	userDto := get_users_list.UsersList{
		Users: userList,
		Meta:  metaData,
	}
	return userDto, nil

}
