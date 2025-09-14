package postgres_monitoring_metrics

type PostgresMonitoringMetricType string

const (
	// db resources (don't need extensions)
	MetricsTypeDbRAM PostgresMonitoringMetricType = "DB_RAM_USAGE"
	MetricsTypeDbIO  PostgresMonitoringMetricType = "DB_IO_USAGE"
)

type PostgresMonitoringMetricValueType string

const (
	MetricsValueTypeByte    PostgresMonitoringMetricValueType = "BYTE"
	MetricsValueTypePercent PostgresMonitoringMetricValueType = "PERCENT"
)
