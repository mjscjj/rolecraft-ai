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
  name: string;
  type: string;
  asyncStatus: string;
  resultSummary: string;
  updatedAt: string;
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

  get: async (id: string): Promise<{ company: Company; stats: CompanyStats; recentOutcomes: CompanyOutcome[] }> => {
    try {
      const res = await client.get<ApiResponse<{ company: Company; stats: CompanyStats; recentOutcomes: CompanyOutcome[] }>>(`/companies/${id}`);
      return res.data.data;
    } catch (error) {
      throw handleApiError(error);
    }
  },
};

export default companyApi;
