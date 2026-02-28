import { useState } from 'react';
import { useNavigate } from 'react-router-dom';
import { User, Lock, Key, Bell, Palette, Save, LogOut } from 'lucide-react';
import client from '../api/client';

export const Settings = () => {
  const navigate = useNavigate();
  const [activeTab, setActiveTab] = useState('profile');
  const [loading, setLoading] = useState(false);
  const [message, setMessage] = useState({ type: '', text: '' });

  // 个人资料
  const [profile, setProfile] = useState({
    name: localStorage.getItem('userName') || '',
    email: localStorage.getItem('userEmail') || '',
  });

  // API 配置
  const [apiConfig, setApiConfig] = useState({
    openRouterKey: localStorage.getItem('openRouterKey') || '',
    preferredModel: localStorage.getItem('preferredModel') || 'google/gemini-3-flash-preview',
  });

  // 主题设置
  const [theme, setTheme] = useState(localStorage.getItem('theme') || 'light');

  const handleSaveProfile = async () => {
    setLoading(true);
    setMessage({ type: '', text: '' });
    try {
      const response = await client.put('/users/me', {
        name: profile.name,
        email: profile.email,
      });

      if (response.data.code === 0) {
        localStorage.setItem('userName', profile.name);
        localStorage.setItem('userEmail', profile.email);
        setMessage({ type: 'success', text: '个人资料已保存' });
      }
    } catch (error: any) {
      setMessage({ type: 'error', text: error.response?.data?.message || '保存失败' });
    } finally {
      setLoading(false);
    }
  };

  const handleChangePassword = async () => {
    // TODO: 实现修改密码
    setMessage({ type: 'info', text: '修改密码功能开发中...' });
  };

  const handleSaveApiConfig = () => {
    localStorage.setItem('openRouterKey', apiConfig.openRouterKey);
    localStorage.setItem('preferredModel', apiConfig.preferredModel);
    setMessage({ type: 'success', text: 'API 配置已保存（本地）' });
  };

  const handleThemeChange = (newTheme: string) => {
    setTheme(newTheme);
    localStorage.setItem('theme', newTheme);
    document.documentElement.classList.toggle('dark', newTheme === 'dark');
    setMessage({ type: 'success', text: `已切换到${newTheme === 'dark' ? '深色' : '浅色'}模式` });
  };

  const handleLogout = () => {
    localStorage.removeItem('token');
    localStorage.removeItem('user');
    navigate('/login');
  };

  const tabs = [
    { id: 'profile', label: '个人资料', icon: User },
    { id: 'security', label: '账号安全', icon: Lock },
    { id: 'api', label: 'API 配置', icon: Key },
    { id: 'theme', label: '主题外观', icon: Palette },
  ];

  return (
    <div className="min-h-screen bg-gray-50">
      {/* Header */}
      <div className="bg-white border-b border-gray-200">
        <div className="max-w-5xl mx-auto px-6 py-4">
          <div className="flex items-center justify-between">
            <div>
              <h1 className="text-2xl font-bold text-gray-900">设置</h1>
              <p className="text-sm text-gray-500 mt-1">管理你的账号和偏好设置</p>
            </div>
            <button
              onClick={handleLogout}
              className="flex items-center gap-2 px-4 py-2 text-red-600 hover:bg-red-50 rounded-lg transition-colors"
            >
              <LogOut className="w-4 h-4" />
              退出登录
            </button>
          </div>
        </div>
      </div>

      <div className="max-w-5xl mx-auto px-6 py-8">
        <div className="flex gap-6">
          {/* Sidebar */}
          <div className="w-64 flex-shrink-0">
            <nav className="space-y-1">
              {tabs.map(tab => {
                const Icon = tab.icon;
                return (
                  <button
                    key={tab.id}
                    onClick={() => setActiveTab(tab.id)}
                    className={`w-full flex items-center gap-3 px-4 py-3 text-sm font-medium rounded-lg transition-colors ${
                      activeTab === tab.id
                        ? 'bg-indigo-50 text-indigo-600'
                        : 'text-gray-600 hover:bg-gray-100'
                    }`}
                  >
                    <Icon className="w-5 h-5" />
                    {tab.label}
                  </button>
                );
              })}
            </nav>
          </div>

          {/* Content */}
          <div className="flex-1">
            <div className="bg-white rounded-xl shadow-sm border border-gray-200 p-6">
              {/* 个人资料 */}
              {activeTab === 'profile' && (
                <div className="space-y-6">
                  <div>
                    <h2 className="text-lg font-semibold text-gray-900 mb-4">个人资料</h2>
                    
                    <div className="space-y-4">
                      <div>
                        <label className="block text-sm font-medium text-gray-700 mb-1">
                          昵称
                        </label>
                        <input
                          type="text"
                          value={profile.name}
                          onChange={(e) => setProfile({ ...profile, name: e.target.value })}
                          className="w-full px-4 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-indigo-500 focus:border-transparent"
                          placeholder="请输入昵称"
                        />
                      </div>

                      <div>
                        <label className="block text-sm font-medium text-gray-700 mb-1">
                          邮箱
                        </label>
                        <input
                          type="email"
                          value={profile.email}
                          onChange={(e) => setProfile({ ...profile, email: e.target.value })}
                          className="w-full px-4 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-indigo-500 focus:border-transparent"
                          placeholder="请输入邮箱"
                        />
                      </div>

                      {message.text && (
                        <div className={`p-3 rounded-lg ${
                          message.type === 'success' ? 'bg-green-50 text-green-700' :
                          message.type === 'error' ? 'bg-red-50 text-red-700' :
                          'bg-blue-50 text-blue-700'
                        }`}>
                          {message.text}
                        </div>
                      )}

                      <button
                        onClick={handleSaveProfile}
                        disabled={loading}
                        className="flex items-center gap-2 px-6 py-2 bg-indigo-600 text-white rounded-lg hover:bg-indigo-700 transition-colors disabled:opacity-50"
                      >
                        <Save className="w-4 h-4" />
                        {loading ? '保存中...' : '保存'}
                      </button>
                    </div>
                  </div>
                </div>
              )}

              {/* 账号安全 */}
              {activeTab === 'security' && (
                <div className="space-y-6">
                  <div>
                    <h2 className="text-lg font-semibold text-gray-900 mb-4">账号安全</h2>
                    
                    <div className="space-y-4">
                      <div>
                        <label className="block text-sm font-medium text-gray-700 mb-1">
                          当前密码
                        </label>
                        <input
                          type="password"
                          className="w-full px-4 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-indigo-500 focus:border-transparent"
                          placeholder="请输入当前密码"
                        />
                      </div>

                      <div>
                        <label className="block text-sm font-medium text-gray-700 mb-1">
                          新密码
                        </label>
                        <input
                          type="password"
                          className="w-full px-4 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-indigo-500 focus:border-transparent"
                          placeholder="请输入新密码"
                        />
                      </div>

                      <div>
                        <label className="block text-sm font-medium text-gray-700 mb-1">
                          确认新密码
                        </label>
                        <input
                          type="password"
                          className="w-full px-4 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-indigo-500 focus:border-transparent"
                          placeholder="请再次输入新密码"
                        />
                      </div>

                      {message.text && (
                        <div className={`p-3 rounded-lg ${
                          message.type === 'success' ? 'bg-green-50 text-green-700' :
                          message.type === 'error' ? 'bg-red-50 text-red-700' :
                          'bg-blue-50 text-blue-700'
                        }`}>
                          {message.text}
                        </div>
                      )}

                      <button
                        onClick={handleChangePassword}
                        className="flex items-center gap-2 px-6 py-2 bg-indigo-600 text-white rounded-lg hover:bg-indigo-700 transition-colors"
                      >
                        <Lock className="w-4 h-4" />
                        修改密码
                      </button>
                    </div>
                  </div>
                </div>
              )}

              {/* API 配置 */}
              {activeTab === 'api' && (
                <div className="space-y-6">
                  <div>
                    <h2 className="text-lg font-semibold text-gray-900 mb-4">API 配置</h2>
                    
                    <div className="space-y-4">
                      <div>
                        <label className="block text-sm font-medium text-gray-700 mb-1">
                          OpenRouter API Key
                        </label>
                        <input
                          type="password"
                          value={apiConfig.openRouterKey}
                          onChange={(e) => setApiConfig({ ...apiConfig, openRouterKey: e.target.value })}
                          className="w-full px-4 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-indigo-500 focus:border-transparent"
                          placeholder="sk-or-v1-..."
                        />
                        <p className="text-xs text-gray-500 mt-1">
                          用于访问 OpenRouter AI 服务，配置后存储在本地
                        </p>
                      </div>

                      <div>
                        <label className="block text-sm font-medium text-gray-700 mb-1">
                          首选模型
                        </label>
                        <select
                          value={apiConfig.preferredModel}
                          onChange={(e) => setApiConfig({ ...apiConfig, preferredModel: e.target.value })}
                          className="w-full px-4 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-indigo-500 focus:border-transparent"
                        >
                          <option value="google/gemini-3-flash-preview">Gemini 3 Flash（快速）</option>
                          <option value="google/gemini-3-pro-preview">Gemini 3 Pro（高质量）</option>
                          <option value="anthropic/claude-opus-4.6">Claude Opus 4.6（深度思考）</option>
                          <option value="deepseek/deepseek-v3.2-speciale">DeepSeek V3.2（中文优化）</option>
                        </select>
                      </div>

                      {message.text && (
                        <div className={`p-3 rounded-lg ${
                          message.type === 'success' ? 'bg-green-50 text-green-700' :
                          message.type === 'error' ? 'bg-red-50 text-red-700' :
                          'bg-blue-50 text-blue-700'
                        }`}>
                          {message.text}
                        </div>
                      )}

                      <button
                        onClick={handleSaveApiConfig}
                        className="flex items-center gap-2 px-6 py-2 bg-indigo-600 text-white rounded-lg hover:bg-indigo-700 transition-colors"
                      >
                        <Key className="w-4 h-4" />
                        保存配置
                      </button>
                    </div>
                  </div>
                </div>
              )}

              {/* 主题外观 */}
              {activeTab === 'theme' && (
                <div className="space-y-6">
                  <div>
                    <h2 className="text-lg font-semibold text-gray-900 mb-4">主题外观</h2>
                    
                    <div className="space-y-4">
                      <div className="flex items-center justify-between p-4 border border-gray-200 rounded-lg">
                        <div className="flex items-center gap-3">
                          <div className="w-12 h-12 bg-white border-2 border-gray-200 rounded-lg"></div>
                          <div>
                            <p className="font-medium text-gray-900">浅色模式</p>
                            <p className="text-sm text-gray-500">明亮清爽的界面</p>
                          </div>
                        </div>
                        <button
                          onClick={() => handleThemeChange('light')}
                          className={`px-4 py-2 rounded-lg transition-colors ${
                            theme === 'light'
                              ? 'bg-indigo-600 text-white'
                              : 'bg-gray-100 text-gray-700 hover:bg-gray-200'
                          }`}
                        >
                          {theme === 'light' ? '当前使用' : '使用'}
                        </button>
                      </div>

                      <div className="flex items-center justify-between p-4 border border-gray-200 rounded-lg">
                        <div className="flex items-center gap-3">
                          <div className="w-12 h-12 bg-gray-800 rounded-lg"></div>
                          <div>
                            <p className="font-medium text-gray-900">深色模式</p>
                            <p className="text-sm text-gray-500">护眼舒适的暗色主题</p>
                          </div>
                        </div>
                        <button
                          onClick={() => handleThemeChange('dark')}
                          className={`px-4 py-2 rounded-lg transition-colors ${
                            theme === 'dark'
                              ? 'bg-indigo-600 text-white'
                              : 'bg-gray-100 text-gray-700 hover:bg-gray-200'
                          }`}
                        >
                          {theme === 'dark' ? '当前使用' : '使用'}
                        </button>
                      </div>

                      {message.text && (
                        <div className={`p-3 rounded-lg ${
                          message.type === 'success' ? 'bg-green-50 text-green-700' :
                          message.type === 'error' ? 'bg-red-50 text-red-700' :
                          'bg-blue-50 text-blue-700'
                        }`}>
                          {message.text}
                        </div>
                      )}
                    </div>
                  </div>
                </div>
              )}
            </div>
          </div>
        </div>
      </div>
    </div>
  );
};
