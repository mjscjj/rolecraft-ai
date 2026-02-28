import React from 'react';
import { AlertCircle } from 'lucide-react';

interface AsyncBoundaryProps {
  children: React.ReactNode;
  fallback?: React.ReactNode;
}

interface AsyncBoundaryState {
  hasError: boolean;
  error: Error | null;
}

/**
 * AsyncBoundary - 异步组件错误边界
 * 用于包装 React.lazy 和其他异步组件
 */
export class AsyncBoundary extends React.Component<AsyncBoundaryProps, AsyncBoundaryState> {
  constructor(props: AsyncBoundaryProps) {
    super(props);
    this.state = {
      hasError: false,
      error: null,
    };
  }

  static getDerivedStateFromError(error: Error): Partial<AsyncBoundaryState> {
    return { hasError: true, error };
  }

  componentDidCatch(error: Error) {
    console.error('AsyncBoundary caught error:', error);
  }

  render() {
    if (this.state.hasError) {
      if (this.props.fallback) {
        return this.props.fallback;
      }

      return (
        <div className="p-4 bg-yellow-50 border border-yellow-200 rounded-lg">
          <div className="flex items-center gap-2">
            <AlertCircle className="w-5 h-5 text-yellow-600" />
            <p className="text-yellow-800">
              组件加载失败：{this.state.error?.message}
            </p>
          </div>
        </div>
      );
    }

    return this.props.children;
  }
}

export default AsyncBoundary;
