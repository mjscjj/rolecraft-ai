import type { ApiResponse } from './client';
import client, { handleApiError } from './client';

// 会话类型
export interface ChatSession {
  id: string;
  roleId: string;
  title: string;
  mode: 'quick' | 'task';
  createdAt: string;
  updatedAt: string;
}

// 消息类型
export interface Message {
  id: string;
  sessionId: string;
  role: 'user' | 'assistant' | 'system';
  content: string;
  createdAt: string;
}

// 对话 API
export const chatApi = {
  // 获取会话列表
  listSessions: async (): Promise<ChatSession[]> => {
    try {
      const response = await client.get<ApiResponse<ChatSession[]>>('/chat-sessions');
      return response.data.data;
    } catch (error) {
      throw handleApiError(error);
    }
  },

  // 创建新会话
  createSession: async (data: {
    roleId: string;
    title?: string;
    mode?: 'quick' | 'task';
  }): Promise<ChatSession> => {
    try {
      const response = await client.post<ApiResponse<ChatSession>>('/chat-sessions', data);
      return response.data.data;
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
      return response.data.data;
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
    onDone: () => void
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
    });

    if (!response.ok) {
      throw new Error(`HTTP ${response.status}`);
    }

    const reader = response.body?.getReader();
    if (!reader) return;

    const decoder = new TextDecoder();

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
            onDone();
          }
        } catch {
          // 忽略解析错误
        }
      }
    }
  },
};

export default chatApi;
