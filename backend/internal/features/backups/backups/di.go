package backups

import (
	"postgresus-backend/internal/features/backups/backups/usecases"
	backups_config "postgresus-backend/internal/features/backups/config"
	"postgresus-backend/internal/features/databases"
	"postgresus-backend/internal/features/notifiers"
	"postgresus-backend/internal/features/storages"
	"postgresus-backend/internal/features/users"
	"postgresus-backend/internal/util/logger"
	"time"
)

var backupRepository = &BackupRepository{}
var backupService = &BackupService{
	databases.GetDatabaseService(),
	storages.GetStorageService(),
	backupRepository,
	notifiers.GetNotifierService(),
	notifiers.GetNotifierService(),
	backups_config.GetBackupConfigService(),
	usecases.GetCreateBackupUsecase(),
	logger.GetLogger(),
	[]BackupRemoveListener{},
}

var backupBackgroundService = &BackupBackgroundService{
	backupService,
	backupRepository,
	backups_config.GetBackupConfigService(),
	storages.GetStorageService(),
	time.Now().UTC(),
	logger.GetLogger(),
}

var backupController = &BackupController{
	backupService,
	users.GetUserService(),
}

func SetupDependencies() {
	backups_config.
		GetBackupConfigService().
		SetDatabaseStorageChangeListener(backupService)

	databases.GetDatabaseService().AddDbRemoveListener(backupService)
	databases.GetDatabaseService().AddDbCopyListener(backups_config.GetBackupConfigService())
}

func GetBackupService() *BackupService {
	return backupService
}

func GetBackupController() *BackupController {
	return backupController
}

func GetBackupBackgroundService() *BackupBackgroundService {
	return backupBackgroundService
}
