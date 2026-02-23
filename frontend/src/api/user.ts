import client, { ApiResponse, handleApiError } from './client';

// 用户类型
export interface User {
  id: string;
  email: string;
  name: string;
  avatar?: string;
  createdAt: string;
}

// 用户 API
export const userApi = {
  // 获取当前用户信息
  getMe: async (): Promise<User> => {
    try {
      const response = await client.get<ApiResponse<User>>('/users/me');
      return response.data.data;
    } catch (error) {
      throw handleApiError(error);
    }
  },

  // 更新用户信息
  updateMe: async (data: {
    name?: string;
    avatar?: string;
    password?: string;
  }): Promise<User> => {
    try {
      const response = await client.put<ApiResponse<User>>('/users/me', data);
      return response.data.data;
    } catch (error) {
      throw handleApiError(error);
    }
  },
};

export default userApi;
