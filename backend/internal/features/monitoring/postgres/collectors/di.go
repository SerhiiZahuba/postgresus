package postgres_monitoring_collectors

import (
	"postgresus-backend/internal/features/databases"
	postgres_monitoring_metrics "postgresus-backend/internal/features/monitoring/postgres/metrics"
	postgres_monitoring_settings "postgresus-backend/internal/features/monitoring/postgres/settings"
	"postgresus-backend/internal/util/logger"
	"sync"
)

var dbMonitoringBackgroundService = &DbMonitoringBackgroundService{
	databases.GetDatabaseService(),
	postgres_monitoring_settings.GetPostgresMonitoringSettingsService(),
	postgres_monitoring_metrics.GetPostgresMonitoringMetricsService(),
	logger.GetLogger(),
	0,
	nil,
	sync.RWMutex{},
}

func GetDbMonitoringBackgroundService() *DbMonitoringBackgroundService {
	return dbMonitoringBackgroundService
}
