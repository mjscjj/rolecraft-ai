import type { ApiResponse } from './client';
import client, { handleApiError } from './client';

export interface WorkspaceTask {
  id: string;
  userId: string;
  companyId?: string;
  name: string;
  description?: string;
  type?: 'general' | 'report' | 'analyze' | string;
  status: 'todo' | 'in_progress' | 'done' | string;
  priority: 'low' | 'medium' | 'high' | string;
  roleId?: string;
  triggerType?: 'manual' | 'once' | 'daily' | 'interval_hours' | string;
  triggerValue?: string;
  timezone?: string;
  nextRunAt?: string;
  lastRunAt?: string;
  asyncStatus?: 'idle' | 'scheduled' | 'running' | 'completed' | 'failed' | string;
  inputSource?: string;
  reportRule?: string;
  resultSummary?: string;
  config?: Record<string, any>;
  createdAt: string;
  updatedAt: string;
}

export type Work = WorkspaceTask;

export interface AgentRunStep {
  agent: string;
  purpose: string;
  output: string;
  durationMs: number;
}

export interface AgentRunAttempt {
  attempt: number;
  status: string;
  durationMs: number;
  error?: string;
  summary?: string;
}

export interface AgentRunPolicy {
  executionMode?: 'serial' | 'parallel' | string;
  timeoutSeconds?: number;
  maxRetries?: number;
  retryDelaySeconds?: number;
  archiveToCompany?: boolean;
  queueRetryOnFailure?: boolean;
  retryWindowMinutes?: number;
  maxFailureCycles?: number;
}

export interface AgentRun {
  id: string;
  workId: string;
  userId: string;
  companyId?: string;
  triggerSource: string;
  status: string;
  summary: string;
  finalAnswer: string;
  confidence: number;
  trace?: {
    steps?: AgentRunStep[];
    nextActions?: string[];
    evidence?: string[];
    attempts?: AgentRunAttempt[];
    policy?: AgentRunPolicy;
    retryQueue?: Record<string, any>;
  };
  errorMessage?: string;
  startedAt?: string;
  finishedAt?: string;
  createdAt: string;
  updatedAt: string;
}

export interface BatchRunItem {
  workId: string;
  status: 'completed' | 'failed' | 'busy' | 'not_found' | string;
  error?: string;
  run?: AgentRun;
}

export interface BatchRunResult {
  items: BatchRunItem[];
  successCount: number;
  failedCount: number;
  busyCount: number;
  notFoundCount: number;
}

export interface WorkspacePayload {
  name: string;
  description?: string;
  companyId?: string;
  type?: string;
  status?: string;
  priority?: string;
  roleId?: string;
  triggerType?: string;
  triggerValue?: string;
  timezone?: string;
  asyncStatus?: string;
  inputSource?: string;
  reportRule?: string;
  resultSummary?: string;
  config?: Record<string, any>;
}

const WORKSPACE_BASE = '/workspaces';

const parseConfig = (value: unknown): WorkspaceTask['config'] | undefined => {
  if (!value) return undefined;
  if (typeof value === 'string') {
    try {
      return JSON.parse(value);
    } catch {
      return undefined;
    }
  }
  if (typeof value === 'object') {
    return value as WorkspaceTask['config'];
  }
  return undefined;
};

const parseTrace = (value: unknown): AgentRun['trace'] | undefined => {
  if (!value) return undefined;
  if (typeof value === 'string') {
    try {
      return JSON.parse(value);
    } catch {
      return undefined;
    }
  }
  if (typeof value === 'object') {
    return value as AgentRun['trace'];
  }
  return undefined;
};

const normalizeRun = (run: any): AgentRun => ({
  ...run,
  trace: parseTrace(run?.trace),
});

const normalizeWork = (work: any): WorkspaceTask => ({
  ...work,
  config: parseConfig(work?.config),
});

export const workApi = {
  list: async (params?: {
    companyId?: string;
    status?: string;
    triggerType?: string;
    asyncStatus?: string;
  }): Promise<WorkspaceTask[]> => {
    try {
      const res = await client.get<ApiResponse<WorkspaceTask[]>>(WORKSPACE_BASE, { params });
      return (res.data.data || []).map(normalizeWork);
    } catch (error) {
      throw handleApiError(error);
    }
  },

  create: async (payload: WorkspacePayload): Promise<WorkspaceTask> => {
    try {
      const res = await client.post<ApiResponse<WorkspaceTask>>(WORKSPACE_BASE, payload);
      return normalizeWork(res.data.data);
    } catch (error) {
      throw handleApiError(error);
    }
  },

  update: async (id: string, payload: WorkspacePayload): Promise<WorkspaceTask> => {
    try {
      const res = await client.put<ApiResponse<WorkspaceTask>>(`${WORKSPACE_BASE}/${id}`, payload);
      return normalizeWork(res.data.data);
    } catch (error) {
      throw handleApiError(error);
    }
  },

  run: async (id: string): Promise<{ work: WorkspaceTask; run: AgentRun }> => {
    try {
      const res = await client.post<ApiResponse<{ work: WorkspaceTask; run: AgentRun }>>(`${WORKSPACE_BASE}/${id}/run`);
      return {
        ...res.data.data,
        work: normalizeWork(res.data.data.work),
        run: normalizeRun(res.data.data.run),
      };
    } catch (error) {
      throw handleApiError(error);
    }
  },

  batchRun: async (ids: string[], maxParallel = 3): Promise<BatchRunResult> => {
    try {
      const res = await client.post<ApiResponse<BatchRunResult>>(`${WORKSPACE_BASE}/batch/run`, {
        ids,
        maxParallel,
      });
      return {
        ...res.data.data,
        items: (res.data.data.items || []).map((item) => ({
          ...item,
          run: item.run ? normalizeRun(item.run) : undefined,
        })),
      };
    } catch (error) {
      throw handleApiError(error);
    }
  },

  listRuns: async (id: string, limit = 20): Promise<AgentRun[]> => {
    try {
      const res = await client.get<ApiResponse<AgentRun[]>>(`${WORKSPACE_BASE}/${id}/runs`, {
        params: { limit },
      });
      return (res.data.data || []).map(normalizeRun);
    } catch (error) {
      throw handleApiError(error);
    }
  },

  getRun: async (id: string, runId: string): Promise<AgentRun> => {
    try {
      const res = await client.get<ApiResponse<AgentRun>>(`${WORKSPACE_BASE}/${id}/runs/${runId}`);
      return normalizeRun(res.data.data);
    } catch (error) {
      throw handleApiError(error);
    }
  },

  delete: async (id: string): Promise<void> => {
    try {
      await client.delete(`${WORKSPACE_BASE}/${id}`);
    } catch (error) {
      throw handleApiError(error);
    }
  },
};

export default workApi;
