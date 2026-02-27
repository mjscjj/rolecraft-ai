import React, { useState, useEffect } from 'react';
import './ThinkingDisplay.css';

// 思考步骤类型
export type ThinkingStepType = 
  | 'understand'
  | 'analyze'
  | 'search'
  | 'organize'
  | 'conclude'
  | 'insight';

// 思考步骤状态
export type ThinkingStepStatus = 
  | 'pending'
  | 'processing'
  | 'completed';

// 思考步骤接口
export interface ThinkingStep {
  id: string;
  type: ThinkingStepType;
  content: string;
  timestamp: number;
  status: ThinkingStepStatus;
  icon: string;
  duration?: number;
}

// 思考过程接口
export interface ThinkingProcess {
  steps: ThinkingStep[];
  startTime: number;
  endTime?: number;
  duration: number;
  isComplete: boolean;
}

// ThinkingStepItem 组件属性
interface ThinkingStepItemProps {
  step: ThinkingStep;
  index: number;
  isExpanded: boolean;
}

// 思考步骤类型配置
const STEP_CONFIG: Record<ThinkingStepType, { label: string; icon: string; color: string }> = {
  understand: { label: '理解问题', icon: '🤔', color: '#667eea' },
  analyze: { label: '分析要素', icon: '🔍', color: '#764ba2' },
  search: { label: '检索知识', icon: '📚', color: '#f093fb' },
  organize: { label: '组织答案', icon: '📝', color: '#f5576c' },
  conclude: { label: '得出结论', icon: '✅', color: '#4facfe' },
  insight: { label: '灵感闪现', icon: '💡', color: '#43e97b' },
};

/**
 * ThinkingStepItem 组件 - 单个思考步骤
 */
export const ThinkingStepItem: React.FC<ThinkingStepItemProps> = ({ step, index, isExpanded }) => {
  const config = STEP_CONFIG[step.type] || STEP_CONFIG.understand;
  const isActive = step.status === 'processing';
  const isCompleted = step.status === 'completed';
  const isPending = step.status === 'pending';

  return (
    <div 
      className={`thinking-step ${step.status} ${isActive ? 'active' : ''}`}
      style={{ animationDelay: `${index * 0.1}s` }}
    >
      <div className="thinking-step-header">
        <span className="thinking-step-icon">{config.icon}</span>
        <span className="thinking-step-label" style={{ color: config.color }}>
          {config.label}
        </span>
        {isCompleted && step.duration && (
          <span className="thinking-step-duration">{step.duration.toFixed(1)}s</span>
        )}
        {isActive && (
          <span className="thinking-step-loading">
            <span className="loading-dot" />
            <span className="loading-dot" />
            <span className="loading-dot" />
          </span>
        )}
      </div>
      
      {isExpanded && (
        <div className="thinking-step-content">
          {step.content}
        </div>
      )}
      
      {isCompleted && (
        <div className="thinking-step-check">✓</div>
      )}
    </div>
  );
};

// ThinkingDisplay 组件属性
interface ThinkingDisplayProps {
  thinkingProcess?: ThinkingProcess;
  isStreaming?: boolean;
  defaultExpanded?: boolean;
  onToggle?: (expanded: boolean) => void;
}

/**
 * ThinkingDisplay 组件 - 思考过程展示
 */
export const ThinkingDisplay: React.FC<ThinkingDisplayProps> = ({
  thinkingProcess,
  isStreaming = false,
  defaultExpanded = true,
  onToggle,
}) => {
  const [isExpanded, setIsExpanded] = useState(defaultExpanded);
  const [displayDuration, setDisplayDuration] = useState(0);

  // 更新显示时长（流式时实时更新）
  useEffect(() => {
    if (!thinkingProcess || !isStreaming) return;

    const interval = setInterval(() => {
      const now = Date.now();
      const elapsed = (now - thinkingProcess.startTime) / 1000;
      setDisplayDuration(elapsed);
    }, 100);

    return () => clearInterval(interval);
  }, [thinkingProcess, isStreaming]);

  // 使用最终时长
  useEffect(() => {
    if (thinkingProcess?.duration) {
      setDisplayDuration(thinkingProcess.duration);
    }
  }, [thinkingProcess?.duration, thinkingProcess?.isComplete]);

  const handleToggle = () => {
    const newExpanded = !isExpanded;
    setIsExpanded(newExpanded);
    onToggle?.(newExpanded);
  };

  if (!thinkingProcess || thinkingProcess.steps.length === 0) {
    return null;
  }

  const { steps, isComplete } = thinkingProcess;

  return (
    <div className="thinking-display">
      <div className="thinking-header" onClick={handleToggle}>
        <div className="thinking-header-left">
          <span className="thinking-header-icon">🧠</span>
          <span className="thinking-header-title">
            深度思考
            {isStreaming && !isComplete && (
              <span className="thinking-streaming-indicator">中...</span>
            )}
          </span>
          <span className="thinking-header-meta">
            ({steps.length}步，{displayDuration.toFixed(1)}s)
          </span>
        </div>
        <button className="thinking-toggle-btn">
          {isExpanded ? '收起' : '展开'}
          <span className={`thinking-toggle-icon ${isExpanded ? 'expanded' : ''}`}>
            ▼
          </span>
        </button>
      </div>

      {isExpanded && (
        <div className="thinking-content">
          {/* 进度条 */}
          {!isComplete && isStreaming && (
            <div className="thinking-progress">
              <div 
                className="thinking-progress-bar"
                style={{ 
                  width: `${(steps.filter(s => s.status === 'completed').length / Math.max(steps.length, 1)) * 100}%` 
                }}
              />
            </div>
          )}

          {/* 思考步骤列表 */}
          <div className="thinking-steps">
            {steps.map((step, index) => (
              <ThinkingStepItem
                key={step.id}
                step={step}
                index={index}
                isExpanded={isExpanded}
              />
            ))}
          </div>

          {/* 完成状态 */}
          {isComplete && !isStreaming && (
            <div className="thinking-complete-badge">
              ✨ 思考完成
            </div>
          )}
        </div>
      )}
    </div>
  );
};

export default ThinkingDisplay;
