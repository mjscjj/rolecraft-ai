import { create } from 'zustand';
import type { ChatSession, Message } from '../api/chat';
import { chatApi } from '../api/chat';

interface ChatState {
  sessions: ChatSession[];
  currentSession: ChatSession | null;
  messages: Message[];
  isLoading: boolean;
  isStreaming: boolean;
  error: string | null;

  // Actions
  fetchSessions: () => Promise<void>;
  createSession: (roleId: string, title?: string, mode?: 'quick' | 'task') => Promise<ChatSession>;
  fetchSession: (id: string) => Promise<void>;
  sendMessage: (content: string) => Promise<void>;
  sendStreamMessage: (content: string) => Promise<void>;
  appendMessage: (message: Message) => void;
  updateLastMessage: (content: string) => void;
  clearMessages: () => void;
  clearError: () => void;
}

export const useChatStore = create<ChatState>((set, get) => ({
  sessions: [],
  currentSession: null,
  messages: [],
  isLoading: false,
  isStreaming: false,
  error: null,

  fetchSessions: async () => {
    set({ isLoading: true, error: null });
    try {
      const sessions = await chatApi.listSessions();
      set({ sessions, isLoading: false });
    } catch (error: any) {
      set({
        error: error.message || 'Failed to fetch sessions',
        isLoading: false,
      });
    }
  },

  createSession: async (roleId, title, mode = 'quick') => {
    set({ isLoading: true, error: null });
    try {
      const session = await chatApi.createSession({ roleId, title, mode });
      set((state) => ({
        sessions: [session, ...state.sessions],
        currentSession: session,
        messages: [],
        isLoading: false,
      }));
      return session;
    } catch (error: any) {
      set({
        error: error.message || 'Failed to create session',
        isLoading: false,
      });
      throw error;
    }
  },

  fetchSession: async (id) => {
    set({ isLoading: true, error: null });
    try {
      const { session, messages } = await chatApi.getSession(id);
      set({
        currentSession: session,
        messages,
        isLoading: false,
      });
    } catch (error: any) {
      set({
        error: error.message || 'Failed to fetch session',
        isLoading: false,
      });
    }
  },

  sendMessage: async (content) => {
    const { currentSession, messages } = get();
    if (!currentSession) return;

    set({ isLoading: true, error: null });

    // 立即添加用户消息
    const tempUserMsg: Message = {
      id: `temp-${Date.now()}`,
      sessionId: currentSession.id,
      role: 'user',
      content,
      createdAt: new Date().toISOString(),
    };
    set({ messages: [...messages, tempUserMsg] });

    try {
      const { assistantMessage } = await chatApi.sendMessage(currentSession.id, content);
      set((state) => ({
        messages: [...state.messages.slice(0, -1), tempUserMsg, assistantMessage],
        isLoading: false,
      }));
    } catch (error: any) {
      set({
        error: error.message || 'Failed to send message',
        isLoading: false,
      });
    }
  },

  sendStreamMessage: async (content) => {
    const { currentSession, messages } = get();
    if (!currentSession) return;

    set({ isStreaming: true, error: null });

    // 添加用户消息
    const userMsg: Message = {
      id: `user-${Date.now()}`,
      sessionId: currentSession.id,
      role: 'user',
      content,
      createdAt: new Date().toISOString(),
    };

    // 添加空的 AI 消息（用于流式填充）
    const aiMsg: Message = {
      id: `ai-${Date.now()}`,
      sessionId: currentSession.id,
      role: 'assistant',
      content: '',
      createdAt: new Date().toISOString(),
    };

    set({ messages: [...messages, userMsg, aiMsg] });

    try {
      await chatApi.streamMessage(
        currentSession.id,
        content,
        (chunk) => {
          // 追加内容
          set((state) => {
            const newMessages = [...state.messages];
            const lastMsg = newMessages[newMessages.length - 1];
            if (lastMsg.role === 'assistant') {
              lastMsg.content += chunk;
            }
            return { messages: newMessages };
          });
        },
        () => {
          set({ isStreaming: false });
        }
      );
    } catch (error: any) {
      set({
        error: error.message || 'Failed to send stream message',
        isStreaming: false,
      });
    }
  },

  appendMessage: (message) => {
    set((state) => ({ messages: [...state.messages, message] }));
  },

  updateLastMessage: (content) => {
    set((state) => {
      const newMessages = [...state.messages];
      const lastMsg = newMessages[newMessages.length - 1];
      if (lastMsg && lastMsg.role === 'assistant') {
        lastMsg.content = content;
      }
      return { messages: newMessages };
    });
  },

  clearMessages: () => set({ messages: [] }),
  clearError: () => set({ error: null }),
}));

export default useChatStore;
