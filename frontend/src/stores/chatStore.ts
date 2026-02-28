import { create } from 'zustand';
import type { ChatSession, Message } from '../api/chat';
import { chatApi } from '../api/chat';

interface ChatState {
  sessions: ChatSession[];
  currentSession: ChatSession | null;
  messages: Message[];
  isLoading: boolean;
  isStreaming: boolean;
  thinkingSteps: string[];
  lastStreamContent: string | null;
  lastStreamMode: 'normal' | 'deep';
  streamController: AbortController | null;
  error: string | null;

  // Actions
  fetchSessions: () => Promise<void>;
  createSession: (
    roleId: string,
    title?: string,
    mode?: 'quick' | 'task',
    modelConfig?: Record<string, any>
  ) => Promise<ChatSession>;
  fetchSession: (id: string) => Promise<void>;
  sendMessage: (content: string) => Promise<void>;
  sendStreamMessage: (content: string) => Promise<void>;
  sendStreamMessageWithThinking: (content: string) => Promise<void>;
  updateSessionConfig: (modelConfig: Record<string, any>) => Promise<void>;
  cancelStream: () => void;
  retryLastStream: () => Promise<void>;
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
  thinkingSteps: [],
  lastStreamContent: null,
  lastStreamMode: 'normal',
  error: null,
  streamController: null as AbortController | null,

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

  createSession: async (roleId, title, mode = 'quick', modelConfig) => {
    set({ isLoading: true, error: null });
    try {
      const session = await chatApi.createSession({ roleId, title, mode, modelConfig });
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
      set((state) => ({
        currentSession: session,
        messages,
        sessions: state.sessions.some((item) => item.id === session.id)
          ? state.sessions.map((item) => (item.id === session.id ? session : item))
          : [session, ...state.sessions],
        isLoading: false,
      }));
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

    const controller = new AbortController();
    set({
      isStreaming: true,
      error: null,
      lastStreamContent: content,
      lastStreamMode: 'normal',
      thinkingSteps: [],
      streamController: controller,
    });

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
        (assistantMessageId) => {
          set((state) => {
            const next = [...state.messages];
            const last = next[next.length - 1];
            if (
              assistantMessageId &&
              last &&
              last.role === 'assistant' &&
              last.id.startsWith('ai-')
            ) {
              last.id = assistantMessageId;
            }
            return { messages: next, isStreaming: false, streamController: null };
          });
        },
        controller.signal
      );
    } catch (error: any) {
      if (error?.name === 'AbortError') {
        set((state) => {
          const next = [...state.messages];
          const last = next[next.length - 1];
          if (last && last.role === 'assistant' && !last.content.trim()) {
            last.content = '已停止生成。';
          }
          return { messages: next, isStreaming: false, streamController: null, thinkingSteps: [] };
        });
        return;
      }
      set((state) => {
        const next = [...state.messages];
        const last = next[next.length - 1];
        if (last && last.role === 'assistant' && !last.content.trim()) {
          last.content = '生成失败，请重试。';
        }
        return {
          messages: next,
          error: error.message || 'Failed to send stream message',
          isStreaming: false,
          streamController: null,
          thinkingSteps: [],
        };
      });
    }
  },

  sendStreamMessageWithThinking: async (content) => {
    const { currentSession, messages } = get();
    if (!currentSession) return;

    const controller = new AbortController();
    set({
      isStreaming: true,
      error: null,
      lastStreamContent: content,
      lastStreamMode: 'deep',
      thinkingSteps: [],
      streamController: controller,
    });

    const userMsg: Message = {
      id: `user-${Date.now()}`,
      sessionId: currentSession.id,
      role: 'user',
      content,
      createdAt: new Date().toISOString(),
    };

    const aiMsg: Message = {
      id: `ai-${Date.now()}`,
      sessionId: currentSession.id,
      role: 'assistant',
      content: '',
      createdAt: new Date().toISOString(),
    };

    set({ messages: [...messages, userMsg, aiMsg] });

    try {
      await chatApi.streamMessageWithThinking(
        currentSession.id,
        content,
        {
          onThinking: (stepContent) => {
            set((state) => ({ thinkingSteps: [...state.thinkingSteps, stepContent] }));
          },
          onChunk: (chunk) => {
            set((state) => {
              const newMessages = [...state.messages];
              const lastMsg = newMessages[newMessages.length - 1];
              if (lastMsg.role === 'assistant') {
                lastMsg.content += chunk;
              }
              return { messages: newMessages };
            });
          },
          onDone: (assistantMessageId) => {
            set((state) => {
              const next = [...state.messages];
              const last = next[next.length - 1];
              if (
                assistantMessageId &&
                last &&
                last.role === 'assistant' &&
                last.id.startsWith('ai-')
              ) {
                last.id = assistantMessageId;
              }
              return { messages: next, isStreaming: false, streamController: null };
            });
          },
        },
        controller.signal
      );
    } catch (error: any) {
      if (error?.name === 'AbortError') {
        set((state) => {
          const next = [...state.messages];
          const last = next[next.length - 1];
          if (last && last.role === 'assistant' && !last.content.trim()) {
            last.content = '已停止生成。';
          }
          return { messages: next, isStreaming: false, streamController: null };
        });
        return;
      }
      set((state) => {
        const next = [...state.messages];
        const last = next[next.length - 1];
        if (last && last.role === 'assistant' && !last.content.trim()) {
          last.content = '生成失败，请重试。';
        }
        return {
          messages: next,
          error: error.message || 'Failed to send deep thinking stream message',
          isStreaming: false,
          streamController: null,
        };
      });
    }
  },

  updateSessionConfig: async (modelConfig) => {
    const { currentSession } = get();
    if (!currentSession) return;

    try {
      await chatApi.updateSessionConfig(currentSession.id, modelConfig);
      set((state) => ({
        currentSession: state.currentSession
          ? { ...state.currentSession, modelConfig }
          : state.currentSession,
        sessions: state.sessions.map((session) =>
          session.id === currentSession.id ? { ...session, modelConfig } : session
        ),
      }));
    } catch (error: any) {
      set({
        error: error.message || 'Failed to update session config',
      });
      throw error;
    }
  },

  cancelStream: () => {
    const { streamController } = get();
    if (streamController) {
      streamController.abort();
    }
    set({ isStreaming: false, streamController: null });
  },

  retryLastStream: async () => {
    const { lastStreamContent, lastStreamMode } = get();
    if (!lastStreamContent) return;
    if (lastStreamMode === 'deep') {
      await get().sendStreamMessageWithThinking(lastStreamContent);
      return;
    }
    await get().sendStreamMessage(lastStreamContent);
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
