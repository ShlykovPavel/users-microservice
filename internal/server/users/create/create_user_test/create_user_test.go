package create_user_test

import (
	"bytes"
	"context"
	"encoding/json"
	"github.com/ShlykovPavel/users-microservice/internal/lib/api/models/users/create_user"
	users "github.com/ShlykovPavel/users-microservice/internal/server/users/create"
	"github.com/ShlykovPavel/users-microservice/internal/storage/database/repositories/users_db"
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
	return args.Error(1)
}

func TestCreateUser(t *testing.T) {
	tests := []struct {
		testName       string
		input          create_user.UserCreate
		setupMock      func(*MockUserRepository)
		expectedStatus int
		expectedBody   string
	}{
		{
			testName: "success creating user",
			input: create_user.UserCreate{
				FirstName: "Ryan",
				LastName:  "Gosling",
				Email:     "ryanGosling@gmail.com",
				Password:  "password",
				Phone:     "+78951235678",
			},
			setupMock: func(mockRepo *MockUserRepository) {
				mockRepo.On("CreateUser", mock.Anything, mock.MatchedBy(func(u *create_user.UserCreate) bool {
					return u.Email == "ryanGosling@gmail.com" && u.FirstName == "Ryan"
				})).Return(int64(123), nil).Once()
			},
			expectedStatus: http.StatusCreated,
			expectedBody:   `{"status":"OK","id":123}`,
		},
		{
			testName: "wrong type in field",
			input: create_user.UserCreate{
				FirstName: "", // Пустое поле вызовет ошибку валидации
				LastName:  "Gosling",
				Email:     "ryanGosling@gmail.com",
				Password:  "password",
				Phone:     "+78951235678",
			},
			setupMock: func(mockRepo *MockUserRepository) {
				//	Не настраиваем мок так как будет ошибка
			},
			expectedStatus: http.StatusBadRequest,
			expectedBody:   `{"status":"ERROR","error":"field FirstName is required"}`,
		},
		{
			testName: "email already exists",
			input: create_user.UserCreate{
				FirstName: "Ryan",
				LastName:  "Gosling",
				Email:     "ryanGosling@gmail.com",
				Password:  "password",
				Phone:     "+78951235678",
			},
			setupMock: func(mockRepo *MockUserRepository) {
				mockRepo.On("CreateUser", mock.Anything, mock.AnythingOfType("*create_user.UserCreate")).
					Return(int64(0), users_db.ErrEmailAlreadyExists).Once()
			},
			expectedStatus: http.StatusBadRequest,
			expectedBody:   `{"status":"ERROR","error":"Пользователь с email уже существует. "}`,
		},
		{
			testName: "timeout error",
			input: create_user.UserCreate{
				FirstName: "Ryan",
				LastName:  "Gosling",
				Email:     "ryanGosling@gmail.com",
				Password:  "password",
				Phone:     "+78951235678",
			},
			setupMock: func(mockRepo *MockUserRepository) {
				mockRepo.On("CreateUser", mock.Anything, mock.AnythingOfType("*create_user.UserCreate")).
					Run(func(args mock.Arguments) {
						time.Sleep(6 * time.Second) // Задержка больше таймаута (5 секунд)
					}).
					Return(int64(0), context.DeadlineExceeded).Once()
			},
			expectedStatus: http.StatusGatewayTimeout,
			expectedBody:   `{"status":"ERROR","error":"Request timed out or canceled"}`,
		},
	}

	//mockRepo := new(MockUserRepository)
	//timeout := time.Duration(5 * time.Second)
	for _, test := range tests {
		t.Run(test.testName, func(t *testing.T) {
			logger := slog.Default()
			timeout := 5 * time.Second
			mockRepo := new(MockUserRepository)
			handler := users.CreateUser(logger, mockRepo, timeout)

			// Настраиваем мок
			test.setupMock(mockRepo)

			// Создаём запрос
			body, _ := json.Marshal(test.input)
			req := httptest.NewRequest(http.MethodPost, "/register", bytes.NewReader(body))
			req = req.WithContext(context.WithValue(context.Background(), middleware.RequestIDKey, "test-id"))
			w := httptest.NewRecorder()

			// Вызываем хендлер
			handler.ServeHTTP(w, req)

			// Проверяем статус
			require.Equal(t, test.expectedStatus, w.Code, "unexpected HTTP status")

			// Проверяем тело ответа
			var respBody map[string]interface{}
			err := json.Unmarshal(w.Body.Bytes(), &respBody)
			require.NoError(t, err, "response should be valid JSON")
			require.Contains(t, w.Body.String(), test.expectedBody, "unexpected response body")

			// Проверяем, что все ожидаемые вызовы мока выполнены
			mockRepo.AssertExpectations(t)
		})
	}

}
