import type { FC } from 'react';
import { useEffect, useMemo, useState } from 'react';
import { useNavigate } from 'react-router-dom';
import { Search, Plus, Filter, Grid3X3, List, Sparkles } from 'lucide-react';
import { RoleCard } from '../components/RoleCard';
import type { Role } from '../types';
import roleApi from '../api/role';

export const RoleMarket: FC = () => {
  const navigate = useNavigate();
  const [roles, setRoles] = useState<Role[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState('');
  const [activeCategory, setActiveCategory] = useState('全部');
  const [viewMode, setViewMode] = useState<'grid' | 'list'>('grid');
  const [searchQuery, setSearchQuery] = useState('');

  useEffect(() => {
    const loadTemplates = async () => {
      setLoading(true);
      setError('');
      try {
        const data = await roleApi.getTemplates();
        const normalized = data.map((role) => ({
          ...role,
          category: role.category || '通用',
          description: role.description || '',
          isTemplate: true,
        }));
        setRoles(normalized);
      } catch (err: any) {
        setError(err?.message || '加载角色模板失败');
      } finally {
        setLoading(false);
      }
    };

    loadTemplates();
  }, []);

  const categories = useMemo(() => {
    const fromData = Array.from(new Set(roles.map((r) => r.category).filter(Boolean)));
    return ['全部', ...fromData];
  }, [roles]);

  const filteredRoles = roles.filter(role => {
    const matchCategory = activeCategory === '全部' || role.category === activeCategory;
    const matchSearch = role.name.toLowerCase().includes(searchQuery.toLowerCase()) ||
                       (role.description || '').toLowerCase().includes(searchQuery.toLowerCase());
    return matchCategory && matchSearch;
  });

  const handleUseTemplate = async (template: Role) => {
    try {
      const created = await roleApi.create({
        name: template.name,
        description: template.description || '',
        category: template.category || '通用',
        systemPrompt: template.systemPrompt || '你是一个有帮助的 AI 助手。',
        welcomeMessage: template.welcomeMessage || `你好！我是${template.name}，很高兴为你服务。`,
      });
      navigate(`/chat/${created.id}`);
    } catch (err: any) {
      alert(err?.message || '使用模板失败，请先登录');
    }
  };

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
      {loading && (
        <div className="py-20 text-center text-slate-500">加载角色模板中...</div>
      )}

      {error && !loading && (
        <div className="rounded-xl border border-red-200 bg-red-50 px-4 py-3 text-sm text-red-600">
          {error}
        </div>
      )}

      <div className={`grid gap-6 ${viewMode === 'grid' ? 'grid-cols-3' : 'grid-cols-1'}`}>
        {filteredRoles.map(role => (
          <RoleCard
            key={role.id}
            role={role}
            onClick={() => handleUseTemplate(role)}
            onUse={() => handleUseTemplate(role)}
          />
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
