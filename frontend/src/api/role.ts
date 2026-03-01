import type { ApiResponse } from './client';
import client, { handleApiError } from './client';

// 角色类型
export interface Role {
  id: string;
  userId?: string;
  companyId?: string;
  name: string;
  description: string;
  avatar?: string;
  category: string;
  systemPrompt: string;
  welcomeMessage: string;
  modelConfig?: Record<string, any>;
  isTemplate: boolean;
  createdAt: string;
  updatedAt: string;
}

// 创建角色请求
export interface CreateRoleRequest {
  name: string;
  description?: string;
  category?: string;
  companyId?: string;
  systemPrompt: string;
  welcomeMessage?: string;
  modelConfig?: Record<string, any>;
}

export interface InstallRoleRequest {
  targetType?: 'personal' | 'company';
  companyId?: string;
  name?: string;
}

// 角色 API
export const roleApi = {
  // 获取角色列表
  list: async (params?: {
    category?: string;
    template?: boolean;
    companyId?: string;
    public?: boolean;
  }): Promise<Role[]> => {
    try {
      const response = await client.get<ApiResponse<Role[]>>('/roles', { params });
      return response.data.data;
    } catch (error) {
      throw handleApiError(error);
    }
  },

  // 获取内置模板
  getTemplates: async (): Promise<Role[]> => {
    try {
      const response = await client.get<ApiResponse<Role[]>>('/roles/templates');
      return response.data.data;
    } catch (error) {
      throw handleApiError(error);
    }
  },

  // 获取单个角色
  get: async (id: string): Promise<Role> => {
    try {
      const response = await client.get<ApiResponse<Role>>(`/roles/${id}`);
      return response.data.data;
    } catch (error) {
      throw handleApiError(error);
    }
  },

  // 创建角色
  create: async (data: CreateRoleRequest): Promise<Role> => {
    try {
      const response = await client.post<ApiResponse<Role>>('/roles', data);
      return response.data.data;
    } catch (error) {
      throw handleApiError(error);
    }
  },

  // 从角色市场安装角色到个人或公司
  installTemplate: async (
    templateId: string,
    data: InstallRoleRequest = {}
  ): Promise<Role> => {
    try {
      const response = await client.post<ApiResponse<{ role: Role }>>(
        `/roles/templates/${templateId}/install`,
        data
      );
      return response.data.data.role;
    } catch (error) {
      throw handleApiError(error);
    }
  },

  // 更新角色
  update: async (id: string, data: Partial<CreateRoleRequest>): Promise<Role> => {
    try {
      const response = await client.put<ApiResponse<Role>>(`/roles/${id}`, data);
      return response.data.data;
    } catch (error) {
      throw handleApiError(error);
    }
  },

  // 删除角色
  delete: async (id: string): Promise<void> => {
    try {
      await client.delete(`/roles/${id}`);
    } catch (error) {
      throw handleApiError(error);
    }
  },

  // 与角色对话
  chat: async (roleId: string, message: string): Promise<{
    role: string;
    message: string;
    reply: string;
  }> => {
    try {
      const response = await client.post<ApiResponse<{
        role: string;
        message: string;
        reply: string;
      }>>(`/roles/${roleId}/chat`, { message });
      return response.data.data;
    } catch (error) {
      throw handleApiError(error);
    }
  },
};

export default roleApi;
