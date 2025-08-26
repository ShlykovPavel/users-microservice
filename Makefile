tidy:
	go mod tidy

run_tests:
	go test -v ./...

add_migration:
	migrate create -ext sql -dir internal/storage/database/migration -seq create_users_table

run_migration_up:
	migrate -path internal/storage/database/migration -database "postgresql://postgres:mysecretpassword@localhost:5433/user_service_db?sslmode=disable" -verbose up

run_migration_down:
	migrate -path internal/storage/database/migration -database "postgresql://postgres:mysecretpassword@localhost:5433/user_service_db?sslmode=disable" -verbose down

force_migration:
	migrate -path internal/storage/database/migration -database "postgresql://postgres:mysecretpassword@localhost:5433/user_service_db?sslmode=disable" force 7

swagger_generation:
	swag init -g cmd/users_service/main.go --output docs --parseDependency --parseInternal