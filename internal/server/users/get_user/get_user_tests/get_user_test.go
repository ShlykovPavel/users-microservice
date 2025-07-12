package get_user_tests

import (
	"context"
	"encoding/json"
	"errors"
	"github.com/ShlykovPavel/users-microservice/internal/lib/api/models/users/create_user"
	"github.com/ShlykovPavel/users-microservice/internal/lib/api/models/users/get_user_by_id"
	"github.com/ShlykovPavel/users-microservice/internal/lib/api/response"
	"github.com/ShlykovPavel/users-microservice/internal/server/users/get_user"
	"github.com/ShlykovPavel/users-microservice/internal/storage/database/repositories/users_db"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

type MockUserRepository struct {
	mock.Mock
}

func (m *MockUserRepository) CreateUser(ctx context.Context, userinfo *create_user.UserCreate) (int64, error) {
	args := m.Called(ctx, userinfo)
	return args.Get(0).(int64), args.Error(1)
}

func (m *MockUserRepository) GetUser(ctx context.Context, userId int64) (users_db.UserInfo, error) {
	args := m.Called(ctx, userId)
	return args.Get(0).(users_db.UserInfo), args.Error(1)
}

func (m *MockUserRepository) CheckAdminInDB(ctx context.Context) (users_db.UserInfo, error) {
	args := m.Called(ctx)
	return args.Get(0).(users_db.UserInfo), args.Error(1)
}

func (m *MockUserRepository) AddFirstAdmin(ctx context.Context, passwordHash string) error {
	args := m.Called(ctx, passwordHash)
	return args.Error(0)
}

func (m *MockUserRepository) GetUserList(ctx context.Context, search string, limit, offset int, sort string) (users_db.UserListResult, error) {
	args := m.Called(ctx)
	return args.Get(0).(users_db.UserListResult), args.Error(1)
}
func (m *MockUserRepository) UpdateUser(ctx context.Context, id int64, firstName, lastName, email, phone, role string) error {
	args := m.Called(ctx, id, firstName, lastName, email, phone, role)
	return args.Error(0)
}

func (m *MockUserRepository) DeleteUser(ctx context.Context, id int64) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func TestGetUser(t *testing.T) {
	tests := []struct {
		name           string
		userId         string
		setupMock      func(*MockUserRepository)
		expectedStatus int
		expectedBody   interface{}
	}{
		{
			name:   "success get user",
			userId: "1",
			setupMock: func(mockRepo *MockUserRepository) {
				mockRepo.On("GetUser", mock.Anything, int64(1)).Return(users_db.UserInfo{
					Email:     "ryanGosling@gmail.com",
					FirstName: "Ryan",
					LastName:  "Gosling",
					Phone:     "+1234567890",
				}, nil).Once()
			},
			expectedStatus: http.StatusOK,
			expectedBody: get_user_by_id.UserInfo{
				Email:     "ryanGosling@gmail.com",
				FirstName: "Ryan",
				LastName:  "Gosling",
				Phone:     "+1234567890",
			},
		},
		{
			name:   "empty user id",
			userId: "",
			setupMock: func(mockRepo *MockUserRepository) {
				// Нет вызова мока, так как хендлер не доходит до обращения к репозиторию
			},
			expectedStatus: http.StatusBadRequest,
			expectedBody:   response.Error("User ID is required"),
		},
		{
			name:   "invalid user id",
			userId: "invalid",
			setupMock: func(mockRepo *MockUserRepository) {
				// Нет вызова мока, так как хендлер не доходит до обращения к репозиторию
			},
			expectedStatus: http.StatusBadRequest,
			expectedBody:   response.Error("Invalid user ID"),
		},
		{
			name:   "user not found",
			userId: "1",
			setupMock: func(mockRepo *MockUserRepository) {
				mockRepo.On("GetUser", mock.Anything, int64(1)).Return(users_db.UserInfo{}, users_db.ErrUserNotFound).Once()
			},
			expectedStatus: http.StatusNotFound,
			expectedBody:   response.Error("User not found"),
		},
		{
			name:   "internal server error",
			userId: "1",
			setupMock: func(mockRepo *MockUserRepository) {
				mockRepo.On("GetUser", mock.Anything, int64(1)).Return(users_db.UserInfo{}, errors.New("database error")).Once()
			},
			expectedStatus: http.StatusInternalServerError,
			expectedBody:   response.Error("Something went wrong, while getting user"),
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			logger := slog.Default()
			timeout := 5 * time.Second
			mockRepo := new(MockUserRepository)
			handler := get_user.GetUserById(logger, mockRepo, timeout)

			// Настраиваем мок
			test.setupMock(mockRepo)

			// Создаем запрос
			req := httptest.NewRequest(http.MethodGet, "/users/"+test.userId, nil)
			rctx := chi.NewRouteContext()
			rctx.URLParams.Add("id", test.userId)
			req = req.WithContext(context.WithValue(context.Background(), chi.RouteCtxKey, rctx))
			req = req.WithContext(context.WithValue(req.Context(), middleware.RequestIDKey, "test-id"))
			w := httptest.NewRecorder()

			// Вызываем хендлер
			handler.ServeHTTP(w, req)

			// Проверяем статус
			require.Equal(t, test.expectedStatus, w.Code, "unexpected HTTP status")

			expectedJSON, err := json.Marshal(test.expectedBody)
			require.NoError(t, err, "failed to marshal expected body to JSON")

			// Сравниваем JSON-строки
			require.JSONEq(t, string(expectedJSON), w.Body.String(), "unexpected response body")

			// Проверяем, что все ожидаемые вызовы мока выполнены
			mockRepo.AssertExpectations(t)
		})
	}
}
