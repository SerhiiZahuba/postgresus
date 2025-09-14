import { InfoCircleOutlined } from '@ant-design/icons';
import { Button, Select, Spin, Switch, Tooltip } from 'antd';
import { useEffect, useState } from 'react';

import type { Database } from '../../../../entity/databases';
import type { PostgresMonitoringSettings } from '../../../../entity/monitoring/settings';
import { monitoringSettingsApi } from '../../../../entity/monitoring/settings';

interface Props {
  database: Database;

  onCancel: () => void;
  onSaved: (monitoringSettings: PostgresMonitoringSettings) => void;
}

const intervalOptions = [
  { label: '15 seconds', value: 15 },
  { label: '30 seconds', value: 30 },
  { label: '1 minute', value: 60 },
  { label: '2 minutes', value: 120 },
  { label: '5 minutes', value: 300 },
  { label: '10 minutes', value: 600 },
  { label: '15 minutes', value: 900 },
  { label: '30 minutes', value: 1800 },
  { label: '1 hour', value: 3600 },
];

export const EditMonitoringSettingsComponent = ({ database, onCancel, onSaved }: Props) => {
  const [monitoringSettings, setMonitoringSettings] = useState<PostgresMonitoringSettings>();
  const [isUnsaved, setIsUnsaved] = useState(false);
  const [isSaving, setIsSaving] = useState(false);

  const updateSettings = (patch: Partial<PostgresMonitoringSettings>) => {
    setMonitoringSettings((prev) => (prev ? { ...prev, ...patch } : prev));
    setIsUnsaved(true);
  };

  const saveSettings = async () => {
    if (!monitoringSettings) return;

    setIsSaving(true);

    try {
      await monitoringSettingsApi.saveSettings(monitoringSettings);
      setIsUnsaved(false);
      onSaved(monitoringSettings);
    } catch (e) {
      alert((e as Error).message);
    }

    setIsSaving(false);
  };

  useEffect(() => {
    monitoringSettingsApi
      .getSettingsByDbID(database.id)
      .then((res) => {
        setMonitoringSettings(res);
        setIsUnsaved(false);
        setIsSaving(false);
      })
      .catch((e) => {
        alert((e as Error).message);
      });
  }, [database]);

  if (!monitoringSettings) return <Spin size="small" />;

  const isAllFieldsValid = true; // All fields have defaults, so always valid

  return (
    <div>
      <div className="mb-1 flex w-full items-center">
        <div className="min-w-[200px]">Database resources monitoring</div>
        <Switch
          checked={monitoringSettings.isDbResourcesMonitoringEnabled}
          onChange={(checked) => updateSettings({ isDbResourcesMonitoringEnabled: checked })}
          size="small"
        />
        <Tooltip
          className="cursor-pointer"
          title="Monitor database-specific metrics like connections, locks, buffer cache hit ratio, and transaction statistics."
        >
          <InfoCircleOutlined className="ml-2" style={{ color: 'gray' }} />
        </Tooltip>
      </div>

      <div className="mt-4 mb-1 flex w-full items-center">
        <div className="min-w-[200px]">Monitoring interval</div>
        <Select
          value={monitoringSettings.monitoringIntervalSeconds}
          onChange={(v) => updateSettings({ monitoringIntervalSeconds: v })}
          size="small"
          className="max-w-[200px] grow"
          options={intervalOptions}
        />
        <Tooltip
          className="cursor-pointer"
          title="How often to collect monitoring metrics. Lower intervals provide more detailed data but use more resources."
        >
          <InfoCircleOutlined className="ml-2" style={{ color: 'gray' }} />
        </Tooltip>
      </div>

      <div className="mt-5 flex">
        <Button danger ghost className="mr-1" onClick={onCancel}>
          Cancel
        </Button>

        <Button
          type="primary"
          className="mr-5 ml-auto"
          onClick={saveSettings}
          loading={isSaving}
          disabled={!isUnsaved || !isAllFieldsValid}
        >
          Save
        </Button>
      </div>
    </div>
  );
};
