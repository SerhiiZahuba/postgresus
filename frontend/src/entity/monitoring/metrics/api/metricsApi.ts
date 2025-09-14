import { getApplicationServer } from '../../../../constants';
import RequestOptions from '../../../../shared/api/RequestOptions';
import { apiHelper } from '../../../../shared/api/apiHelper';
import type { GetMetricsRequest } from '../model/GetMetricsRequest';
import type { PostgresMonitoringMetric } from '../model/PostgresMonitoringMetric';

export const metricsApi = {
  async getMetrics(request: GetMetricsRequest): Promise<PostgresMonitoringMetric[]> {
    const requestOptions: RequestOptions = new RequestOptions();
    requestOptions.setBody(JSON.stringify(request));
    return apiHelper.fetchPostJson<PostgresMonitoringMetric[]>(
      `${getApplicationServer()}/api/v1/postgres-monitoring-metrics/get`,
      requestOptions,
    );
  },
};
