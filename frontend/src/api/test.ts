import type { ApiResponse } from './client';
import client, { handleApiError } from './client';

// 测试消息请求
export interface TestMessageRequest {
  content: string;
  systemPrompt?: string;
  modelConfig?: Record<string, any>;
  roleName?: string;
}

// 测试消息响应
export interface TestMessageResponse {
  content: string;
  responseTime: number;
  tokens: number;
  model: string;
  metadata?: Record<string, any>;
}

// A/B 测试版本
export interface ABTestVersion {
  versionId: string;
  versionName: string;
  systemPrompt: string;
  modelConfig?: Record<string, any>;
}

// A/B 测试请求
export interface ABTestRequest {
  versions: ABTestVersion[];
  question: string;
}

// A/B 测试结果项
export interface ABTestResultItem {
  versionId: string;
  versionName: string;
  response: string;
  responseTime: number;
  score: number;
  rating: number;
  feedback: string;
}

// A/B 测试结果
export interface ABTestResult {
  testId: string;
  question: string;
  results: ABTestResultItem[];
  winnerId?: string;
  createdAt: string;
}

// 测试历史
export interface TestHistory {
  testId: string;
  roleId: string;
  roleName: string;
  testType: string;
  question: string;
  response: string;
  rating: number;
  feedback: string;
  createdAt: string;
}

// 测试报告
export interface TestReport {
  roleId: string;
  roleName: string;
  totalTests: number;
  averageRating: number;
  passRate: number;
  testsByRating: Record<number, number>;
  improvementTrend: Array<{
    date: string;
    avgRating: number;
    testCount: number;
  }>;
  suggestions: string[];
  exportUrl: string;
}

// 测试 API
export const testApi = {
  // 发送测试消息
  sendMessage: async (req: TestMessageRequest): Promise<TestMessageResponse> => {
    try {
      const response = await client.post<ApiResponse<TestMessageResponse>>(
        '/test/message',
        req
      );
      return response.data.data;
    } catch (error) {
      throw handleApiError(error);
    }
  },

  // 运行 A/B 测试
  runABTest: async (req: ABTestRequest): Promise<ABTestResult> => {
    try {
      const response = await client.post<ApiResponse<ABTestResult>>(
        '/test/ab',
        req
      );
      return response.data.data;
    } catch (error) {
      throw handleApiError(error);
    }
  },

  // 对比多个版本
  compareVersions: async (versionIds: string[], question: string): Promise<{
    question: string;
    results: ABTestResultItem[];
  }> => {
    try {
      const response = await client.post<ApiResponse<{
        question: string;
        results: ABTestResultItem[];
      }>>(
        '/test/compare',
        { versionIds, question }
      );
      return response.data.data;
    } catch (error) {
      throw handleApiError(error);
    }
  },

  // 保存测试结果
  saveTestResult: async (data: {
    roleId: string;
    roleName: string;
    testType: string;
    question: string;
    response: string;
    rating: number;
    feedback: string;
  }): Promise<{ saved: boolean; testId: string }> => {
    try {
      const response = await client.post<ApiResponse<{
        saved: boolean;
        testId: string;
      }>>(
        '/test/save',
        data
      );
      return response.data.data;
    } catch (error) {
      throw handleApiError(error);
    }
  },

  // 获取测试历史
  getTestHistory: async (roleId?: string): Promise<{
    history: TestHistory[];
    total: number;
  }> => {
    try {
      const response = await client.get<ApiResponse<{
        history: TestHistory[];
        total: number;
      }>>('/test/history', {
        params: { roleId },
      });
      return response.data.data;
    } catch (error) {
      throw handleApiError(error);
    }
  },

  // 获取测试报告
  getTestReport: async (roleId: string): Promise<TestReport> => {
    try {
      const response = await client.get<ApiResponse<TestReport>>(
        '/test/report',
        {
          params: { roleId },
        }
      );
      return response.data.data;
    } catch (error) {
      throw handleApiError(error);
    }
  },

  // 导出测试报告
  exportTestReport: async (roleId: string, format: string = 'pdf'): Promise<Blob> => {
    try {
      const response = await client.get(`/test/export/${roleId}`, {
        params: { format },
        responseType: 'blob',
      });
      return response.data;
    } catch (error) {
      throw handleApiError(error);
    }
  },

  // 评分测试回复
  rateTestResponse: async (data: {
    testId: string;
    rating: number;
    feedback?: string;
  }): Promise<{ rated: boolean; testId: string; rating: number }> => {
    try {
      const response = await client.post<ApiResponse<{
        rated: boolean;
        testId: string;
        rating: number;
      }>>(
        '/test/rate',
        data
      );
      return response.data.data;
    } catch (error) {
      throw handleApiError(error);
    }
  },
};

export default testApi;
