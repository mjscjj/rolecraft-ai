import type { FC } from 'react';
import { useState, useEffect } from 'react';
import { 
  Search, 
  Plus, 
  Filter, 
  Grid3X3, 
  List, 
  Star, 
  TrendingUp, 
  Download, 
  Upload,
  Share2,
  TestTube,
  BarChart3,
  Lightbulb,
  Zap,
  Award,
  Users,
  MessageSquare,
  Clock,
  ChevronRight,
  X,
  Play,
  Save,
  Copy,
  CheckCircle,
  AlertCircle,
  Info
} from 'lucide-react';
import { RoleCard } from '../components/RoleCard';
import type { Role } from '../types';

// 能力雷达图组件
const CapabilityRadar: FC<{ capabilities: RoleCapability }> = ({ capabilities }) => {
  const axes = [
    { key: 'creativity', label: '创造性' },
    { key: 'logic', label: '逻辑性' },
    { key: 'professionalism', label: '专业性' },
    { key: 'empathy', label: '共情力' },
    { key: 'efficiency', label: '效率' },
    { key: 'adaptability', label: '适应性' },
  ];

  const angleStep = (Math.PI * 2) / axes.length;
  const radius = 80;
  const center = 100;

  const getPoint = (value: number, index: number) => {
    const angle = index * angleStep - Math.PI / 2;
    const r = (value / 100) * radius;
    return {
      x: center + r * Math.cos(angle),
      y: center + r * Math.sin(angle),
    };
  };

  const points = axes.map((axis, i) => {
    const value = capabilities[axis.key as keyof RoleCapability] as number;
    return getPoint(value, i);
  });

  const pathD = points.map((p, i) => (i === 0 ? `M ${p.x} ${p.y}` : `L ${p.x} ${p.y}`)).join(' ') + ' Z';

  return (
    <div className="relative w-64 h-64">
      <svg viewBox="0 0 200 200" className="w-full h-full">
        {/* 背景网格 */}
        {[20, 40, 60, 80, 100].map((level) => {
          const gridPoints = axes.map((_, i) => getPoint(level, i));
          const gridPath = gridPoints.map((p, i) => (i === 0 ? `M ${p.x} ${p.y}` : `L ${p.x} ${p.y}`)).join(' ') + ' Z';
          return (
            <path
              key={level}
              d={gridPath}
              fill="none"
              stroke="#e2e8f0"
              strokeWidth="1"
            />
          );
        })}
        
        {/* 轴线 */}
        {axes.map((_, i) => {
          const p = getPoint(100, i);
          return (
            <line
              key={i}
              x1={center}
              y1={center}
              x2={p.x}
              y2={p.y}
              stroke="#e2e8f0"
              strokeWidth="1"
            />
          );
        })}
        
        {/* 数据区域 */}
        <path
          d={pathD}
          fill="rgba(59, 130, 246, 0.2)"
          stroke="#3b82f6"
          strokeWidth="2"
        />
        
        {/* 数据点 */}
        {points.map((p, i) => (
          <circle
            key={i}
            cx={p.x}
            cy={p.y}
            r="4"
            fill="#3b82f6"
          />
        ))}
        
        {/* 标签 */}
        {axes.map((axis, i) => {
          const labelPos = getPoint(115, i);
          return (
            <text
              key={i}
              x={labelPos.x}
              y={labelPos.y}
              textAnchor="middle"
              dominantBaseline="middle"
              className="text-xs fill-slate-600"
            >
              {axis.label}
            </text>
          );
        })}
      </svg>
    </div>
  );
};

interface RoleCapability {
  creativity: number;
  logic: number;
  professionalism: number;
  empathy: number;
  efficiency: number;
  adaptability: number;
}

interface UsageStats {
  totalChats: number;
  totalMessages: number;
  avgSessionTime: number;
  activeUsers: number;
  favoriteCount: number;
  shareCount: number;
  popularityRank: number;
  categoryRank: number;
}

interface OptimizationSuggestion {
  type: string;
  priority: string;
  title: string;
  description: string;
  example: string;
}

interface TestReport {
  testId: string;
  roleName: string;
  testCases: TestCaseResult[];
  overallScore: number;
  passRate: number;
  summary: string;
}

interface TestCaseResult {
  caseId: string;
  caseName: string;
  input: string;
  expectedOutput: string;
  actualOutput: string;
  score: number;
  passed: boolean;
  feedback: string;
}

interface EnhancedRoleTemplate {
  id: string;
  name: string;
  description: string;
  category: string;
  systemPrompt: string;
  welcomeMessage: string;
  capabilities: RoleCapability;
  tags: string[];
  rating: number;
  usageCount: number;
  isPremium: boolean;
  exampleConversations: string[];
}

const categories = ['全部', '通用', '营销', '法律', '财务', '技术', '人事', '行政', '健康', '教育', '生活', '职业'];

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
];

export const Roles: FC = () => {
  const [activeCategory, setActiveCategory] = useState('全部');
  const [viewMode, setViewMode] = useState<'grid' | 'list'>('grid');
  const [searchQuery, setSearchQuery] = useState('');
  const [selectedRole, setSelectedRole] = useState<Role | null>(null);
  const [showEvaluation, setShowEvaluation] = useState(false);
  const [showSuggestions, setShowSuggestions] = useState(false);
  const [showTest, setShowTest] = useState(false);
  const [showExport, setShowExport] = useState(false);
  const [showTemplates, setShowTemplates] = useState(false);
  
  // 模拟数据
  const [capabilities] = useState<RoleCapability>({
    creativity: 75,
    logic: 85,
    professionalism: 90,
    empathy: 70,
    efficiency: 80,
    adaptability: 75,
  });
  
  const [usageStats] = useState<UsageStats>({
    totalChats: 1234,
    totalMessages: 15678,
    avgSessionTime: 23.5,
    activeUsers: 456,
    favoriteCount: 234,
    shareCount: 89,
    popularityRank: 12,
    categoryRank: 3,
  });
  
  const [suggestions] = useState<OptimizationSuggestion[]>([
    {
      type: 'prompt',
      priority: 'high',
      title: '增强专业性描述',
      description: '建议补充角色的专业背景和资质认证，提升可信度',
      example: '你是一位拥有 10 年经验的资深专家，持有 PMP、Scrum Master 认证...',
    },
    {
      type: 'prompt',
      priority: 'medium',
      title: '提升创造性思维',
      description: '营销类角色需要更强的创造性，建议加入创新思维指导',
      example: '你善于提出创新性的解决方案，能够从多个角度思考问题...',
    },
    {
      type: 'skill',
      priority: 'low',
      title: '添加技能标签',
      description: '建议为角色添加更多技能标签，便于用户发现',
      example: '',
    },
  ]);
  
  const [testReport, setTestReport] = useState<TestReport | null>(null);
  const [isRunningTest, setIsRunningTest] = useState(false);

  const filteredRoles = mockRoles.filter(role => {
    const matchCategory = activeCategory === '全部' || role.category === activeCategory;
    const matchSearch = role.name.toLowerCase().includes(searchQuery.toLowerCase()) ||
                       role.description.toLowerCase().includes(searchQuery.toLowerCase());
    return matchCategory && matchSearch;
  });

  const handleRunTest = () => {
    setIsRunningTest(true);
    // 模拟测试运行
    setTimeout(() => {
      setTestReport({
        testId: 'test_001',
        roleName: '智能助理',
        testCases: [
          {
            caseId: '1',
            caseName: '基础问候测试',
            input: '你好，请介绍一下你自己',
            expectedOutput: '友好的自我介绍',
            actualOutput: '你好！我是你的智能助理...',
            score: 92,
            passed: true,
            feedback: '回复友好专业，符合预期',
          },
          {
            caseId: '2',
            caseName: '专业能力测试',
            input: '帮我写一封会议邀请邮件',
            expectedOutput: '专业的邮件模板',
            actualOutput: '好的，请问会议的时间、地点...',
            score: 88,
            passed: true,
            feedback: '回复实用，但可以更详细',
          },
          {
            caseId: '3',
            caseName: '边界情况测试',
            input: '我不太确定该怎么描述...',
            expectedOutput: '耐心引导',
            actualOutput: '没关系，让我帮你梳理一下...',
            score: 95,
            passed: true,
            feedback: '引导非常到位',
          },
        ],
        overallScore: 91.7,
        passRate: 100,
        summary: '完成 3 个测试用例，通过率 100%',
      });
      setIsRunningTest(false);
    }, 2000);
  };

  const handleExport = () => {
    const exportData = {
      roleId: selectedRole?.id,
      name: selectedRole?.name,
      description: selectedRole?.description,
      category: selectedRole?.category,
      systemPrompt: selectedRole?.systemPrompt,
      welcomeMessage: '你好！',
      version: '1.0',
      exportedAt: new Date().toISOString(),
    };
    
    const blob = new Blob([JSON.stringify(exportData, null, 2)], { type: 'application/json' });
    const url = URL.createObjectURL(blob);
    const a = document.createElement('a');
    a.href = url;
    a.download = `${selectedRole?.name}_role.json`;
    a.click();
    URL.revokeObjectURL(url);
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
          <button 
            onClick={() => setShowTemplates(true)}
            className="flex items-center gap-2 px-4 py-2 bg-white border border-slate-200 rounded-lg hover:bg-slate-50 transition-colors"
          >
            <Award className="w-5 h-5 text-slate-600" />
            <span className="text-slate-700">模板库</span>
          </button>
          <a 
            href="/roles/create"
            className="flex items-center gap-2 px-4 py-2 bg-slate-900 text-white rounded-lg hover:bg-slate-800 transition-colors"
          >
            <Plus className="w-5 h-5" />
            创建角色
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
          <div key={role.id} className="relative group">
            <RoleCard 
              role={role} 
              onClick={() => setSelectedRole(role)}
              onUse={() => console.log('Use role', role.id)}
            />
            {/* Quick Actions */}
            <div className="absolute top-2 right-2 flex gap-1 opacity-0 group-hover:opacity-100 transition-opacity">
              <button
                onClick={() => { setSelectedRole(role); setShowEvaluation(true); }}
                className="p-2 bg-white rounded-lg shadow-md hover:bg-slate-50 transition-colors"
                title="能力评估"
              >
                <BarChart3 className="w-4 h-4 text-slate-600" />
              </button>
              <button
                onClick={() => { setSelectedRole(role); setShowSuggestions(true); }}
                className="p-2 bg-white rounded-lg shadow-md hover:bg-slate-50 transition-colors"
                title="优化建议"
              >
                <Lightbulb className="w-4 h-4 text-slate-600" />
              </button>
              <button
                onClick={() => { setSelectedRole(role); setShowTest(true); }}
                className="p-2 bg-white rounded-lg shadow-md hover:bg-slate-50 transition-colors"
                title="测试角色"
              >
                <TestTube className="w-4 h-4 text-slate-600" />
              </button>
              <button
                onClick={() => { setSelectedRole(role); setShowExport(true); }}
                className="p-2 bg-white rounded-lg shadow-md hover:bg-slate-50 transition-colors"
                title="导出配置"
              >
                <Download className="w-4 h-4 text-slate-600" />
              </button>
            </div>
          </div>
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

      {/* Role Detail Modal */}
      {selectedRole && (
        <div className="fixed inset-0 bg-black/50 flex items-center justify-center z-50 p-4">
          <div className="bg-white rounded-2xl max-w-4xl w-full max-h-[90vh] overflow-y-auto">
            <div className="sticky top-0 bg-white border-b border-slate-100 px-6 py-4 flex items-center justify-between">
              <h2 className="text-xl font-bold text-slate-900">{selectedRole.name}</h2>
              <button
                onClick={() => setSelectedRole(null)}
                className="p-2 hover:bg-slate-100 rounded-lg transition-colors"
              >
                <X className="w-5 h-5 text-slate-500" />
              </button>
            </div>
            
            <div className="p-6 space-y-6">
              {/* Basic Info */}
              <div className="flex items-start gap-4">
                <div className="w-20 h-20 rounded-full bg-gradient-to-br from-primary to-primary-dark flex items-center justify-center text-white text-3xl font-bold">
                  {selectedRole.name.charAt(0)}
                </div>
                <div>
                  <h3 className="text-lg font-semibold text-slate-900">{selectedRole.name}</h3>
                  <span className="text-sm text-slate-500">{selectedRole.category}</span>
                  <p className="text-slate-600 mt-2">{selectedRole.description}</p>
                </div>
              </div>

              {/* Stats */}
              <div className="grid grid-cols-4 gap-4">
                <div className="bg-slate-50 rounded-xl p-4">
                  <div className="flex items-center gap-2 text-slate-500 mb-1">
                    <MessageSquare className="w-4 h-4" />
                    <span className="text-xs">总对话数</span>
                  </div>
                  <p className="text-2xl font-bold text-slate-900">{usageStats.totalChats}</p>
                </div>
                <div className="bg-slate-50 rounded-xl p-4">
                  <div className="flex items-center gap-2 text-slate-500 mb-1">
                    <Users className="w-4 h-4" />
                    <span className="text-xs">活跃用户</span>
                  </div>
                  <p className="text-2xl font-bold text-slate-900">{usageStats.activeUsers}</p>
                </div>
                <div className="bg-slate-50 rounded-xl p-4">
                  <div className="flex items-center gap-2 text-slate-500 mb-1">
                    <Clock className="w-4 h-4" />
                    <span className="text-xs">平均时长</span>
                  </div>
                  <p className="text-2xl font-bold text-slate-900">{usageStats.avgSessionTime}m</p>
                </div>
                <div className="bg-slate-50 rounded-xl p-4">
                  <div className="flex items-center gap-2 text-slate-500 mb-1">
                    <TrendingUp className="w-4 h-4" />
                    <span className="text-xs">总排名</span>
                  </div>
                  <p className="text-2xl font-bold text-slate-900">#{usageStats.popularityRank}</p>
                </div>
              </div>

              {/* Capability Radar */}
              <div className="border border-slate-100 rounded-xl p-6">
                <h3 className="text-lg font-semibold text-slate-900 mb-4">能力评估</h3>
                <div className="flex items-center gap-8">
                  <CapabilityRadar capabilities={capabilities} />
                  <div className="space-y-3">
                    <div className="flex items-center justify-between gap-8">
                      <span className="text-sm text-slate-600">创造性</span>
                      <div className="flex items-center gap-2">
                        <div className="w-32 h-2 bg-slate-100 rounded-full">
                          <div className="h-full bg-blue-500 rounded-full" style={{ width: `${capabilities.creativity}%` }} />
                        </div>
                        <span className="text-sm font-medium w-10">{capabilities.creativity}</span>
                      </div>
                    </div>
                    <div className="flex items-center justify-between gap-8">
                      <span className="text-sm text-slate-600">逻辑性</span>
                      <div className="flex items-center gap-2">
                        <div className="w-32 h-2 bg-slate-100 rounded-full">
                          <div className="h-full bg-green-500 rounded-full" style={{ width: `${capabilities.logic}%` }} />
                        </div>
                        <span className="text-sm font-medium w-10">{capabilities.logic}</span>
                      </div>
                    </div>
                    <div className="flex items-center justify-between gap-8">
                      <span className="text-sm text-slate-600">专业性</span>
                      <div className="flex items-center gap-2">
                        <div className="w-32 h-2 bg-slate-100 rounded-full">
                          <div className="h-full bg-purple-500 rounded-full" style={{ width: `${capabilities.professionalism}%` }} />
                        </div>
                        <span className="text-sm font-medium w-10">{capabilities.professionalism}</span>
                      </div>
                    </div>
                    <div className="flex items-center justify-between gap-8">
                      <span className="text-sm text-slate-600">共情力</span>
                      <div className="flex items-center gap-2">
                        <div className="w-32 h-2 bg-slate-100 rounded-full">
                          <div className="h-full bg-pink-500 rounded-full" style={{ width: `${capabilities.empathy}%` }} />
                        </div>
                        <span className="text-sm font-medium w-10">{capabilities.empathy}</span>
                      </div>
                    </div>
                    <div className="flex items-center justify-between gap-8">
                      <span className="text-sm text-slate-600">效率</span>
                      <div className="flex items-center gap-2">
                        <div className="w-32 h-2 bg-slate-100 rounded-full">
                          <div className="h-full bg-yellow-500 rounded-full" style={{ width: `${capabilities.efficiency}%` }} />
                        </div>
                        <span className="text-sm font-medium w-10">{capabilities.efficiency}</span>
                      </div>
                    </div>
                    <div className="flex items-center justify-between gap-8">
                      <span className="text-sm text-slate-600">适应性</span>
                      <div className="flex items-center gap-2">
                        <div className="w-32 h-2 bg-slate-100 rounded-full">
                          <div className="h-full bg-orange-500 rounded-full" style={{ width: `${capabilities.adaptability}%` }} />
                        </div>
                        <span className="text-sm font-medium w-10">{capabilities.adaptability}</span>
                      </div>
                    </div>
                  </div>
                </div>
              </div>

              {/* System Prompt */}
              <div className="border border-slate-100 rounded-xl p-6">
                <h3 className="text-lg font-semibold text-slate-900 mb-2">系统提示词</h3>
                <pre className="bg-slate-50 rounded-lg p-4 text-sm text-slate-700 whitespace-pre-wrap font-mono">
                  {selectedRole.systemPrompt}
                </pre>
              </div>

              {/* Actions */}
              <div className="flex gap-3">
                <button
                  onClick={() => setShowEvaluation(true)}
                  className="flex-1 flex items-center justify-center gap-2 px-4 py-3 bg-slate-900 text-white rounded-lg hover:bg-slate-800 transition-colors"
                >
                  <BarChart3 className="w-5 h-5" />
                  查看评估报告
                </button>
                <button
                  onClick={() => setShowSuggestions(true)}
                  className="flex-1 flex items-center justify-center gap-2 px-4 py-3 bg-white border border-slate-200 text-slate-700 rounded-lg hover:bg-slate-50 transition-colors"
                >
                  <Lightbulb className="w-5 h-5" />
                  获取优化建议
                </button>
                <button
                  onClick={() => setShowTest(true)}
                  className="flex-1 flex items-center justify-center gap-2 px-4 py-3 bg-white border border-slate-200 text-slate-700 rounded-lg hover:bg-slate-50 transition-colors"
                >
                  <TestTube className="w-5 h-5" />
                  测试角色
                </button>
              </div>
            </div>
          </div>
        </div>
      )}

      {/* Evaluation Modal */}
      {showEvaluation && selectedRole && (
        <div className="fixed inset-0 bg-black/50 flex items-center justify-center z-50 p-4">
          <div className="bg-white rounded-2xl max-w-3xl w-full max-h-[90vh] overflow-y-auto">
            <div className="sticky top-0 bg-white border-b border-slate-100 px-6 py-4 flex items-center justify-between">
              <h2 className="text-xl font-bold text-slate-900">角色评估报告</h2>
              <button
                onClick={() => setShowEvaluation(false)}
                className="p-2 hover:bg-slate-100 rounded-lg transition-colors"
              >
                <X className="w-5 h-5 text-slate-500" />
              </button>
            </div>
            
            <div className="p-6 space-y-6">
              <div className="text-center">
                <h3 className="text-2xl font-bold text-slate-900">{selectedRole.name}</h3>
                <p className="text-slate-500 mt-1">综合能力评分</p>
                <div className="text-5xl font-bold text-primary mt-4">
                  {((capabilities.creativity + capabilities.logic + capabilities.professionalism + capabilities.empathy + capabilities.efficiency + capabilities.adaptability) / 6).toFixed(1)}
                </div>
                <p className="text-sm text-slate-500 mt-2">满分 100</p>
              </div>

              <CapabilityRadar capabilities={capabilities} />

              <div className="grid grid-cols-2 gap-4">
                <div className="bg-green-50 rounded-xl p-4">
                  <h4 className="font-semibold text-green-800 mb-2 flex items-center gap-2">
                    <CheckCircle className="w-5 h-5" />
                    优势
                  </h4>
                  <ul className="space-y-1 text-sm text-green-700">
                    <li>• 逻辑分析能力强 ({capabilities.logic})</li>
                    <li>• 专业知识扎实 ({capabilities.professionalism})</li>
                    <li>• 响应效率高 ({capabilities.efficiency})</li>
                  </ul>
                </div>
                <div className="bg-orange-50 rounded-xl p-4">
                  <h4 className="font-semibold text-orange-800 mb-2 flex items-center gap-2">
                    <AlertCircle className="w-5 h-5" />
                    待改进
                  </h4>
                  <ul className="space-y-1 text-sm text-orange-700">
                    <li>• 共情力可以提升 ({capabilities.empathy})</li>
                    <li>• 适应性有提升空间 ({capabilities.adaptability})</li>
                  </ul>
                </div>
              </div>

              <div className="bg-slate-50 rounded-xl p-4">
                <h4 className="font-semibold text-slate-900 mb-2 flex items-center gap-2">
                  <Info className="w-5 h-5" />
                  使用统计
                </h4>
                <div className="grid grid-cols-3 gap-4 text-sm">
                  <div>
                    <span className="text-slate-500">总对话数</span>
                    <p className="font-semibold text-slate-900">{usageStats.totalChats}</p>
                  </div>
                  <div>
                    <span className="text-slate-500">活跃用户</span>
                    <p className="font-semibold text-slate-900">{usageStats.activeUsers}</p>
                  </div>
                  <div>
                    <span className="text-slate-500">平均时长</span>
                    <p className="font-semibold text-slate-900">{usageStats.avgSessionTime} 分钟</p>
                  </div>
                </div>
              </div>
            </div>
          </div>
        </div>
      )}

      {/* Suggestions Modal */}
      {showSuggestions && selectedRole && (
        <div className="fixed inset-0 bg-black/50 flex items-center justify-center z-50 p-4">
          <div className="bg-white rounded-2xl max-w-3xl w-full max-h-[90vh] overflow-y-auto">
            <div className="sticky top-0 bg-white border-b border-slate-100 px-6 py-4 flex items-center justify-between">
              <h2 className="text-xl font-bold text-slate-900">优化建议</h2>
              <button
                onClick={() => setShowSuggestions(false)}
                className="p-2 hover:bg-slate-100 rounded-lg transition-colors"
              >
                <X className="w-5 h-5 text-slate-500" />
              </button>
            </div>
            
            <div className="p-6 space-y-4">
              {suggestions.map((suggestion, index) => (
                <div key={index} className="border border-slate-100 rounded-xl p-4">
                  <div className="flex items-start justify-between">
                    <div className="flex items-start gap-3">
                      <Lightbulb className={`w-5 h-5 mt-0.5 ${
                        suggestion.priority === 'high' ? 'text-red-500' :
                        suggestion.priority === 'medium' ? 'text-yellow-500' :
                        'text-blue-500'
                      }`} />
                      <div>
                        <h4 className="font-semibold text-slate-900">{suggestion.title}</h4>
                        <p className="text-sm text-slate-600 mt-1">{suggestion.description}</p>
                        {suggestion.example && (
                          <div className="mt-3 bg-slate-50 rounded-lg p-3">
                            <p className="text-xs text-slate-500 mb-1">示例：</p>
                            <p className="text-sm text-slate-700">{suggestion.example}</p>
                          </div>
                        )}
                      </div>
                    </div>
                    <span className={`text-xs px-2 py-1 rounded-full ${
                      suggestion.priority === 'high' ? 'bg-red-100 text-red-700' :
                      suggestion.priority === 'medium' ? 'bg-yellow-100 text-yellow-700' :
                      'bg-blue-100 text-blue-700'
                    }`}>
                      {suggestion.priority === 'high' ? '高优先级' :
                       suggestion.priority === 'medium' ? '中优先级' : '低优先级'}
                    </span>
                  </div>
                </div>
              ))}
            </div>
          </div>
        </div>
      )}

      {/* Test Modal */}
      {showTest && selectedRole && (
        <div className="fixed inset-0 bg-black/50 flex items-center justify-center z-50 p-4">
          <div className="bg-white rounded-2xl max-w-4xl w-full max-h-[90vh] overflow-y-auto">
            <div className="sticky top-0 bg-white border-b border-slate-100 px-6 py-4 flex items-center justify-between">
              <h2 className="text-xl font-bold text-slate-900">角色测试</h2>
              <button
                onClick={() => setShowTest(false)}
                className="p-2 hover:bg-slate-100 rounded-lg transition-colors"
              >
                <X className="w-5 h-5 text-slate-500" />
              </button>
            </div>
            
            <div className="p-6 space-y-6">
              {!testReport ? (
                <>
                  <div className="text-center py-8">
                    <TestTube className="w-16 h-16 text-slate-300 mx-auto mb-4" />
                    <h3 className="text-lg font-semibold text-slate-900">准备测试角色</h3>
                    <p className="text-slate-500 mt-2">将运行 3 个预设测试用例，评估角色的各项能力</p>
                  </div>
                  
                  <div className="space-y-3">
                    <div className="flex items-center gap-3 p-3 bg-slate-50 rounded-lg">
                      <div className="w-8 h-8 rounded-full bg-blue-100 flex items-center justify-center text-blue-600 font-semibold">1</div>
                      <div>
                        <p className="font-medium text-slate-900">基础问候测试</p>
                        <p className="text-sm text-slate-500">测试角色的基本交互能力</p>
                      </div>
                    </div>
                    <div className="flex items-center gap-3 p-3 bg-slate-50 rounded-lg">
                      <div className="w-8 h-8 rounded-full bg-green-100 flex items-center justify-center text-green-600 font-semibold">2</div>
                      <div>
                        <p className="font-medium text-slate-900">专业能力测试</p>
                        <p className="text-sm text-slate-500">测试角色的专业知识</p>
                      </div>
                    </div>
                    <div className="flex items-center gap-3 p-3 bg-slate-50 rounded-lg">
                      <div className="w-8 h-8 rounded-full bg-purple-100 flex items-center justify-center text-purple-600 font-semibold">3</div>
                      <div>
                        <p className="font-medium text-slate-900">边界情况测试</p>
                        <p className="text-sm text-slate-500">测试角色处理模糊问题的能力</p>
                      </div>
                    </div>
                  </div>
                  
                  <button
                    onClick={handleRunTest}
                    disabled={isRunningTest}
                    className="w-full flex items-center justify-center gap-2 px-4 py-3 bg-primary text-white rounded-lg hover:bg-primary-dark transition-colors disabled:opacity-50"
                  >
                    {isRunningTest ? (
                      <>
                        <div className="w-5 h-5 border-2 border-white border-t-transparent rounded-full animate-spin" />
                        运行测试中...
                      </>
                    ) : (
                      <>
                        <Play className="w-5 h-5" />
                        开始测试
                      </>
                    )}
                  </button>
                </>
              ) : (
                <>
                  <div className="text-center">
                    <div className="text-5xl font-bold text-primary mb-2">{testReport.overallScore.toFixed(1)}</div>
                    <p className="text-slate-500">综合得分</p>
                    <p className="text-sm text-slate-600 mt-2">{testReport.summary}</p>
                  </div>
                  
                  <div className="space-y-3">
                    {testReport.testCases.map((testCase) => (
                      <div key={testCase.caseId} className={`border rounded-xl p-4 ${testCase.passed ? 'border-green-200 bg-green-50' : 'border-red-200 bg-red-50'}`}>
                        <div className="flex items-start justify-between">
                          <div>
                            <h4 className="font-semibold text-slate-900">{testCase.caseName}</h4>
                            <p className="text-sm text-slate-600 mt-1">输入：{testCase.input}</p>
                            <p className="text-xs text-slate-500 mt-2">{testCase.feedback}</p>
                          </div>
                          <div className="text-right">
                            <div className={`text-2xl font-bold ${testCase.passed ? 'text-green-600' : 'text-red-600'}`}>
                              {testCase.score}
                            </div>
                            <span className={`text-xs px-2 py-1 rounded-full ${testCase.passed ? 'bg-green-200 text-green-700' : 'bg-red-200 text-red-700'}`}>
                              {testCase.passed ? '通过' : '未通过'}
                            </span>
                          </div>
                        </div>
                      </div>
                    ))}
                  </div>
                  
                  <button
                    onClick={handleRunTest}
                    className="w-full flex items-center justify-center gap-2 px-4 py-3 bg-slate-900 text-white rounded-lg hover:bg-slate-800 transition-colors"
                  >
                    <Play className="w-5 h-5" />
                    重新测试
                  </button>
                </>
              )}
            </div>
          </div>
        </div>
      )}

      {/* Export Modal */}
      {showExport && selectedRole && (
        <div className="fixed inset-0 bg-black/50 flex items-center justify-center z-50 p-4">
          <div className="bg-white rounded-2xl max-w-lg w-full">
            <div className="border-b border-slate-100 px-6 py-4 flex items-center justify-between">
              <h2 className="text-xl font-bold text-slate-900">导出角色配置</h2>
              <button
                onClick={() => setShowExport(false)}
                className="p-2 hover:bg-slate-100 rounded-lg transition-colors"
              >
                <X className="w-5 h-5 text-slate-500" />
              </button>
            </div>
            
            <div className="p-6 space-y-4">
              <div className="bg-slate-50 rounded-xl p-4">
                <h3 className="font-semibold text-slate-900 mb-2">{selectedRole.name}</h3>
                <p className="text-sm text-slate-600">{selectedRole.description}</p>
              </div>
              
              <div className="space-y-3">
                <label className="flex items-center justify-between p-3 border border-slate-200 rounded-lg cursor-pointer hover:bg-slate-50">
                  <div className="flex items-center gap-3">
                    <input type="checkbox" defaultChecked className="w-4 h-4 rounded border-slate-300 text-primary focus:ring-primary" />
                    <span className="text-sm text-slate-700">包含系统提示词</span>
                  </div>
                  <CheckCircle className="w-5 h-5 text-primary" />
                </label>
                <label className="flex items-center justify-between p-3 border border-slate-200 rounded-lg cursor-pointer hover:bg-slate-50">
                  <div className="flex items-center gap-3">
                    <input type="checkbox" defaultChecked className="w-4 h-4 rounded border-slate-300 text-primary focus:ring-primary" />
                    <span className="text-sm text-slate-700">包含技能配置</span>
                  </div>
                  <CheckCircle className="w-5 h-5 text-primary" />
                </label>
                <label className="flex items-center justify-between p-3 border border-slate-200 rounded-lg cursor-pointer hover:bg-slate-50">
                  <div className="flex items-center gap-3">
                    <input type="checkbox" className="w-4 h-4 rounded border-slate-300 text-primary focus:ring-primary" />
                    <span className="text-sm text-slate-700">包含对话历史</span>
                  </div>
                </label>
              </div>
              
              <div className="flex gap-3">
                <button
                  onClick={handleExport}
                  className="flex-1 flex items-center justify-center gap-2 px-4 py-3 bg-primary text-white rounded-lg hover:bg-primary-dark transition-colors"
                >
                  <Download className="w-5 h-5" />
                  导出 JSON
                </button>
                <button
                  onClick={() => {
                    navigator.clipboard.writeText(JSON.stringify({ roleId: selectedRole.id }, null, 2));
                  }}
                  className="flex-1 flex items-center justify-center gap-2 px-4 py-3 bg-white border border-slate-200 text-slate-700 rounded-lg hover:bg-slate-50 transition-colors"
                >
                  <Copy className="w-5 h-5" />
                  复制配置
                </button>
              </div>
            </div>
          </div>
        </div>
      )}

      {/* Templates Modal */}
      {showTemplates && (
        <div className="fixed inset-0 bg-black/50 flex items-center justify-center z-50 p-4">
          <div className="bg-white rounded-2xl max-w-5xl w-full max-h-[90vh] overflow-y-auto">
            <div className="sticky top-0 bg-white border-b border-slate-100 px-6 py-4 flex items-center justify-between">
              <h2 className="text-xl font-bold text-slate-900">角色模板库</h2>
              <button
                onClick={() => setShowTemplates(false)}
                className="p-2 hover:bg-slate-100 rounded-lg transition-colors"
              >
                <X className="w-5 h-5 text-slate-500" />
              </button>
            </div>
            
            <div className="p-6 space-y-6">
              <div className="flex items-center gap-2 overflow-x-auto pb-2">
                {['全部', '通用', '营销', '法律', '财务', '技术', '健康', '教育'].map(cat => (
                  <button
                    key={cat}
                    className={`px-4 py-2 rounded-full text-sm font-medium whitespace-nowrap transition-colors ${
                      cat === '全部'
                        ? 'bg-slate-900 text-white'
                        : 'bg-white text-slate-600 hover:bg-slate-100 border border-slate-200'
                    }`}
                  >
                    {cat}
                  </button>
                ))}
              </div>
              
              <div className="grid grid-cols-2 gap-4">
                {[
                  { name: '智能助理', category: '通用', rating: 4.8, usage: 15234, premium: false },
                  { name: '营销专家', category: '营销', rating: 4.9, usage: 12456, premium: false },
                  { name: '法务顾问', category: '法律', rating: 4.7, usage: 8932, premium: true },
                  { name: '心理咨询师', category: '健康', rating: 4.9, usage: 23456, premium: false },
                  { name: '编程导师', category: '技术', rating: 4.8, usage: 18765, premium: false },
                  { name: '财务规划师', category: '财务', rating: 4.7, usage: 9876, premium: true },
                  { name: '学术研究员', category: '教育', rating: 4.6, usage: 7654, premium: false },
                  { name: '健身教练', category: '健康', rating: 4.8, usage: 14532, premium: false },
                  { name: '旅行规划师', category: '生活', rating: 4.7, usage: 11234, premium: false },
                  { name: '职业规划师', category: '职业', rating: 4.8, usage: 13567, premium: true },
                ].map((template, index) => (
                  <div key={index} className="border border-slate-100 rounded-xl p-4 hover:shadow-md transition-shadow cursor-pointer">
                    <div className="flex items-start justify-between mb-3">
                      <div className="w-12 h-12 rounded-full bg-gradient-to-br from-primary to-primary-dark flex items-center justify-center text-white font-bold">
                        {template.name.charAt(0)}
                      </div>
                      {template.premium && (
                        <span className="text-xs px-2 py-1 bg-yellow-100 text-yellow-700 rounded-full flex items-center gap-1">
                          <Star className="w-3 h-3" />
                          高级
                        </span>
                      )}
                    </div>
                    <h3 className="font-semibold text-slate-900">{template.name}</h3>
                    <p className="text-xs text-slate-500 mt-1">{template.category}</p>
                    <div className="flex items-center justify-between mt-3">
                      <div className="flex items-center gap-1">
                        <Star className="w-4 h-4 text-yellow-500 fill-yellow-500" />
                        <span className="text-sm font-medium text-slate-700">{template.rating}</span>
                      </div>
                      <div className="flex items-center gap-1 text-xs text-slate-500">
                        <Users className="w-3 h-3" />
                        {template.usage.toLocaleString()}
                      </div>
                    </div>
                    <button className="w-full mt-3 py-2 text-sm text-primary hover:bg-primary/5 rounded-lg transition-colors">
                      使用此模板
                    </button>
                  </div>
                ))}
              </div>
            </div>
          </div>
        </div>
      )}
    </div>
  );
};
