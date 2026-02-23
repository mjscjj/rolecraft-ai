export interface User {
  id: string;
  email: string;
  name: string;
  avatar?: string;
}

export interface Workspace {
  id: string;
  name: string;
  type: 'personal' | 'team' | 'enterprise';
  ownerId: string;
}

export interface Role {
  id: string;
  name: string;
  avatar?: string;
  description: string;
  category: string;
  systemPrompt: string;
  welcomeMessage?: string;
  isTemplate?: boolean;
  skills?: Skill[];
}

export interface Skill {
  id: string;
  name: string;
  description: string;
}

export interface Document {
  id: string;
  name: string;
  fileType: string;
  fileSize: number;
  status: 'pending' | 'processing' | 'completed' | 'failed';
  createdAt: string;
}

export interface ChatSession {
  id: string;
  roleId: string;
  title: string;
  mode: 'quick' | 'task';
  updatedAt: string;
}

export interface Message {
  id: string;
  role: 'user' | 'assistant' | 'system';
  content: string;
  sources?: string[];
  createdAt: string;
}