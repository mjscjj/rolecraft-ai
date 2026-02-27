import type { ApiResponse } from './client';
import client, { handleApiError } from './client';

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

export interface Folder {
  id: string;
  name: string;
  parentId?: string;
  documentCount?: number;
  createdAt: string;
}

export interface DocumentSearchResult {
  query: string;
  documents: Document[];
  total: number;
  searchTimeMs: number;
  vectorResults: number;
}

// 文档状态类型
export interface DocumentStatus {
  id: string;
  status: 'pending' | 'processing' | 'completed' | 'failed';
  progress: number; // 0-100, -1 for failed
  message: string;
  updatedAt: string;
}

// 文档 API
export const documentApi = {
  // 获取文档列表
  list: async (params?: {
    status?: string;
    type?: string;
    folder?: string;
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

  uploadWithFolder: async (file: File, folderId?: string): Promise<Document | Document[]> => {
    try {
      const formData = new FormData();
      formData.append('file', file);
      if (folderId) formData.append('folderId', folderId);

      const response = await client.post<ApiResponse<Document | Document[]>>('/documents', formData, {
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

  // 获取文档处理状态
  getStatus: async (id: string): Promise<DocumentStatus> => {
    try {
      const response = await client.get<ApiResponse<DocumentStatus>>(`/documents/${id}/status`);
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

  update: async (
    id: string,
    payload: { name?: string; description?: string; tags?: string[]; folderId?: string }
  ): Promise<Document> => {
    try {
      const response = await client.put<ApiResponse<Document>>(`/documents/${id}`, payload);
      return response.data.data;
    } catch (error) {
      throw handleApiError(error);
    }
  },

  search: async (payload: {
    query: string;
    topN?: number;
    filters?: Record<string, string>;
    sortBy?: string;
    sortOrder?: 'asc' | 'desc';
  }): Promise<DocumentSearchResult> => {
    try {
      const response = await client.post<ApiResponse<DocumentSearchResult>>('/documents/search', payload);
      return response.data.data;
    } catch (error) {
      throw handleApiError(error);
    }
  },

  batchDelete: async (ids: string[]): Promise<void> => {
    try {
      await client.delete('/documents/batch', { data: { ids } });
    } catch (error) {
      throw handleApiError(error);
    }
  },

  batchMove: async (ids: string[], folderId?: string): Promise<void> => {
    try {
      await client.put('/documents/batch/move', { ids, folderId: folderId || '' });
    } catch (error) {
      throw handleApiError(error);
    }
  },

  batchUpdateTags: async (ids: string[], tags: string[]): Promise<void> => {
    try {
      await client.put('/documents/batch/tags', { ids, tags });
    } catch (error) {
      throw handleApiError(error);
    }
  },

  listFolders: async (): Promise<Folder[]> => {
    try {
      const response = await client.get<ApiResponse<Folder[]>>('/folders');
      return response.data.data;
    } catch (error) {
      throw handleApiError(error);
    }
  },

  createFolder: async (name: string, parentId?: string): Promise<Folder> => {
    try {
      const response = await client.post<ApiResponse<Folder>>('/folders', { name, parentId: parentId || '' });
      return response.data.data;
    } catch (error) {
      throw handleApiError(error);
    }
  },

  deleteFolder: async (id: string): Promise<void> => {
    try {
      await client.delete(`/folders/${id}`);
    } catch (error) {
      throw handleApiError(error);
    }
  },

  preview: async (id: string, fileType: string): Promise<{ content?: string; blob?: Blob }> => {
    try {
      if (fileType === 'pdf') {
        const response = await client.get(`/documents/${id}/preview`, { responseType: 'blob' });
        return { blob: response.data as Blob };
      }
      const response = await client.get<ApiResponse<{ content: string }>>(`/documents/${id}/preview`);
      return { content: response.data.data?.content || '' };
    } catch (error) {
      throw handleApiError(error);
    }
  },

  download: async (id: string): Promise<Blob> => {
    try {
      const response = await client.get(`/documents/${id}/download`, { responseType: 'blob' });
      return response.data as Blob;
    } catch (error) {
      throw handleApiError(error);
    }
  },
};

export default documentApi;
