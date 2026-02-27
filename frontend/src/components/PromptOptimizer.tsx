import { useState, useCallback, useEffect } from 'react';
import { api } from '../utils/api';

// 类型定义
export interface PromptVersion {
  id: string;
  content: string;
  score: number;
  features: string[];
  scenarios: string[];
  isRecommended?: boolean;
}

export interface OptimizationResult {
  versions: PromptVersion[];
  suggestions: OptimizationSuggestion[];
  originalLength: number;
  optimizedLength: number;
  improvementScore: number;
}

export interface OptimizationSuggestion {
  type: 'specificity' | 'example' | 'tone' | 'completeness';
  message: string;
  suggestion: string;
}

export interface PromptOptimizerProps {
  initialPrompt?: string;
  onOptimize?: (optimizedPrompt: string) => void;
  onClose?: () => void;
}

interface OptimizerState {
  status: 'idle' | 'optimizing' | 'completed' | 'error';
  progress: number;
  error?: string;
}

export const PromptOptimizer: React.FC<PromptOptimizerProps> = ({
  initialPrompt = '',
  onOptimize,
  onClose,
}) => {
  const [inputPrompt, setInputPrompt] = useState(initialPrompt);
  const [state, setState] = useState<OptimizerState>({
    status: 'idle',
    progress: 0,
  });
  const [result, setResult] = useState<OptimizationResult | null>(null);
  const [selectedVersion, setSelectedVersion] = useState<string | null>(null);
  const [showSuggestions, setShowSuggestions] = useState(true);

  // 一键优化
  const handleOptimize = useCallback(async () => {
    if (!inputPrompt.trim()) {
      setState({ status: 'error', progress: 0, error: '请输入提示词内容' });
      return;
    }

    setState({ status: 'optimizing', progress: 10 });

    try {
      // 模拟优化进度
      const progressInterval = setInterval(() => {
        setState(prev => ({
          ...prev,
          progress: Math.min(prev.progress + 10, 90),
        }));
      }, 300);

      const response = await api.post<OptimizationResult>('/api/prompt/optimize', {
        prompt: inputPrompt,
        generateVersions: 3,
        includeSuggestions: true,
      });

      clearInterval(progressInterval);
      setState({ status: 'completed', progress: 100 });
      setResult(response.data);

      // 自动选择推荐版本
      const recommended = response.data.versions.find(v => v.isRecommended);
      if (recommended) {
        setSelectedVersion(recommended.id);
      }
    } catch (error) {
      console.error('优化失败:', error);
      setState({
        status: 'error',
        progress: 0,
        error: error instanceof Error ? error.message : '优化失败，请重试',
      });
    }
  }, [inputPrompt]);

  // 应用选定版本
  const handleApplyVersion = useCallback((versionId: string) => {
    const version = result?.versions.find(v => v.id === versionId);
    if (version && onOptimize) {
      onOptimize(version.content);
    }
  }, [result, onOptimize]);

  // 实时建议（编辑过程中）
  const handleInputChange = useCallback((e: React.ChangeEvent<HTMLTextAreaElement>) => {
    const newValue = e.target.value;
    setInputPrompt(newValue);

    // 简单的实时检查
    if (newValue.length < 20) {
      // 可以显示"描述不够具体"的提示
    }
  }, []);

  // 渲染版本卡片
  const renderVersionCard = (version: PromptVersion) => (
    <div
      key={version.id}
      className={`p-4 rounded-lg border-2 cursor-pointer transition-all ${
        selectedVersion === version.id
          ? 'border-primary bg-primary/5'
          : 'border-slate-200 hover:border-primary/50'
      }`}
      onClick={() => setSelectedVersion(version.id)}
    >
      <div className="flex items-center justify-between mb-2">
        <div className="flex items-center gap-2">
          <span className="font-semibold text-slate-900">版本 {version.id}</span>
          {version.isRecommended && (
            <span className="text-xs px-2 py-0.5 bg-green-100 text-green-700 rounded-full">
              推荐
            </span>
          )}
        </div>
        <div className="flex items-center gap-2">
          <span className="text-sm text-slate-600">评分:</span>
          <span className="font-bold text-primary">{version.score}/100</span>
        </div>
      </div>

      <div className="text-sm text-slate-700 mb-3 line-clamp-3">{version.content}</div>

      <div className="space-y-2">
        <div>
          <span className="text-xs text-slate-500">特点:</span>
          <div className="flex flex-wrap gap-1 mt-1">
            {version.features.map((feature, idx) => (
              <span key={idx} className="text-xs px-2 py-0.5 bg-blue-50 text-blue-700 rounded">
                {feature}
              </span>
            ))}
          </div>
        </div>

        <div>
          <span className="text-xs text-slate-500">适用场景:</span>
          <div className="flex flex-wrap gap-1 mt-1">
            {version.scenarios.map((scenario, idx) => (
              <span key={idx} className="text-xs px-2 py-0.5 bg-purple-50 text-purple-700 rounded">
                {scenario}
              </span>
            ))}
          </div>
        </div>
      </div>

      {selectedVersion === version.id && (
        <button
          className="w-full mt-3 py-2 bg-primary text-white rounded-lg hover:bg-primary-dark transition-colors"
          onClick={(e) => {
            e.stopPropagation();
            handleApplyVersion(version.id);
          }}
        >
          应用此版本
        </button>
      )}
    </div>
  );

  // 渲染实时建议
  const renderSuggestions = () => (
    <div className="bg-amber-50 border border-amber-200 rounded-lg p-4">
      <div className="flex items-center justify-between mb-3">
        <h4 className="font-semibold text-amber-900">💡 优化建议</h4>
        <button
          className="text-sm text-amber-700 hover:text-amber-900"
          onClick={() => setShowSuggestions(!showSuggestions)}
        >
          {showSuggestions ? '收起' : '展开'}
        </button>
      </div>

      {showSuggestions && result?.suggestions && result.suggestions.length > 0 ? (
        <div className="space-y-2">
          {result.suggestions.map((suggestion, idx) => (
            <div key={idx} className="text-sm text-amber-800">
              <span className="font-medium">{getSuggestionIcon(suggestion.type)}</span>{' '}
              {suggestion.message}
            </div>
          ))}
        </div>
      ) : (
        <div className="text-sm text-amber-700">
          {inputPrompt.length > 0
            ? '输入提示词后获取 AI 优化建议'
            : '开始输入以获取实时建议'}
        </div>
      )}
    </div>
  );

  return (
    <div className="fixed inset-0 bg-black/50 flex items-center justify-center z-50 p-4">
      <div className="bg-white rounded-2xl w-full max-w-4xl max-h-[90vh] overflow-y-auto">
        {/* 头部 */}
        <div className="p-6 border-b border-slate-200 flex items-center justify-between sticky top-0 bg-white rounded-t-2xl">
          <div>
            <h2 className="text-2xl font-bold text-slate-900">✨ AI 提示词优化器</h2>
            <p className="text-sm text-slate-600 mt-1">
              一键生成专业版本，多版本对比选择
            </p>
          </div>
          {onClose && (
            <button
              onClick={onClose}
              className="p-2 hover:bg-slate-100 rounded-lg transition-colors"
            >
              <svg className="w-6 h-6" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M6 18L18 6M6 6l12 12" />
              </svg>
            </button>
          )}
        </div>

        {/* 内容区 */}
        <div className="p-6 space-y-6">
          {/* 输入区 */}
          <div>
            <label className="block text-sm font-medium text-slate-700 mb-2">
              原始提示词
            </label>
            <textarea
              value={inputPrompt}
              onChange={handleInputChange}
              placeholder="简单描述你的需求，例如：帮我写一个 Python 脚本，用于分析销售数据..."
              className="w-full h-32 p-3 border border-slate-300 rounded-lg focus:ring-2 focus:ring-primary focus:border-transparent resize-none"
            />
            <div className="flex items-center justify-between mt-2">
              <span className="text-xs text-slate-500">
                {inputPrompt.length} 字符
              </span>
              <button
                onClick={handleOptimize}
                disabled={state.status === 'optimizing' || !inputPrompt.trim()}
                className="px-6 py-2 bg-gradient-to-r from-primary to-primary-dark text-white rounded-lg hover:from-primary-dark hover:to-primary transition-all disabled:opacity-50 disabled:cursor-not-allowed flex items-center gap-2"
              >
                {state.status === 'optimizing' ? (
                  <>
                    <svg className="animate-spin h-4 w-4" viewBox="0 0 24 24">
                      <circle className="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" strokeWidth="4" fill="none" />
                      <path className="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z" />
                    </svg>
                    优化中...
                  </>
                ) : (
                  <>
                    <svg className="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                      <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M13 10V3L4 14h7v7l9-11h-7z" />
                    </svg>
                    AI 优化
                  </>
                )}
              </button>
            </div>
          </div>

          {/* 进度条 */}
          {state.status === 'optimizing' && (
            <div className="space-y-2">
              <div className="flex items-center justify-between text-sm">
                <span className="text-slate-600">正在优化提示词...</span>
                <span className="text-primary font-medium">{state.progress}%</span>
              </div>
              <div className="h-2 bg-slate-200 rounded-full overflow-hidden">
                <div
                  className="h-full bg-gradient-to-r from-primary to-primary-dark transition-all duration-300"
                  style={{ width: `${state.progress}%` }}
                />
              </div>
            </div>
          )}

          {/* 错误提示 */}
          {state.status === 'error' && (
            <div className="bg-red-50 border border-red-200 rounded-lg p-4 text-red-700">
              <div className="flex items-center gap-2">
                <svg className="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                  <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M12 8v4m0 4h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z" />
                </svg>
                {state.error}
              </div>
            </div>
          )}

          {/* 实时建议 */}
          {renderSuggestions()}

          {/* 优化结果 */}
          {state.status === 'completed' && result && (
            <div className="space-y-4">
              <div className="flex items-center justify-between">
                <h3 className="text-lg font-semibold text-slate-900">
                  生成 {result.versions.length} 个优化版本
                </h3>
                <div className="flex items-center gap-4 text-sm">
                  <span className="text-slate-600">
                    原始：{result.originalLength} 字符
                  </span>
                  <span className="text-slate-600">
                    优化：{result.optimizedLength} 字符
                  </span>
                  <span className="text-green-600 font-medium">
                    提升 {result.improvementScore}%
                  </span>
                </div>
              </div>

              <div className="grid grid-cols-1 md:grid-cols-3 gap-4">
                {result.versions.map(version => renderVersionCard(version))}
              </div>
            </div>
          )}
        </div>

        {/* 底部操作栏 */}
        <div className="p-6 border-t border-slate-200 bg-slate-50 rounded-b-2xl flex items-center justify-between">
          <button
            onClick={onClose}
            className="px-4 py-2 text-slate-600 hover:bg-slate-200 rounded-lg transition-colors"
          >
            取消
          </button>
          <div className="flex items-center gap-3">
            {selectedVersion && (
              <span className="text-sm text-slate-600">
                已选择版本 {selectedVersion}
              </span>
            )}
            <button
              onClick={() => selectedVersion && handleApplyVersion(selectedVersion)}
              disabled={!selectedVersion}
              className="px-6 py-2 bg-primary text-white rounded-lg hover:bg-primary-dark transition-colors disabled:opacity-50 disabled:cursor-not-allowed"
            >
              确认应用
            </button>
          </div>
        </div>
      </div>
    </div>
  );
};

// 辅助函数：获取建议图标
function getSuggestionIcon(type: string): string {
  switch (type) {
    case 'specificity':
      return '🎯';
    case 'example':
      return '📝';
    case 'tone':
      return '💬';
    case 'completeness':
      return '✅';
    default:
      return '💡';
  }
}

export default PromptOptimizer;
