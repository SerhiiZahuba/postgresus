package databases

import (
	"errors"
	"log/slog"
	"postgresus-backend/internal/features/databases/databases/postgresql"
	"postgresus-backend/internal/features/notifiers"
	users_models "postgresus-backend/internal/features/users/models"
	"time"

	"github.com/google/uuid"
)

type DatabaseService struct {
	dbRepository    *DatabaseRepository
	notifierService *notifiers.NotifierService
	logger          *slog.Logger

	dbCreationListener []DatabaseCreationListener
	dbRemoveListener   []DatabaseRemoveListener
	dbCopyListener     []DatabaseCopyListener
}

func (s *DatabaseService) AddDbCreationListener(
	dbCreationListener DatabaseCreationListener,
) {
	s.dbCreationListener = append(s.dbCreationListener, dbCreationListener)
}

func (s *DatabaseService) AddDbRemoveListener(
	dbRemoveListener DatabaseRemoveListener,
) {
	s.dbRemoveListener = append(s.dbRemoveListener, dbRemoveListener)
}

func (s *DatabaseService) AddDbCopyListener(
	dbCopyListener DatabaseCopyListener,
) {
	s.dbCopyListener = append(s.dbCopyListener, dbCopyListener)
}

func (s *DatabaseService) CreateDatabase(
	user *users_models.User,
	database *Database,
) (*Database, error) {
	database.UserID = user.ID

	if err := database.Validate(); err != nil {
		return nil, err
	}

	database, err := s.dbRepository.Save(database)
	if err != nil {
		return nil, err
	}

	for _, listener := range s.dbCreationListener {
		listener.OnDatabaseCreated(database.ID)
	}

	return database, nil
}

func (s *DatabaseService) UpdateDatabase(
	user *users_models.User,
	database *Database,
) error {
	if database.ID == uuid.Nil {
		return errors.New("database ID is required for update")
	}

	existingDatabase, err := s.dbRepository.FindByID(database.ID)
	if err != nil {
		return err
	}

	if existingDatabase.UserID != user.ID {
		return errors.New("you have not access to this database")
	}

	// Validate the update
	if err := database.ValidateUpdate(*existingDatabase, *database); err != nil {
		return err
	}

	if err := database.Validate(); err != nil {
		return err
	}

	_, err = s.dbRepository.Save(database)
	if err != nil {
		return err
	}

	return nil
}

func (s *DatabaseService) DeleteDatabase(
	user *users_models.User,
	id uuid.UUID,
) error {
	existingDatabase, err := s.dbRepository.FindByID(id)
	if err != nil {
		return err
	}

	if existingDatabase.UserID != user.ID {
		return errors.New("you have not access to this database")
	}

	for _, listener := range s.dbRemoveListener {
		if err := listener.OnBeforeDatabaseRemove(id); err != nil {
			return err
		}
	}

	return s.dbRepository.Delete(id)
}

func (s *DatabaseService) GetDatabase(
	user *users_models.User,
	id uuid.UUID,
) (*Database, error) {
	database, err := s.dbRepository.FindByID(id)
	if err != nil {
		return nil, err
	}

	if database.UserID != user.ID {
		return nil, errors.New("you have not access to this database")
	}

	return database, nil
}

func (s *DatabaseService) GetDatabasesByUser(
	user *users_models.User,
) ([]*Database, error) {
	return s.dbRepository.FindByUserID(user.ID)
}

func (s *DatabaseService) IsNotifierUsing(
	user *users_models.User,
	notifierID uuid.UUID,
) (bool, error) {
	_, err := s.notifierService.GetNotifier(user, notifierID)
	if err != nil {
		return false, err
	}

	return s.dbRepository.IsNotifierUsing(notifierID)
}

func (s *DatabaseService) TestDatabaseConnection(
	user *users_models.User,
	databaseID uuid.UUID,
) error {
	database, err := s.dbRepository.FindByID(databaseID)
	if err != nil {
		return err
	}

	if database.UserID != user.ID {
		return errors.New("you have not access to this database")
	}

	err = database.TestConnection(s.logger)
	if err != nil {
		lastSaveError := err.Error()
		database.LastBackupErrorMessage = &lastSaveError
		return err
	}

	database.LastBackupErrorMessage = nil

	_, err = s.dbRepository.Save(database)
	if err != nil {
		return err
	}

	return nil
}

func (s *DatabaseService) TestDatabaseConnectionDirect(
	database *Database,
) error {
	return database.TestConnection(s.logger)
}

func (s *DatabaseService) GetDatabaseByID(
	id uuid.UUID,
) (*Database, error) {
	return s.dbRepository.FindByID(id)
}

func (s *DatabaseService) GetAllDatabases() ([]*Database, error) {
	return s.dbRepository.GetAllDatabases()
}

func (s *DatabaseService) SetBackupError(databaseID uuid.UUID, errorMessage string) error {
	database, err := s.dbRepository.FindByID(databaseID)
	if err != nil {
		return err
	}

	database.LastBackupErrorMessage = &errorMessage
	_, err = s.dbRepository.Save(database)
	if err != nil {
		return err
	}

	return nil
}

func (s *DatabaseService) SetLastBackupTime(databaseID uuid.UUID, backupTime time.Time) error {
	database, err := s.dbRepository.FindByID(databaseID)
	if err != nil {
		return err
	}

	database.LastBackupTime = &backupTime
	database.LastBackupErrorMessage = nil // Clear any previous error
	_, err = s.dbRepository.Save(database)
	if err != nil {
		return err
	}

	return nil
}

func (s *DatabaseService) CopyDatabase(
	user *users_models.User,
	databaseID uuid.UUID,
) (*Database, error) {
	existingDatabase, err := s.dbRepository.FindByID(databaseID)
	if err != nil {
		return nil, err
	}

	if existingDatabase.UserID != user.ID {
		return nil, errors.New("you have not access to this database")
	}

	newDatabase := &Database{
		ID:                     uuid.Nil,
		UserID:                 user.ID,
		Name:                   existingDatabase.Name + " (Copy)",
		Type:                   existingDatabase.Type,
		Notifiers:              existingDatabase.Notifiers,
		LastBackupTime:         nil,
		LastBackupErrorMessage: nil,
		HealthStatus:           existingDatabase.HealthStatus,
	}

	switch existingDatabase.Type {
	case DatabaseTypePostgres:
		if existingDatabase.Postgresql != nil {
			newDatabase.Postgresql = &postgresql.PostgresqlDatabase{
				ID:         uuid.Nil,
				DatabaseID: nil,
				Version:    existingDatabase.Postgresql.Version,
				Host:       existingDatabase.Postgresql.Host,
				Port:       existingDatabase.Postgresql.Port,
				Username:   existingDatabase.Postgresql.Username,
				Password:   existingDatabase.Postgresql.Password,
				Database:   existingDatabase.Postgresql.Database,
				IsHttps:    existingDatabase.Postgresql.IsHttps,
			}
		}
	}

	if err := newDatabase.Validate(); err != nil {
		return nil, err
	}

	copiedDatabase, err := s.dbRepository.Save(newDatabase)
	if err != nil {
		return nil, err
	}

	for _, listener := range s.dbCreationListener {
		listener.OnDatabaseCreated(copiedDatabase.ID)
	}

	for _, listener := range s.dbCopyListener {
		listener.OnDatabaseCopied(databaseID, copiedDatabase.ID)
	}

	return copiedDatabase, nil
}

func (s *DatabaseService) SetHealthStatus(
	databaseID uuid.UUID,
	healthStatus *HealthStatus,
) error {
	database, err := s.dbRepository.FindByID(databaseID)
	if err != nil {
		return err
	}

	database.HealthStatus = healthStatus
	_, err = s.dbRepository.Save(database)
	if err != nil {
		return err
	}

	return nil
}
