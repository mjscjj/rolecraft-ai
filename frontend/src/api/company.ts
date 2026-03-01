import type { ApiResponse } from './client';
import client, { handleApiError } from './client';

export interface Company {
  id: string;
  ownerId: string;
  name: string;
  description?: string;
  createdAt: string;
  updatedAt: string;
}

export interface CompanyStats {
  roleCount: number;
  workCount: number;
  workspaceCount: number;
  outcomeCount: number;
  docCount: number;
}

export interface CompanyOutcome {
  id: string;
  workId: string;
  workName?: string;
  status: string;
  confidence: number;
  resultSummary: string;
  updatedAt: string;
}

export interface CompanyDelivery {
  id: string;
  workId: string;
  workName?: string;
  summary: string;
  finalAnswer: string;
  confidence: number;
  stepCount: number;
  nextActions: string[];
  evidence: string[];
  updatedAt: string;
}

export type CompanyExportFormat = 'json' | 'markdown';

export interface CompanyExportFilters {
  keyword?: string;
  minConfidence?: number;
  from?: string;
  to?: string;
}

export interface CompanyExportRecord {
  id: string;
  companyId: string;
  companyName: string;
  format: CompanyExportFormat | string;
  fileName: string;
  deliveryCount: number;
  filters?: Record<string, any>;
  createdAt: string;
  updatedAt: string;
}

export interface CompanyExportDetail extends CompanyExportRecord {
  content: string;
}

export const companyApi = {
  list: async (): Promise<Company[]> => {
    try {
      const res = await client.get<ApiResponse<Company[]>>('/companies');
      return res.data.data;
    } catch (error) {
      throw handleApiError(error);
    }
  },

  create: async (payload: { name: string; description?: string }): Promise<Company> => {
    try {
      const res = await client.post<ApiResponse<Company>>('/companies', payload);
      return res.data.data;
    } catch (error) {
      throw handleApiError(error);
    }
  },

  get: async (id: string): Promise<{ company: Company; stats: CompanyStats; recentOutcomes: CompanyOutcome[]; deliveryBoard: CompanyDelivery[] }> => {
    try {
      const res = await client.get<ApiResponse<{ company: Company; stats: CompanyStats; recentOutcomes: CompanyOutcome[]; deliveryBoard: CompanyDelivery[] }>>(`/companies/${id}`);
      return res.data.data;
    } catch (error) {
      throw handleApiError(error);
    }
  },

  listExports: async (id: string, limit = 20): Promise<CompanyExportRecord[]> => {
    try {
      const res = await client.get<ApiResponse<CompanyExportRecord[]>>(`/companies/${id}/exports`, {
        params: { limit },
      });
      return res.data.data || [];
    } catch (error) {
      throw handleApiError(error);
    }
  },

  getExport: async (id: string, exportId: string): Promise<CompanyExportDetail> => {
    try {
      const res = await client.get<ApiResponse<CompanyExportDetail>>(`/companies/${id}/exports/${exportId}`);
      return res.data.data;
    } catch (error) {
      throw handleApiError(error);
    }
  },

  createExport: async (
    id: string,
    payload: CompanyExportFilters & { format: CompanyExportFormat }
  ): Promise<CompanyExportDetail> => {
    try {
      const res = await client.post<ApiResponse<CompanyExportDetail>>(`/companies/${id}/exports`, payload);
      return res.data.data;
    } catch (error) {
      throw handleApiError(error);
    }
  },
};

export default companyApi;
