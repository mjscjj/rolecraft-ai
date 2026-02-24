// 用户类型
export interface User {
  id: string;
  email: string;
  name: string;
  avatar?: string;
  createdAt?: string;
}

// 角色类型
export interface Role {
  id: string;
  name: string;
  description?: string;
  avatar?: string;
  category?: string;
  systemPrompt?: string;
  welcomeMessage?: string;
  modelConfig?: Record<string, any>;
  isTemplate?: boolean;
  createdAt?: string;
  updatedAt?: string;
}

// 创建角色请求
export interface CreateRoleRequest {
  name: string;
  description?: string;
  category?: string;
  systemPrompt: string;
  welcomeMessage?: string;
  modelConfig?: Record<string, any>;
}

// 文档类型
export interface Document {
  id: string;
  name: string;
  fileType: string;
  fileSize: number;
  status: 'pending' | 'processing' | 'completed' | 'failed';
  createdAt?: string;
}

// 对话会话
export interface ChatSession {
  id: string;
  roleId: string;
  title: string;
  mode: 'quick' | 'task';
  createdAt: string;
  updatedAt: string;
}

// 消息
export interface Message {
  id: string;
  sessionId: string;
  role: 'user' | 'assistant' | 'system';
  content: string;
  createdAt: string;
}

// 技能
export interface Skill {
  id: string;
  name: string;
  icon?: string;
  description?: string;
}

// API 响应
export interface ApiResponse<T = any> {
  code: number;
  message: string;
  data: T;
}