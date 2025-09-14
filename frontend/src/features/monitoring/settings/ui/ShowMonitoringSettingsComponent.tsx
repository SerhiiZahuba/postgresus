import { Switch } from 'antd';
import { useEffect, useState } from 'react';

import type { Database } from '../../../../entity/databases';
import type { PostgresMonitoringSettings } from '../../../../entity/monitoring/settings';
import { monitoringSettingsApi } from '../../../../entity/monitoring/settings';

interface Props {
  database: Database;
}

const intervalLabels = {
  15: '15 seconds',
  30: '30 seconds',
  60: '1 minute',
  120: '2 minutes',
  300: '5 minutes',
  600: '10 minutes',
  900: '15 minutes',
  1800: '30 minutes',
  3600: '1 hour',
};

export const ShowMonitoringSettingsComponent = ({ database }: Props) => {
  const [monitoringSettings, setMonitoringSettings] = useState<PostgresMonitoringSettings>();

  useEffect(() => {
    if (database.id) {
      monitoringSettingsApi.getSettingsByDbID(database.id).then((res) => {
        setMonitoringSettings(res);
      });
    }
  }, [database]);

  if (!monitoringSettings) return <div />;

  return (
    <div>
      <div className="mb-1 flex w-full items-center">
        <div className="min-w-[200px]">Database resources monitoring</div>
        <Switch checked={monitoringSettings.isDbResourcesMonitoringEnabled} disabled size="small" />
      </div>

      <div className="mb-1 flex w-full items-center">
        <div className="min-w-[200px]">Monitoring interval</div>
        <div>
          {intervalLabels[
            monitoringSettings.monitoringIntervalSeconds as keyof typeof intervalLabels
          ] || `${monitoringSettings.monitoringIntervalSeconds} seconds`}
        </div>
      </div>
    </div>
  );
};
