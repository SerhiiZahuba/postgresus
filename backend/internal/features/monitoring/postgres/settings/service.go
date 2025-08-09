package postgres_monitoring_settings

import (
	"errors"
	"postgresus-backend/internal/features/databases"
	users_models "postgresus-backend/internal/features/users/models"
	"postgresus-backend/internal/util/logger"
	"postgresus-backend/internal/util/tools"

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
		DatabaseID:                         dbID,
		IsSystemResourcesMonitoringEnabled: true,
		IsDbResourcesMonitoringEnabled:     true,
		IsQueriesMonitoringEnabled:         true,
		MonitoringIntervalSeconds:          15,
	}

	err = s.ensureExtensionsInstalled(
		dbID,
		[]tools.PostgresqlExtension{tools.PostgresqlExtensionPgProctab},
	)
	if err != nil {
		settings.IsSystemResourcesMonitoringEnabled = false
	} else {
		settings.AddInstalledExtensions([]tools.PostgresqlExtension{tools.PostgresqlExtensionPgProctab})
	}

	err = s.ensureExtensionsInstalled(
		dbID,
		[]tools.PostgresqlExtension{tools.PostgresqlExtensionPgStatMonitor},
	)
	if err != nil {
		settings.IsQueriesMonitoringEnabled = false
	} else {
		settings.AddInstalledExtensions([]tools.PostgresqlExtension{tools.PostgresqlExtensionPgStatMonitor})
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

	existingSettings, err := s.postgresMonitoringSettingsRepository.GetByDbID(settings.DatabaseID)
	if err != nil {
		return err
	}

	if existingSettings != nil &&
		settings.IsSystemResourcesMonitoringEnabled &&
		!existingSettings.IsSystemResourcesMonitoringEnabled {
		err := s.ensureExtensionsInstalled(
			settings.DatabaseID,
			[]tools.PostgresqlExtension{tools.PostgresqlExtensionPgProctab},
		)
		if err != nil {
			return errors.New(
				"failed to install pg_proctab extension, system resources is not possible (please, disable it)",
			)
		}

		settings.AddInstalledExtensions(
			[]tools.PostgresqlExtension{tools.PostgresqlExtensionPgProctab},
		)
	}

	if existingSettings != nil &&
		settings.IsQueriesMonitoringEnabled &&
		!existingSettings.IsQueriesMonitoringEnabled {
		err := s.ensureExtensionsInstalled(
			settings.DatabaseID,
			[]tools.PostgresqlExtension{tools.PostgresqlExtensionPgStatMonitor},
		)
		if err != nil {
			return errors.New(
				"failed to install pg_stat_monitor extension, queries monitoring is not possible (please, disable it)",
			)
		}

		settings.AddInstalledExtensions(
			[]tools.PostgresqlExtension{tools.PostgresqlExtensionPgStatMonitor},
		)
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

func (s *PostgresMonitoringSettingsService) ensureExtensionsInstalled(
	dbID uuid.UUID,
	extensions []tools.PostgresqlExtension,
) error {
	database, err := s.databaseService.GetDatabaseByID(dbID)
	if err != nil {
		return err
	}

	if database.Type != databases.DatabaseTypePostgres {
		return errors.New("database is not a postgres database")
	}

	if database.Postgresql == nil {
		return errors.New("database is not a postgres database")
	}

	if database.Postgresql.Version < tools.PostgresqlVersion16 {
		return errors.New("system monitoring extensions supported for postgres 16+")
	}

	err = database.Postgresql.InstallExtensions(extensions)
	if err != nil {
		return err
	}

	return nil
}
