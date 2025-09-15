import {
  ClearOutlined,
  DownloadOutlined,
  InfoCircleOutlined,
  PlayCircleOutlined,
} from '@ant-design/icons';
import { Alert, Button, Input, InputNumber, Space, Table, Tag, Tooltip, Typography } from 'antd';
import { useMemo, useState } from 'react';

import { type ExecuteResponse, sqlQueryApi } from '../../../entity/query';
import { ToastHelper } from '../../../shared/toast';

const { TextArea } = Input;
const { Text } = Typography;

interface Props {
  databaseId: string;
}

function toCsv(resp: ExecuteResponse): string {
  const esc = (v: unknown) => {
    if (v === null || v === undefined) return '';
    const s = typeof v === 'object' ? JSON.stringify(v) : String(v);
    if (/[",\n\r]/.test(s)) return `"${s.replace(/"/g, '""')}"`;
    return s;
  };
  const header = resp.columns.map((c: string) => esc(c)).join(',');
  const rows = resp.rows.map((r: unknown[]) => r.map(esc).join(','));
  return [header, ...rows].join('\n');
}

function friendlyError(msg: string): string {
  const m = msg || 'Unexpected error';
  if (/SQLSTATE\s*42P01/i.test(m) || /relation .* does not exist/i.test(m)) {
    return `${m}\n\nHint: table (relation) not found. Check schema/name or use schema-qualified name (e.g. public.users).`;
  }
  return m;
}

export const SqlQueryComponent = ({ databaseId }: Props) => {
  const [sql, setSql] = useState<string>('SELECT now() AS ts');
  const [maxRows, setMaxRows] = useState<number>(1000);
  const [timeoutSec, setTimeoutSec] = useState<number>(30);
  const [loading, setLoading] = useState(false);

  const [resp, setResp] = useState<ExecuteResponse | null>(null);
  const [errorText, setErrorText] = useState<string | null>(null);

  const columns = useMemo(() => {
    if (!resp) return [];
    return resp.columns.map((name: string, idx: number) => ({
      title: name || `col_${idx + 1}`,
      dataIndex: `c${idx}`,
      key: `c${idx}`,
      ellipsis: true,
    }));
  }, [resp]);

  const dataSource = useMemo(() => {
    if (!resp) return [];
    return resp.rows.map((row: unknown[], i: number) => {
      const record: Record<string, unknown> = { key: i };
      row.forEach((v: unknown, idx: number) => {
        record[`c${idx}`] = v;
      });
      return record;
    });
  }, [resp]);

  const run = async () => {
    if (!sql.trim()) {
      ToastHelper.showToast({ title: 'Empty SQL', description: 'Please enter a query' });
      return;
    }

    setLoading(true);
    setResp(null);
    setErrorText(null);
    try {
      const r = await sqlQueryApi.execute({
        databaseId: databaseId,
        sql,
        maxRows,
        timeoutSec,
      });
      setResp(r);
    } catch (e: unknown) {
      let msg = 'Unexpected error';
      if (e instanceof Error) msg = e.message;
      else if (typeof e === 'string') msg = e;
      setErrorText(friendlyError(msg));
    } finally {
      setLoading(false);
    }
  };

  const clearAll = () => {
    setResp(null);
    setErrorText(null);
  };

  const downloadCsv = () => {
    if (!resp) return;
    const csv = toCsv(resp);
    const blob = new Blob([csv], { type: 'text/csv;charset=utf-8;' });
    const url = URL.createObjectURL(blob);
    const a = document.createElement('a');
    a.href = url;
    a.download = 'result.csv';
    a.click();
    URL.revokeObjectURL(url);
  };

  const downloadJson = () => {
    if (!resp) return;
    const items = resp.rows.map((row: unknown[]) => {
      const obj: Record<string, unknown> = {};
      resp.columns.forEach((name: string, i: number) => {
        obj[name || `col_${i + 1}`] = row[i];
      });
      return obj;
    });
    const blob = new Blob([JSON.stringify(items, null, 2)], { type: 'application/json' });
    const url = URL.createObjectURL(blob);
    const a = document.createElement('a');
    a.href = url;
    a.download = 'result.json';
    a.click();
    URL.revokeObjectURL(url);
  };

  return (
    <div className="mt-2 rounded border border-gray-200 bg-white p-3">
      <div className="mb-2 flex items-center">
        <div className="font-bold">SQL query</div>
        <Tooltip title="Single statement. Results may be limited by server policy.">
          <InfoCircleOutlined className="ml-2 text-gray-500" />
        </Tooltip>
      </div>

      <TextArea
        value={sql}
        onChange={(e) => setSql(e.target.value)}
        autoSize={{ minRows: 4, maxRows: 12 }}
        placeholder="Write your SQL here…"
      />

      <div className="mt-2 flex flex-wrap items-center gap-3">
        <Space>
          <Button type="primary" icon={<PlayCircleOutlined />} loading={loading} onClick={run}>
            Run
          </Button>
          <Button
            icon={<ClearOutlined />}
            onClick={clearAll}
            disabled={loading && !resp && !errorText}
          >
            Clear result
          </Button>
        </Space>

        <div className="mx-2 h-5 w-px bg-gray-300" />

        <Space size="small">
          <Text type="secondary">Max rows:</Text>
          <InputNumber
            min={1}
            max={10000}
            value={maxRows}
            onChange={(v) => setMaxRows(v ?? 1000)}
          />
        </Space>

        <Space size="small">
          <Text type="secondary">Timeout (s):</Text>
          <InputNumber
            min={1}
            max={120}
            value={timeoutSec}
            onChange={(v) => setTimeoutSec(v ?? 30)}
          />
        </Space>

        {resp && (
          <>
            <div className="mx-2 h-5 w-px bg-gray-300" />
            <Space size="small">
              <Button icon={<DownloadOutlined />} onClick={downloadCsv}>
                CSV
              </Button>
              <Button icon={<DownloadOutlined />} onClick={downloadJson}>
                JSON
              </Button>
            </Space>
          </>
        )}
      </div>

      {errorText && (
        <div className="mt-3">
          <Alert
            type="error"
            showIcon
            message="Query error"
            description={<pre className="m-0 break-words whitespace-pre-wrap">{errorText}</pre>}
          />
        </div>
      )}

      {resp && (
        <div className="mt-3">
          <div className="mb-2 flex flex-wrap items-center gap-3 text-sm">
            <Text type="secondary">Rows:</Text>
            <Text strong>{resp.rowCount}</Text>

            <Text type="secondary">Time:</Text>
            <Text strong>{resp.executionMs} ms</Text>

            {resp.truncated && <Tag color="orange">Truncated</Tag>}
          </div>

          <Table
            size="small"
            bordered
            loading={loading}
            columns={columns}
            dataSource={dataSource}
            scroll={{ x: true }}
            pagination={{ pageSize: 50 }}
          />
        </div>
      )}
    </div>
  );
};
