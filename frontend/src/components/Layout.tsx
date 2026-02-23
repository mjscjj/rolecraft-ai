import { FC, ReactNode } from 'react';
import { 
  LayoutDashboard, 
  Users, 
  BookOpen, 
  MessageSquare, 
  Settings,
  Plus,
  ChevronDown,
  Search,
  Bell
} from 'lucide-react';

interface LayoutProps {
  children: ReactNode;
}

const navItems = [
  { icon: LayoutDashboard, label: '仪表板', path: '/' },
  { icon: Users, label: '角色中心', path: '/roles' },
  { icon: BookOpen, label: '知识库', path: '/documents' },
  { icon: MessageSquare, label: '对话', path: '/chat' },
  { icon: Settings, label: '设置', path: '/settings' },
];

export const Layout: FC<LayoutProps> = ({ children }) => {
  return (
    <div className="min-h-screen bg-slate-50 flex">
      {/* Sidebar */}
      <aside className="w-64 bg-white border-r border-slate-200 flex flex-col fixed h-full">
        {/* Logo */}
        <div className="p-6 border-b border-slate-100">
          <div className="flex items-center gap-3">
            <div className="w-10 h-10 rounded-xl bg-gradient-to-br from-primary to-primary-dark flex items-center justify-center">
              <span className="text-white font-bold text-lg">R</span>
            </div>
            <div>
              <h1 className="font-bold text-slate-900">RoleCraft</h1>
              <p className="text-xs text-slate-500">AI 角色平台</p>
            </div>
          </div>
        </div>

        {/* New Chat Button */}
        <div className="p-4">
          <button className="w-full flex items-center justify-center gap-2 py-3 px-4 bg-slate-900 text-white rounded-xl hover:bg-slate-800 transition-colors font-medium">
            <Plus className="w-5 h-5" />
            新建对话
          </button>
        </div>

        {/* Navigation */}
        <nav className="flex-1 px-3 py-2 space-y-1">
          {navItems.map(item => (
            <a
              key={item.path}
              href={item.path}
              className="flex items-center gap-3 px-3 py-2.5 text-slate-600 hover:bg-slate-50 hover:text-slate-900 rounded-lg transition-colors"
            >
              <item.icon className="w-5 h-5" />
              <span className="font-medium">{item.label}</span>
            </a>
          ))}
        </nav>

        {/* User Profile */}
        <div className="p-4 border-t border-slate-100">
          <div className="flex items-center gap-3 p-2 hover:bg-slate-50 rounded-lg cursor-pointer">
            <div className="w-9 h-9 rounded-full bg-gradient-to-br from-primary/20 to-primary-dark/20 flex items-center justify-center">
              <span className="text-primary font-semibold">U</span>
            </div>
            <div className="flex-1 min-w-0">
              <p className="text-sm font-medium text-slate-900 truncate">用户</p>
              <p className="text-xs text-slate-500 truncate">个人版</p>
            </div>
            <ChevronDown className="w-4 h-4 text-slate-400" />
          </div>
        </div>
      </aside>

      {/* Main Content */}
      <main className="flex-1 ml-64">
        {/* Header */}
        <header className="h-16 bg-white border-b border-slate-200 flex items-center justify-between px-8 sticky top-0 z-10">
          <div className="flex items-center gap-4 flex-1 max-w-xl">
            <Search className="w-5 h-5 text-slate-400" />
            <input
              type="text"
              placeholder="搜索角色、文档或对话..."
              className="flex-1 bg-transparent outline-none text-slate-700 placeholder:text-slate-400"
            />
          </div>
          <div className="flex items-center gap-4">
            <button className="relative p-2 text-slate-500 hover:bg-slate-100 rounded-lg transition-colors">
              <Bell className="w-5 h-5" />
              <span className="absolute top-1.5 right-1.5 w-2 h-2 bg-red-500 rounded-full"></span>
            </button>
          </div>
        </header>

        {/* Page Content */}
        <div className="p-8">
          {children}
        </div>
      </main>
    </div>
  );
};