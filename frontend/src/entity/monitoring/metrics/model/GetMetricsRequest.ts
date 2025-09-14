import type { PostgresMonitoringMetricType } from './PostgresMonitoringMetricType';

export interface GetMetricsRequest {
  databaseId: string;
  metricType: PostgresMonitoringMetricType;
  from: string;
  to: string;
}
