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

	sqlRepo = NewRepository()
	sqlService = NewService(sqlRepo)


	userSvc := users.GetUserService()
	dbSvc := databases.GetDatabaseService()


	sqlPgController = NewPostgresSqlQueryController(sqlService, userSvc, dbSvc)
}

func GetSqlQueryRepository() *Repository { return sqlRepo }
func GetSqlQueryService() *Service       { return sqlService }


func GetPostgresSqlQueryController() *PostgresSqlQueryController {
	return sqlPgController
}
