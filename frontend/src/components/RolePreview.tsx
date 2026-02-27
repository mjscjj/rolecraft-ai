import React, { useState, useEffect } from 'react';
import { Role, RoleCapability } from '../types';

interface RolePreviewProps {
  role?: Partial<Role>;
  systemPrompt?: string;
  modelName?: string;
  category?: string;
  onTestChat?: (message: string) => Promise<string>;
}

interface CapabilityItem {
  name: string;
  value: number;
  color: string;
}

export const RolePreview: React.FC<RolePreviewProps> = ({
  role,
  systemPrompt,
  modelName,
  category,
  onTestChat,
}) => {
  const [capabilities, setCapabilities] = useState<CapabilityItem[]>([]);
  const [testMessage, setTestMessage] = useState('');
  const [testResponse, setTestResponse] = useState('');
  const [isLoading, setIsLoading] = useState(false);
  const [rating, setRating] = useState(0);
  const [hoverRating, setHoverRating] = useState(0);

  // 分析能力雷达图数据
  useEffect(() => {
    if (systemPrompt || role?.systemPrompt) {
      const caps = analyzeCapabilities(systemPrompt || role?.systemPrompt || '');
      setCapabilities(caps);
    }
  }, [systemPrompt, role?.systemPrompt]);

  const analyzeCapabilities = (prompt: string): CapabilityItem[] => {
    const promptLower = prompt.toLowerCase();
    
    const calcScore = (keywords: string[]) => {
      let score = 50;
      keywords.forEach(keyword => {
        if (promptLower.includes(keyword)) {
          score += 10;
        }
      });
      return Math.min(100, Math.max(0, score));
    };

    return [
      {
        name: '创造性',
        value: calcScore(['创意', '创新', '创造', '想象', '设计', '艺术', '写作']),
        color: '#FF6B6B',
      },
      {
        name: '逻辑性',
        value: calcScore(['逻辑', '分析', '推理', '数据', '结构', '系统', '算法']),
        color: '#4ECDC4',
      },
      {
        name: '专业性',
        value: calcScore(['专业', '专家', '资深', '精通', '认证', '经验']),
        color: '#45B7D1',
      },
      {
        name: '共情力',
        value: calcScore(['理解', '关心', '支持', '帮助', '耐心', '友好', '温暖']),
        color: '#FFA07A',
      },
      {
        name: '效率',
        value: calcScore(['快速', '高效', '及时', '简洁', '优化', '自动化']),
        color: '#98D8C8',
      },
    ];
  };

  const handleTestChat = async () => {
    if (!testMessage.trim() || !onTestChat) return;
    
    setIsLoading(true);
    try {
      const response = await onTestChat(testMessage);
      setTestResponse(response);
      setTestMessage('');
    } catch (error) {
      setTestResponse('测试失败，请稍后重试');
    } finally {
      setIsLoading(false);
    }
  };

  const handleRating = (value: number) => {
    setRating(value);
  };

  const getPreviewData = () => {
    const name = role?.name || modelName || '未命名角色';
    const desc = role?.description || category ? `${category || '通用'}类 AI 助手` : '智能助手';
    const avatar = role?.avatar || name.charAt(0);
    
    return { name, desc, avatar };
  };

  const { name, desc, avatar } = getPreviewData();

  return (
    <div className="bg-white rounded-xl shadow-lg p-6 space-y-6">
      {/* 角色形象展示 */}
      <div className="flex items-start gap-4">
        <div className="w-20 h-20 rounded-full bg-gradient-to-br from-primary to-primary-dark flex items-center justify-center text-white text-3xl font-bold flex-shrink-0">
          {avatar}
        </div>
        <div className="flex-1">
          <h2 className="text-2xl font-bold text-slate-900">{name}</h2>
          <p className="text-slate-600 mt-1">{desc}</p>
          {category && (
            <span className="inline-block mt-2 text-xs px-2 py-1 bg-primary/10 text-primary rounded-full">
              {category}
            </span>
          )}
        </div>
      </div>

      {/* 能力雷达图 */}
      <div className="bg-slate-50 rounded-lg p-4">
        <h3 className="text-lg font-semibold text-slate-900 mb-4">能力雷达图</h3>
        <div className="grid grid-cols-2 md:grid-cols-3 lg:grid-cols-5 gap-4">
          {capabilities.map((cap) => (
            <div key={cap.name} className="text-center">
              <div className="relative w-16 h-16 mx-auto mb-2">
                <svg className="w-16 h-16 transform -rotate-90">
                  <circle
                    cx="32"
                    cy="32"
                    r="28"
                    stroke="#e2e8f0"
                    strokeWidth="6"
                    fill="none"
                  />
                  <circle
                    cx="32"
                    cy="32"
                    r="28"
                    stroke={cap.color}
                    strokeWidth="6"
                    fill="none"
                    strokeDasharray={`${(cap.value / 100) * 176} 176`}
                    strokeLinecap="round"
                  />
                </svg>
                <div className="absolute inset-0 flex items-center justify-center text-xs font-bold text-slate-700">
                  {Math.round(cap.value)}
                </div>
              </div>
              <p className="text-xs text-slate-600">{cap.name}</p>
            </div>
          ))}
        </div>
      </div>

      {/* 预计效果描述 */}
      <div className="bg-blue-50 border border-blue-200 rounded-lg p-4">
        <h3 className="text-lg font-semibold text-blue-900 mb-2">预计效果</h3>
        <ul className="space-y-2 text-sm text-blue-800">
          <li className="flex items-start gap-2">
            <span className="text-blue-500 mt-1">✓</span>
            <span>基于系统提示词，AI 将扮演 <strong>{name}</strong> 的角色</span>
          </li>
          <li className="flex items-start gap-2">
            <span className="text-blue-500 mt-1">✓</span>
            <span>回复风格：专业、友好、有针对性的建议</span>
          </li>
          <li className="flex items-start gap-2">
            <span className="text-blue-500 mt-1">✓</span>
            <span>适用场景：{category || '通用'}咨询、问题解答、专业建议</span>
          </li>
          {capabilities.some(c => c.value >= 70) && (
            <li className="flex items-start gap-2">
              <span className="text-blue-500 mt-1">★</span>
              <span>
                优势能力：{capabilities.filter(c => c.value >= 70).map(c => c.name).join('、')}
              </span>
            </li>
          )}
        </ul>
      </div>

      {/* 测试对话框 */}
      <div className="border border-slate-200 rounded-lg p-4">
        <h3 className="text-lg font-semibold text-slate-900 mb-4">测试对话</h3>
        
        {/* 预设测试问题 */}
        <div className="mb-4">
          <p className="text-sm text-slate-600 mb-2">预设问题：</p>
          <div className="flex flex-wrap gap-2">
            {[
              '你好，请介绍一下你自己',
              `我在${category || '工作'}中遇到一个问题...`,
              '能给我一些专业建议吗？',
            ].map((preset, idx) => (
              <button
                key={idx}
                onClick={() => setTestMessage(preset)}
                className="text-xs px-3 py-1.5 bg-slate-100 hover:bg-slate-200 text-slate-700 rounded-full transition-colors"
              >
                {preset}
              </button>
            ))}
          </div>
        </div>

        {/* 自定义问题输入 */}
        <div className="mb-4">
          <textarea
            value={testMessage}
            onChange={(e) => setTestMessage(e.target.value)}
            placeholder="输入你的测试问题..."
            className="w-full px-4 py-2 border border-slate-300 rounded-lg focus:ring-2 focus:ring-primary focus:border-transparent resize-none"
            rows={3}
          />
          <button
            onClick={handleTestChat}
            disabled={isLoading || !testMessage.trim()}
            className="mt-2 w-full py-2 bg-primary text-white rounded-lg hover:bg-primary-dark disabled:opacity-50 disabled:cursor-not-allowed transition-colors"
          >
            {isLoading ? '测试中...' : '发送测试'}
          </button>
        </div>

        {/* AI 回复 */}
        {testResponse && (
          <div className="bg-slate-50 rounded-lg p-4 mb-4">
            <div className="flex items-start gap-3">
              <div className="w-8 h-8 rounded-full bg-primary flex items-center justify-center text-white text-sm font-bold flex-shrink-0">
                {avatar}
              </div>
              <div className="flex-1">
                <p className="text-sm text-slate-800">{testResponse}</p>
                <p className="text-xs text-slate-500 mt-2">回复时间：{isLoading ? '...' : '< 1s'}</p>
              </div>
            </div>
          </div>
        )}

        {/* 满意度评分 */}
        {testResponse && (
          <div className="border-t border-slate-200 pt-4">
            <p className="text-sm text-slate-600 mb-2">满意度评分：</p>
            <div className="flex gap-2">
              {[1, 2, 3, 4, 5].map((star) => (
                <button
                  key={star}
                  onClick={() => handleRating(star)}
                  onMouseEnter={() => setHoverRating(star)}
                  onMouseLeave={() => setHoverRating(0)}
                  className="text-2xl transition-transform hover:scale-110"
                >
                  {star <= (hoverRating || rating) ? '⭐' : '☆'}
                </button>
              ))}
            </div>
            {rating > 0 && (
              <p className="text-xs text-slate-500 mt-2">
                感谢评分：{rating} 星
              </p>
            )}
          </div>
        )}
      </div>

      {/* 配置变更实时更新提示 */}
      <div className="text-xs text-slate-500 text-center">
        💡 提示：修改角色配置后，预览将实时更新
      </div>
    </div>
  );
};

export default RolePreview;
