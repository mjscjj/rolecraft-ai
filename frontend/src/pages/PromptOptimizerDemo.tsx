import { useState } from 'react';
import { PromptOptimizer } from '../components/PromptOptimizer';
import type { OptimizationResult } from '../components/PromptOptimizer';

export const PromptOptimizerDemo = () => {
  const [showOptimizer, setShowOptimizer] = useState(false);
  const [optimizedPrompt, setOptimizedPrompt] = useState('');
  const [history, setHistory] = useState<Array<{
    original: string;
    optimized: string;
    timestamp: string;
  }>>([]);

  const handleOptimize = (optimized: string) => {
    setOptimizedPrompt(optimized);
    setShowOptimizer(false);
    
    // 记录历史
    setHistory(prev => [{
      original: '帮我写一个 Python 脚本',
      optimized: optimized,
      timestamp: new Date().toLocaleString(),
    }, ...prev]);
  };

  return (
    <div className="min-h-screen bg-gradient-to-br from-slate-50 to-slate-100 p-8">
      <div className="max-w-6xl mx-auto">
        {/* 头部 */}
        <div className="text-center mb-12">
          <h1 className="text-4xl font-bold text-slate-900 mb-4">
            ✨ AI 提示词优化器
          </h1>
          <p className="text-lg text-slate-600">
            一键生成专业提示词，多版本对比，实时优化建议
          </p>
        </div>

        {/* 功能展示区 */}
        <div className="grid grid-cols-1 md:grid-cols-2 gap-8 mb-12">
          {/* 功能卡片 1 */}
          <div className="bg-white rounded-xl p-6 shadow-md">
            <div className="flex items-center gap-3 mb-4">
              <div className="w-12 h-12 rounded-lg bg-primary/10 flex items-center justify-center">
                <svg className="w-6 h-6 text-primary" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                  <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M13 10V3L4 14h7v7l9-11h-7z" />
                </svg>
              </div>
              <h3 className="text-xl font-semibold text-slate-900">一键优化</h3>
            </div>
            <p className="text-slate-600 mb-4">
              简单描述你的需求，AI 自动生成专业版本的提示词
            </p>
            <button
              onClick={() => setShowOptimizer(true)}
              className="w-full py-3 bg-primary text-white rounded-lg hover:bg-primary-dark transition-colors font-medium"
            >
              立即体验
            </button>
          </div>

          {/* 功能卡片 2 */}
          <div className="bg-white rounded-xl p-6 shadow-md">
            <div className="flex items-center gap-3 mb-4">
              <div className="w-12 h-12 rounded-lg bg-green-100 flex items-center justify-center">
                <svg className="w-6 h-6 text-green-600" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                  <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M9 12l2 2 4-4m6 2a9 9 0 11-18 0 9 9 0 0118 0z" />
                </svg>
              </div>
              <h3 className="text-xl font-semibold text-slate-900">多版本对比</h3>
            </div>
            <p className="text-slate-600 mb-4">
              生成 3 个不同风格的版本，包含评分和适用场景说明
            </p>
            <div className="flex gap-2">
              <span className="px-3 py-1 bg-blue-50 text-blue-700 rounded-full text-sm">结构化</span>
              <span className="px-3 py-1 bg-purple-50 text-purple-700 rounded-full text-sm">详细版</span>
              <span className="px-3 py-1 bg-amber-50 text-amber-700 rounded-full text-sm">简洁版</span>
            </div>
          </div>

          {/* 功能卡片 3 */}
          <div className="bg-white rounded-xl p-6 shadow-md">
            <div className="flex items-center gap-3 mb-4">
              <div className="w-12 h-12 rounded-lg bg-amber-100 flex items-center justify-center">
                <svg className="w-6 h-6 text-amber-600" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                  <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M9.663 17h4.673M12 3v1m6.364 1.636l-.707.707M21 12h-1M4 12H3m3.343-5.657l-.707-.707m2.828 9.9a5 5 0 117.072 0l-.548.547A3.374 3.374 0 0014 18.469V19a2 2 0 11-4 0v-.531c0-.895-.356-1.754-.988-2.386l-.548-.547z" />
                </svg>
              </div>
              <h3 className="text-xl font-semibold text-slate-900">实时建议</h3>
            </div>
            <p className="text-slate-600 mb-4">
              编辑过程中提供智能建议，帮助你完善提示词
            </p>
            <div className="space-y-2 text-sm text-slate-600">
              <div className="flex items-center gap-2">
                <span>🎯</span>
                <span>描述具体化建议</span>
              </div>
              <div className="flex items-center gap-2">
                <span>📝</span>
                <span>添加示例建议</span>
              </div>
              <div className="flex items-center gap-2">
                <span>💬</span>
                <span>语气优化建议</span>
              </div>
            </div>
          </div>

          {/* 功能卡片 4 */}
          <div className="bg-white rounded-xl p-6 shadow-md">
            <div className="flex items-center gap-3 mb-4">
              <div className="w-12 h-12 rounded-lg bg-indigo-100 flex items-center justify-center">
                <svg className="w-6 h-6 text-indigo-600" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                  <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M9 19v-6a2 2 0 00-2-2H5a2 2 0 00-2 2v6a2 2 0 002 2h2a2 2 0 002-2zm0 0V9a2 2 0 012-2h2a2 2 0 012 2v10m-6 0a2 2 0 002 2h2a2 2 0 002-2m0 0V5a2 2 0 012-2h2a2 2 0 012 2v14a2 2 0 01-2 2h-2a2 2 0 01-2-2z" />
                </svg>
              </div>
              <h3 className="text-xl font-semibold text-slate-900">学习机制</h3>
            </div>
            <p className="text-slate-600 mb-4">
              记录用户选择，持续优化推荐算法
            </p>
            <div className="text-sm text-slate-600">
              <div className="flex items-center justify-between py-1">
                <span>已优化次数</span>
                <span className="font-semibold">{history.length}</span>
              </div>
              <div className="flex items-center justify-between py-1">
                <span>平均提升</span>
                <span className="font-semibold text-green-600">+45%</span>
              </div>
            </div>
          </div>
        </div>

        {/* 优化结果展示 */}
        {optimizedPrompt && (
          <div className="bg-white rounded-xl p-8 shadow-lg mb-12">
            <h2 className="text-2xl font-bold text-slate-900 mb-6">优化结果</h2>
            <div className="space-y-4">
              <div>
                <label className="block text-sm font-medium text-slate-700 mb-2">
                  原始提示词
                </label>
                <div className="p-4 bg-slate-50 rounded-lg text-slate-700">
                  帮我写一个 Python 脚本
                </div>
              </div>
              <div>
                <label className="block text-sm font-medium text-slate-700 mb-2">
                  优化后提示词
                </label>
                <div className="p-4 bg-green-50 border border-green-200 rounded-lg text-slate-800 whitespace-pre-wrap">
                  {optimizedPrompt}
                </div>
              </div>
              <button
                onClick={() => setShowOptimizer(true)}
                className="px-6 py-3 bg-primary text-white rounded-lg hover:bg-primary-dark transition-colors"
              >
                继续优化
              </button>
            </div>
          </div>
        )}

        {/* 使用场景 */}
        <div className="bg-white rounded-xl p-8 shadow-md">
          <h2 className="text-2xl font-bold text-slate-900 mb-6">适用场景</h2>
          <div className="grid grid-cols-1 md:grid-cols-3 gap-6">
            <div className="text-center p-4">
              <div className="w-16 h-16 mx-auto mb-4 rounded-full bg-blue-100 flex items-center justify-center">
                <span className="text-3xl">📧</span>
              </div>
              <h3 className="font-semibold text-slate-900 mb-2">邮件撰写</h3>
              <p className="text-sm text-slate-600">生成专业的商务邮件提示词</p>
            </div>
            <div className="text-center p-4">
              <div className="w-16 h-16 mx-auto mb-4 rounded-full bg-purple-100 flex items-center justify-center">
                <span className="text-3xl">📊</span>
              </div>
              <h3 className="font-semibold text-slate-900 mb-2">数据分析</h3>
              <p className="text-sm text-slate-600">优化数据分析任务描述</p>
            </div>
            <div className="text-center p-4">
              <div className="w-16 h-16 mx-auto mb-4 rounded-full bg-green-100 flex items-center justify-center">
                <span className="text-3xl">📝</span>
              </div>
              <h3 className="font-semibold text-slate-900 mb-2">内容创作</h3>
              <p className="text-sm text-slate-600">提升内容生成质量</p>
            </div>
          </div>
        </div>
      </div>

      {/* 优化器弹窗 */}
      {showOptimizer && (
        <PromptOptimizer
          initialPrompt="帮我写一个 Python 脚本"
          onOptimize={handleOptimize}
          onClose={() => setShowOptimizer(false)}
        />
      )}
    </div>
  );
};

export default PromptOptimizerDemo;
