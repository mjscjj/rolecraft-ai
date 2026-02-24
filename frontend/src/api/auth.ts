import type { ApiResponse } from './client';
import client, { handleApiError } from './client';

// 用户类型
export interface User {
  id: string;
  email: string;
  name: string;
  avatar?: string;
  createdAt: string;
}

// 登录响应
export interface LoginResponse {
  user: User;
  token: string;
}

// 认证 API
export const authApi = {
  // 注册
  register: async (data: {
    email: string;
    password: string;
    name: string;
  }): Promise<LoginResponse> => {
    try {
      const response = await client.post<ApiResponse<LoginResponse>>('/auth/register', data);
      return response.data.data;
    } catch (error) {
      throw handleApiError(error);
    }
  },

  // 登录
  login: async (data: {
    email: string;
    password: string;
  }): Promise<LoginResponse> => {
    try {
      const response = await client.post<ApiResponse<LoginResponse>>('/auth/login', data);
      return response.data.data;
    } catch (error) {
      throw handleApiError(error);
    }
  },

  // 刷新 Token
  refresh: async (): Promise<{ token: string }> => {
    try {
      const response = await client.post<ApiResponse<{ token: string }>>('/auth/refresh');
      return response.data.data;
    } catch (error) {
      throw handleApiError(error);
    }
  },
};

export default authApi;
