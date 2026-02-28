import { useState, useEffect } from 'react';
import { Search, Trash2, Archive, MessageSquare, Clock, ChevronLeft } from 'lucide-react';
import client from '../api/client';

interface ChatSession {
  id: string;
  title: string;
  roleId: string;
  roleName?: string;
  createdAt: string;
  updatedAt: string;
  isArchived?: boolean;
  messageCount?: number;
}

interface ChatHistoryProps {
  onSelectSession: (sessionId: string) => void;
  currentSessionId?: string;
  isOpen: boolean;
  onClose: () => void;
}

export const ChatHistory: React.FC<ChatHistoryProps> = ({
  onSelectSession,
  currentSessionId,
  isOpen,
  onClose,
}) => {
  const [sessions, setSessions] = useState<ChatSession[]>([]);
  const [loading, setLoading] = useState(true);
  const [searchQuery, setSearchQuery] = useState('');
  const [filter, setFilter] = useState<'all' | 'active' | 'archived'>('active');

  useEffect(() => {
    if (isOpen) {
      loadSessions();
    }
  }, [isOpen, filter]);

  const loadSessions = async () => {
    try {
      setLoading(true);
      const response = await client.get('/chat-sessions');
      if (response.data.code === 0) {
        let filteredSessions = response.data.data || [];
        
        // 过滤
        if (filter === 'active') {
          filteredSessions = filteredSessions.filter((s: any) => !s.config?.isArchived);
        } else if (filter === 'archived') {
          filteredSessions = filteredSessions.filter((s: any) => s.config?.isArchived);
        }
        
        // 搜索
        if (searchQuery) {
          filteredSessions = filteredSessions.filter((s: any) =>
            s.title.toLowerCase().includes(searchQuery.toLowerCase())
          );
        }
        
        setSessions(filteredSessions);
      }
    } catch (error) {
      console.error('加载会话列表失败:', error);
    } finally {
      setLoading(false);
    }
  };

  const handleDelete = async (sessionId: string, e: React.MouseEvent) => {
    e.stopPropagation();
    if (!confirm('确定要删除这个会话吗？')) return;
    
    try {
      await client.delete(`/chat-sessions/${sessionId}`);
      setSessions(prev => prev.filter(s => s.id !== sessionId));
    } catch (error) {
      console.error('删除失败:', error);
      alert('删除失败');
    }
  };

  const handleArchive = async (sessionId: string, e: React.MouseEvent) => {
    e.stopPropagation();
    try {
      await client.put(`/chat-sessions/${sessionId}/archive`, { isArchived: true });
      setSessions(prev => prev.filter(s => s.id !== sessionId));
    } catch (error) {
      console.error('归档失败:', error);
    }
  };

  const filteredSessions = sessions.filter(session => {
    if (!searchQuery) return true;
    return session.title.toLowerCase().includes(searchQuery.toLowerCase());
  });

  if (!isOpen) return null;

  return (
    <div className="fixed inset-0 z-50 flex">
      {/* 遮罩层 */}
      <div className="absolute inset-0 bg-black/50" onClick={onClose} />
      
      {/* 侧边栏 */}
      <div className="relative w-80 bg-white shadow-2xl flex flex-col h-full">
        {/* Header */}
        <div className="p-4 border-b border-gray-200">
          <div className="flex items-center justify-between mb-3">
            <h2 className="text-lg font-bold text-gray-900">对话历史</h2>
            <button
              onClick={onClose}
              className="p-1 hover:bg-gray-100 rounded"
            >
              <ChevronLeft className="w-5 h-5" />
            </button>
          </div>
          
          {/* 搜索框 */}
          <div className="relative">
            <Search className="absolute left-3 top-1/2 -translate-y-1/2 w-4 h-4 text-gray-400" />
            <input
              type="text"
              value={searchQuery}
              onChange={(e) => setSearchQuery(e.target.value)}
              placeholder="搜索会话..."
              className="w-full pl-10 pr-4 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-indigo-500 focus:border-transparent text-sm"
            />
          </div>
          
          {/* 过滤标签 */}
          <div className="flex gap-2 mt-3">
            <button
              onClick={() => setFilter('active')}
              className={`flex-1 px-3 py-1.5 text-sm rounded-lg transition-colors ${
                filter === 'active'
                  ? 'bg-indigo-600 text-white'
                  : 'bg-gray-100 text-gray-600 hover:bg-gray-200'
              }`}
            >
              进行中
            </button>
            <button
              onClick={() => setFilter('archived')}
              className={`flex-1 px-3 py-1.5 text-sm rounded-lg transition-colors ${
                filter === 'archived'
                  ? 'bg-indigo-600 text-white'
                  : 'bg-gray-100 text-gray-600 hover:bg-gray-200'
              }`}
            >
              已归档
            </button>
            <button
              onClick={() => setFilter('all')}
              className={`flex-1 px-3 py-1.5 text-sm rounded-lg transition-colors ${
                filter === 'all'
                  ? 'bg-indigo-600 text-white'
                  : 'bg-gray-100 text-gray-600 hover:bg-gray-200'
              }`}
            >
              全部
            </button>
          </div>
        </div>

        {/* 会话列表 */}
        <div className="flex-1 overflow-y-auto p-2">
          {loading ? (
            <div className="text-center py-8 text-gray-500">加载中...</div>
          ) : filteredSessions.length === 0 ? (
            <div className="text-center py-8 text-gray-500">
              <MessageSquare className="w-12 h-12 mx-auto mb-2 opacity-50" />
              <p>暂无会话</p>
            </div>
          ) : (
            <div className="space-y-1">
              {filteredSessions.map(session => (
                <div
                  key={session.id}
                  onClick={() => {
                    onSelectSession(session.id);
                    onClose();
                  }}
                  className={`group p-3 rounded-lg cursor-pointer transition-colors ${
                    currentSessionId === session.id
                      ? 'bg-indigo-50 border-indigo-200'
                      : 'hover:bg-gray-50 border-transparent'
                  } border`}
                >
                  <div className="flex items-start justify-between">
                    <div className="flex-1 min-w-0">
                      <h3 className="font-medium text-gray-900 truncate text-sm">
                        {session.title || '新对话'}
                      </h3>
                      <div className="flex items-center gap-2 mt-1">
                        <Clock className="w-3 h-3 text-gray-400" />
                        <span className="text-xs text-gray-500">
                          {new Date(session.updatedAt).toLocaleDateString('zh-CN')}
                        </span>
                      </div>
                    </div>
                    
                    {/* 操作按钮 */}
                    <div className="flex items-center gap-1 opacity-0 group-hover:opacity-100 transition-opacity">
                      <button
                        onClick={(e) => handleArchive(session.id, e)}
                        className="p-1 hover:bg-gray-200 rounded"
                        title="归档"
                      >
                        <Archive className="w-3 h-3 text-gray-500" />
                      </button>
                      <button
                        onClick={(e) => handleDelete(session.id, e)}
                        className="p-1 hover:bg-red-100 rounded"
                        title="删除"
                      >
                        <Trash2 className="w-3 h-3 text-red-500" />
                      </button>
                    </div>
                  </div>
                </div>
              ))}
            </div>
          )}
        </div>

        {/* Footer */}
        <div className="p-3 border-t border-gray-200 text-center text-xs text-gray-500">
          共 {filteredSessions.length} 个会话
        </div>
      </div>
    </div>
  );
};
