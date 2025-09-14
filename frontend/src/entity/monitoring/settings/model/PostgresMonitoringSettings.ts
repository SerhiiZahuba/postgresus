import type { Database } from '../../../databases';
import { PostgresqlExtension } from './PostgresqlExtension';

export interface PostgresMonitoringSettings {
  databaseId: string;
  database?: Database;

  isDbResourcesMonitoringEnabled: boolean;
  monitoringIntervalSeconds: number;

  installedExtensions: PostgresqlExtension[];
  installedExtensionsRaw?: string;
}
