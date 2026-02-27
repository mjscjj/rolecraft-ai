import type { ApiResponse } from './client';
import client, { handleApiError } from './client';

// 提示词版本
export interface PromptVersion {
  id: string;
  content: string;
  score: number;
  features: string[];
  scenarios: string[];
  isRecommended?: boolean;
}

// 优化建议
export interface OptimizationSuggestion {
  type: 'specificity' | 'example' | 'tone' | 'completeness';
  message: string;
  suggestion: string;
}

// 优化结果
export interface OptimizationResult {
  versions: PromptVersion[];
  suggestions: OptimizationSuggestion[];
  originalLength: number;
  optimizedLength: number;
  improvementScore: number;
}

// 提示词优化 API
export const promptApi = {
  // 优化提示词
  optimize: async (
    prompt: string,
    generateVersions: number = 3,
    includeSuggestions: boolean = true
  ): Promise<OptimizationResult> => {
    try {
      const response = await client.post<ApiResponse<OptimizationResult>>('/prompt/optimize', {
        prompt,
        generateVersions,
        includeSuggestions,
      });
      return response.data.data;
    } catch (error) {
      throw handleApiError(error);
    }
  },

  // 获取实时建议
  getSuggestions: async (prompt: string): Promise<OptimizationSuggestion[]> => {
    try {
      const response = await client.post<ApiResponse<OptimizationSuggestion[]>>('/prompt/suggestions', {
        prompt,
      });
      return response.data.data;
    } catch (error) {
      throw handleApiError(error);
    }
  },

  // 记录用户选择（用于学习机制）
  logSelection: async (
    originalPrompt: string,
    selectedVersion: string,
    userID: string,
    rating: number = 5
  ): Promise<void> => {
    try {
      await client.post('/prompt/log', {
        originalPrompt,
        selectedVersion,
        userID,
        rating,
      });
    } catch (error) {
      throw handleApiError(error);
    }
  },
};

export default promptApi;
