import { InfoCircleOutlined } from '@ant-design/icons';
import {
  Button,
  Checkbox,
  Input,
  InputNumber,
  Modal,
  Select,
  Spin,
  Switch,
  TimePicker,
  Tooltip,
} from 'antd';
import dayjs, { Dayjs } from 'dayjs';
import { useEffect, useMemo, useState } from 'react';

import { type BackupConfig, backupConfigApi } from '../../../entity/backups';
import { BackupNotificationType } from '../../../entity/backups/model/BackupNotificationType';
import type { Database } from '../../../entity/databases';
import { Period } from '../../../entity/databases/model/Period';
import { type Interval, IntervalType } from '../../../entity/intervals';
import { type Storage, getStorageLogoFromType, storageApi } from '../../../entity/storages';
import {
  getLocalDayOfMonth,
  getLocalWeekday,
  getUserTimeFormat,
  getUtcDayOfMonth,
  getUtcWeekday,
} from '../../../shared/time/utils';
import { ConfirmationComponent } from '../../../shared/ui';
import { EditStorageComponent } from '../../storages/ui/edit/EditStorageComponent';

interface Props {
  database: Database;

  isShowBackButton: boolean;
  onBack: () => void;

  isShowCancelButton?: boolean;
  onCancel: () => void;

  saveButtonText?: string;
  isSaveToApi: boolean;
  onSaved: (backupConfig: BackupConfig) => void;
}

const weekdayOptions = [
  { value: 1, label: 'Mon' },
  { value: 2, label: 'Tue' },
  { value: 3, label: 'Wed' },
  { value: 4, label: 'Thu' },
  { value: 5, label: 'Fri' },
  { value: 6, label: 'Sat' },
  { value: 7, label: 'Sun' },
];

export const EditBackupConfigComponent = ({
  database,

  isShowBackButton,
  onBack,

  isShowCancelButton,
  onCancel,
  saveButtonText,
  isSaveToApi,
  onSaved,
}: Props) => {
  const [backupConfig, setBackupConfig] = useState<BackupConfig>();
  const [isUnsaved, setIsUnsaved] = useState(false);
  const [isSaving, setIsSaving] = useState(false);

  const [storages, setStorages] = useState<Storage[]>([]);
  const [isStoragesLoading, setIsStoragesLoading] = useState(false);
  const [isShowCreateStorage, setShowCreateStorage] = useState(false);

  const [isShowWarn, setIsShowWarn] = useState(false);
  const [isShowBackupDisableConfirm, setIsShowBackupDisableConfirm] = useState(false);

  const timeFormat = useMemo(() => {
    const is12 = getUserTimeFormat();
    return { use12Hours: is12, format: is12 ? 'h:mm A' : 'HH:mm' };
  }, []);

  const updateBackupConfig = (patch: Partial<BackupConfig>) => {
    setBackupConfig((prev) => (prev ? { ...prev, ...patch } : prev));
    setIsUnsaved(true);
  };

  const saveInterval = (patch: Partial<Interval>) => {
    setBackupConfig((prev) => {
      if (!prev) return prev;

      const updatedBackupInterval = { ...(prev.backupInterval ?? {}), ...patch };

      if (!updatedBackupInterval.id && prev.backupInterval?.id) {
        updatedBackupInterval.id = prev.backupInterval.id;
      }

      return { ...prev, backupInterval: updatedBackupInterval as Interval };
    });

    setIsUnsaved(true);
  };

  const saveBackupConfig = async () => {
    if (!backupConfig) return;

    if (isSaveToApi) {
      setIsSaving(true);
      try {
        await backupConfigApi.saveBackupConfig(backupConfig);
        setIsUnsaved(false);
      } catch (e) {
        alert((e as Error).message);
      }
      setIsSaving(false);
    }

    onSaved(backupConfig);
  };

  const loadStorages = async () => {
    setIsStoragesLoading(true);

    try {
      const storages = await storageApi.getStorages();
      setStorages(storages);
    } catch (e) {
      alert((e as Error).message);
    }

    setIsStoragesLoading(false);
  };

  useEffect(() => {
    if (database.id) {
      backupConfigApi.getBackupConfigByDbID(database.id).then((res) => {
        setBackupConfig(res);
        setIsUnsaved(false);
        setIsSaving(false);
      });
    } else {
      setBackupConfig({
        databaseId: database.id,
        isBackupsEnabled: true,
        backupInterval: {
          id: undefined as unknown as string,
          interval: IntervalType.DAILY,
          timeOfDay: '00:00',
        },
        storage: undefined,
        cpuCount: 1,
        storePeriod: Period.WEEK,
        sendNotificationsOn: [],
        isRetryIfFailed: true,
        maxFailedTriesCount: 3,
      });
    }
    loadStorages();
  }, [database]);

  if (!backupConfig) return <div />;

  if (isStoragesLoading) {
    return (
      <div className="mb-5 flex items-center">
        <Spin />
      </div>
    );
  }

  const { backupInterval } = backupConfig;

  // UTC → local conversions for display
  const localTime: Dayjs | undefined = backupInterval?.timeOfDay
    ? dayjs.utc(backupInterval.timeOfDay, 'HH:mm').local()
    : undefined;

  const displayedWeekday: number | undefined =
    backupInterval?.interval === IntervalType.WEEKLY &&
    backupInterval.weekday &&
    backupInterval.timeOfDay
      ? getLocalWeekday(backupInterval.weekday, backupInterval.timeOfDay)
      : backupInterval?.weekday;

  const displayedDayOfMonth: number | undefined =
    backupInterval?.interval === IntervalType.MONTHLY &&
    backupInterval.dayOfMonth &&
    backupInterval.timeOfDay
      ? getLocalDayOfMonth(backupInterval.dayOfMonth, backupInterval.timeOfDay)
      : backupInterval?.dayOfMonth;

  // mandatory-field check
  const isAllFieldsFilled =
    !backupConfig.isBackupsEnabled ||
    (Boolean(backupConfig.storePeriod) &&
      Boolean(backupConfig.storage?.id) &&
      Boolean(backupConfig.cpuCount) &&
      Boolean(backupInterval?.interval) &&
      (!backupInterval ||
        ((backupInterval.interval !== IntervalType.WEEKLY || displayedWeekday) &&
          (backupInterval.interval !== IntervalType.MONTHLY || displayedDayOfMonth) &&
          (backupInterval.interval !== IntervalType.CRON || !!backupInterval.cronExpr))));

  return (
    <div>
      <div className="mb-1 flex w-full items-center">
        <div className="min-w-[150px]">Backups enabled</div>
        <Switch
          checked={backupConfig.isBackupsEnabled}
          onChange={(checked) => {
            // If disabling backups on existing database, show confirmation
            if (!checked && database.id && backupConfig.isBackupsEnabled) {
              setIsShowBackupDisableConfirm(true);
            } else {
              updateBackupConfig({ isBackupsEnabled: checked });
            }
          }}
          size="small"
        />
      </div>

      {backupConfig.isBackupsEnabled && (
        <>
          <div className="mt-4 mb-1 flex w-full items-center">
            <div className="min-w-[150px]">Backup interval</div>
            <Select
              value={backupInterval?.interval}
              onChange={(v) => saveInterval({ interval: v })}
              size="small"
              className="max-w-[200px] grow"
              options={[
                { label: 'Hourly', value: IntervalType.HOURLY },
                { label: 'Daily', value: IntervalType.DAILY },
                { label: 'Weekly', value: IntervalType.WEEKLY },
                { label: 'Monthly', value: IntervalType.MONTHLY },
                { label: 'Custom (CRON)', value: IntervalType.CRON },
              ]}
            />
          </div>

          {backupInterval?.interval === IntervalType.WEEKLY && (
            <div className="mb-1 flex w-full items-center">
              <div className="min-w-[150px]">Backup weekday</div>
              <Select
                value={displayedWeekday}
                onChange={(localWeekday) => {
                  if (!localWeekday) return;
                  const ref = localTime ?? dayjs();
                  saveInterval({ weekday: getUtcWeekday(localWeekday, ref) });
                }}
                size="small"
                className="max-w-[200px] grow"
                options={weekdayOptions}
              />
            </div>
          )}

          {backupInterval?.interval === IntervalType.MONTHLY && (
            <div className="mb-1 flex w-full items-center">
              <div className="min-w-[150px]">Backup day of month</div>
              <InputNumber
                min={1}
                max={31}
                value={displayedDayOfMonth}
                onChange={(localDom) => {
                  if (!localDom) return;
                  const ref = localTime ?? dayjs();
                  saveInterval({ dayOfMonth: getUtcDayOfMonth(localDom, ref) });
                }}
                size="small"
                className="max-w-[200px] grow"
              />
            </div>
          )}

          {backupInterval?.interval !== IntervalType.HOURLY && (
            <div className="mb-1 flex w-full items-center">
              <div className="min-w-[150px]">Backup time of day</div>
              <TimePicker
                value={localTime}
                format={timeFormat.format}
                use12Hours={timeFormat.use12Hours}
                allowClear={false}
                size="small"
                className="max-w-[200px] grow"
                onChange={(t) => {
                  if (!t) return;
                  const patch: Partial<Interval> = { timeOfDay: t.utc().format('HH:mm') };

                  if (backupInterval?.interval === IntervalType.WEEKLY && displayedWeekday) {
                    patch.weekday = getUtcWeekday(displayedWeekday, t);
                  }
                  if (backupInterval?.interval === IntervalType.MONTHLY && displayedDayOfMonth) {
                    patch.dayOfMonth = getUtcDayOfMonth(displayedDayOfMonth, t);
                  }

                  saveInterval(patch);
                }}
              />
            </div>
          )}

          {backupInterval?.interval === IntervalType.CRON && (
            <div className="mb-1 flex w-full items-center">
              <div className="min-w-[150px]">CRON expression</div>
              <Input
                value={backupInterval?.cronExpr}
                onChange={(e: React.ChangeEvent<HTMLInputElement>) =>
                  saveInterval({ cronExpr: e.target.value })
                }
                size="small"
                placeholder="e.g. 0 9 * * *"
                className="max-w-[250px] grow"
              />
              <Tooltip
                className="ml-2 cursor-pointer"
                title="Standard CRON format. Example: '0 9 * * *' = every day at 09:00"
              >
                <InfoCircleOutlined style={{ color: 'gray' }} />
              </Tooltip>
            </div>
          )}

          <div className="mt-4 mb-1 flex w-full items-center">
            <div className="min-w-[150px]">Retry backup if failed</div>
            <Switch
              size="small"
              checked={backupConfig.isRetryIfFailed}
              onChange={(checked) => updateBackupConfig({ isRetryIfFailed: checked })}
            />

            <Tooltip
              className="cursor-pointer"
              title="Automatically retry failed backups. Backups can fail due to network failures, storage issues or temporary database unavailability."
            >
              <InfoCircleOutlined className="ml-2" style={{ color: 'gray' }} />
            </Tooltip>
          </div>

          {backupConfig.isRetryIfFailed && (
            <div className="mb-1 flex w-full items-center">
              <div className="min-w-[150px]">Max failed tries count</div>
              <InputNumber
                min={1}
                max={10}
                value={backupConfig.maxFailedTriesCount}
                onChange={(value) => updateBackupConfig({ maxFailedTriesCount: value || 1 })}
                size="small"
                className="max-w-[200px] grow"
              />

              <Tooltip
                className="cursor-pointer"
                title="Maximum number of retry attempts for failed backups. You will receive a notification when all tries have failed."
              >
                <InfoCircleOutlined className="ml-2" style={{ color: 'gray' }} />
              </Tooltip>
            </div>
          )}

          <div className="mt-5 mb-1 flex w-full items-center">
            <div className="min-w-[150px]">CPU count</div>
            <InputNumber
              min={1}
              max={16}
              value={backupConfig.cpuCount}
              onChange={(value) => updateBackupConfig({ cpuCount: value || 1 })}
              size="small"
              className="max-w-[200px] grow"
            />

            <Tooltip
              className="cursor-pointer"
              title="Number of CPU cores to use for restore processing. Higher values may speed up restores, but use more resources."
            >
              <InfoCircleOutlined className="ml-2" style={{ color: 'gray' }} />
            </Tooltip>
          </div>

          <div className="mb-1 flex w-full items-center">
            <div className="min-w-[150px]">Store period</div>
            <Select
              value={backupConfig.storePeriod}
              onChange={(v) => updateBackupConfig({ storePeriod: v })}
              size="small"
              className="max-w-[200px] grow"
              options={[
                { label: '1 day', value: Period.DAY },
                { label: '1 week', value: Period.WEEK },
                { label: '1 month', value: Period.MONTH },
                { label: '3 months', value: Period.THREE_MONTH },
                { label: '6 months', value: Period.SIX_MONTH },
                { label: '1 year', value: Period.YEAR },
                { label: '2 years', value: Period.TWO_YEARS },
                { label: '3 years', value: Period.THREE_YEARS },
                { label: '4 years', value: Period.FOUR_YEARS },
                { label: '5 years', value: Period.FIVE_YEARS },
                { label: 'Forever', value: Period.FOREVER },
              ]}
            />

            <Tooltip
              className="cursor-pointer"
              title="How long to keep the backups? Make sure you have enough storage space."
            >
              <InfoCircleOutlined className="ml-2" style={{ color: 'gray' }} />
            </Tooltip>
          </div>

          <div className="mt-5 mb-1 flex w-full items-center">
            <div className="min-w-[150px]">Storage</div>
            <Select
              value={backupConfig.storage?.id}
              onChange={(storageId) => {
                if (storageId.includes('create-new-storage')) {
                  setShowCreateStorage(true);
                  return;
                }

                const selectedStorage = storages.find((s) => s.id === storageId);
                updateBackupConfig({ storage: selectedStorage });

                if (backupConfig.storage?.id) {
                  setIsShowWarn(true);
                }
              }}
              size="small"
              className="mr-2 max-w-[200px] grow"
              options={[
                ...storages.map((s) => ({ label: s.name, value: s.id })),
                { label: 'Create new storage', value: 'create-new-storage' },
              ]}
              placeholder="Select storage"
            />

            {backupConfig.storage?.type && (
              <img
                src={getStorageLogoFromType(backupConfig.storage.type)}
                alt="storageIcon"
                className="ml-1 h-4 w-4"
              />
            )}
          </div>

          <div className="mt-4 mb-1 flex w-full items-start">
            <div className="mt-1 min-w-[150px]">Notifications</div>
            <div className="flex flex-col space-y-2">
              <Checkbox
                checked={backupConfig.sendNotificationsOn.includes(
                  BackupNotificationType.BackupSuccess,
                )}
                onChange={(e) => {
                  const notifications = [...backupConfig.sendNotificationsOn];
                  const index = notifications.indexOf(BackupNotificationType.BackupSuccess);
                  if (e.target.checked && index === -1) {
                    notifications.push(BackupNotificationType.BackupSuccess);
                  } else if (!e.target.checked && index > -1) {
                    notifications.splice(index, 1);
                  }
                  updateBackupConfig({ sendNotificationsOn: notifications });
                }}
              >
                Backup success
              </Checkbox>

              <Checkbox
                checked={backupConfig.sendNotificationsOn.includes(
                  BackupNotificationType.BackupFailed,
                )}
                onChange={(e) => {
                  const notifications = [...backupConfig.sendNotificationsOn];
                  const index = notifications.indexOf(BackupNotificationType.BackupFailed);
                  if (e.target.checked && index === -1) {
                    notifications.push(BackupNotificationType.BackupFailed);
                  } else if (!e.target.checked && index > -1) {
                    notifications.splice(index, 1);
                  }
                  updateBackupConfig({ sendNotificationsOn: notifications });
                }}
              >
                Backup failed
              </Checkbox>
            </div>
          </div>
        </>
      )}

      <div className="mt-5 flex">
        {isShowBackButton && (
          <Button className="mr-1" onClick={onBack}>
            Back
          </Button>
        )}

        {isShowCancelButton && (
          <Button danger ghost className="mr-1" onClick={onCancel}>
            Cancel
          </Button>
        )}

        <Button
          type="primary"
          className={`${isShowCancelButton ? 'ml-1' : 'ml-auto'} mr-5`}
          onClick={saveBackupConfig}
          loading={isSaving}
          disabled={!isUnsaved || !isAllFieldsFilled}
        >
          {saveButtonText || 'Save'}
        </Button>
      </div>

      {isShowCreateStorage && (
        <Modal
          title="Add storage"
          footer={<div />}
          open={isShowCreateStorage}
          onCancel={() => setShowCreateStorage(false)}
        >
          <div className="my-3 max-w-[275px] text-gray-500">
            Storage - is a place where backups will be stored (local disk, S3, Google Drive, etc.)
          </div>

          <EditStorageComponent
            isShowName
            isShowClose={false}
            onClose={() => setShowCreateStorage(false)}
            onChanged={() => {
              loadStorages();
              setShowCreateStorage(false);
            }}
          />
        </Modal>
      )}

      {isShowWarn && (
        <ConfirmationComponent
          onConfirm={() => {
            setIsShowWarn(false);
          }}
          onDecline={() => {
            setIsShowWarn(false);
          }}
          description="If you change the storage, all backups in this storage will be deleted."
          actionButtonColor="red"
          actionText="I understand"
          cancelText="Cancel"
          hideCancelButton
        />
      )}

      {isShowBackupDisableConfirm && (
        <ConfirmationComponent
          onConfirm={() => {
            updateBackupConfig({ isBackupsEnabled: false });
            setIsShowBackupDisableConfirm(false);
          }}
          onDecline={() => {
            setIsShowBackupDisableConfirm(false);
          }}
          description="All current backups will be removed? Are you sure?"
          actionButtonColor="red"
          actionText="Yes, disable backing up and remove all existing backup files"
          cancelText="Cancel"
        />
      )}
    </div>
  );
};
