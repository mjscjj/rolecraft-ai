import { useAuthStore } from '../stores/authStore';
import { useRoleStore } from '../stores/roleStore';
import { useChatStore } from '../stores/chatStore';
import { roleApi } from '../api/role';
import { chatApi } from '../api/chat';

// 认证相关 Hook
export const useAuth = () => {
  const store = useAuthStore();
  
  return {
    user: store.user,
    isAuthenticated: store.isAuthenticated,
    isLoading: store.isLoading,
    error: store.error,
    login: store.login,
    register: store.register,
    logout: store.logout,
    clearError: store.clearError,
  };
};

// 角色相关 Hook
export const useRoles = () => {
  const store = useRoleStore();
  
  return {
    roles: store.roles,
    templates: store.templates,
    currentRole: store.currentRole,
    isLoading: store.isLoading,
    error: store.error,
    fetchRoles: store.fetchRoles,
    fetchTemplates: store.fetchTemplates,
    createRole: store.createRole,
    updateRole: store.updateRole,
    deleteRole: store.deleteRole,
    setCurrentRole: store.setCurrentRole,
    clearError: store.clearError,
  };
};

// 对话相关 Hook
export const useChat = () => {
  const store = useChatStore();
  
  return {
    sessions: store.sessions,
    currentSession: store.currentSession,
    messages: store.messages,
    isLoading: store.isLoading,
    isStreaming: store.isStreaming,
    error: store.error,
    fetchSessions: store.fetchSessions,
    createSession: store.createSession,
    fetchSession: store.fetchSession,
    sendMessage: store.sendMessage,
    sendStreamMessage: store.sendStreamMessage,
    clearMessages: store.clearMessages,
    clearError: store.clearError,
  };
};

// 角色对话 Hook (便捷组合)
export const useRoleChat = (_roleId?: string) => {
  const roleStore = useRoleStore();
  const chatStore = useChatStore();
  
  const startChat = async (id: string) => {
    const role = await roleApi.get(id);
    roleStore.setCurrentRole(role);
    const session = await chatApi.createSession({ roleId: id, mode: 'quick' });
    chatStore.fetchSession(session.id);
    return session;
  };
  
  return {
    role: roleStore.currentRole,
    messages: chatStore.messages,
    isStreaming: chatStore.isStreaming,
    startChat,
    sendMessage: chatStore.sendStreamMessage,
  };
};

export default {
  useAuth,
  useRoles,
  useChat,
  useRoleChat,
};
