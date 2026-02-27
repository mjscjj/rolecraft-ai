import { useState, useEffect } from 'react';
import { useNavigate } from 'react-router-dom';
import { Bot, Plus, MessageSquare, Sparkles } from 'lucide-react';
import { RoleCard } from '../components/RoleCard';
import type { Role } from '../types';
import client from '../api/client';

export const Dashboard = () => {
  const navigate = useNavigate();
  const [roles, setRoles] = useState<Role[]>([]);
  const [loading, setLoading] = useState(true);
  const [stats, setStats] = useState({ totalRoles: 0, totalChats: 0 });

  useEffect(() => {
    loadRoles();
  }, []);

  const loadRoles = async () => {
    try {
      const token = localStorage.getItem('token');
      if (!token) {
        navigate('/login');
        return;
      }

      const response = await client.get('/roles');
      if (response.data.code === 200 || response.data.code === 0) {
        setRoles(response.data.data || []);
        setStats({
          totalRoles: response.data.data?.length || 0,
          totalChats: 0
        });
      }
    } catch (error) {
      console.error('加载角色失败:', error);
    } finally {
      setLoading(false);
    }
  };

  const templateRoles: Role[] = [
    { id: 't1', name: '🤖 智能助理', description: '全能型办公助手，处理日常事务', category: '通用', systemPrompt: '你是一个专业、友好的 AI 助理，帮助用户处理各种任务。', welcomeMessage: '你好！我是你的智能助理，有什么可以帮你的？', isTemplate: true },
    { id: 't2', name: '✍️ 文案专家', description: '专业文案创作，营销内容撰写', category: '营销', systemPrompt: '你是一名专业的文案专家，擅长创作吸引人的营销文案和内容。', welcomeMessage: '你好！让我帮你创作精彩的文案！', isTemplate: true },
    { id: 't3', name: '💻 编程助手', description: '代码编写、调试和技术咨询', category: '技术', systemPrompt: '你是一名经验丰富的程序员，帮助用户编写代码、调试问题和解答技术疑问。', welcomeMessage: '你好！有什么编程问题我可以帮你？', isTemplate: true },
    { id: 't4', name: '📚 学习导师', description: '知识讲解、学习规划和答疑', category: '教育', systemPrompt: '你是一名耐心的老师，帮助学生理解知识、制定学习计划和解答疑问。', welcomeMessage: '你好！今天想学习什么？', isTemplate: true },
  ];

  const handleUseTemplate = async (template: Role) => {
    try {
      const response = await client.post('/roles', {
        name: template.name,
        description: template.description,
        category: template.category,
        systemPrompt: template.systemPrompt,
        welcomeMessage: template.welcomeMessage,
      });

      if (response.data.code === 200 || response.data.code === 0) {
        const newRole = response.data.data;
        navigate(`/chat/${newRole.id}`);
      }
    } catch (error) {
      console.error('创建角色失败:', error);
      alert('创建失败，请重试');
    }
  };

  const handleCreateRole = () => {
    navigate('/roles/create');
  };

  if (loading) {
    return (
      <div className="flex items-center justify-center h-64">
        <div className="text-center">
          <div className="w-12 h-12 border-4 border-primary border-t-transparent rounded-full animate-spin mx-auto mb-4"></div>
          <p className="text-slate-500">加载中...</p>
        </div>
      </div>
    );
  }

  return (
    <div className="space-y-8">
      {/* Header */}
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-3xl font-bold text-slate-900">我的 AI 团队</h1>
          <p className="text-slate-500 mt-1">管理你的 AI 角色，开始对话</p>
        </div>
        <button
          onClick={handleCreateRole}
          className="flex items-center gap-2 px-6 py-3 bg-gradient-to-r from-indigo-600 to-purple-600 text-white rounded-xl hover:shadow-lg transition-all"
        >
          <Plus className="w-5 h-5" />
          创建角色
        </button>
      </div>

      {/* Stats */}
      <div className="grid grid-cols-1 md:grid-cols-3 gap-6">
        <div className="bg-white p-6 rounded-xl shadow-sm border border-slate-100">
          <div className="flex items-center gap-4">
            <div className="w-12 h-12 bg-blue-100 rounded-xl flex items-center justify-center">
              <Bot className="w-6 h-6 text-blue-600" />
            </div>
            <div>
              <p className="text-sm text-slate-500">AI 角色</p>
              <p className="text-2xl font-bold">{stats.totalRoles}</p>
            </div>
          </div>
        </div>
        <div className="bg-white p-6 rounded-xl shadow-sm border border-slate-100">
          <div className="flex items-center gap-4">
            <div className="w-12 h-12 bg-green-100 rounded-xl flex items-center justify-center">
              <MessageSquare className="w-6 h-6 text-green-600" />
            </div>
            <div>
              <p className="text-sm text-slate-500">对话次数</p>
              <p className="text-2xl font-bold">{stats.totalChats}</p>
            </div>
          </div>
        </div>
        <div className="bg-white p-6 rounded-xl shadow-sm border border-slate-100">
          <div className="flex items-center gap-4">
            <div className="w-12 h-12 bg-purple-100 rounded-xl flex items-center justify-center">
              <Sparkles className="w-6 h-6 text-purple-600" />
            </div>
            <div>
              <p className="text-sm text-slate-500">使用 OpenRouter</p>
              <p className="text-2xl font-bold">Gemini 3</p>
            </div>
          </div>
        </div>
      </div>

      {/* My Roles */}
      <div>
        <h2 className="text-xl font-bold text-slate-900 mb-4">我的角色</h2>
        {roles.length === 0 ? (
          <div className="bg-white rounded-xl p-12 text-center border border-slate-100">
            <Bot className="w-16 h-16 text-slate-300 mx-auto mb-4" />
            <p className="text-slate-500 mb-6">还没有创建角色</p>
            <button
              onClick={handleCreateRole}
              className="px-6 py-3 bg-primary text-white rounded-xl hover:shadow-lg transition-all"
            >
              创建第一个角色
            </button>
          </div>
        ) : (
          <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6">
            {roles.map(role => (
              <RoleCard
                key={role.id}
                role={role}
                onClick={() => navigate(`/chat/${role.id}`)}
                onUse={() => navigate(`/chat/${role.id}`)}
              />
            ))}
          </div>
        )}
      </div>

      {/* Template Roles */}
      <div>
        <h2 className="text-xl font-bold text-slate-900 mb-4">快速创建 - 使用模板</h2>
        <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-6">
          {templateRoles.map(template => (
            <div
              key={template.id}
              className="bg-white p-6 rounded-xl shadow-sm border border-slate-100 hover:shadow-md transition-all cursor-pointer"
              onClick={() => handleUseTemplate(template)}
            >
              <div className="text-2xl mb-2">{template.name.split(' ')[0]}</div>
              <h3 className="font-bold text-slate-900 mb-2">{template.name}</h3>
              <p className="text-sm text-slate-500 mb-4">{template.description}</p>
              <div className="flex items-center gap-2 text-sm text-primary">
                <Plus className="w-4 h-4" />
                <span>使用此模板</span>
              </div>
            </div>
          ))}
        </div>
      </div>
    </div>
  );
};
