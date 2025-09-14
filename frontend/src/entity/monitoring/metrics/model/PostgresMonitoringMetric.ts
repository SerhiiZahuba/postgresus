import type { PostgresMonitoringMetricType } from './PostgresMonitoringMetricType';
import type { PostgresMonitoringMetricValueType } from './PostgresMonitoringMetricValueType';

export interface PostgresMonitoringMetric {
  id: string;
  databaseId: string;
  metric: PostgresMonitoringMetricType;
  valueType: PostgresMonitoringMetricValueType;
  value: number;
  createdAt: string;
}
