import type { FC, ReactNode } from 'react';
import { NavLink, useNavigate } from 'react-router-dom';
import { 
  LayoutDashboard, 
  BookOpen, 
  MessageSquare, 
  Settings,
  Plus,
  Bot,
  BarChart3,
  Building2,
  Briefcase
} from 'lucide-react';

interface LayoutProps {
  children: ReactNode;
}

const navItems = [
  { icon: LayoutDashboard, label: '仪表盘', path: '/' },
  { icon: Bot, label: '角色市场', path: '/roles' },
  { icon: MessageSquare, label: '对话', path: '/chat' },
  { icon: Building2, label: '我的公司', path: '/companies' },
  { icon: Briefcase, label: '工作区', path: '/workspaces' },
  { icon: BookOpen, label: '知识库', path: '/documents' },
  { icon: BarChart3, label: '数据分析', path: '/analytics' },
  { icon: Settings, label: '设置', path: '/settings' },
];

export const Layout: FC<LayoutProps> = ({ children }) => {
  const navigate = useNavigate();
  const userRaw = typeof window !== 'undefined' ? localStorage.getItem('user') : null;
  const user = userRaw ? JSON.parse(userRaw) : null;

  const handleLogout = () => {
    localStorage.removeItem('token');
    localStorage.removeItem('user');
    navigate('/login');
  };

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
          
          {/* Navigation */}
          <nav className="space-y-1">
            {navItems.map((item) => (
              <NavLink
                key={item.path}
                to={item.path}
                className={({ isActive }) =>
                  `flex items-center gap-3 px-3 py-2 rounded-lg transition-colors ${
                    isActive ? 'bg-indigo-50 text-indigo-700' : 'text-gray-600 hover:bg-gray-100'
                  }`
                }
              >
                <item.icon className="w-5 h-5" />
                <span>{item.label}</span>
              </NavLink>
            ))}
          </nav>
        </div>
        
        {/* Create Button */}
        <div className="absolute bottom-0 left-0 right-0 p-4">
          <button
            onClick={() => navigate('/roles/create')}
            className="w-full flex items-center justify-center gap-2 px-4 py-2 bg-indigo-600 text-white rounded-lg hover:bg-indigo-700 transition-colors"
          >
            <Plus className="w-5 h-5" />
            创建角色
          </button>
          <div className="mt-3 rounded-lg border border-slate-200 bg-white px-3 py-2">
            <p className="truncate text-sm font-medium text-slate-700">{user?.name || '用户'}</p>
            <p className="truncate text-xs text-slate-400">{user?.email || ''}</p>
            <button
              onClick={handleLogout}
              className="mt-2 w-full rounded-md bg-slate-100 px-2 py-1 text-xs text-slate-600 hover:bg-slate-200"
            >
              退出登录
            </button>
          </div>
        </div>
      </aside>
      
      {/* Main Content */}
      <main className="ml-64 p-8">
        {children}
      </main>
    </div>
  );
};
