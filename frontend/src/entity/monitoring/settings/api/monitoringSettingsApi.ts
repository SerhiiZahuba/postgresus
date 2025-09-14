import { getApplicationServer } from '../../../../constants';
import RequestOptions from '../../../../shared/api/RequestOptions';
import { apiHelper } from '../../../../shared/api/apiHelper';
import type { PostgresMonitoringSettings } from '../model/PostgresMonitoringSettings';

export const monitoringSettingsApi = {
  async saveSettings(settings: PostgresMonitoringSettings) {
    const requestOptions: RequestOptions = new RequestOptions();
    requestOptions.setBody(JSON.stringify(settings));
    return apiHelper.fetchPostJson<PostgresMonitoringSettings>(
      `${getApplicationServer()}/api/v1/postgres-monitoring-settings/save`,
      requestOptions,
    );
  },

  async getSettingsByDbID(databaseId: string) {
    const requestOptions: RequestOptions = new RequestOptions();
    return apiHelper.fetchGetJson<PostgresMonitoringSettings>(
      `${getApplicationServer()}/api/v1/postgres-monitoring-settings/database/${databaseId}`,
      requestOptions,
      true,
    );
  },
};
