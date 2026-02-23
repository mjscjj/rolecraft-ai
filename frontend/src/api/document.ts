import client, { ApiResponse, handleApiError } from './client';

// 文档类型
export interface Document {
  id: string;
  name: string;
  fileType: string;
  fileSize: number;
  filePath: string;
  status: 'pending' | 'processing' | 'completed' | 'failed';
  chunkCount?: number;
  createdAt: string;
}

// 文档 API
export const documentApi = {
  // 获取文档列表
  list: async (params?: {
    status?: string;
    type?: string;
  }): Promise<Document[]> => {
    try {
      const response = await client.get<ApiResponse<Document[]>>('/documents', { params });
      return response.data.data;
    } catch (error) {
      throw handleApiError(error);
    }
  },

  // 上传文档
  upload: async (file: File): Promise<Document> => {
    try {
      const formData = new FormData();
      formData.append('file', file);
      
      const response = await client.post<ApiResponse<Document>>('/documents', formData, {
        headers: {
          'Content-Type': 'multipart/form-data',
        },
      });
      return response.data.data;
    } catch (error) {
      throw handleApiError(error);
    }
  },

  // 获取文档详情
  get: async (id: string): Promise<Document> => {
    try {
      const response = await client.get<ApiResponse<Document>>(`/documents/${id}`);
      return response.data.data;
    } catch (error) {
      throw handleApiError(error);
    }
  },

  // 删除文档
  delete: async (id: string): Promise<void> => {
    try {
      await client.delete(`/documents/${id}`);
    } catch (error) {
      throw handleApiError(error);
    }
  },
};

export default documentApi;
