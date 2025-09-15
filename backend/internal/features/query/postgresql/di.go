package sqlquery

import (
	"postgresus-backend/internal/features/databases"
	"postgresus-backend/internal/features/users"
)

var (
	sqlRepo        *Repository
	sqlService     *Service
	sqlPgController *PostgresSqlQueryController
)

func init() {
	// репозиторій та сервіс
	sqlRepo = NewRepository()
	sqlService = NewService(sqlRepo)

	// залежності з сусідніх модулів через їх DI
	userSvc := users.GetUserService()
	dbSvc := databases.GetDatabaseService()

	// контролер
	sqlPgController = NewPostgresSqlQueryController(sqlService, userSvc, dbSvc)
}

func GetSqlQueryRepository() *Repository { return sqlRepo }
func GetSqlQueryService() *Service       { return sqlService }

// Головне: цей контролер реєструємо у router
func GetPostgresSqlQueryController() *PostgresSqlQueryController {
	return sqlPgController
}
