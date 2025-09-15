export interface ExecuteResponse {
  columns: string[];
  rows: unknown[][];
  rowCount: number; // was row_count
  truncated: boolean;
  executionMs: number; // was execution_ms
}
