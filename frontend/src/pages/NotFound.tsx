import { useNavigate } from 'react-router-dom';
import { Home, Search, AlertTriangle } from 'lucide-react';

export const NotFound = () => {
  const navigate = useNavigate();

  return (
    <div className="min-h-screen bg-gray-50 flex items-center justify-center p-4">
      <div className="max-w-md w-full text-center">
        {/* 404 图标 */}
        <div className="w-24 h-24 mx-auto mb-6 bg-gradient-to-br from-indigo-600 to-purple-600 rounded-full flex items-center justify-center">
          <AlertTriangle className="w-12 h-12 text-white" />
        </div>

        {/* 404 文字 */}
        <h1 className="text-6xl font-bold text-gray-900 mb-4">
          404
        </h1>

        <h2 className="text-2xl font-semibold text-gray-700 mb-4">
          页面未找到
        </h2>

        <p className="text-gray-600 mb-8">
          抱歉，您访问的页面不存在或已被移除。
        </p>

        {/* 搜索框 */}
        <div className="mb-6">
          <div className="relative">
            <Search className="absolute left-3 top-1/2 -translate-y-1/2 w-5 h-5 text-gray-400" />
            <input
              type="text"
              placeholder="搜索内容..."
              className="w-full pl-10 pr-4 py-3 border border-gray-300 rounded-lg focus:ring-2 focus:ring-indigo-500 focus:border-transparent"
            />
          </div>
        </div>

        {/* 操作按钮 */}
        <div className="flex gap-3 justify-center">
          <button
            onClick={() => navigate('/')}
            className="flex items-center gap-2 px-6 py-3 bg-indigo-600 text-white rounded-lg hover:bg-indigo-700 transition-colors font-medium"
          >
            <Home className="w-5 h-5" />
            返回首页
          </button>

          <button
            onClick={() => navigate(-1)}
            className="flex items-center gap-2 px-6 py-3 bg-gray-200 text-gray-700 rounded-lg hover:bg-gray-300 transition-colors font-medium"
          >
            返回上一页
          </button>
        </div>

        {/* 快捷链接 */}
        <div className="mt-8 pt-8 border-t border-gray-200">
          <p className="text-sm text-gray-500 mb-4">热门页面：</p>
          <div className="flex flex-wrap gap-2 justify-center">
            <button
              onClick={() => navigate('/')}
              className="px-4 py-2 text-sm bg-white border border-gray-300 rounded-lg hover:bg-gray-50 transition-colors"
            >
              📊 仪表盘
            </button>
            <button
              onClick={() => navigate('/roles')}
              className="px-4 py-2 text-sm bg-white border border-gray-300 rounded-lg hover:bg-gray-50 transition-colors"
            >
              🎭 角色市场
            </button>
            <button
              onClick={() => navigate('/documents')}
              className="px-4 py-2 text-sm bg-white border border-gray-300 rounded-lg hover:bg-gray-50 transition-colors"
            >
              📚 知识库
            </button>
            <button
              onClick={() => navigate('/settings')}
              className="px-4 py-2 text-sm bg-white border border-gray-300 rounded-lg hover:bg-gray-50 transition-colors"
            >
              ⚙️ 设置
            </button>
          </div>
        </div>

        {/* 帮助信息 */}
        <div className="mt-8 text-sm text-gray-500">
          <p>需要帮助？</p>
          <a
            href="mailto:support@rolecraft.ai"
            className="text-indigo-600 hover:underline"
          >
            联系技术支持
          </a>
        </div>
      </div>
    </div>
  );
};
