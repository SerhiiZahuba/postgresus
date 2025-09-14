package postgres_monitoring_collectors

import (
	"context"
	"fmt"
	"log/slog"
	"postgresus-backend/internal/config"
	"postgresus-backend/internal/features/databases"
	"postgresus-backend/internal/features/databases/databases/postgresql"
	postgres_monitoring_metrics "postgresus-backend/internal/features/monitoring/postgres/metrics"
	postgres_monitoring_settings "postgresus-backend/internal/features/monitoring/postgres/settings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

type DbMonitoringBackgroundService struct {
	databaseService           *databases.DatabaseService
	monitoringSettingsService *postgres_monitoring_settings.PostgresMonitoringSettingsService
	metricsService            *postgres_monitoring_metrics.PostgresMonitoringMetricService
	logger                    *slog.Logger
	isRunning                 int32
	lastRunTimes              map[uuid.UUID]time.Time
	lastRunTimesMutex         sync.RWMutex
}

func (s *DbMonitoringBackgroundService) Run() {
	for {
		if config.IsShouldShutdown() {
			s.logger.Info("stopping background monitoring tasks")
			return
		}

		s.processMonitoringTasks()
		time.Sleep(1 * time.Second)
	}
}

func (s *DbMonitoringBackgroundService) processMonitoringTasks() {
	if !atomic.CompareAndSwapInt32(&s.isRunning, 0, 1) {
		s.logger.Warn("skipping background task execution, previous task still running")
		return
	}
	defer atomic.StoreInt32(&s.isRunning, 0)

	dbsWithEnabledDbMonitoring, err := s.monitoringSettingsService.GetAllDbsWithEnabledDbMonitoring()
	if err != nil {
		s.logger.Error("failed to get all databases with enabled db monitoring", "error", err)
		return
	}

	for _, dbSettings := range dbsWithEnabledDbMonitoring {
		s.processDatabase(&dbSettings)
	}
}

func (s *DbMonitoringBackgroundService) processDatabase(
	settings *postgres_monitoring_settings.PostgresMonitoringSettings,
) {
	db, err := s.databaseService.GetDatabaseByID(settings.DatabaseID)
	if err != nil {
		s.logger.Error("failed to get database by id", "error", err)
		return
	}

	if db.Type != databases.DatabaseTypePostgres {
		return
	}

	if !s.isReadyForNextRun(settings) {
		return
	}

	err = s.collectAndSaveMetrics(db, settings)
	if err != nil {
		s.logger.Error("failed to collect and save metrics", "error", err)
		return
	}

	s.updateLastRunTime(db)
}

func (s *DbMonitoringBackgroundService) collectAndSaveMetrics(
	db *databases.Database,
	settings *postgres_monitoring_settings.PostgresMonitoringSettings,
) error {
	if db.Postgresql == nil {
		return nil
	}

	s.logger.Debug("collecting metrics for database", "database_id", db.ID)

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	conn, err := s.connectToDatabase(ctx, db)
	if err != nil {
		return fmt.Errorf("failed to connect to database: %w", err)
	}

	if conn == nil {
		return nil
	}

	defer func() {
		if closeErr := conn.Close(ctx); closeErr != nil {
			s.logger.Error("Failed to close connection", "error", closeErr)
		}
	}()

	var metrics []postgres_monitoring_metrics.PostgresMonitoringMetric
	now := time.Now().UTC()

	if settings.IsDbResourcesMonitoringEnabled {
		dbMetrics, err := s.collectDatabaseResourceMetrics(ctx, conn, db.ID, now)
		if err != nil {
			s.logger.Error("failed to collect database resource metrics", "error", err)
		} else {
			metrics = append(metrics, dbMetrics...)
		}
	}

	if len(metrics) > 0 {
		if err := s.metricsService.Insert(metrics); err != nil {
			return fmt.Errorf("failed to insert metrics: %w", err)
		}
		s.logger.Debug(
			"successfully collected and saved metrics",
			"count",
			len(metrics),
			"database_id",
			db.ID,
		)
	}

	return nil
}

func (s *DbMonitoringBackgroundService) isReadyForNextRun(
	settings *postgres_monitoring_settings.PostgresMonitoringSettings,
) bool {
	s.lastRunTimesMutex.RLock()
	defer s.lastRunTimesMutex.RUnlock()

	if s.lastRunTimes == nil {
		return true
	}

	lastRun, exists := s.lastRunTimes[settings.DatabaseID]
	if !exists {
		return true
	}

	return time.Since(lastRun) >= time.Duration(settings.MonitoringIntervalSeconds)*time.Second
}

func (s *DbMonitoringBackgroundService) updateLastRunTime(db *databases.Database) {
	s.lastRunTimesMutex.Lock()
	defer s.lastRunTimesMutex.Unlock()

	if s.lastRunTimes == nil {
		s.lastRunTimes = make(map[uuid.UUID]time.Time)
	}
	s.lastRunTimes[db.ID] = time.Now().UTC()
}

func (s *DbMonitoringBackgroundService) connectToDatabase(
	ctx context.Context,
	db *databases.Database,
) (*pgx.Conn, error) {
	if db.Postgresql == nil {
		return nil, nil
	}

	if db.Postgresql.Database == nil || *db.Postgresql.Database == "" {
		return nil, nil
	}

	connStr := s.buildConnectionString(db.Postgresql)
	return pgx.Connect(ctx, connStr)
}

func (s *DbMonitoringBackgroundService) buildConnectionString(
	pg *postgresql.PostgresqlDatabase,
) string {
	sslMode := "disable"
	if pg.IsHttps {
		sslMode = "require"
	}

	return fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		pg.Host,
		pg.Port,
		pg.Username,
		pg.Password,
		*pg.Database,
		sslMode,
	)
}

func (s *DbMonitoringBackgroundService) collectDatabaseResourceMetrics(
	ctx context.Context,
	conn *pgx.Conn,
	databaseID uuid.UUID,
	timestamp time.Time,
) ([]postgres_monitoring_metrics.PostgresMonitoringMetric, error) {
	var metrics []postgres_monitoring_metrics.PostgresMonitoringMetric

	// Collect I/O statistics
	ioMetrics, err := s.collectIOMetrics(ctx, conn, databaseID, timestamp)
	if err != nil {
		s.logger.Warn("failed to collect I/O metrics", "error", err)
	} else {
		metrics = append(metrics, ioMetrics...)
	}

	// Collect memory usage (approximation based on buffer usage)
	ramMetric, err := s.collectRAMUsageMetric(ctx, conn, databaseID, timestamp)
	if err != nil {
		s.logger.Warn("failed to collect RAM usage metric", "error", err)
	} else {
		metrics = append(metrics, ramMetric)
	}

	return metrics, nil
}

func (s *DbMonitoringBackgroundService) collectIOMetrics(
	ctx context.Context,
	conn *pgx.Conn,
	databaseID uuid.UUID,
	timestamp time.Time,
) ([]postgres_monitoring_metrics.PostgresMonitoringMetric, error) {
	var blocksRead, blocksHit int64
	query := `
		SELECT 
			COALESCE(SUM(blks_read), 0) as total_reads,
			COALESCE(SUM(blks_hit), 0) as total_hits
		FROM pg_stat_database 
		WHERE datname = current_database()
	`

	err := conn.QueryRow(ctx, query).Scan(&blocksRead, &blocksHit)
	if err != nil {
		return nil, err
	}

	// Calculate I/O activity as total blocks accessed (PostgreSQL block size is typically 8KB)
	const pgBlockSize = 8192 // 8KB
	totalIOBytes := float64((blocksRead + blocksHit) * pgBlockSize)

	return []postgres_monitoring_metrics.PostgresMonitoringMetric{
		{
			DatabaseID: databaseID,
			Metric:     postgres_monitoring_metrics.MetricsTypeDbIO,
			ValueType:  postgres_monitoring_metrics.MetricsValueTypeByte,
			Value:      totalIOBytes,
			CreatedAt:  timestamp,
		},
	}, nil
}

func (s *DbMonitoringBackgroundService) collectRAMUsageMetric(
	ctx context.Context,
	conn *pgx.Conn,
	databaseID uuid.UUID,
	timestamp time.Time,
) (postgres_monitoring_metrics.PostgresMonitoringMetric, error) {
	var sharedBuffers int64
	query := `
		SELECT 
			COALESCE(SUM(blks_hit), 0) * 8192 as buffer_usage
		FROM pg_stat_database 
		WHERE datname = current_database()
	`

	err := conn.QueryRow(ctx, query).Scan(&sharedBuffers)
	if err != nil {
		return postgres_monitoring_metrics.PostgresMonitoringMetric{}, err
	}

	return postgres_monitoring_metrics.PostgresMonitoringMetric{
		DatabaseID: databaseID,
		Metric:     postgres_monitoring_metrics.MetricsTypeDbRAM,
		ValueType:  postgres_monitoring_metrics.MetricsValueTypeByte,
		Value:      float64(sharedBuffers),
		CreatedAt:  timestamp,
	}, nil
}
