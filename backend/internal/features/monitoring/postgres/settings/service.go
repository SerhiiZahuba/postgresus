package postgres_monitoring_settings

import (
	"errors"
	"postgresus-backend/internal/features/databases"
	users_models "postgresus-backend/internal/features/users/models"
	"postgresus-backend/internal/util/logger"

	"github.com/google/uuid"
)

var log = logger.GetLogger()

type PostgresMonitoringSettingsService struct {
	databaseService                      *databases.DatabaseService
	postgresMonitoringSettingsRepository *PostgresMonitoringSettingsRepository
}

func (s *PostgresMonitoringSettingsService) OnDatabaseCreated(dbID uuid.UUID) {
	db, err := s.databaseService.GetDatabaseByID(dbID)
	if err != nil {
		return
	}

	if db.Type != databases.DatabaseTypePostgres {
		return
	}

	settings := &PostgresMonitoringSettings{
		DatabaseID:                     dbID,
		IsDbResourcesMonitoringEnabled: true,
		MonitoringIntervalSeconds:      60,
	}

	err = s.postgresMonitoringSettingsRepository.Save(settings)
	if err != nil {
		log.Error("failed to save postgres monitoring settings", "error", err)
	}
}

func (s *PostgresMonitoringSettingsService) Save(
	user *users_models.User,
	settings *PostgresMonitoringSettings,
) error {
	db, err := s.databaseService.GetDatabaseByID(settings.DatabaseID)
	if err != nil {
		return err
	}

	if db.UserID != user.ID {
		return errors.New("user does not have access to this database")
	}

	return s.postgresMonitoringSettingsRepository.Save(settings)
}

func (s *PostgresMonitoringSettingsService) GetByDbID(
	user *users_models.User,
	dbID uuid.UUID,
) (*PostgresMonitoringSettings, error) {
	dbSettings, err := s.postgresMonitoringSettingsRepository.GetByDbIDWithRelations(dbID)
	if err != nil {
		return nil, err
	}

	if dbSettings == nil {
		s.OnDatabaseCreated(dbID)

		dbSettings, err := s.postgresMonitoringSettingsRepository.GetByDbIDWithRelations(dbID)
		if err != nil {
			return nil, err
		}

		if dbSettings == nil {
			return nil, errors.New("postgres monitoring settings not found")
		}

		return s.GetByDbID(user, dbID)
	}

	if dbSettings.Database.UserID != user.ID {
		return nil, errors.New("user does not have access to this database")
	}

	return dbSettings, nil
}

func (s *PostgresMonitoringSettingsService) GetAllDbsWithEnabledDbMonitoring() (
	[]PostgresMonitoringSettings, error,
) {
	return s.postgresMonitoringSettingsRepository.GetAllDbsWithEnabledDbMonitoring()
}
