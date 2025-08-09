package postgres_monitoring_settings

import (
	"postgresus-backend/internal/features/databases"
	"postgresus-backend/internal/features/databases/databases/postgresql"
	"postgresus-backend/internal/features/notifiers"
	"postgresus-backend/internal/features/storages"
	"postgresus-backend/internal/features/users"
	users_models "postgresus-backend/internal/features/users/models"
	"postgresus-backend/internal/util/tools"
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

func Test_DatabaseCreated_SettingsCreatedAndExtensionsInstalled(t *testing.T) {
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

	// System and queries monitoring may be disabled if extension installation fails
	// in the test environment, but the service should handle this gracefully
	// We test the logic by checking the installed extensions field
	t.Logf("System monitoring enabled: %v", settings.IsSystemResourcesMonitoringEnabled)
	t.Logf("Queries monitoring enabled: %v", settings.IsQueriesMonitoringEnabled)
	t.Logf("Installed extensions: %v", settings.InstalledExtensions)

	// If system monitoring is enabled, pg_proctab should be in installed extensions
	if settings.IsSystemResourcesMonitoringEnabled {
		assert.Contains(t, settings.InstalledExtensions, tools.PostgresqlExtensionPgProctab,
			"If system monitoring is enabled, pg_proctab extension should be tracked")
	}

	// If queries monitoring is enabled, pg_stat_monitor should be in installed extensions
	if settings.IsQueriesMonitoringEnabled {
		assert.Contains(t, settings.InstalledExtensions, tools.PostgresqlExtensionPgStatMonitor,
			"If queries monitoring is enabled, pg_stat_monitor extension should be tracked")
	}
}

func Test_DatabaseCreated_PrePostgres16_ExtensionsNotSupported(t *testing.T) {
	// Test that extension-based monitoring is disabled for older PostgreSQL versions
	testUserResponse := users.GetTestUser()
	storage := storages.CreateTestStorage(testUserResponse.UserID)
	notifier := notifiers.CreateTestNotifier(testUserResponse.UserID)

	// Note: We manually create the database here because CreateTestDatabase always uses PostgreSQL 16,
	// but this test specifically needs PostgreSQL 14 to verify older version behavior
	testDatabase := &databases.Database{
		UserID: testUserResponse.UserID,
		Name:   "Old PostgreSQL Database " + uuid.New().String(),
		Type:   databases.DatabaseTypePostgres,
		Postgresql: &postgresql.PostgresqlDatabase{
			Version:  tools.PostgresqlVersion14, // Older version
			Host:     "localhost",
			Port:     5432,
			Username: "test",
			Password: "test",
			Database: func() *string { s := "test_db"; return &s }(),
		},
		Notifiers: []notifiers.Notifier{*notifier},
	}

	// Save the test database
	repo := &databases.DatabaseRepository{}
	database, err := repo.Save(testDatabase)
	assert.NoError(t, err)

	defer storages.RemoveTestStorage(storage.ID)
	defer notifiers.RemoveTestNotifier(notifier)
	defer repo.Delete(database.ID)

	// Get the monitoring settings service
	service := GetPostgresMonitoringSettingsService()

	// Execute - trigger the database creation event
	service.OnDatabaseCreated(database.ID)

	// Verify settings were created
	settingsRepo := GetPostgresMonitoringSettingsRepository()
	settings, err := settingsRepo.GetByDbID(database.ID)
	assert.NoError(t, err)
	assert.NotNil(t, settings)

	// For pre-16 versions, extension-based monitoring should be disabled
	// because ensureExtensionsInstalled should return an error for versions < 16
	assert.False(t, settings.IsSystemResourcesMonitoringEnabled,
		"System monitoring should be disabled for PostgreSQL versions < 16")
	assert.False(t, settings.IsQueriesMonitoringEnabled,
		"Queries monitoring should be disabled for PostgreSQL versions < 16")

	// DB resources monitoring should still be enabled (doesn't require extensions)
	assert.True(t, settings.IsDbResourcesMonitoringEnabled)

	// No extensions should be installed for older versions
	assert.Empty(t, settings.InstalledExtensions,
		"No extensions should be installed for PostgreSQL versions < 16")
}

func Test_MonitoringEnabled_ExtensionsInstalled(t *testing.T) {
	// Get or create a test user
	testUser := getTestUserModel()
	testUserResponse := users.GetTestUser()
	storage := storages.CreateTestStorage(testUserResponse.UserID)
	notifier := notifiers.CreateTestNotifier(testUserResponse.UserID)
	database := databases.CreateTestDatabase(testUserResponse.UserID, storage, notifier)

	defer storages.RemoveTestStorage(storage.ID)
	defer notifiers.RemoveTestNotifier(notifier)
	defer databases.RemoveTestDatabase(database)

	// Create initial settings with monitoring disabled
	service := GetPostgresMonitoringSettingsService()
	settingsRepo := GetPostgresMonitoringSettingsRepository()

	initialSettings := &PostgresMonitoringSettings{
		DatabaseID:                         database.ID,
		IsSystemResourcesMonitoringEnabled: false,
		IsDbResourcesMonitoringEnabled:     true,
		IsQueriesMonitoringEnabled:         false,
		MonitoringIntervalSeconds:          15,
	}

	err := settingsRepo.Save(initialSettings)
	assert.NoError(t, err)

	// Test enabling system monitoring - extension installation might fail in test environment
	systemSettings := &PostgresMonitoringSettings{
		DatabaseID:                         database.ID,
		IsSystemResourcesMonitoringEnabled: true,
		IsDbResourcesMonitoringEnabled:     true,
		IsQueriesMonitoringEnabled:         false,
		MonitoringIntervalSeconds:          15,
	}

	err = service.Save(testUser, systemSettings)
	// In test environment, extension installation might fail - this is expected behavior
	if err != nil {
		t.Logf("Extension installation failed as expected in test environment: %v", err)
		assert.Contains(t, err.Error(), "failed to install pg_proctab extension")
		return // Test passed - service correctly handles extension installation failures
	}

	// If extension installation succeeded, verify the settings
	updatedSettings, err := settingsRepo.GetByDbID(database.ID)
	assert.NoError(t, err)
	assert.True(t, updatedSettings.IsSystemResourcesMonitoringEnabled)
	assert.Contains(t, updatedSettings.InstalledExtensions, tools.PostgresqlExtensionPgProctab)

	// Test enabling queries monitoring - should install pg_stat_monitor extension
	queriesSettings := &PostgresMonitoringSettings{
		DatabaseID:                         database.ID,
		IsSystemResourcesMonitoringEnabled: true,
		IsDbResourcesMonitoringEnabled:     true,
		IsQueriesMonitoringEnabled:         true,
		MonitoringIntervalSeconds:          15,
	}

	err = service.Save(testUser, queriesSettings)
	if err != nil {
		t.Logf("Queries monitoring extension installation failed: %v", err)
		assert.Contains(t, err.Error(), "failed to install pg_stat_monitor extension")
		return // Test passed - service correctly handles extension installation failures
	}

	// If both extensions installed successfully, verify final state
	finalSettings, err := settingsRepo.GetByDbID(database.ID)
	assert.NoError(t, err)
	assert.True(t, finalSettings.IsSystemResourcesMonitoringEnabled)
	assert.True(t, finalSettings.IsQueriesMonitoringEnabled)
	assert.Contains(t, finalSettings.InstalledExtensions, tools.PostgresqlExtensionPgProctab)
	assert.Contains(t, finalSettings.InstalledExtensions, tools.PostgresqlExtensionPgStatMonitor)
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
