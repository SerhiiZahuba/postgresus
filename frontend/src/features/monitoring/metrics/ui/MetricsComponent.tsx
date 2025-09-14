import { InfoCircleOutlined } from '@ant-design/icons';
import { Button, Spin } from 'antd';
import dayjs from 'dayjs';
import { useEffect, useState } from 'react';
import {
  CartesianGrid,
  Line,
  LineChart,
  ResponsiveContainer,
  Tooltip,
  XAxis,
  YAxis,
} from 'recharts';

import type { GetMetricsRequest, PostgresMonitoringMetric } from '../../../../entity/monitoring';
import { PostgresMonitoringMetricType, metricsApi } from '../../../../entity/monitoring';

interface Props {
  databaseId: string;
}

type Period = '1H' | '24H' | '7D' | '1M';

interface ChartDataPoint {
  timestamp: string;
  displayTime: string;
  ramUsage?: number;
  ioUsage?: number;
}

const formatBytes = (bytes: number): string => {
  const units = ['B', 'KB', 'MB', 'GB', 'TB'];
  let size = bytes;
  let unitIndex = 0;

  while (size >= 1024 && unitIndex < units.length - 1) {
    size /= 1024;
    unitIndex++;
  }

  return `${size.toFixed(1)} ${units[unitIndex]}`;
};

const getDateRange = (period: Period) => {
  const now = dayjs();
  let from: dayjs.Dayjs;

  switch (period) {
    case '1H':
      from = now.subtract(1, 'hour');
      break;
    case '24H':
      from = now.subtract(24, 'hours');
      break;
    case '7D':
      from = now.subtract(7, 'days');
      break;
    case '1M':
      from = now.subtract(1, 'month');
      break;
    default:
      from = now.subtract(24, 'hours');
  }

  return {
    from: from.toISOString(),
    to: now.toISOString(),
  };
};

const getDisplayTime = (timestamp: string, period: Period): string => {
  const date = dayjs(timestamp);

  switch (period) {
    case '1H':
      return date.format('HH:mm');
    case '24H':
      return date.format('HH:mm');
    case '7D':
      return date.format('MM/DD HH:mm');
    case '1M':
      return date.format('MM/DD');
    default:
      return date.format('HH:mm');
  }
};

export const MetricsComponent = ({ databaseId }: Props) => {
  const [selectedPeriod, setSelectedPeriod] = useState<Period>('24H');
  const [isLoading, setIsLoading] = useState(false);
  const [ramData, setRamData] = useState<PostgresMonitoringMetric[]>([]);
  const [ioData, setIoData] = useState<PostgresMonitoringMetric[]>([]);

  const loadMetrics = async (period: Period) => {
    setIsLoading(true);

    try {
      const { from, to } = getDateRange(period);

      const [ramMetrics, ioMetrics] = await Promise.all([
        metricsApi.getMetrics({
          databaseId,
          metricType: PostgresMonitoringMetricType.DB_RAM_USAGE,
          from,
          to,
        } as GetMetricsRequest),
        metricsApi.getMetrics({
          databaseId,
          metricType: PostgresMonitoringMetricType.DB_IO_USAGE,
          from,
          to,
        } as GetMetricsRequest),
      ]);

      setRamData(ramMetrics);
      setIoData(ioMetrics);
    } catch (error) {
      alert((error as Error).message);
    } finally {
      setIsLoading(false);
    }
  };

  useEffect(() => {
    loadMetrics(selectedPeriod);
  }, [databaseId, selectedPeriod]);

  const prepareChartData = (): ChartDataPoint[] => {
    // Create a map for easier lookup
    const ramMap = new Map(ramData.map((item) => [item.createdAt, item.value]));
    const ioMap = new Map(ioData.map((item) => [item.createdAt, item.value]));

    // Get all unique timestamps and sort them
    const allTimestamps = Array.from(
      new Set([...ramData.map((d) => d.createdAt), ...ioData.map((d) => d.createdAt)]),
    ).sort();

    return allTimestamps.map((timestamp) => ({
      timestamp,
      displayTime: getDisplayTime(timestamp, selectedPeriod),
      ramUsage: ramMap.get(timestamp),
      ioUsage: ioMap.get(timestamp),
    }));
  };

  const chartData = prepareChartData();

  const periodButtons: Period[] = ['1H', '24H', '7D', '1M'];

  if (isLoading) {
    return (
      <div className="flex justify-center p-8">
        <Spin size="large" />
      </div>
    );
  }

  return (
    <div className="w-full rounded-tr-md rounded-br-md rounded-bl-md bg-white p-5 shadow">
      <div className="mb-6">
        <h3 className="mb-4 text-lg font-bold">Database Metrics</h3>

        <div className="mb-4 rounded-md border border-yellow-300 bg-yellow-50 p-3 text-sm text-yellow-800">
          <div className="flex items-center">
            <InfoCircleOutlined className="mr-2 text-yellow-600" />
            This feature is in development. Do not consider it as production ready.
          </div>
        </div>

        <div className="flex gap-2">
          {periodButtons.map((period) => (
            <Button
              key={period}
              type={selectedPeriod === period ? 'primary' : 'default'}
              onClick={() => setSelectedPeriod(period)}
              size="small"
            >
              {period}
            </Button>
          ))}
        </div>
      </div>

      <div className="grid grid-cols-1 gap-6 lg:grid-cols-2">
        {/* RAM Usage Chart */}
        <div>
          <h4 className="mb-3 text-base font-semibold">RAM Usage (cumulative)</h4>
          <div style={{ width: '100%', height: '300px' }}>
            <ResponsiveContainer width="100%" height="100%">
              <LineChart data={chartData} margin={{ top: 10, right: 30, left: 20, bottom: 5 }}>
                <CartesianGrid strokeDasharray="3 3" stroke="#f0f0f0" />
                <XAxis dataKey="displayTime" tick={{ fontSize: 12 }} stroke="#666" />
                <YAxis tickFormatter={formatBytes} tick={{ fontSize: 12 }} stroke="#666" />
                <Tooltip
                  formatter={(value: number) => [formatBytes(value), 'RAM Usage']}
                  labelStyle={{ color: '#666' }}
                />
                <Line
                  type="monotone"
                  dataKey="ramUsage"
                  stroke="#1890ff"
                  strokeWidth={2}
                  dot={{ fill: '#1890ff', strokeWidth: 2, r: 3 }}
                  connectNulls={false}
                />
              </LineChart>
            </ResponsiveContainer>
          </div>
        </div>

        {/* IO Usage Chart */}
        <div>
          <h4 className="mb-3 text-base font-semibold">IO Usage (cumulative)</h4>
          <div style={{ width: '100%', height: '300px' }}>
            <ResponsiveContainer width="100%" height="100%">
              <LineChart data={chartData} margin={{ top: 10, right: 30, left: 20, bottom: 5 }}>
                <CartesianGrid strokeDasharray="3 3" stroke="#f0f0f0" />
                <XAxis dataKey="displayTime" tick={{ fontSize: 12 }} stroke="#666" />
                <YAxis tickFormatter={formatBytes} tick={{ fontSize: 12 }} stroke="#666" />
                <Tooltip
                  formatter={(value: number) => [formatBytes(value), 'IO Usage']}
                  labelStyle={{ color: '#666' }}
                />
                <Line
                  type="monotone"
                  dataKey="ioUsage"
                  stroke="#52c41a"
                  strokeWidth={2}
                  dot={{ fill: '#52c41a', strokeWidth: 2, r: 3 }}
                  connectNulls={false}
                />
              </LineChart>
            </ResponsiveContainer>
          </div>
        </div>
      </div>

      {chartData.length === 0 && (
        <div className="mt-6 text-center text-gray-500">
          No metrics data available for the selected period
        </div>
      )}
    </div>
  );
};
