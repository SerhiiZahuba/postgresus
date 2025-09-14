package postgres_monitoring_settings

import (
	"postgresus-backend/internal/features/databases"
	"postgresus-backend/internal/features/notifiers"
	"postgresus-backend/internal/features/storages"
	"postgresus-backend/internal/features/users"
	users_models "postgresus-backend/internal/features/users/models"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

// Helper function to get a proper users_models.User for testing
func getTestUserModel() *users_models.User {
	signInResponse := users.GetTestUser()

	// Get the user service to retrieve the full user model
	userService := users.GetUserService()
	user, err := userService.GetFirstUser()
	if err != nil {
		panic(err)
	}

	// Verify we got the right user
	if user.ID != signInResponse.UserID {
		panic("user ID mismatch")
	}

	return user
}

func Test_DatabaseCreated_SettingsCreated(t *testing.T) {
	// Get or create a test user
	testUserResponse := users.GetTestUser()
	storage := storages.CreateTestStorage(testUserResponse.UserID)
	notifier := notifiers.CreateTestNotifier(testUserResponse.UserID)
	database := databases.CreateTestDatabase(testUserResponse.UserID, storage, notifier)

	defer storages.RemoveTestStorage(storage.ID)
	defer notifiers.RemoveTestNotifier(notifier)
	defer databases.RemoveTestDatabase(database)

	// Get the monitoring settings service
	service := GetPostgresMonitoringSettingsService()

	// Execute - trigger the database creation event
	service.OnDatabaseCreated(database.ID)

	// Verify settings were created by attempting to retrieve them
	// Note: Since we can't easily mock the extension installation without major changes,
	// we focus on testing the settings creation and default values logic
	settingsRepo := GetPostgresMonitoringSettingsRepository()
	settings, err := settingsRepo.GetByDbID(database.ID)
	assert.NoError(t, err)
	assert.NotNil(t, settings)

	// Verify default settings values
	assert.Equal(t, database.ID, settings.DatabaseID)
	assert.Equal(t, int64(15), settings.MonitoringIntervalSeconds)
	assert.True(t, settings.IsDbResourcesMonitoringEnabled) // Always enabled
}

func Test_GetSettingsByDbID_SettingsReturned(t *testing.T) {
	// Get or create a test user
	testUser := getTestUserModel()
	testUserResponse := users.GetTestUser()
	storage := storages.CreateTestStorage(testUserResponse.UserID)
	notifier := notifiers.CreateTestNotifier(testUserResponse.UserID)
	database := databases.CreateTestDatabase(testUserResponse.UserID, storage, notifier)

	defer storages.RemoveTestStorage(storage.ID)
	defer notifiers.RemoveTestNotifier(notifier)
	defer databases.RemoveTestDatabase(database)

	service := GetPostgresMonitoringSettingsService()

	// Test 1: Get settings that don't exist yet - should auto-create them
	settings, err := service.GetByDbID(testUser, database.ID)
	assert.NoError(t, err)
	assert.NotNil(t, settings)
	assert.Equal(t, database.ID, settings.DatabaseID)
	assert.Equal(t, int64(15), settings.MonitoringIntervalSeconds)
	assert.True(t, settings.IsDbResourcesMonitoringEnabled) // Always enabled

	// Test 2: Get settings that already exist
	existingSettings, err := service.GetByDbID(testUser, database.ID)
	assert.NoError(t, err)
	assert.NotNil(t, existingSettings)
	assert.Equal(t, settings.DatabaseID, existingSettings.DatabaseID)
	assert.Equal(t, settings.MonitoringIntervalSeconds, existingSettings.MonitoringIntervalSeconds)

	// Test 3: Access control - create another user and test they can't access this database
	anotherUser := &users_models.User{
		ID: uuid.New(),
		// Other fields can be empty for this test
	}

	_, err = service.GetByDbID(anotherUser, database.ID)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "user does not have access to this database")

	// Test 4: Try to get settings for non-existent database
	nonExistentDbID := uuid.New()
	_, err = service.GetByDbID(testUser, nonExistentDbID)
	assert.Error(t, err) // Should fail because database doesn't exist
}
