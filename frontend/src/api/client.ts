import type { AxiosInstance, AxiosError } from 'axios';
import axios from 'axios';

// API 配置
const rawApiURL = String(import.meta.env.VITE_API_URL || 'http://localhost:8080').replace(/\/+$/, '');
const API_ROOT = rawApiURL.endsWith('/api/v1') ? rawApiURL.slice(0, -7) : rawApiURL;
export const API_BASE_URL = `${API_ROOT}/api/v1`;

// 创建 Axios 实例
const client: AxiosInstance = axios.create({
  baseURL: API_BASE_URL,
  timeout: 30000,
  headers: {
    'Content-Type': 'application/json',
  },
});

// 请求拦截器 - 添加 JWT Token
client.interceptors.request.use(
  (config) => {
    const token = localStorage.getItem('token');
    if (token) {
      config.headers.Authorization = `Bearer ${token}`;
    }
    return config;
  },
  (error) => Promise.reject(error)
);

// 响应拦截器 - 处理错误和 Token 刷新
client.interceptors.response.use(
  (response) => response,
  async (error: AxiosError) => {
    const originalRequest = error.config;
    
    // 401 错误且不是刷新 Token 请求
    if (error.response?.status === 401 && originalRequest) {
      // 清除本地存储
      localStorage.removeItem('token');
      localStorage.removeItem('user');
      
      // 跳转到登录页
      window.location.href = '/login';
    }
    
    return Promise.reject(error);
  }
);

// 统一响应类型
export interface ApiResponse<T = any> {
  code: number;
  message: string;
  data: T;
}

// 错误处理
export interface ApiError {
  message: string;
  status: number;
}

export const handleApiError = (error: unknown): ApiError => {
  if (axios.isAxiosError(error)) {
    return {
      message: error.response?.data?.error || error.message,
      status: error.response?.status || 500,
    };
  }
  return {
    message: 'An unexpected error occurred',
    status: 500,
  };
};

export default client;
