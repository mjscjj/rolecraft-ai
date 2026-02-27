import React, { useState, useRef, useEffect } from 'react';
import Sidebar from '../components/WebUI/Sidebar';
import Message from '../components/WebUI/Message';
import ChatInput from '../components/WebUI/ChatInput';
import { chatApi, type ChatSession, type Message as MessageType } from '../api/chat';
import { useChatStore } from '../stores/chatStore';
import '../styles/webui.css';

interface ChatWebUIProps {
  initialRoleId?: string;
}

const ChatWebUI: React.FC<ChatWebUIProps> = ({ initialRoleId }) => {
  // State from store
  const {
    sessions,
    currentSession,
    messages,
    isLoading,
    isStreaming,
    error,
    fetchSessions,
    createSession,
    fetchSession,
    sendStreamMessage,
    clearError,
  } = useChatStore();

  // Local state
  const [isSidebarCollapsed, setIsSidebarCollapsed] = useState(false);
  const [isDarkTheme, setIsDarkTheme] = useState(false);
  const [selectedModel, setSelectedModel] = useState('qwen-plus');
  const [selectedKnowledge, setSelectedKnowledge] = useState('none');
  const [showWelcome, setShowWelcome] = useState(true);
  const [userInitials, setUserInitials] = useState('U');

  const messagesEndRef = useRef<HTMLDivElement>(null);

  // Initialize
  useEffect(() => {
    fetchSessions();
    
    // Get user info for avatar
    const userName = localStorage.getItem('user_name') || 'User';
    setUserInitials(userName.split(' ').map(n => n[0]).join('').toUpperCase().slice(0, 2));

    // Check theme preference
    const savedTheme = localStorage.getItem('theme');
    if (savedTheme === 'dark') {
      setIsDarkTheme(true);
      document.documentElement.setAttribute('data-theme', 'dark');
    }
  }, []);

  // Auto-scroll to bottom when new messages arrive
  useEffect(() => {
    messagesEndRef.current?.scrollIntoView({ behavior: 'smooth' });
  }, [messages]);

  // Update session when currentSession changes
  useEffect(() => {
    if (currentSession) {
      setShowWelcome(false);
    } else {
      setShowWelcome(true);
    }
  }, [currentSession]);

  // Handle new chat
  const handleNewChat = async () => {
    if (initialRoleId) {
      try {
        await createSession(initialRoleId, '新对话', 'quick');
        setShowWelcome(false);
      } catch (error) {
        console.error('Failed to create session:', error);
      }
    } else {
      // If no initial role, show role selector or create default
      alert('请先选择一个 AI 角色');
    }
  };

  // Handle select session
  const handleSelectSession = async (sessionId: string) => {
    try {
      await fetchSession(sessionId);
      setShowWelcome(false);
    } catch (error) {
      console.error('Failed to fetch session:', error);
    }
  };

  // Handle send message
  const handleSendMessage = async (content: string) => {
    if (!currentSession) {
      // Create new session if none exists
      if (initialRoleId) {
        try {
          await createSession(initialRoleId, content.slice(0, 30) + '...', 'quick');
          // Message will be sent after session is created
          setTimeout(() => sendStreamMessage(content), 100);
        } catch (error) {
          console.error('Failed to create session:', error);
        }
      } else {
        alert('请先选择一个 AI 角色');
      }
    } else {
      await sendStreamMessage(content);
    }
  };

  // Handle rename session
  const handleRenameSession = async (sessionId: string, newTitle: string) => {
    try {
      await chatApi.updateTitle(sessionId, newTitle);
      await fetchSessions();
    } catch (error) {
      console.error('Failed to rename session:', error);
      alert('重命名失败');
    }
  };

  // Handle delete session
  const handleDeleteSession = async (sessionId: string) => {
    try {
      await chatApi.deleteSession(sessionId);
      await fetchSessions();
      if (currentSession?.id === sessionId) {
        setShowWelcome(true);
      }
    } catch (error) {
      console.error('Failed to delete session:', error);
      alert('删除失败');
    }
  };

  // Handle archive session
  const handleArchiveSession = async (sessionId: string, isArchived: boolean) => {
    try {
      await chatApi.archive(sessionId, isArchived);
      await fetchSessions();
    } catch (error) {
      console.error('Failed to archive session:', error);
      alert('归档失败');
    }
  };

  // Handle edit message
  const handleEditMessage = async (messageId: string, newContent: string) => {
    if (!currentSession) return;
    try {
      await chatApi.editMessage(currentSession.id, messageId, newContent);
      await fetchSession(currentSession.id);
    } catch (error) {
      console.error('Failed to edit message:', error);
      alert('编辑失败');
    }
  };

  // Handle regenerate message
  const handleRegenerateMessage = async (messageId: string) => {
    if (!currentSession) return;
    try {
      await chatApi.regenerate(currentSession.id, messageId);
      await fetchSession(currentSession.id);
    } catch (error) {
      console.error('Failed to regenerate message:', error);
      alert('重新生成失败');
    }
  };

  // Handle rate message
  const handleRateMessage = async (messageId: string, rating: 'up' | 'down') => {
    try {
      await chatApi.rateMessage(messageId, rating);
      // Show feedback
      console.log('Rated:', rating);
    } catch (error) {
      console.error('Failed to rate message:', error);
    }
  };

  // Handle copy message
  const handleCopyMessage = (content: string) => {
    // Could show a toast notification here
    console.log('Copied to clipboard');
  };

  // Handle attach files
  const handleAttachFiles = (files: FileList) => {
    console.log('Files to attach:', files);
    // Implement file upload logic
    alert(`已选择 ${files.length} 个文件 (功能开发中)`);
  };

  // Handle voice input
  const handleVoiceInput = () => {
    console.log('Voice input requested');
    // Implement speech recognition
    alert('语音输入功能开发中');
  };

  // Toggle theme
  const toggleTheme = () => {
    const newTheme = !isDarkTheme;
    setIsDarkTheme(newTheme);
    localStorage.setItem('theme', newTheme ? 'dark' : 'light');
    document.documentElement.setAttribute('data-theme', newTheme ? 'dark' : 'light');
  };

  // Export conversation
  const handleExport = async (format: 'markdown' | 'json' | 'pdf' = 'markdown') => {
    if (!currentSession) return;
    try {
      const blob = await chatApi.export(currentSession.id, format);
      const url = URL.createObjectURL(blob);
      const a = document.createElement('a');
      a.href = url;
      a.download = `conversation-${currentSession.id}.${format === 'markdown' ? 'md' : format}`;
      a.click();
      URL.revokeObjectURL(url);
    } catch (error) {
      console.error('Failed to export:', error);
      alert('导出失败');
    }
  };

  // Suggestion cards for welcome screen
  const suggestions = [
    { icon: '💡', text: '解释量子计算的基本原理' },
    { icon: '✍️', text: '帮我写一封商务邮件' },
    { icon: '💻', text: '创建一个 React 组件示例' },
    { icon: '📊', text: '分析这份数据的关键趋势' },
  ];

  return (
    <div className="webui-container">
      {/* Header */}
      <header className="webui-header">
        <div className="header-left">
          <button
            className="sidebar-toggle"
            onClick={() => setIsSidebarCollapsed(!isSidebarCollapsed)}
            title={isSidebarCollapsed ? '展开侧边栏' : '折叠侧边栏'}
          >
            {isSidebarCollapsed ? '➡️' : '⬅️'}
          </button>
          <div className="logo">
            <span className="logo-icon">🎭</span>
            <span>RoleCraft</span>
          </div>
        </div>

        <div className="header-center">
          {/* Model Selector */}
          <select
            className="model-selector"
            value={selectedModel}
            onChange={(e) => setSelectedModel(e.target.value)}
            title="选择模型"
          >
            <option value="qwen-plus">Qwen Plus</option>
            <option value="qwen-max">Qwen Max</option>
            <option value="qwen-turbo">Qwen Turbo</option>
            <option value="gpt-4">GPT-4</option>
            <option value="gpt-3.5-turbo">GPT-3.5</option>
          </select>

          {/* Knowledge Base Selector */}
          <select
            className="knowledge-selector"
            value={selectedKnowledge}
            onChange={(e) => setSelectedKnowledge(e.target.value)}
            title="选择知识库"
          >
            <option value="none">无知识库</option>
            <option value="kb1">产品知识库</option>
            <option value="kb2">技术文档</option>
            <option value="kb3">客服手册</option>
          </select>
        </div>

        <div className="header-right">
          {/* Export Button */}
          <button
            className="icon-button"
            onClick={() => handleExport('markdown')}
            title="导出对话"
            disabled={!currentSession}
          >
            📥
          </button>

          {/* Theme Toggle */}
          <button
            className="icon-button"
            onClick={toggleTheme}
            title={isDarkTheme ? '切换到浅色主题' : '切换到深色主题'}
          >
            {isDarkTheme ? '☀️' : '🌙'}
          </button>

          {/* Settings */}
          <button
            className="icon-button"
            title="设置"
          >
            ⚙️
          </button>

          {/* User Menu */}
          <div className="user-menu" title="用户菜单">
            <div className="user-avatar">{userInitials}</div>
          </div>
        </div>
      </header>

      {/* Main Content */}
      <main className="webui-main">
        {/* Sidebar */}
        <Sidebar
          sessions={sessions}
          currentSessionId={currentSession?.id}
          isCollapsed={isSidebarCollapsed}
          onNewChat={handleNewChat}
          onSelectSession={handleSelectSession}
          onRenameSession={handleRenameSession}
          onDeleteSession={handleDeleteSession}
          onArchiveSession={handleArchiveSession}
          onToggleCollapse={() => setIsSidebarCollapsed(!isSidebarCollapsed)}
        />

        {/* Chat Area */}
        <div className="webui-chat-area">
          {showWelcome || !currentSession ? (
            /* Welcome Screen */
            <div className="welcome-screen">
              <div className="welcome-icon">🤖</div>
              <h1 className="welcome-title">欢迎使用 RoleCraft</h1>
              <p className="welcome-subtitle">
                选择一个 AI 角色开始对话，或从左侧边栏继续之前的对话
              </p>
              
              {suggestions.length > 0 && (
                <div className="suggestion-grid">
                  {suggestions.map((suggestion, index) => (
                    <div
                      key={index}
                      className="suggestion-card"
                      onClick={() => handleSendMessage(suggestion.text)}
                    >
                      <div className="suggestion-icon">{suggestion.icon}</div>
                      <div className="suggestion-text">{suggestion.text}</div>
                    </div>
                  ))}
                </div>
              )}
            </div>
          ) : (
            /* Messages */
            <>
              <div className="messages-container">
                <div className="messages-wrapper">
                  {messages.map((message) => (
                    <Message
                      key={message.id}
                      id={message.id}
                      role={message.role}
                      content={message.content}
                      createdAt={message.createdAt}
                      isStreaming={isStreaming && message.id === messages[messages.length - 1].id && message.role === 'assistant' && !message.content}
                      onEdit={handleEditMessage}
                      onRegenerate={handleRegenerateMessage}
                      onCopy={handleCopyMessage}
                      onRate={handleRateMessage}
                    />
                  ))}
                  
                  {/* Typing Indicator */}
                  {isStreaming && messages.length > 0 && messages[messages.length - 1].role === 'user' && (
                    <div className="message ai">
                      <div className="message-avatar">🤖</div>
                      <div className="message-content">
                        <div className="message-bubble">
                          <div className="typing-indicator">
                            <div className="typing-dot" />
                            <div className="typing-dot" />
                            <div className="typing-dot" />
                          </div>
                        </div>
                      </div>
                    </div>
                  )}
                  
                  <div ref={messagesEndRef} />
                </div>
              </div>

              {/* Input Area */}
              <ChatInput
                onSend={handleSendMessage}
                onAttach={handleAttachFiles}
                onVoiceInput={handleVoiceInput}
                disabled={isLoading}
                isStreaming={isStreaming}
                placeholder={isStreaming ? 'AI 正在思考...' : '输入消息... (Shift+Enter 换行)'}
              />
            </>
          )}
        </div>
      </main>

      {/* Error Toast */}
      {error && (
        <div
          style={{
            position: 'fixed',
            bottom: '20px',
            right: '20px',
            padding: '12px 20px',
            background: 'var(--error-color)',
            color: 'white',
            borderRadius: '8px',
            boxShadow: 'var(--shadow-lg)',
            zIndex: 1000,
            display: 'flex',
            alignItems: 'center',
            gap: '12px',
          }}
        >
          <span>⚠️</span>
          <span>{error}</span>
          <button
            onClick={clearError}
            style={{
              background: 'transparent',
              border: 'none',
              color: 'white',
              cursor: 'pointer',
              fontSize: '18px',
            }}
          >
            ×
          </button>
        </div>
      )}
    </div>
  );
};

export default ChatWebUI;
