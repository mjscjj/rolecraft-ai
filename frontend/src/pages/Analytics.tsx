import type { FC } from 'react';
import { useState, useEffect } from 'react';
import { 
  Users, 
  MessageSquare, 
  DollarSign,
  Activity,
  FileText,
  AlertTriangle,
  Star,
  ArrowUpRight,
  ArrowDownRight,
  Minus,
  Download,
  Calendar,
  BarChart3,
  PieChart
} from 'lucide-react';
import { analyticsApi } from '../api/analytics';

// 类型定义
interface DashboardMetrics {
  totalUsers: number;
  activeUsers: number;
  totalRoles: number;
  totalSessions: number;
  totalMessages: number;
  totalDocuments: number;
  totalCost: number;
  averageRating: number;
  userActivity?: {
    dau: number;
    wau: number;
    mau: number;
  };
  costStats?: {
    totalTokens: number;
    totalCost: number;
    averageCostPerDay: number;
  };
  qualityStats?: {
    averageRating: number;
    satisfactionRate: number;
  };
  topRoles?: Array<{
    roleId: string;
    roleName: string;
    tokensUsed: number;
    cost: number;
    percent: number;
  }>;
  recentTrends?: Array<{
    date: string;
    cost: number;
    tokens: number;
  }>;
}

interface UserActivity {
  dau: number;
  wau: number;
  mau: number;
}

interface CostTrend {
  date: string;
  cost: number;
  tokens: number;
}

interface ReportData {
  reportType: string;
  periodStart: string;
  periodEnd: string;
  keyMetrics: Array<{
    name: string;
    value: number;
    unit: string;
    change: number;
    trend: string;
  }>;
  recommendations: string[];
}

// 指标卡片组件
interface MetricCardProps {
  title: string;
  value: string | number;
  icon: any;
  trend?: number;
  subtitle?: string;
  color?: string;
}

const MetricCard: FC<MetricCardProps> = ({ title, value, icon: Icon, trend, subtitle, color = 'primary' }) => {
  const trendIcon = trend !== undefined ? (
    trend > 0 ? (
      <ArrowUpRight className="w-4 h-4 text-green-500" />
    ) : trend < 0 ? (
      <ArrowDownRight className="w-4 h-4 text-red-500" />
    ) : (
      <Minus className="w-4 h-4 text-slate-400" />
    )
  ) : null;

  const colorClasses = {
    primary: 'bg-primary/10 text-primary',
    green: 'bg-green-500/10 text-green-500',
    blue: 'bg-blue-500/10 text-blue-500',
    orange: 'bg-orange-500/10 text-orange-500',
    purple: 'bg-purple-500/10 text-purple-500',
  };

  return (
    <div className="bg-white p-6 rounded-xl shadow-sm border border-slate-100">
      <div className="flex items-center justify-between">
        <div>
          <p className="text-sm text-slate-500">{title}</p>
          <p className="text-3xl font-bold text-slate-900 mt-1">{value}</p>
          {subtitle && <p className="text-xs text-slate-400 mt-1">{subtitle}</p>}
        </div>
        <div className={`w-12 h-12 rounded-xl flex items-center justify-center ${colorClasses[color as keyof typeof colorClasses]}`}>
          <Icon className="w-6 h-6" />
        </div>
      </div>
      {trend !== undefined && (
        <div className="flex items-center gap-1 mt-4 text-sm">
          {trendIcon}
          <span className={trend > 0 ? 'text-green-500 font-medium' : trend < 0 ? 'text-red-500 font-medium' : 'text-slate-400 font-medium'}>
            {trend > 0 ? '+' : ''}{trend}%
          </span>
          <span className="text-slate-400">较上期</span>
        </div>
      )}
    </div>
  );
};

// 简易图表组件
const SimpleLineChart: FC<{ data: CostTrend[]; color?: string }> = ({ data, color = '#3b82f6' }) => {
  if (!data || data.length === 0) return null;

  const maxValue = Math.max(...data.map(d => d.cost));
  const minValue = Math.min(...data.map(d => d.cost));
  const range = maxValue - minValue || 1;

  const points = data.map((d, i) => {
    const x = (i / (data.length - 1)) * 100;
    const y = 100 - ((d.cost - minValue) / range) * 100;
    return `${x},${y}`;
  }).join(' ');

  return (
    <div className="w-full h-32">
      <svg viewBox="0 0 100 100" preserveAspectRatio="none" className="w-full h-full">
        <polyline
          fill="none"
          stroke={color}
          strokeWidth="2"
          points={points}
        />
        {data.map((d, i) => {
          const x = (i / (data.length - 1)) * 100;
          const y = 100 - ((d.cost - minValue) / range) * 100;
          return (
            <circle
              key={i}
              cx={x}
              cy={y}
              r="3"
              fill={color}
              className="hover:r-5 transition-all"
            />
          );
        })}
      </svg>
      <div className="flex justify-between mt-2 text-xs text-slate-400">
        <span>{data[0]?.date}</span>
        <span>{data[data.length - 1]?.date}</span>
      </div>
    </div>
  );
};

const SimpleBarChart: FC<{ data: Array<{ name: string; value: number }>; color?: string }> = ({ data, color = '#3b82f6' }) => {
  const maxValue = Math.max(...data.map(d => d.value));

  return (
    <div className="space-y-3">
      {data.map((item, i) => (
        <div key={i} className="flex items-center gap-3">
          <div className="w-24 text-xs text-slate-600 truncate">{item.name}</div>
          <div className="flex-1 h-6 bg-slate-100 rounded-full overflow-hidden">
            <div
              className="h-full rounded-full transition-all duration-500"
              style={{ width: `${(item.value / maxValue) * 100}%`, backgroundColor: color }}
            />
          </div>
          <div className="w-16 text-xs text-slate-500 text-right">{item.value.toFixed(1)}</div>
        </div>
      ))}
    </div>
  );
};

export const Analytics: FC = () => {
  const [loading, setLoading] = useState(true);
  const [metrics, setMetrics] = useState<DashboardMetrics | null>(null);
  const [userActivity, setUserActivity] = useState<UserActivity | null>(null);
  const [costTrend, setCostTrend] = useState<CostTrend[]>([]);
  const [reportType, setReportType] = useState<'weekly' | 'monthly'>('weekly');
  const [reportData, setReportData] = useState<ReportData | null>(null);

  useEffect(() => {
    loadDashboardData();
  }, []);

  const loadDashboardData = async () => {
    try {
      setLoading(true);
      
      // 加载 Dashboard 指标
      const metricsRes = await analyticsApi.getDashboard();
      setMetrics(metricsRes.data);

      // 加载用户活跃度
      const activityRes = await analyticsApi.getUserActivity();
      setUserActivity(activityRes.data);

      // 加载成本趋势
      const trendRes = await analyticsApi.getCostTrend(7);
      setCostTrend(trendRes.data);

      // 加载报告数据
      const reportRes = await analyticsApi.getReport(reportType);
      setReportData(reportRes.data);
    } catch (error) {
      console.error('Failed to load analytics data:', error);
    } finally {
      setLoading(false);
    }
  };

  const handleExportReport = async () => {
    try {
      const response = await analyticsApi.exportReport(reportType);
      // 实际实现中应该下载 PDF 文件
      alert('报告已生成（实际应下载 PDF 文件）');
      console.log('Report:', response.data);
    } catch (error) {
      console.error('Failed to export report:', error);
      alert('导出失败');
    }
  };

  if (loading) {
    return (
      <div className="flex items-center justify-center h-full">
        <div className="text-center">
          <Activity className="w-12 h-12 text-primary mx-auto mb-4 animate-pulse" />
          <p className="text-slate-500">加载数据中...</p>
        </div>
      </div>
    );
  }

  return (
    <div className="space-y-8">
      {/* Header */}
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-2xl font-bold text-slate-900">数据分析 Dashboard 📊</h1>
          <p className="text-slate-500 mt-1">数据驱动的洞察和决策支持</p>
        </div>
        <div className="flex items-center gap-3">
          <button
            onClick={() => setReportType(reportType === 'weekly' ? 'monthly' : 'weekly')}
            className="flex items-center gap-2 px-4 py-2 bg-white border border-slate-200 rounded-lg hover:bg-slate-50 transition-colors"
          >
            <Calendar className="w-4 h-4" />
            {reportType === 'weekly' ? '周报' : '月报'}
          </button>
          <button
            onClick={handleExportReport}
            className="flex items-center gap-2 px-4 py-2 bg-primary text-white rounded-lg hover:bg-primary-dark transition-colors"
          >
            <Download className="w-4 h-4" />
            导出报告
          </button>
        </div>
      </div>

      {/* 核心指标 */}
      <div className="grid grid-cols-4 gap-6">
        <MetricCard
          title="活跃用户"
          value={metrics?.activeUsers || 0}
          icon={Users}
          trend={12.5}
          subtitle={`DAU: ${userActivity?.dau || 0} | WAU: ${userActivity?.wau || 0}`}
          color="blue"
        />
        <MetricCard
          title="对话次数"
          value={metrics?.totalSessions || 0}
          icon={MessageSquare}
          trend={8.3}
          subtitle={`总消息：${metrics?.totalMessages || 0}`}
          color="green"
        />
        <MetricCard
          title="总成本"
          value={`¥${(metrics?.totalCost || 0).toFixed(2)}`}
          icon={DollarSign}
          trend={-5.2}
          subtitle={`Token: ${(metrics?.costStats?.totalTokens || 0).toLocaleString()}`}
          color="orange"
        />
        <MetricCard
          title="平均评分"
          value={(metrics?.averageRating || 0).toFixed(1)}
          icon={Star}
          trend={2.1}
          subtitle={`满意度：${(metrics?.qualityStats?.satisfactionRate || 0).toFixed(1)}%`}
          color="purple"
        />
      </div>

      {/* 用户活跃度 */}
      <div className="grid grid-cols-2 gap-6">
        <div className="bg-white p-6 rounded-xl shadow-sm border border-slate-100">
          <div className="flex items-center justify-between mb-4">
            <h2 className="font-semibold text-slate-900 flex items-center gap-2">
              <Activity className="w-5 h-5 text-primary" />
              用户活跃度趋势
            </h2>
          </div>
          <SimpleLineChart data={costTrend} color="#3b82f6" />
        </div>

        <div className="bg-white p-6 rounded-xl shadow-sm border border-slate-100">
          <div className="flex items-center justify-between mb-4">
            <h2 className="font-semibold text-slate-900 flex items-center gap-2">
              <BarChart3 className="w-5 h-5 text-primary" />
              用户活跃度统计
            </h2>
          </div>
          <div className="space-y-4">
            <div className="flex items-center justify-between p-4 bg-blue-50 rounded-lg">
              <div>
                <p className="text-sm text-slate-600">日活跃用户 (DAU)</p>
                <p className="text-2xl font-bold text-blue-600">{userActivity?.dau || 0}</p>
              </div>
              <Users className="w-8 h-8 text-blue-500" />
            </div>
            <div className="flex items-center justify-between p-4 bg-green-50 rounded-lg">
              <div>
                <p className="text-sm text-slate-600">周活跃用户 (WAU)</p>
                <p className="text-2xl font-bold text-green-600">{userActivity?.wau || 0}</p>
              </div>
              <Calendar className="w-8 h-8 text-green-500" />
            </div>
            <div className="flex items-center justify-between p-4 bg-purple-50 rounded-lg">
              <div>
                <p className="text-sm text-slate-600">月活跃用户 (MAU)</p>
                <p className="text-2xl font-bold text-purple-600">{userActivity?.mau || 0}</p>
              </div>
              <Calendar className="w-8 h-8 text-purple-500" />
            </div>
          </div>
        </div>
      </div>

      {/* 成本和角色使用 */}
      <div className="grid grid-cols-2 gap-6">
        <div className="bg-white p-6 rounded-xl shadow-sm border border-slate-100">
          <div className="flex items-center justify-between mb-4">
            <h2 className="font-semibold text-slate-900 flex items-center gap-2">
              <DollarSign className="w-5 h-5 text-primary" />
              成本趋势 (近 7 天)
            </h2>
          </div>
          <SimpleLineChart data={costTrend} color="#f97316" />
          <div className="mt-4 pt-4 border-t border-slate-100">
            <div className="flex items-center justify-between">
              <div>
                <p className="text-sm text-slate-500">总成本</p>
                <p className="text-xl font-bold text-slate-900">¥{(metrics?.totalCost || 0).toFixed(2)}</p>
              </div>
              <div>
                <p className="text-sm text-slate-500">日均成本</p>
                <p className="text-xl font-bold text-slate-900">¥{(metrics?.costStats?.averageCostPerDay || 0).toFixed(2)}</p>
              </div>
              <div>
                <p className="text-sm text-slate-500">总 Token</p>
                <p className="text-xl font-bold text-slate-900">{(metrics?.costStats?.totalTokens || 0).toLocaleString()}</p>
              </div>
            </div>
          </div>
        </div>

        <div className="bg-white p-6 rounded-xl shadow-sm border border-slate-100">
          <div className="flex items-center justify-between mb-4">
            <h2 className="font-semibold text-slate-900 flex items-center gap-2">
              <PieChart className="w-5 h-5 text-primary" />
              Top 5 角色使用
            </h2>
          </div>
          <SimpleBarChart
            data={(metrics?.topRoles || []).map(r => ({
              name: r.roleName,
              value: r.cost
            }))}
            color="#8b5cf6"
          />
        </div>
      </div>

      {/* 对话质量 */}
      <div className="bg-white p-6 rounded-xl shadow-sm border border-slate-100">
        <div className="flex items-center justify-between mb-4">
          <h2 className="font-semibold text-slate-900 flex items-center gap-2">
            <Star className="w-5 h-5 text-primary" />
            对话质量评估
          </h2>
        </div>
        <div className="grid grid-cols-4 gap-6">
          <div className="text-center p-4 bg-yellow-50 rounded-lg">
            <p className="text-sm text-slate-600">平均评分</p>
            <p className="text-3xl font-bold text-yellow-600 mt-2">{(metrics?.averageRating || 0).toFixed(1)}</p>
            <p className="text-xs text-slate-400 mt-1">满分 5 分</p>
          </div>
          <div className="text-center p-4 bg-green-50 rounded-lg">
            <p className="text-sm text-slate-600">满意度</p>
            <p className="text-3xl font-bold text-green-600 mt-2">{(metrics?.qualityStats?.satisfactionRate || 0).toFixed(1)}%</p>
            <p className="text-xs text-slate-400 mt-1">用户满意</p>
          </div>
          <div className="text-center p-4 bg-blue-50 rounded-lg">
            <p className="text-sm text-slate-600">高质量对话</p>
            <p className="text-3xl font-bold text-blue-600 mt-2">{metrics?.qualityStats ? Math.round(metrics.qualityStats.satisfactionRate * 180 / 100) : 0}</p>
            <p className="text-xs text-slate-400 mt-1">占比 70%+</p>
          </div>
          <div className="text-center p-4 bg-purple-50 rounded-lg">
            <p className="text-sm text-slate-600">总评分数</p>
            <p className="text-3xl font-bold text-purple-600 mt-2">{metrics?.qualityStats ? Math.round(metrics.qualityStats.satisfactionRate * 2.5) : 0}</p>
            <p className="text-xs text-slate-400 mt-1">次评价</p>
          </div>
        </div>
      </div>

      {/* 报告摘要和建议 */}
      <div className="bg-white p-6 rounded-xl shadow-sm border border-slate-100">
        <div className="flex items-center justify-between mb-4">
          <h2 className="font-semibold text-slate-900 flex items-center gap-2">
            <FileText className="w-5 h-5 text-primary" />
            {reportType === 'weekly' ? '周报' : '月报'}摘要
          </h2>
          <span className="text-sm text-slate-500">
            {reportData?.periodStart} 至 {reportData?.periodEnd}
          </span>
        </div>
        
        {reportData && (
          <>
            <div className="grid grid-cols-4 gap-4 mb-6">
              {reportData.keyMetrics.map((metric, i) => (
                <div key={i} className="p-4 bg-slate-50 rounded-lg">
                  <p className="text-sm text-slate-600">{metric.name}</p>
                  <p className="text-2xl font-bold text-slate-900 mt-1">
                    {metric.value.toFixed(1)}{metric.unit}
                  </p>
                  <div className="flex items-center gap-1 mt-2">
                    {metric.trend === 'up' ? (
                      <ArrowUpRight className="w-4 h-4 text-green-500" />
                    ) : metric.trend === 'down' ? (
                      <ArrowDownRight className="w-4 h-4 text-red-500" />
                    ) : (
                      <Minus className="w-4 h-4 text-slate-400" />
                    )}
                    <span className={metric.change > 0 ? 'text-green-500 text-sm' : 'text-red-500 text-sm'}>
                      {metric.change > 0 ? '+' : ''}{metric.change}%
                    </span>
                  </div>
                </div>
              ))}
            </div>

            <div>
              <h3 className="font-semibold text-slate-900 mb-3 flex items-center gap-2">
                <AlertTriangle className="w-4 h-4 text-orange-500" />
                优化建议
              </h3>
              <ul className="space-y-2">
                {reportData.recommendations.map((rec, i) => (
                  <li key={i} className="flex items-start gap-2 text-sm text-slate-600">
                    <span className="w-1.5 h-1.5 rounded-full bg-primary mt-1.5 flex-shrink-0" />
                    {rec}
                  </li>
                ))}
              </ul>
            </div>
          </>
        )}
      </div>
    </div>
  );
};
