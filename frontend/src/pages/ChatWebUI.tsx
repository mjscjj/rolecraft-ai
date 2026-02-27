import React, { useState, useRef, useEffect } from 'react';
import { useParams } from 'react-router-dom';
import Sidebar from '../components/WebUI/Sidebar';
import Message from '../components/WebUI/Message';
import ChatInput from '../components/WebUI/ChatInput';
import { chatApi } from '../api/chat';
import { documentApi } from '../api/document';
import type { Document } from '../api/document';
import roleApi from '../api/role';
import { useChatStore } from '../stores/chatStore';
import '../styles/webui.css';

interface ChatWebUIProps {
  initialRoleId?: string;
}

const ChatWebUI: React.FC<ChatWebUIProps> = ({ initialRoleId }) => {
  const { roleId: routeRoleId } = useParams<{ roleId: string }>();
  const [selectedRoleId, setSelectedRoleId] = useState('');
  const effectiveRoleId = initialRoleId || routeRoleId || selectedRoleId;

  // State from store
  const {
    sessions,
    currentSession,
    messages,
    isLoading,
    isStreaming,
    thinkingSteps,
    error,
    fetchSessions,
    createSession,
    fetchSession,
    sendStreamMessage,
    sendStreamMessageWithThinking,
    updateSessionConfig,
    cancelStream,
    retryLastStream,
    clearError,
  } = useChatStore();

  // Local state
  const [isSidebarCollapsed, setIsSidebarCollapsed] = useState(false);
  const [isDarkTheme, setIsDarkTheme] = useState(false);
  const [selectedModel, setSelectedModel] = useState('qwen-plus');
  const [selectedKnowledge, setSelectedKnowledge] = useState('none');
  const [temperature, setTemperature] = useState(0.7);
  const [chatMode, setChatMode] = useState<'normal' | 'deep'>('normal');
  const [knowledgeFolders, setKnowledgeFolders] = useState<Array<{ id: string; name: string }>>([]);
  const [availableRoles, setAvailableRoles] = useState<Array<{ id: string; name: string }>>([]);
  const [isUploadingFiles, setIsUploadingFiles] = useState(false);
  const [pendingAttachmentNames, setPendingAttachmentNames] = useState<string[]>([]);
  const [showWelcome, setShowWelcome] = useState(true);
  const [userInitials, setUserInitials] = useState('U');

  const messagesEndRef = useRef<HTMLDivElement>(null);
  const lastSavedConfigRef = useRef('');

  // Initialize
  useEffect(() => {
    fetchSessions();
    
    // Get user info for avatar
    let userName = 'User';
    const userRaw = localStorage.getItem('user');
    if (userRaw) {
      try {
        const parsedUser = JSON.parse(userRaw);
        userName = parsedUser?.name || 'User';
      } catch {
        userName = 'User';
      }
    }
    setUserInitials(userName.split(' ').map((n: string) => n[0]).join('').toUpperCase().slice(0, 2));

    // Check theme preference
    const savedTheme = localStorage.getItem('theme');
    if (savedTheme === 'dark') {
      setIsDarkTheme(true);
      document.documentElement.setAttribute('data-theme', 'dark');
    }

    const preferredModel = localStorage.getItem('preferredModel');
    const preferredTemperature = localStorage.getItem('preferredTemperature');
    const preferredKnowledge = localStorage.getItem('preferredKnowledgeScope');
    const preferredChatMode = localStorage.getItem('preferredChatMode');
    if (preferredModel) setSelectedModel(preferredModel);
    if (preferredTemperature) setTemperature(Number(preferredTemperature));
    if (preferredKnowledge) setSelectedKnowledge(preferredKnowledge);
    if (preferredChatMode === 'deep' || preferredChatMode === 'normal') {
      setChatMode(preferredChatMode);
    }

    documentApi.listFolders()
      .then((folders) => {
        setKnowledgeFolders(folders.map((folder) => ({ id: folder.id, name: folder.name })));
      })
      .catch((err) => {
        console.warn('Failed to load knowledge folders:', err);
      });

    roleApi.list()
      .then((roles) => {
        const normalizedRoles = roles.map((role) => ({ id: role.id, name: role.name }));
        setAvailableRoles(normalizedRoles);
        if (!initialRoleId && !routeRoleId && normalizedRoles.length > 0) {
          setSelectedRoleId(normalizedRoles[0].id);
        }
      })
      .catch((err) => {
        console.warn('Failed to load roles:', err);
      });
  }, []);

  const buildSessionConfig = () => ({
    preferredModel: selectedModel,
    preferredTemperature: temperature,
    knowledgeScope: selectedKnowledge,
    chatMode,
    customAPIKey: localStorage.getItem('customAPIKey') || '',
  });

  // Auto-scroll to bottom when new messages arrive
  useEffect(() => {
    messagesEndRef.current?.scrollIntoView({ behavior: 'smooth' });
  }, [messages]);

  // Update session when currentSession changes
  useEffect(() => {
    if (currentSession) {
      setShowWelcome(false);
      const sessionConfig = currentSession.modelConfig || {};
      const preferredModel = typeof sessionConfig.preferredModel === 'string' ? sessionConfig.preferredModel : '';
      const preferredTemperature = Number(sessionConfig.preferredTemperature);
      const knowledgeScope = typeof sessionConfig.knowledgeScope === 'string' ? sessionConfig.knowledgeScope : '';
      const sessionChatMode = sessionConfig.chatMode === 'deep' ? 'deep' : 'normal';

      if (preferredModel) {
        setSelectedModel(preferredModel);
      }
      if (!Number.isNaN(preferredTemperature) && preferredTemperature >= 0 && preferredTemperature <= 2) {
        setTemperature(preferredTemperature);
      }
      if (knowledgeScope) {
        setSelectedKnowledge(knowledgeScope);
      }
      setChatMode(sessionChatMode);

      const initialConfig = JSON.stringify({
        preferredModel: preferredModel || selectedModel,
        preferredTemperature: !Number.isNaN(preferredTemperature) && preferredTemperature >= 0 && preferredTemperature <= 2
          ? preferredTemperature
          : temperature,
        knowledgeScope: knowledgeScope || selectedKnowledge,
        chatMode: sessionChatMode || chatMode,
        customAPIKey: localStorage.getItem('customAPIKey') || '',
      });
      lastSavedConfigRef.current = initialConfig;
    } else {
      setShowWelcome(true);
    }
  }, [currentSession]);

  useEffect(() => {
    localStorage.setItem('preferredModel', selectedModel);
    localStorage.setItem('preferredTemperature', String(temperature));
    localStorage.setItem('preferredKnowledgeScope', selectedKnowledge);
    localStorage.setItem('preferredChatMode', chatMode);

    if (!currentSession) return;

    const nextConfig = buildSessionConfig();
    const nextConfigText = JSON.stringify(nextConfig);
    if (nextConfigText === lastSavedConfigRef.current) {
      return;
    }

    const timer = setTimeout(async () => {
      try {
        await updateSessionConfig(nextConfig);
        lastSavedConfigRef.current = nextConfigText;
      } catch (error) {
        console.error('Failed to update session config:', error);
      }
    }, 300);

    return () => clearTimeout(timer);
  }, [selectedModel, selectedKnowledge, temperature, chatMode, currentSession?.id]);

  // Handle new chat
  const handleNewChat = async () => {
    if (effectiveRoleId) {
      try {
        await createSession(effectiveRoleId, '新对话', 'quick', buildSessionConfig());
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
    const attachmentPrefix = pendingAttachmentNames.length
      ? `本轮新增知识文档：${pendingAttachmentNames.join('、')}\n请结合这些文档回答。\n\n`
      : '';
    const composedContent = `${attachmentPrefix}${content}`;

    const sendByMode = async (text: string) => {
      if (chatMode === 'deep') {
        await sendStreamMessageWithThinking(text);
      } else {
        await sendStreamMessage(text);
      }
    };

    if (!currentSession) {
      // Create new session if none exists
      if (effectiveRoleId) {
        try {
          await createSession(effectiveRoleId, content.slice(0, 30) + '...', 'quick', buildSessionConfig());
          // Message will be sent after session is created
          setTimeout(() => {
            sendByMode(composedContent);
          }, 100);
          setPendingAttachmentNames([]);
        } catch (error) {
          console.error('Failed to create session:', error);
        }
      } else {
        alert('请先选择一个 AI 角色');
      }
    } else {
      await sendByMode(composedContent);
      setPendingAttachmentNames([]);
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
  const handleAttachFiles = async (files: FileList) => {
    const fileArray = Array.from(files);
    if (fileArray.length === 0) return;

    setIsUploadingFiles(true);
    try {
      const results = await Promise.allSettled(fileArray.map((file) => documentApi.uploadWithFolder(file)));
      const successCount = results.filter((result) => result.status === 'fulfilled').length;
      const failCount = results.length - successCount;
      const uploadedNames: string[] = [];

      results.forEach((result) => {
        if (result.status !== 'fulfilled') return;
        const payload = result.value;
        const docs = Array.isArray(payload) ? payload : [payload];
        docs.forEach((doc) => {
          const typedDoc = doc as Document;
          if (typedDoc?.name) {
            uploadedNames.push(typedDoc.name);
          }
        });
      });

      if (successCount > 0) {
        setSelectedKnowledge('all');
        if (uploadedNames.length > 0) {
          setPendingAttachmentNames(uploadedNames);
        }
      }

      if (failCount > 0) {
        alert(`上传完成：成功 ${successCount} 个，失败 ${failCount} 个`);
      } else {
        alert(`上传成功，共 ${successCount} 个文件`);
      }
    } catch (error) {
      console.error('Failed to upload files:', error);
      alert('文件上传失败');
    } finally {
      setIsUploadingFiles(false);
    }
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
          {!initialRoleId && !routeRoleId && (
            <select
              className="model-selector"
              value={selectedRoleId}
              onChange={(e) => setSelectedRoleId(e.target.value)}
              title="选择角色"
            >
              {availableRoles.length === 0 && (
                <option value="">暂无可用角色</option>
              )}
              {availableRoles.map((role) => (
                <option key={role.id} value={role.id}>
                  {role.name}
                </option>
              ))}
            </select>
          )}

          {/* Model Selector */}
          <select
            className="model-selector"
            value={chatMode}
            onChange={(e) => setChatMode(e.target.value as 'normal' | 'deep')}
            title="对话模式"
          >
            <option value="normal">标准模式</option>
            <option value="deep">深度思考(搜索)</option>
          </select>

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
            <option value="all">全部文档</option>
            {knowledgeFolders.map((folder) => (
              <option key={folder.id} value={`folder:${folder.id}`}>
                文件夹: {folder.name}
              </option>
            ))}
          </select>
          <input
            type="number"
            min="0"
            max="2"
            step="0.1"
            value={temperature}
            onChange={(e) => setTemperature(Number(e.target.value))}
            title="温度参数"
            style={{ width: 72 }}
            className="model-selector"
          />
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
          {isStreaming && (
            <button className="icon-button" onClick={cancelStream} title="停止生成">
              ⏹️
            </button>
          )}
          {!isStreaming && error && (
            <button className="icon-button" onClick={retryLastStream} title="重试上一次发送">
              🔁
            </button>
          )}

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
                  {thinkingSteps.length > 0 && (
                    <div
                      style={{
                        margin: '8px 0 16px',
                        padding: '12px 14px',
                        borderRadius: '10px',
                        border: '1px solid var(--border-color)',
                        background: 'var(--bg-secondary)',
                      }}
                    >
                      <div style={{ fontSize: 13, color: 'var(--text-secondary)', marginBottom: 8 }}>
                        深度思考过程
                      </div>
                      <ol style={{ margin: 0, paddingLeft: 18, color: 'var(--text-primary)', fontSize: 13 }}>
                        {thinkingSteps.map((step, index) => (
                          <li key={`${step}-${index}`} style={{ marginBottom: 6 }}>
                            {step}
                          </li>
                        ))}
                      </ol>
                    </div>
                  )}
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
              {pendingAttachmentNames.length > 0 && (
                <div
                  style={{
                    padding: '8px 20px 0',
                    color: 'var(--text-secondary)',
                    fontSize: 13,
                  }}
                >
                  本轮将携带文档上下文: {pendingAttachmentNames.join('、')}
                </div>
              )}
              <ChatInput
                onSend={handleSendMessage}
                onAttach={handleAttachFiles}
                onVoiceInput={handleVoiceInput}
                disabled={isLoading || isUploadingFiles}
                isStreaming={isStreaming}
                placeholder={
                  isUploadingFiles
                    ? '文件上传中...'
                    : isStreaming
                      ? 'AI 正在思考...'
                      : '输入消息... (Shift+Enter 换行)'
                }
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
