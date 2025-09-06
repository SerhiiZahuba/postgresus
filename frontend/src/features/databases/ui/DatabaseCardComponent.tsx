import { InfoCircleOutlined } from '@ant-design/icons';
import dayjs from 'dayjs';

import { type Database, DatabaseType } from '../../../entity/databases';
import { HealthStatus } from '../../../entity/databases/model/HealthStatus';
import { getUserShortTimeFormat } from '../../../shared/time/getUserTimeFormat';

interface Props {
  database: Database;
  selectedDatabaseId?: string;
  setSelectedDatabaseId: (databaseId: string) => void;
}

export const DatabaseCardComponent = ({
  database,
  selectedDatabaseId,
  setSelectedDatabaseId,
}: Props) => {
  let databaseIcon = '';
  let databaseType = '';

  if (database.type === DatabaseType.POSTGRES) {
    databaseIcon = '/icons/databases/postgresql.svg';
    databaseType = 'PostgreSQL';
  }

  return (
    <div
      className={`mb-3 cursor-pointer rounded p-3 shadow ${selectedDatabaseId === database.id ? 'bg-blue-100' : 'bg-white'}`}
      onClick={() => setSelectedDatabaseId(database.id)}
    >
      <div className="flex">
        <div className="mb-1 font-bold">{database.name}</div>

        {database.healthStatus && (
          <div className="ml-auto pl-1">
            <div
              className={`rounded px-[6px] py-[2px] text-[10px] text-white ${
                database.healthStatus === HealthStatus.AVAILABLE ? 'bg-green-500' : 'bg-red-500'
              }`}
            >
              {database.healthStatus === HealthStatus.AVAILABLE ? 'Available' : 'Unavailable'}
            </div>
          </div>
        )}
      </div>

      <div className="mb flex items-center">
        <div className="text-sm text-gray-500">Database type: {databaseType}</div>

        <img src={databaseIcon} alt="databaseIcon" className="ml-1 h-4 w-4" />
      </div>

        {database.storage && (
            <div className="mb flex items-center">
                <div className="text-sm text-gray-500">Storage: {database.storage.name}</div>
                <img
                    src={getStorageLogoFromType(database.storage.type)}
                    alt="storageIcon"
                    className="ml-1 h-4 w-4"
                />
            </div>
        )}

      {database.lastBackupTime && (
        <div className="mt-3 mb-1 text-xs text-gray-500">
          <span className="font-bold">Last backup</span>
          <br />
          {dayjs(database.lastBackupTime).format(getUserShortTimeFormat().format)} (
          {dayjs(database.lastBackupTime).fromNow()})
        </div>
      )}

      {database.lastBackupErrorMessage && (
        <div className="mt-1 flex items-center text-sm text-red-600 underline">
          <InfoCircleOutlined className="mr-1" style={{ color: 'red' }} />
          Has backup error
        </div>
      )}
    </div>
  );
};
