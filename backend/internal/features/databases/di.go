package databases

import (
	"postgresus-backend/internal/features/notifiers"
	"postgresus-backend/internal/features/users"
	"postgresus-backend/internal/util/logger"
)

var databaseRepository = &DatabaseRepository{}

var databaseService = &DatabaseService{
	databaseRepository,
	notifiers.GetNotifierService(),
	logger.GetLogger(),
	[]DatabaseCreationListener{},
	[]DatabaseRemoveListener{},
	[]DatabaseCopyListener{},
}

var databaseController = &DatabaseController{
	databaseService,
	users.GetUserService(),
}

func GetDatabaseService() *DatabaseService {
	return databaseService
}

func GetDatabaseController() *DatabaseController {
	return databaseController
}
