import axios from 'axios';

const API_BASE = 'http://localhost:8080/api/v1';

// 获取认证 token
const getAuthHeaders = () => {
  const token = localStorage.getItem('token');
  return {
    headers: {
      Authorization: token ? `Bearer ${token}` : '',
    },
  };
};

export const analyticsApi = {
  // Dashboard 核心指标
  getDashboard: () => {
    return axios.get(`${API_BASE}/analytics/dashboard`, getAuthHeaders());
  },

  // 用户活跃度
  getUserActivity: () => {
    return axios.get(`${API_BASE}/analytics/user-activity`, getAuthHeaders());
  },

  // 功能使用率
  getFeatureUsage: () => {
    return axios.get(`${API_BASE}/analytics/feature-usage`, getAuthHeaders());
  },

  // 留存率
  getRetentionRate: () => {
    return axios.get(`${API_BASE}/analytics/retention`, getAuthHeaders());
  },

  // 流失风险用户
  getChurnRiskUsers: () => {
    return axios.get(`${API_BASE}/analytics/churn-risk`, getAuthHeaders());
  },

  // 对话质量
  getConversationQuality: () => {
    return axios.get(`${API_BASE}/analytics/conversation-quality`, getAuthHeaders());
  },

  // 回复质量
  getReplyQuality: () => {
    return axios.get(`${API_BASE}/analytics/reply-quality`, getAuthHeaders());
  },

  // 常见问题
  getFAQStats: () => {
    return axios.get(`${API_BASE}/analytics/faq`, getAuthHeaders());
  },

  // 敏感词检测
  getSensitiveWords: () => {
    return axios.get(`${API_BASE}/analytics/sensitive-words`, getAuthHeaders());
  },

  // 成本统计
  getCostStats: () => {
    return axios.get(`${API_BASE}/analytics/cost`, getAuthHeaders());
  },

  // 按角色分类成本
  getCostByRole: () => {
    return axios.get(`${API_BASE}/analytics/cost/by-role`, getAuthHeaders());
  },

  // 按用户分类成本
  getCostByUser: () => {
    return axios.get(`${API_BASE}/analytics/cost/by-user`, getAuthHeaders());
  },

  // 成本趋势
  getCostTrend: (days: number = 30) => {
    return axios.get(`${API_BASE}/analytics/cost/trend?days=${days}`, getAuthHeaders());
  },

  // 成本预测
  getCostPrediction: (period: 'week' | 'month' = 'month') => {
    return axios.get(`${API_BASE}/analytics/cost/prediction?period=${period}`, getAuthHeaders());
  },

  // 生成报告
  getReport: (type: 'weekly' | 'monthly' = 'weekly') => {
    return axios.get(`${API_BASE}/analytics/report?type=${type}`, getAuthHeaders());
  },

  // 导出报告
  exportReport: (type: 'weekly' | 'monthly' = 'weekly') => {
    return axios.get(`${API_BASE}/analytics/report/export?type=${type}`, {
      ...getAuthHeaders(),
      responseType: 'blob',
    });
  },
};
