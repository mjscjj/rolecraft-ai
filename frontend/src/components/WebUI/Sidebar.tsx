import React, { useState, useMemo } from 'react';
import type { ChatSession } from '../../api/chat';

interface SidebarProps {
  sessions: ChatSession[];
  currentSessionId?: string;
  isCollapsed: boolean;
  onNewChat: () => void;
  onSelectSession: (sessionId: string) => void;
  onRenameSession: (sessionId: string, newTitle: string) => void;
  onDeleteSession: (sessionId: string) => void;
  onArchiveSession: (sessionId: string, isArchived: boolean) => void;
  onToggleCollapse: () => void;
  onSearch?: (query: string) => void;
}

interface ChatGroup {
  title: string;
  sessions: ChatSession[];
}

const Sidebar: React.FC<SidebarProps> = ({
  sessions,
  currentSessionId,
  isCollapsed,
  onNewChat,
  onSelectSession,
  onRenameSession,
  onDeleteSession,
  onArchiveSession,
  onToggleCollapse,
  onSearch,
}) => {
  const [searchQuery, setSearchQuery] = useState('');
  const [editingId, setEditingId] = useState<string | null>(null);
  const [editTitle, setEditTitle] = useState('');

  // Group sessions by time
  const groupedSessions = useMemo(() => {
    const now = new Date();
    const today = new Date(now.getFullYear(), now.getMonth(), now.getDate());
    const yesterday = new Date(today);
    yesterday.setDate(yesterday.getDate() - 1);
    const lastWeek = new Date(today);
    lastWeek.setDate(lastWeek.getDate() - 7);

    const groups: ChatGroup[] = [
      { title: '今天', sessions: [] },
      { title: '昨天', sessions: [] },
      { title: '更早', sessions: [] },
    ];

    sessions.forEach((session) => {
      const updatedAt = new Date(session.updatedAt);
      
      if (updatedAt >= today) {
        groups[0].sessions.push(session);
      } else if (updatedAt >= yesterday) {
        groups[1].sessions.push(session);
      } else {
        groups[2].sessions.push(session);
      }
    });

    // Filter out empty groups and sort sessions within groups
    return groups
      .filter((group) => group.sessions.length > 0)
      .map((group) => ({
        ...group,
        sessions: group.sessions.sort(
          (a, b) => new Date(b.updatedAt).getTime() - new Date(a.updatedAt).getTime()
        ),
      }));
  }, [sessions]);

  // Filter sessions by search query
  const filteredGroups = useMemo(() => {
    if (!searchQuery.trim()) {
      return groupedSessions;
    }

    const query = searchQuery.toLowerCase();
    return groupedSessions
      .map((group) => ({
        ...group,
        sessions: group.sessions.filter(
          (session) =>
            session.title.toLowerCase().includes(query) ||
            session.mode.toLowerCase().includes(query)
        ),
      }))
      .filter((group) => group.sessions.length > 0);
  }, [groupedSessions, searchQuery]);

  const handleStartEdit = (session: ChatSession) => {
    setEditingId(session.id);
    setEditTitle(session.title);
  };

  const handleSaveEdit = (sessionId: string) => {
    if (editTitle.trim()) {
      onRenameSession(sessionId, editTitle.trim());
    }
    setEditingId(null);
    setEditTitle('');
  };

  const handleCancelEdit = () => {
    setEditingId(null);
    setEditTitle('');
  };

  const handleKeyDown = (e: React.KeyboardEvent, sessionId: string) => {
    if (e.key === 'Enter') {
      handleSaveEdit(sessionId);
    } else if (e.key === 'Escape') {
      handleCancelEdit();
    }
  };

  const getSessionIcon = (mode: string) => {
    return mode === 'task' ? '📋' : '💬';
  };

  if (isCollapsed) {
    return (
      <div className="webui-sidebar collapsed">
        <div style={{ padding: '16px', textAlign: 'center' }}>
          <button
            onClick={onToggleCollapse}
            style={{
              width: '40px',
              height: '40px',
              background: 'var(--bg-hover)',
              border: 'none',
              borderRadius: '8px',
              cursor: 'pointer',
              display: 'flex',
              alignItems: 'center',
              justifyContent: 'center',
              fontSize: '20px',
            }}
            title="展开侧边栏"
          >
            ➡️
          </button>
        </div>
      </div>
    );
  }

  return (
    <div className="webui-sidebar">
      {/* Sidebar Header */}
      <div className="sidebar-header">
        <button className="new-chat-btn" onClick={onNewChat}>
          <span>+</span>
          <span>新建对话</span>
        </button>
      </div>

      {/* Search */}
      <div className="sidebar-search">
        <input
          type="text"
          className="search-input"
          placeholder="搜索对话..."
          value={searchQuery}
          onChange={(e) => {
            setSearchQuery(e.target.value);
            onSearch?.(e.target.value);
          }}
        />
      </div>

      {/* Session List */}
      <div className="sidebar-content">
        {filteredGroups.length === 0 ? (
          <div
            style={{
              padding: '20px',
              textAlign: 'center',
              color: 'var(--text-tertiary)',
              fontSize: '14px',
            }}
          >
            {searchQuery ? '没有找到匹配的对话' : '暂无对话历史'}
          </div>
        ) : (
          filteredGroups.map((group) => (
            <div key={group.title} className="chat-group">
              <div className="chat-group-title">{group.title}</div>
              <div className="chat-list">
                {group.sessions.map((session) => (
                  <div
                    key={session.id}
                    className={`chat-item ${session.id === currentSessionId ? 'active' : ''}`}
                    onClick={() => onSelectSession(session.id)}
                  >
                    <span className="chat-item-icon">{getSessionIcon(session.mode)}</span>
                    
                    {editingId === session.id ? (
                      <input
                        type="text"
                        value={editTitle}
                        onChange={(e) => setEditTitle(e.target.value)}
                        onBlur={() => handleSaveEdit(session.id)}
                        onKeyDown={(e) => handleKeyDown(e, session.id)}
                        onClick={(e) => e.stopPropagation()}
                        autoFocus
                        style={{
                          flex: 1,
                          padding: '4px 8px',
                          background: 'var(--bg-primary)',
                          border: '1px solid var(--accent-color)',
                          borderRadius: '4px',
                          fontSize: '14px',
                          color: 'var(--text-primary)',
                        }}
                      />
                    ) : (
                      <span className="chat-item-title" title={session.title}>
                        {session.title}
                      </span>
                    )}

                    {/* Actions */}
                    <div className="chat-item-actions">
                      <button
                        className="chat-item-action"
                        onClick={(e) => {
                          e.stopPropagation();
                          handleStartEdit(session);
                        }}
                        title="重命名"
                      >
                        ✏️
                      </button>
                      <button
                        className="chat-item-action"
                        onClick={(e) => {
                          e.stopPropagation();
                          onArchiveSession(session.id, true);
                        }}
                        title="归档"
                      >
                        📦
                      </button>
                      <button
                        className="chat-item-action delete"
                        onClick={(e) => {
                          e.stopPropagation();
                          if (confirm('确定要删除这个对话吗？')) {
                            onDeleteSession(session.id);
                          }
                        }}
                        title="删除"
                      >
                        🗑️
                      </button>
                    </div>
                  </div>
                ))}
              </div>
            </div>
          ))
        )}
      </div>

      {/* Sidebar Footer */}
      <div className="sidebar-footer">
        <button
          className="theme-toggle"
          onClick={onToggleCollapse}
          title="折叠侧边栏"
        >
          <span>☰ 折叠侧边栏</span>
          <span>⬅️</span>
        </button>
      </div>
    </div>
  );
};

export default Sidebar;
