package postgres_monitoring_metrics

import (
	"postgresus-backend/internal/features/databases"
	"postgresus-backend/internal/features/notifiers"
	"postgresus-backend/internal/features/storages"
	"postgresus-backend/internal/features/users"
	users_models "postgresus-backend/internal/features/users/models"
	"testing"
	"time"

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

func Test_GetMetrics_MetricsReturned(t *testing.T) {
	// Setup test data
	testUser := getTestUserModel()
	testUserResponse := users.GetTestUser()
	storage := storages.CreateTestStorage(testUserResponse.UserID)
	notifier := notifiers.CreateTestNotifier(testUserResponse.UserID)
	database := databases.CreateTestDatabase(testUserResponse.UserID, storage, notifier)

	defer storages.RemoveTestStorage(storage.ID)
	defer notifiers.RemoveTestNotifier(notifier)
	defer databases.RemoveTestDatabase(database)

	// Get service and repository
	service := GetPostgresMonitoringMetricsService()
	repository := GetPostgresMonitoringMetricsRepository()

	// Create test metrics
	now := time.Now().UTC()
	testMetrics := []PostgresMonitoringMetric{
		{
			DatabaseID: database.ID,
			Metric:     MetricsTypeDbRAM,
			ValueType:  MetricsValueTypeByte,
			Value:      1024000,
			CreatedAt:  now.Add(-2 * time.Hour),
		},
		{
			DatabaseID: database.ID,
			Metric:     MetricsTypeDbRAM,
			ValueType:  MetricsValueTypeByte,
			Value:      2048000,
			CreatedAt:  now.Add(-1 * time.Hour),
		},
		{
			DatabaseID: database.ID,
			Metric:     MetricsTypeSystemCPU,
			ValueType:  MetricsValueTypePercent,
			Value:      75.5,
			CreatedAt:  now.Add(-30 * time.Minute),
		},
	}

	// Insert test metrics
	err := repository.Insert(testMetrics)
	assert.NoError(t, err)

	// Test getting DB RAM metrics
	from := now.Add(-3 * time.Hour)
	to := now

	metrics, err := service.GetMetrics(testUser, database.ID, MetricsTypeDbRAM, from, to)
	assert.NoError(t, err)
	assert.Len(t, metrics, 2)

	// Verify metrics are ordered by created_at DESC
	assert.True(t, metrics[0].CreatedAt.After(metrics[1].CreatedAt))
	assert.Equal(t, float64(2048000), metrics[0].Value)
	assert.Equal(t, float64(1024000), metrics[1].Value)
	assert.Equal(t, MetricsTypeDbRAM, metrics[0].Metric)
	assert.Equal(t, MetricsValueTypeByte, metrics[0].ValueType)

	// Test getting CPU metrics
	cpuMetrics, err := service.GetMetrics(testUser, database.ID, MetricsTypeSystemCPU, from, to)
	assert.NoError(t, err)
	assert.Len(t, cpuMetrics, 1)
	assert.Equal(t, float64(75.5), cpuMetrics[0].Value)
	assert.Equal(t, MetricsTypeSystemCPU, cpuMetrics[0].Metric)
	assert.Equal(t, MetricsValueTypePercent, cpuMetrics[0].ValueType)

	// Test access control - create another user and test they can't access this database
	anotherUser := &users_models.User{
		ID: uuid.New(),
	}

	_, err = service.GetMetrics(anotherUser, database.ID, MetricsTypeDbRAM, from, to)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "database not found")

	// Test with non-existent database
	nonExistentDbID := uuid.New()
	_, err = service.GetMetrics(testUser, nonExistentDbID, MetricsTypeDbRAM, from, to)
	assert.Error(t, err)
}

func Test_GetMetricsWithPagination_PaginationWorks(t *testing.T) {
	// Setup test data
	testUser := getTestUserModel()
	testUserResponse := users.GetTestUser()
	storage := storages.CreateTestStorage(testUserResponse.UserID)
	notifier := notifiers.CreateTestNotifier(testUserResponse.UserID)
	database := databases.CreateTestDatabase(testUserResponse.UserID, storage, notifier)

	defer storages.RemoveTestStorage(storage.ID)
	defer notifiers.RemoveTestNotifier(notifier)
	defer databases.RemoveTestDatabase(database)

	// Get repository and service
	repository := GetPostgresMonitoringMetricsRepository()
	service := GetPostgresMonitoringMetricsService()

	// Create many test metrics for pagination testing
	now := time.Now().UTC()
	testMetrics := []PostgresMonitoringMetric{}

	for i := 0; i < 25; i++ {
		testMetrics = append(testMetrics, PostgresMonitoringMetric{
			DatabaseID: database.ID,
			Metric:     MetricsTypeDbRAM,
			ValueType:  MetricsValueTypeByte,
			Value:      float64(1000000 + i*100000),
			CreatedAt:  now.Add(-time.Duration(i) * time.Minute),
		})
	}

	// Insert test metrics
	err := repository.Insert(testMetrics)
	assert.NoError(t, err)

	// Test getting all metrics via service (should return all 25)
	from := now.Add(-30 * time.Minute)
	to := now

	allMetrics, err := service.GetMetrics(testUser, database.ID, MetricsTypeDbRAM, from, to)
	assert.NoError(t, err)
	assert.Len(t, allMetrics, 25)

	// Verify they are ordered by created_at DESC (most recent first)
	for i := 0; i < len(allMetrics)-1; i++ {
		assert.True(t, allMetrics[i].CreatedAt.After(allMetrics[i+1].CreatedAt) ||
			allMetrics[i].CreatedAt.Equal(allMetrics[i+1].CreatedAt))
	}

	// Note: Since the current repository doesn't have pagination methods,
	// this test demonstrates the need for pagination but tests current behavior.
	// TODO: Add GetByMetricsWithLimit method to repository and update service
	t.Logf("All metrics count: %d (pagination methods should be added)", len(allMetrics))
}

func Test_GetMetricsWithFilterByType_FilterWorks(t *testing.T) {
	// Setup test data
	testUser := getTestUserModel()
	testUserResponse := users.GetTestUser()
	storage := storages.CreateTestStorage(testUserResponse.UserID)
	notifier := notifiers.CreateTestNotifier(testUserResponse.UserID)
	database := databases.CreateTestDatabase(testUserResponse.UserID, storage, notifier)

	defer storages.RemoveTestStorage(storage.ID)
	defer notifiers.RemoveTestNotifier(notifier)
	defer databases.RemoveTestDatabase(database)

	// Get service and repository
	service := GetPostgresMonitoringMetricsService()
	repository := GetPostgresMonitoringMetricsRepository()

	// Create test metrics of different types
	now := time.Now().UTC()
	testMetrics := []PostgresMonitoringMetric{
		// DB RAM metrics
		{
			DatabaseID: database.ID,
			Metric:     MetricsTypeDbRAM,
			ValueType:  MetricsValueTypeByte,
			Value:      1024000,
			CreatedAt:  now.Add(-2 * time.Hour),
		},
		{
			DatabaseID: database.ID,
			Metric:     MetricsTypeDbRAM,
			ValueType:  MetricsValueTypeByte,
			Value:      2048000,
			CreatedAt:  now.Add(-1 * time.Hour),
		},
		// DB ROM metrics
		{
			DatabaseID: database.ID,
			Metric:     MetricsTypeDbROM,
			ValueType:  MetricsValueTypeByte,
			Value:      5000000,
			CreatedAt:  now.Add(-90 * time.Minute),
		},
		{
			DatabaseID: database.ID,
			Metric:     MetricsTypeDbROM,
			ValueType:  MetricsValueTypeByte,
			Value:      5500000,
			CreatedAt:  now.Add(-30 * time.Minute),
		},
		// System CPU metrics
		{
			DatabaseID: database.ID,
			Metric:     MetricsTypeSystemCPU,
			ValueType:  MetricsValueTypePercent,
			Value:      75.5,
			CreatedAt:  now.Add(-45 * time.Minute),
		},
		// System RAM metrics
		{
			DatabaseID: database.ID,
			Metric:     MetricsTypeSystemRAM,
			ValueType:  MetricsValueTypePercent,
			Value:      65.2,
			CreatedAt:  now.Add(-25 * time.Minute),
		},
	}

	// Insert test metrics
	err := repository.Insert(testMetrics)
	assert.NoError(t, err)

	from := now.Add(-3 * time.Hour)
	to := now

	// Test filtering by DB RAM type
	ramMetrics, err := service.GetMetrics(testUser, database.ID, MetricsTypeDbRAM, from, to)
	assert.NoError(t, err)
	assert.Len(t, ramMetrics, 2)
	for _, metric := range ramMetrics {
		assert.Equal(t, MetricsTypeDbRAM, metric.Metric)
		assert.Equal(t, MetricsValueTypeByte, metric.ValueType)
	}

	// Test filtering by DB ROM type
	romMetrics, err := service.GetMetrics(testUser, database.ID, MetricsTypeDbROM, from, to)
	assert.NoError(t, err)
	assert.Len(t, romMetrics, 2)
	for _, metric := range romMetrics {
		assert.Equal(t, MetricsTypeDbROM, metric.Metric)
		assert.Equal(t, MetricsValueTypeByte, metric.ValueType)
	}

	// Test filtering by System CPU type
	cpuMetrics, err := service.GetMetrics(testUser, database.ID, MetricsTypeSystemCPU, from, to)
	assert.NoError(t, err)
	assert.Len(t, cpuMetrics, 1)
	for _, metric := range cpuMetrics {
		assert.Equal(t, MetricsTypeSystemCPU, metric.Metric)
		assert.Equal(t, MetricsValueTypePercent, metric.ValueType)
	}

	// Test filtering by System RAM type
	systemRamMetrics, err := service.GetMetrics(testUser, database.ID, MetricsTypeSystemRAM, from, to)
	assert.NoError(t, err)
	assert.Len(t, systemRamMetrics, 1)
	for _, metric := range systemRamMetrics {
		assert.Equal(t, MetricsTypeSystemRAM, metric.Metric)
		assert.Equal(t, MetricsValueTypePercent, metric.ValueType)
	}

	// Test filtering by non-existent metric type (should return empty)
	ioMetrics, err := service.GetMetrics(testUser, database.ID, MetricsTypeDbIO, from, to)
	assert.NoError(t, err)
	assert.Len(t, ioMetrics, 0)

	// Test time filtering - get only recent metrics (last hour)
	recentFrom := now.Add(-1 * time.Hour)
	recentRamMetrics, err := service.GetMetrics(testUser, database.ID, MetricsTypeDbRAM, recentFrom, to)
	assert.NoError(t, err)
	assert.Len(t, recentRamMetrics, 1) // Only the metric from 1 hour ago
	assert.Equal(t, float64(2048000), recentRamMetrics[0].Value)
}
