import type { ApiResponse } from './client';
import client, { handleApiError } from './client';

// 会话类型
export interface ChatSession {
  id: string;
  roleId: string;
  title: string;
  mode: 'quick' | 'task';
  modelConfig?: Record<string, any>;
  createdAt: string;
  updatedAt: string;
}

const parseModelConfig = (value: unknown): Record<string, any> | undefined => {
  if (!value) return undefined;
  if (typeof value === 'string') {
    try {
      const parsed = JSON.parse(value);
      if (parsed && typeof parsed === 'object') return parsed as Record<string, any>;
    } catch {
      return undefined;
    }
    return undefined;
  }
  if (typeof value === 'object') {
    return value as Record<string, any>;
  }
  return undefined;
};

const normalizeSession = (session: any): ChatSession => ({
  ...session,
  modelConfig: parseModelConfig(session?.modelConfig),
});

// 消息类型
export interface Message {
  id: string;
  sessionId: string;
  role: 'user' | 'assistant' | 'system';
  content: string;
  createdAt: string;
}

interface StreamWithThinkingHandlers {
  onThinking?: (content: string) => void;
  onChunk: (chunk: string) => void;
  onDone: () => void;
}

// 对话 API
export const chatApi = {
  // 获取会话列表
  listSessions: async (): Promise<ChatSession[]> => {
    try {
      const response = await client.get<ApiResponse<ChatSession[]>>('/chat-sessions');
      return (response.data.data || []).map(normalizeSession);
    } catch (error) {
      throw handleApiError(error);
    }
  },

  // 创建新会话
  createSession: async (data: {
    roleId: string;
    title?: string;
    mode?: 'quick' | 'task';
    modelConfig?: Record<string, any>;
  }): Promise<ChatSession> => {
    try {
      const response = await client.post<ApiResponse<ChatSession>>('/chat-sessions', data);
      return normalizeSession(response.data.data);
    } catch (error) {
      throw handleApiError(error);
    }
  },

  // 获取会话详情（包含消息历史）
  getSession: async (id: string): Promise<{
    session: ChatSession;
    messages: Message[];
  }> => {
    try {
      const response = await client.get<ApiResponse<{
        session: ChatSession;
        messages: Message[];
      }>>(`/chat-sessions/${id}`);
      return {
        ...response.data.data,
        session: normalizeSession(response.data.data.session),
      };
    } catch (error) {
      throw handleApiError(error);
    }
  },

  // 发送消息（普通响应）
  sendMessage: async (sessionId: string, content: string): Promise<{
    userMessage: Message;
    assistantMessage: Message;
  }> => {
    try {
      const response = await client.post<ApiResponse<{
        userMessage: Message;
        assistantMessage: Message;
      }>>(`/chat/${sessionId}/complete`, { content });
      return response.data.data;
    } catch (error) {
      throw handleApiError(error);
    }
  },

  // 发送消息（流式响应）
  streamMessage: async (
    sessionId: string,
    content: string,
    onChunk: (chunk: string) => void,
    onDone: () => void,
    signal?: AbortSignal
  ): Promise<void> => {
    const token = localStorage.getItem('token');
    const API_BASE_URL = import.meta.env.VITE_API_URL || 'http://localhost:8080';
    
    const response = await fetch(`${API_BASE_URL}/api/v1/chat/${sessionId}/stream`, {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
        'Authorization': `Bearer ${token}`,
      },
      body: JSON.stringify({ content }),
      signal,
    });

    if (!response.ok) {
      throw new Error(`HTTP ${response.status}`);
    }

    const reader = response.body?.getReader();
    if (!reader) return;

    const decoder = new TextDecoder();
    let finished = false;
    const doneOnce = () => {
      if (finished) {
        return;
      }
      finished = true;
      onDone();
    };

    while (true) {
      const { done, value } = await reader.read();
      if (done) break;

      const chunk = decoder.decode(value);
      const lines = chunk.split('\n').filter(line => line.startsWith('data: '));

      for (const line of lines) {
        const data = line.slice(6);
        try {
          const parsed = JSON.parse(data);
          if (parsed.content) {
            onChunk(parsed.content);
          }
          if (parsed.done) {
            doneOnce();
          }
        } catch {
          // 忽略解析错误
        }
      }
    }

    doneOnce();
  },

  // 发送消息（流式 + 深度思考过程）
  streamMessageWithThinking: async (
    sessionId: string,
    content: string,
    handlers: StreamWithThinkingHandlers,
    signal?: AbortSignal
  ): Promise<void> => {
    const token = localStorage.getItem('token');
    const API_BASE_URL = import.meta.env.VITE_API_URL || 'http://localhost:8080';

    const response = await fetch(`${API_BASE_URL}/api/v1/chat/${sessionId}/stream-with-thinking`, {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
        'Authorization': `Bearer ${token}`,
      },
      body: JSON.stringify({ content }),
      signal,
    });

    if (!response.ok) {
      throw new Error(`HTTP ${response.status}`);
    }

    const reader = response.body?.getReader();
    if (!reader) return;

    const decoder = new TextDecoder();
    let finished = false;
    const doneOnce = () => {
      if (finished) {
        return;
      }
      finished = true;
      handlers.onDone();
    };

    while (true) {
      const { done, value } = await reader.read();
      if (done) break;

      const chunk = decoder.decode(value);
      const lines = chunk.split('\n').filter((line) => line.startsWith('data: '));

      for (const line of lines) {
        const data = line.slice(6);
        try {
          const parsed = JSON.parse(data);

          if (parsed.type === 'error') {
            const message = parsed?.data?.message || '深度思考请求失败';
            throw new Error(message);
          }

          if (parsed.type === 'thinking' && parsed.step?.content) {
            handlers.onThinking?.(parsed.step.content);
            continue;
          }

          if (parsed.type === 'answer' && parsed.content) {
            handlers.onChunk(parsed.content);
            continue;
          }

          if (parsed.content) {
            handlers.onChunk(parsed.content);
          }

          if (parsed.done || parsed.type === 'done') {
            doneOnce();
          }
        } catch (err: any) {
          if (err instanceof Error) {
            throw err;
          }
          // 忽略单条解析错误，继续处理后续流
        }
      }
    }

    doneOnce();
  },

  // 切换角色
  switchRole: async (sessionId: string, roleId: string): Promise<{
    sessionId: string;
    oldRoleId: string;
    newRoleId: string;
    newRoleName: string;
  }> => {
    try {
      const response = await client.post<ApiResponse<{
        sessionId: string;
        oldRoleId: string;
        newRoleId: string;
        newRoleName: string;
      }>>(`/chat-sessions/${sessionId}/switch-role`, { roleId });
      return response.data.data;
    } catch (error) {
      throw handleApiError(error);
    }
  },

  // 更新会话标题
  updateTitle: async (sessionId: string, title: string): Promise<{
    sessionId: string;
    title: string;
  }> => {
    try {
      const response = await client.put<ApiResponse<{
        sessionId: string;
        title: string;
      }>>(`/chat-sessions/${sessionId}/title`, { title });
      return response.data.data;
    } catch (error) {
      throw handleApiError(error);
    }
  },

  // 更新会话配置
  updateSessionConfig: async (sessionId: string, modelConfig: Record<string, any>): Promise<{
    sessionId: string;
    modelConfig: Record<string, any>;
  }> => {
    try {
      const response = await client.put<ApiResponse<{
        sessionId: string;
        modelConfig: Record<string, any>;
      }>>(`/chat-sessions/${sessionId}/config`, { modelConfig });
      return response.data.data;
    } catch (error) {
      throw handleApiError(error);
    }
  },

  // 归档/取消归档会话
  archive: async (sessionId: string, isArchived: boolean): Promise<{
    sessionId: string;
    isArchived: boolean;
  }> => {
    try {
      const response = await client.post<ApiResponse<{
        sessionId: string;
        isArchived: boolean;
      }>>(`/chat-sessions/${sessionId}/archive`, { isArchived });
      return response.data.data;
    } catch (error) {
      throw handleApiError(error);
    }
  },

  // 导出会话
  export: async (sessionId: string, format: 'markdown' | 'json' | 'pdf'): Promise<Blob> => {
    const token = localStorage.getItem('token');
    const API_BASE_URL = import.meta.env.VITE_API_URL || 'http://localhost:8080';
    
    const response = await fetch(`${API_BASE_URL}/api/v1/chat-sessions/${sessionId}/export`, {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
        'Authorization': `Bearer ${token}`,
      },
      body: JSON.stringify({ format }),
    });

    if (!response.ok) {
      throw new Error(`HTTP ${response.status}`);
    }

    return await response.blob();
  },

  // 搜索会话
  search: async (query: string): Promise<ChatSession[]> => {
    try {
      const response = await client.post<ApiResponse<ChatSession[]>>('/chat-sessions/search', { query });
      return (response.data.data || []).map(normalizeSession);
    } catch (error) {
      throw handleApiError(error);
    }
  },

  // 编辑消息
  editMessage: async (sessionId: string, messageId: string, content: string): Promise<{
    messageId: string;
    content: string;
  }> => {
    try {
      const response = await client.put<ApiResponse<{
        messageId: string;
        content: string;
      }>>(`/chat/${sessionId}/messages/${messageId}`, { content });
      return response.data.data;
    } catch (error) {
      throw handleApiError(error);
    }
  },

  // 重新生成回复
  regenerate: async (sessionId: string, messageId: string): Promise<{
    assistantMessage: Message;
  }> => {
    try {
      const response = await client.post<ApiResponse<{
        assistantMessage: Message;
      }>>(`/chat/${sessionId}/messages/${messageId}/regenerate`);
      return response.data.data;
    } catch (error) {
      throw handleApiError(error);
    }
  },

  // 点赞/点踩
  rateMessage: async (messageId: string, rating: 'up' | 'down'): Promise<{
    messageId: string;
    rating: string;
  }> => {
    try {
      const response = await client.post<ApiResponse<{
        messageId: string;
        rating: string;
      }>>(`/chat/messages/${messageId}/rate`, { rating });
      return response.data.data;
    } catch (error) {
      throw handleApiError(error);
    }
  },

  // 删除会话
  deleteSession: async (sessionId: string): Promise<void> => {
    try {
      await client.delete(`/chat-sessions/${sessionId}`);
    } catch (error) {
      throw handleApiError(error);
    }
  },
};

export default chatApi;
