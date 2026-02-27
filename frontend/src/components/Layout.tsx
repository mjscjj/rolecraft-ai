import type { FC, ReactNode } from 'react';
import { 
  LayoutDashboard, 
  BookOpen, 
  MessageSquare, 
  Settings,
  Plus,
  Search,
  Bot,
  BarChart3
} from 'lucide-react';

interface LayoutProps {
  children: ReactNode;
}

const navItems = [
  { icon: LayoutDashboard, label: '仪表盘', path: '/' },
  { icon: Bot, label: '角色市场', path: '/roles' },
  { icon: MessageSquare, label: '对话', path: '/chat' },
  { icon: BookOpen, label: '知识库', path: '/documents' },
  { icon: BarChart3, label: '数据分析', path: '/analytics' },
  { icon: Settings, label: '设置', path: '/settings' },
];

export const Layout: FC<LayoutProps> = ({ children }) => {
  return (
    <div className="min-h-screen bg-gray-50">
      {/* Sidebar */}
      <aside className="fixed left-0 top-0 h-full w-64 bg-white border-r border-gray-200 shadow-sm">
        <div className="p-6">
          <div className="flex items-center gap-2 mb-8">
            <div className="w-10 h-10 bg-gradient-to-br from-indigo-500 to-purple-600 rounded-lg flex items-center justify-center">
              <Bot className="w-6 h-6 text-white" />
            </div>
            <span className="text-xl font-bold bg-gradient-to-r from-indigo-600 to-purple-600 bg-clip-text text-transparent">
              RoleCraft AI
            </span>
          </div>
          
          {/* Search */}
          <div className="relative mb-6">
            <Search className="absolute left-3 top-1/2 -translate-y-1/2 w-4 h-4 text-gray-400" />
            <input
              type="text"
              placeholder="搜索..."
              className="w-full pl-10 pr-4 py-2 bg-gray-100 rounded-lg text-sm focus:outline-none focus:ring-2 focus:ring-indigo-500"
            />
          </div>
          
          {/* Navigation */}
          <nav className="space-y-1">
            {navItems.map((item) => (
              <a
                key={item.path}
                href={item.path}
                className="flex items-center gap-3 px-3 py-2 text-gray-600 hover:bg-gray-100 rounded-lg transition-colors"
              >
                <item.icon className="w-5 h-5" />
                <span>{item.label}</span>
              </a>
            ))}
          </nav>
        </div>
        
        {/* Create Button */}
        <div className="absolute bottom-0 left-0 right-0 p-4">
          <button className="w-full flex items-center justify-center gap-2 px-4 py-2 bg-indigo-600 text-white rounded-lg hover:bg-indigo-700 transition-colors">
            <Plus className="w-5 h-5" />
            创建角色
          </button>
        </div>
      </aside>
      
      {/* Main Content */}
      <main className="ml-64 p-8">
        {children}
      </main>
    </div>
  );
};