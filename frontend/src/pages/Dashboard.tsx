import { FC } from 'react';
import { 
  Users, 
  FileText, 
  MessageSquare, 
  TrendingUp,
  Clock,
  ArrowRight
} from 'lucide-react';
import { RoleCard } from '../components/RoleCard';
import { Role, ChatSession } from '../types';

const stats = [
  { label: '我的角色', value: 12, icon: Users, trend: '+2' },
  { label: '知识文档', value: 48, icon: FileText, trend: '+5' },
  { label: '对话次数', value: 256, icon: MessageSquare, trend: '+12' },
];

const recentRoles: Role[] = [
  {
    id: '1',
    name: '营销专家',
    description: '专业的营销策划助手，帮助制定营销策略、撰写文案',
    category: '营销',
    systemPrompt: '',
    skills: [{ id: '1', name: '文案撰写', description: '' }, { id: '2', name: '市场分析', description: '' }]
  },
  {
    id: '2',
    name: '法务顾问',
    description: '合同审查与法律咨询专家',
    category: '法律',
    systemPrompt: '',
    skills: [{ id: '3', name: '合同审核', description: '' }]
  },
  {
    id: '3',
    name: '智能助理',
    description: '全能型办公助手，处理日常事务',
    category: '通用',
    systemPrompt: '',
  }
];

const recentChats: ChatSession[] = [
  { id: '1', roleId: '1', title: 'Q1营销方案讨论', mode: 'task', updatedAt: '10分钟前' },
  { id: '2', roleId: '2', title: '劳动合同条款审查', mode: 'quick', updatedAt: '2小时前' },
  { id: '3', roleId: '3', title: '周报整理', mode: 'quick', updatedAt: '昨天' },
];

const templateRoles: Role[] = [
  { id: 't1', name: '智能助理', description: '全能型办公助手', category: '通用', systemPrompt: '', isTemplate: true },
  { id: 't2', name: '法务顾问', description: '合同审查与法律咨询', category: '法律', systemPrompt: '', isTemplate: true },
  { id: 't3', name: '营销专家', description: '营销策划与内容创作', category: '营销', systemPrompt: '', isTemplate: true },
  { id: 't4', name: 'HR专员', description: '招聘与员工关系', category: '人事', systemPrompt: '', isTemplate: true },
];

export const Dashboard: FC = () => {
  return (
    <div className="space-y-8">
      {/* Welcome */}
      <div>
        <h1 className="text-2xl font-bold text-slate-900">欢迎回来 👋</h1>
        <p className="text-slate-500 mt-1">这是你的 AI 团队今日概览</p>
      </div>

      {/* Stats */}
      <div className="grid grid-cols-3 gap-6">
        {stats.map(stat => (
          <div key={stat.label} className="bg-white p-6 rounded-xl shadow-sm border border-slate-100">
            <div className="flex items-center justify-between">
              <div>
                <p className="text-sm text-slate-500">{stat.label}</p>
                <p className="text-3xl font-bold text-slate-900 mt-1">{stat.value}</p>
              </div>
              <div className="w-12 h-12 bg-primary/10 rounded-xl flex items-center justify-center">
                <stat.icon className="w-6 h-6 text-primary" />
              </div>
            </div>
            <div className="flex items-center gap-1 mt-4 text-sm">
              <TrendingUp className="w-4 h-4 text-green-500" />
              <span className="text-green-500 font-medium">{stat.trend}</span>
              <span className="text-slate-400">本周新增</span>
            </div>
          </div>
        ))}
      </div>

      {/* Recent & Chats */}
      <div className="grid grid-cols-2 gap-6">
        {/* Recent Roles */}
        <div className="bg-white rounded-xl shadow-sm border border-slate-100 p-6">
          <div className="flex items-center justify-between mb-4">
            <h2 className="font-semibold text-slate-900">最近使用的角色</h2>
            <a href="/roles" className="text-sm text-primary hover:underline flex items-center gap-1">
              查看全部 <ArrowRight className="w-4 h-4" />
            </a>
          </div>
          <div className="space-y-3">
            {recentRoles.map(role => (
              <div key={role.id} className="flex items-center gap-3 p-3 hover:bg-slate-50 rounded-lg cursor-pointer transition-colors">
                <div className="w-10 h-10 rounded-full bg-gradient-to-br from-primary to-primary-dark flex items-center justify-center text-white font-semibold">
                  {role.name.charAt(0)}
                </div>
                <div className="flex-1">
                  <p className="font-medium text-slate-900">{role.name}</p>
                  <p className="text-xs text-slate-500">{role.category}</p>
                </div>
              </div>
            ))}
          </div>
        </div>

        {/* Recent Chats */}
        <div className="bg-white rounded-xl shadow-sm border border-slate-100 p-6">
          <div className="flex items-center justify-between mb-4">
            <h2 className="font-semibold text-slate-900">最近对话</h2>
            <a href="/chat" className="text-sm text-primary hover:underline flex items-center gap-1">
              查看全部 <ArrowRight className="w-4 h-4" />
            </a>
          </div>
          <div className="space-y-3">
            {recentChats.map(chat => (
              <div key={chat.id} className="flex items-center gap-3 p-3 hover:bg-slate-50 rounded-lg cursor-pointer transition-colors">
                <div className="w-10 h-10 rounded-full bg-slate-100 flex items-center justify-center">
                  <MessageSquare className="w-5 h-5 text-slate-500" />
                </div>
                <div className="flex-1">
                  <p className="font-medium text-slate-900">{chat.title}</p>
                  <div className="flex items-center gap-2 text-xs text-slate-500">
                    <Clock className="w-3 h-3" />
                    {chat.updatedAt}
                  </div>
                </div>
              </div>
            ))}
          </div>
        </div>
      </div>

      {/* Template Roles */}
      <div>
        <h2 className="font-semibold text-slate-900 mb-4">推荐角色模板</h2>
        <div className="grid grid-cols-4 gap-6">
          {templateRoles.map(role => (
            <RoleCard key={role.id} role={role} />
          ))}
        </div>
      </div>
    </div>
  );
};