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

export const workApi = {
  list: async (params?: {
    companyId?: string;
    status?: string;
    triggerType?: string;
    asyncStatus?: string;
  }): Promise<WorkspaceTask[]> => {
    try {
      const res = await client.get<ApiResponse<WorkspaceTask[]>>(WORKSPACE_BASE, { params });
      return res.data.data;
    } catch (error) {
      throw handleApiError(error);
    }
  },

  create: async (payload: WorkspacePayload): Promise<WorkspaceTask> => {
    try {
      const res = await client.post<ApiResponse<WorkspaceTask>>(WORKSPACE_BASE, payload);
      return res.data.data;
    } catch (error) {
      throw handleApiError(error);
    }
  },

  update: async (id: string, payload: WorkspacePayload): Promise<WorkspaceTask> => {
    try {
      const res = await client.put<ApiResponse<WorkspaceTask>>(`${WORKSPACE_BASE}/${id}`, payload);
      return res.data.data;
    } catch (error) {
      throw handleApiError(error);
    }
  },

  run: async (id: string): Promise<WorkspaceTask> => {
    try {
      const res = await client.post<ApiResponse<WorkspaceTask>>(`${WORKSPACE_BASE}/${id}/run`);
      return res.data.data;
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
