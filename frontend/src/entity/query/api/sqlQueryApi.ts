// src/entity/api/sqlQueryApi.ts
import { apiHelper } from '../../../shared/api/apiHelper';
import RequestOptions from '../../../shared/api/RequestOptions';
import type { ExecuteResponse } from '../model/ExecuteResponse';

export type ExecuteRequest = {
    databaseId: string; // was database_id
    sql: string;
    maxRows?: number;    // was max_rows
    timeoutSec?: number; // was timeout_sec
};


export const sqlQueryApi = {
    execute: async (payload: ExecuteRequest): Promise<ExecuteResponse> => {
        const opts = new RequestOptions().setBody(JSON.stringify(payload));
        return apiHelper.fetchPostJson<ExecuteResponse>('/api/v1/sqlquery/execute', opts);
    },
};
