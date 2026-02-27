import type { FC } from 'react'; import { useState } from 'react';
import { Search, Plus, Filter, Grid3X3, List, Sparkles } from 'lucide-react';
import { RoleCard } from '../components/RoleCard';
import type { Role } from '../types';

const categories = ['全部', '通用', '营销', '法律', '财务', '技术', '人事', '行政'];

const mockRoles: Role[] = [
  {
    id: '1',
    name: '智能助理',
    description: '全能型办公助手，帮助处理日常事务、撰写邮件、安排日程、整理资料等',
    category: '通用',
    systemPrompt: 'You are a helpful assistant...',
    isTemplate: true,
    skills: [
      { id: '1', name: '邮件撰写', description: '' },
      { id: '2', name: '日程管理', description: '' },
      { id: '3', name: '资料整理', description: '' },
    ]
  },
  {
    id: '2',
    name: '营销专家',
    description: '专业的营销策划助手，帮助制定营销策略、撰写广告文案、分析市场趋势',
    category: '营销',
    systemPrompt: 'You are a marketing expert...',
    isTemplate: true,
    skills: [
      { id: '4', name: '文案撰写', description: '' },
      { id: '5', name: '活动策划', description: '' },
      { id: '6', name: '市场分析', description: '' },
    ]
  },
  {
    id: '3',
    name: '法务顾问',
    description: '合同审查与法律咨询专家，协助审查合同条款、解答法律问题',
    category: '法律',
    systemPrompt: 'You are a legal advisor...',
    isTemplate: true,
    skills: [
      { id: '7', name: '合同审核', description: '' },
      { id: '8', name: '法规查询', description: '' },
      { id: '9', name: '风险评估', description: '' },
    ]
  },
  {
    id: '4',
    name: '财务助手',
    description: '财务报表分析与税务咨询专家，帮助分析财务数据、规划税务',
    category: '财务',
    systemPrompt: 'You are a financial assistant...',
    isTemplate: true,
    skills: [
      { id: '10', name: '报表分析', description: '' },
      { id: '11', name: '税务规划', description: '' },
    ]
  },
  {
    id: '5',
    name: '技术支持',
    description: 'IT 问题诊断与解决专家，帮助排查技术故障、审查代码',
    category: '技术',
    systemPrompt: 'You are a tech support...',
    skills: [
      { id: '12', name: '故障排查', description: '' },
      { id: '13', name: '代码审查', description: '' },
    ]
  },
  {
    id: '6',
    name: 'HR 专员',
    description: '招聘与员工关系专家，协助简历筛选、面试安排、政策解答',
    category: '人事',
    systemPrompt: 'You are an HR specialist...',
    isTemplate: true,
    skills: [
      { id: '14', name: '简历筛选', description: '' },
      { id: '15', name: '面试安排', description: '' },
    ]
  },
];

export const RoleMarket: FC = () => {
  const [activeCategory, setActiveCategory] = useState('全部');
  const [viewMode, setViewMode] = useState<'grid' | 'list'>('grid');
  const [searchQuery, setSearchQuery] = useState('');

  const filteredRoles = mockRoles.filter(role => {
    const matchCategory = activeCategory === '全部' || role.category === activeCategory;
    const matchSearch = role.name.toLowerCase().includes(searchQuery.toLowerCase()) ||
                       role.description.toLowerCase().includes(searchQuery.toLowerCase());
    return matchCategory && matchSearch;
  });

  return (
    <div className="space-y-6">
      {/* Header */}
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-2xl font-bold text-slate-900">角色中心</h1>
          <p className="text-slate-500 mt-1">发现和使用 AI 角色，或创建属于你自己的角色</p>
        </div>
        <div className="flex items-center gap-3">
          <a 
            href="/roles/wizard"
            className="flex items-center gap-2 px-4 py-2 bg-gradient-to-r from-primary to-primary-dark text-white rounded-lg hover:from-primary-dark hover:to-primary transition-all shadow-md shadow-primary/20"
          >
            <Sparkles className="w-5 h-5" />
            零提示词创建
          </a>
          <a 
            href="/roles/create"
            className="flex items-center gap-2 px-4 py-2 bg-slate-900 text-white rounded-lg hover:bg-slate-800 transition-colors"
          >
            <Plus className="w-5 h-5" />
            传统创建
          </a>
        </div>
      </div>

      {/* Search & Filter */}
      <div className="flex items-center gap-4">
        <div className="flex-1 relative">
          <Search className="absolute left-3 top-1/2 -translate-y-1/2 w-5 h-5 text-slate-400" />
          <input
            type="text"
            placeholder="搜索角色..."
            value={searchQuery}
            onChange={(e) => setSearchQuery(e.target.value)}
            className="w-full pl-10 pr-4 py-2.5 bg-white border border-slate-200 rounded-xl outline-none focus:border-primary focus:ring-2 focus:ring-primary/20 transition-all"
          />
        </div>
        <button className="flex items-center gap-2 px-4 py-2.5 bg-white border border-slate-200 rounded-xl hover:bg-slate-50 transition-colors">
          <Filter className="w-5 h-5 text-slate-500" />
          <span className="text-slate-700">筛选</span>
        </button>
        <div className="flex items-center gap-1 bg-white border border-slate-200 rounded-xl p-1">
          <button 
            onClick={() => setViewMode('grid')}
            className={`p-2 rounded-lg transition-colors ${viewMode === 'grid' ? 'bg-slate-100 text-slate-900' : 'text-slate-400 hover:text-slate-600'}`}
          >
            <Grid3X3 className="w-5 h-5" />
          </button>
          <button 
            onClick={() => setViewMode('list')}
            className={`p-2 rounded-lg transition-colors ${viewMode === 'list' ? 'bg-slate-100 text-slate-900' : 'text-slate-400 hover:text-slate-600'}`}
          >
            <List className="w-5 h-5" />
          </button>
        </div>
      </div>

      {/* Categories */}
      <div className="flex items-center gap-2 overflow-x-auto pb-2">
        {categories.map(category => (
          <button
            key={category}
            onClick={() => setActiveCategory(category)}
            className={`px-4 py-2 rounded-full text-sm font-medium whitespace-nowrap transition-colors ${
              activeCategory === category
                ? 'bg-slate-900 text-white'
                : 'bg-white text-slate-600 hover:bg-slate-100 border border-slate-200'
            }`}
          >
            {category}
          </button>
        ))}
      </div>

      {/* Role Grid */}
      <div className={`grid gap-6 ${viewMode === 'grid' ? 'grid-cols-3' : 'grid-cols-1'}`}>
        {filteredRoles.map(role => (
          <RoleCard key={role.id} role={role} />
        ))}
      </div>

      {filteredRoles.length === 0 && (
        <div className="text-center py-16">
          <div className="w-16 h-16 bg-slate-100 rounded-full flex items-center justify-center mx-auto mb-4">
            <Search className="w-8 h-8 text-slate-400" />
          </div>
          <p className="text-slate-500">没有找到匹配的角色</p>
        </div>
      )}
    </div>
  );
};