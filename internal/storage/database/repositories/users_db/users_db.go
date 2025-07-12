package users_db

import (
	"context"
	"errors"
	"fmt"
	"github.com/ShlykovPavel/users-microservice/internal/lib/api/models/users/create_user"
	"github.com/ShlykovPavel/users-microservice/internal/storage/database"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
	"log/slog"
	"strings"
)

var ErrEmailAlreadyExists = errors.New("Пользователь с email уже существует. ")
var ErrUserNotFound = errors.New("Пользователь не найден ")

type UserRepository interface {
	CreateUser(ctx context.Context, userinfo *create_user.UserCreate) (int64, error)
	GetUser(ctx context.Context, userId int64) (UserInfo, error)
	GetUserList(ctx context.Context, search string, limit, offset int, sort string) (UserListResult, error)
	CheckAdminInDB(ctx context.Context) (UserInfo, error)
	AddFirstAdmin(ctx context.Context, passwordHash string) error
}

type UserRepositoryImpl struct {
	db  *pgxpool.Pool
	log *slog.Logger
}

// UserInfo Структура с информацие о пользователе
type UserInfo struct {
	ID           int64
	FirstName    string
	LastName     string
	Email        string
	PasswordHash string
	Role         string
	Phone        string
}
type UserListResult struct {
	Users []UserInfo
	Total int64
}

func NewUsersDB(dbPoll *pgxpool.Pool, log *slog.Logger) *UserRepositoryImpl {
	return &UserRepositoryImpl{
		db:  dbPoll,
		log: log,
	}
}

// CreateUser Создание пользователя
// Принимает:
// ctx - внешний контекст, что б вызывающая сторона могла контролировать запрос (например выставить таймаут)
// userinfo - структуру UserInfo с необходимыми полями для добавления
//
// После запроса возвращается Id созданного пользователя
func (us *UserRepositoryImpl) CreateUser(ctx context.Context, userinfo *create_user.UserCreate) (int64, error) {
	query := `
INSERT INTO users (first_name, last_name, email, password, Role, phone)
VALUES ($1, $2, $3, $4, 'user', $5)
RETURNING id`
	var id int64
	err := us.db.QueryRow(ctx, query, userinfo.FirstName, userinfo.LastName, userinfo.Email, userinfo.Password, userinfo.Phone).Scan(&id)
	if err != nil {
		if ctxErr := database.DbCtxError(ctx, err, us.log); ctxErr != nil {
			return 0, ctxErr
		}
		dbErr := database.PsqlErrorHandler(err)
		if pgErr, ok := err.(*pgconn.PgError); ok && pgErr.Code == database.PSQLUniqueError {
			return 0, ErrEmailAlreadyExists
		}
		return 0, dbErr
	}

	return id, nil
}

func (us *UserRepositoryImpl) GetUser(ctx context.Context, userId int64) (UserInfo, error) {
	query := `SELECT first_name, last_name, email, password, role, phone FROM users WHERE id = $1`

	var user UserInfo
	err := us.db.QueryRow(ctx, query, userId).Scan(
		&user.FirstName,
		&user.LastName,
		&user.Email,
		&user.PasswordHash,
		&user.Role,
		&user.Phone)
	if errors.Is(err, pgx.ErrNoRows) {
		return UserInfo{}, ErrUserNotFound
	}
	if err != nil {
		if ctxErr := database.DbCtxError(ctx, err, us.log); ctxErr != nil {
			return UserInfo{}, ctxErr
		}
		dbErr := database.PsqlErrorHandler(err)
		return UserInfo{}, dbErr
	}
	return user, nil
}

func (us *UserRepositoryImpl) GetUserList(ctx context.Context, search string, limit, offset int, sort string) (UserListResult, error) {
	// Базовый SQL-запрос для пользователей
	query := "SELECT id, first_name, last_name, email, role, phone FROM users"
	countQuery := "SELECT COUNT(*) FROM users"
	args := []interface{}{}
	countArgs := []interface{}{}

	// Фильтрация по search
	if search != "" {
		query += " WHERE first_name ILIKE $1 OR last_name ILIKE $1 OR email ILIKE $1"
		countQuery += " WHERE first_name ILIKE $1 OR last_name ILIKE $1 OR email ILIKE $1"
		args = append(args, "%"+search+"%")
		countArgs = append(countArgs, "%"+search+"%")
	}

	// Сортировка
	if sort != "" {
		parts := strings.Split(sort, ":")
		if len(parts) == 2 && (parts[1] == "asc" || parts[1] == "desc") {
			// Простая проверка допустимых полей
			switch parts[0] {
			case "id", "first_name", "last_name", "email", "role", "phone":
				query += fmt.Sprintf(" ORDER BY %s %s", parts[0], strings.ToUpper(parts[1]))
			default:
				us.log.Warn("Invalid sort field", slog.String("field", parts[0]))
				return UserListResult{}, fmt.Errorf("invalid sort field: %s", parts[0])
			}
		} else {
			us.log.Warn("Invalid sort format", slog.String("sort", sort))
			return UserListResult{}, fmt.Errorf("invalid sort format: %s", sort)
		}
	}

	// Пагинация
	query += fmt.Sprintf(" LIMIT $%d OFFSET $%d", len(args)+1, len(args)+2)
	args = append(args, limit, offset)

	// Подсчёт total
	var total int64
	err := us.db.QueryRow(ctx, countQuery, countArgs...).Scan(&total)
	if err != nil {
		us.log.Error("Failed to count users", slog.Any("error", err))
		return UserListResult{}, fmt.Errorf("failed to count users: %w", err)
	}

	// Получение пользователей
	rows, err := us.db.Query(ctx, query, args...)
	if err != nil {
		us.log.Error("Failed to query users", slog.Any("error", err))
		return UserListResult{}, fmt.Errorf("failed to query users: %w", err)
	}
	defer rows.Close()

	var users []UserInfo
	for rows.Next() {
		var user UserInfo
		if err := rows.Scan(&user.ID, &user.FirstName, &user.LastName, &user.Email, &user.Role, &user.Phone); err != nil {
			us.log.Error("Error scanning user row", slog.Any("error", err))
			return UserListResult{}, fmt.Errorf("error scanning user row: %w", err)
		}
		users = append(users, user)
	}
	if err := rows.Err(); err != nil {
		us.log.Error("Error reading rows", slog.Any("error", err))
		return UserListResult{}, fmt.Errorf("error reading rows: %w", err)
	}

	return UserListResult{
		Users: users,
		Total: total,
	}, nil
}

func (us *UserRepositoryImpl) CheckAdminInDB(ctx context.Context) (UserInfo, error) {
	query := `SELECT id, first_name, last_name, email, password FROM users WHERE Role LIKE '%admin%'`

	var user UserInfo
	err := us.db.QueryRow(ctx, query).Scan(
		&user.ID,
		&user.FirstName,
		&user.LastName,
		&user.Email,
		&user.PasswordHash,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return UserInfo{}, pgx.ErrNoRows
	}
	if err != nil {
		dbErr := database.PsqlErrorHandler(err)
		return UserInfo{}, dbErr
	}
	return user, nil
}

func (us *UserRepositoryImpl) AddFirstAdmin(ctx context.Context, passwordHash string) error {
	query := `INSERT INTO users (id, first_name, last_name, email, password, Role, phone) VALUES ($1, $2, $3, $4, $5, $6, $7)`

	_, err := us.db.Exec(ctx, query, 0, "Admin first name", "Admin last name", "admin@admin.com", passwordHash, "admin", +78901234567)
	if err != nil {
		dbErr := database.PsqlErrorHandler(err)
		return dbErr
	}
	return nil
}

func (us *UserRepositoryImpl) SetAdminRole(ctx context.Context, id int64) error {
	query := `UPDATE users SET Role = 'admin' WHERE id = $1`
	result, err := us.db.Exec(ctx, query, id)
	if err != nil {
		dbErr := database.PsqlErrorHandler(err)
		return dbErr
	}
	if result.RowsAffected() == 0 {
		return ErrUserNotFound
	}
	return nil
}
