import React, { Component, ErrorInfo, ReactNode } from 'react';
import { AlertTriangle, RefreshCw, Bug, FileText } from 'lucide-react';

interface Props {
  children: ReactNode;
  fallback?: ReactNode;
  onError?: (error: Error, errorInfo: ErrorInfo) => void;
}

interface State {
  hasError: boolean;
  error: Error | null;
  errorInfo: ErrorInfo | null;
}

export class ErrorBoundary extends Component<Props, State> {
  constructor(props: Props) {
    super(props);
    this.state = {
      hasError: false,
      error: null,
      errorInfo: null,
    };
  }

  static getDerivedStateFromError(error: Error): Partial<State> {
    return { hasError: true, error };
  }

  componentDidCatch(error: Error, errorInfo: ErrorInfo) {
    this.setState({ errorInfo });

    // 错误日志上报
    console.error('🚨 ErrorBoundary caught an error:', error, errorInfo);
    
    // 调用 onError 回调
    if (this.props.onError) {
      this.props.onError(error, errorInfo);
    }

    // 可以发送到错误监控服务
    // this.reportError(error, errorInfo);
  }

  // 错误上报（可扩展）
  private async reportError(error: Error, errorInfo: ErrorInfo) {
    try {
      await fetch('/api/v1/logs/error', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({
          error: {
            message: error.message,
            stack: error.stack,
          },
          component: errorInfo.componentStack,
          timestamp: new Date().toISOString(),
          userAgent: navigator.userAgent,
          url: window.location.href,
        }),
      });
    } catch (e) {
      console.error('Failed to report error:', e);
    }
  }

  // 重置错误状态
  handleReset = () => {
    this.setState({
      hasError: false,
      error: null,
      errorInfo: null,
    });
  };

  // 复制错误信息
  handleCopyError = () => {
    const errorText = `
错误信息：${this.state.error?.message}
组件堆栈：${this.state.errorInfo?.componentStack}
时间：${new Date().toISOString()}
`.trim();

    navigator.clipboard.writeText(errorText);
    alert('错误信息已复制到剪贴板');
  };

  render() {
    if (this.state.hasError) {
      // 使用自定义 fallback
      if (this.props.fallback) {
        return this.props.fallback;
      }

      // 默认错误 UI
      return (
        <div className="min-h-[400px] flex items-center justify-center bg-gray-50 rounded-lg border-2 border-red-100 p-8">
          <div className="max-w-md w-full text-center">
            <div className="w-16 h-16 mx-auto mb-4 bg-red-100 rounded-full flex items-center justify-center">
              <AlertTriangle className="w-8 h-8 text-red-600" />
            </div>
            
            <h2 className="text-xl font-bold text-gray-900 mb-2">
              出错了
            </h2>
            
            <p className="text-gray-600 mb-4">
              抱歉，页面出现了一些问题。请尝试刷新或联系技术支持。
            </p>

            {/* 错误详情（开发环境） */}
            {process.env.NODE_ENV === 'development' && this.state.error && (
              <div className="mb-4 p-4 bg-red-50 rounded-lg text-left">
                <h3 className="text-sm font-semibold text-red-800 mb-2">
                  错误详情：
                </h3>
                <p className="text-xs text-red-700 font-mono break-all">
                  {this.state.error.message}
                </p>
                {this.state.errorInfo && (
                  <details className="mt-2">
                    <summary className="text-xs text-red-700 cursor-pointer">
                      查看组件堆栈
                    </summary>
                    <pre className="mt-2 text-xs text-red-600 overflow-auto max-h-40">
                      {this.state.errorInfo.componentStack}
                    </pre>
                  </details>
                )}
              </div>
            )}

            {/* 操作按钮 */}
            <div className="flex gap-3 justify-center">
              <button
                onClick={this.handleReset}
                className="flex items-center gap-2 px-4 py-2 bg-indigo-600 text-white rounded-lg hover:bg-indigo-700 transition-colors"
              >
                <RefreshCw className="w-4 h-4" />
                重试
              </button>
              
              <button
                onClick={() => window.location.reload()}
                className="flex items-center gap-2 px-4 py-2 bg-gray-200 text-gray-700 rounded-lg hover:bg-gray-300 transition-colors"
              >
                刷新页面
              </button>

              {this.state.error && (
                <button
                  onClick={this.handleCopyError}
                  className="flex items-center gap-2 px-4 py-2 bg-white border border-gray-300 text-gray-700 rounded-lg hover:bg-gray-50 transition-colors"
                >
                  <FileText className="w-4 h-4" />
                  复制错误
                </button>
              )}
            </div>

            {/* 技术支持 */}
            <div className="mt-6 pt-6 border-t border-gray-200">
              <p className="text-sm text-gray-500">
                需要帮助？{' '}
                <a
                  href="mailto:support@rolecraft.ai"
                  className="text-indigo-600 hover:underline"
                >
                  联系技术支持
                </a>
              </p>
            </div>
          </div>
        </div>
      );
    }

    return this.props.children;
  }
}

// 函数组件错误边界包装器
export function withErrorBoundary<P extends object>(
  WrappedComponent: React.ComponentType<P>,
  displayName?: string
) {
  const ComponentWithBoundary = (props: P) => (
    <ErrorBoundary>
      <WrappedComponent {...props} />
    </ErrorBoundary>
  );

  if (displayName) {
    ComponentWithBoundary.displayName = displayName;
  }

  return ComponentWithBoundary;
}

// 异步组件错误边界
export function withAsyncErrorBoundary<P extends object>(
  asyncComponent: () => Promise<{ default: React.ComponentType<P> }>,
  displayName?: string
) {
  return withErrorBoundary(
    React.lazy(asyncComponent),
    displayName
  );
}
