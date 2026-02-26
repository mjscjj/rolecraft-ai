import type { FC } from 'react';
import { useState, useRef, useEffect } from 'react';
import { Send, Paperclip, MoreVertical, Copy, RotateCcw, ThumbsUp, ThumbsDown, ChevronDown, Sparkles } from 'lucide-react';
import type { Message } from '../types';
import { roleApi } from '../api/role';
import { chatApi } from '../api/chat';
import './Chat.css';

const API_BASE = 'http://localhost:8080/api/v1';

interface ChatProps {
  roleId?: string;
  roleName?: string;
}

export const Chat: FC<ChatProps> = ({ roleId, roleName = 'AI 助手' }) => {
  const [messages, setMessages] = useState<Message[]>([]);
  const [inputValue, setInputValue] = useState('');
  const [isLoading, setIsLoading] = useState(false);
  const [sessionId, setSessionId] = useState<string | null>(null);
  const [currentRoleId, setCurrentRoleId] = useState<string>(roleId || '');
  const [currentRoleName, setCurrentRoleName] = useState<string>(roleName);
  const [availableRoles, setAvailableRoles] = useState<any[]>([]);
  const [showRoleSelector, setShowRoleSelector] = useState(false);
  const [isSwitching, setIsSwitching] = useState(false);
  const messagesEndRef = useRef<HTMLDivElement>(null);
  const roleSelectorRef = useRef<HTMLDivElement>(null);

  const scrollToBottom = () => {
    messagesEndRef.current?.scrollIntoView({ behavior: 'smooth' });
  };

  useEffect(() => {
    scrollToBottom();
  }, [messages]);

  // 点击外部关闭角色选择器
  useEffect(() => {
    const handleClickOutside = (event: MouseEvent) => {
      if (roleSelectorRef.current && !roleSelectorRef.current.contains(event.target as Node)) {
        setShowRoleSelector(false);
      }
    };

    if (showRoleSelector) {
      document.addEventListener('mousedown', handleClickOutside);
    }

    return () => {
      document.removeEventListener('mousedown', handleClickOutside);
    };
  }, [showRoleSelector]);

  // 加载可用角色列表
  useEffect(() => {
    const loadRoles = async () => {
      try {
        const token = localStorage.getItem('token');
        if (!token) return;

        const roles = await roleApi.list();
        setAvailableRoles(roles);
      } catch (err) {
        console.error('Failed to load roles:', err);
      }
    };

    loadRoles();
  }, []);

  // 初始化会话
  useEffect(() => {
    const initSession = async () => {
      const token = localStorage.getItem('token');
      if (!token || !roleId) return;

      try {
        const res = await fetch(`${API_BASE}/chat-sessions`, {
          method: 'POST',
          headers: {
            'Content-Type': 'application/json',
            'Authorization': `Bearer ${token}`,
          },
          body: JSON.stringify({ roleId, mode: 'quick' }),
        });
        const data = await res.json();
        if (data.data?.id) {
          setSessionId(data.data.id);
          setCurrentRoleId(roleId);
          // 加载欢迎消息
          if (data.data.role?.welcomeMessage) {
            setMessages([{
              id: 'welcome',
              role: 'assistant',
              content: data.data.role.welcomeMessage,
              createdAt: new Date().toISOString(),
            }]);
          }
        }
      } catch (err) {
        console.error('Failed to create session:', err);
      }
    };

    initSession();
  }, [roleId]);

  const handleSend = async () => {
    if (!inputValue.trim() || !sessionId || isLoading) return;

    const token = localStorage.getItem('token');
    if (!token) {
      alert('请先登录');
      return;
    }

    const newMessage: Message = {
      id: Date.now().toString(),
      role: 'user',
      content: inputValue,
      createdAt: new Date().toISOString(),
    };

    setMessages(prev => [...prev, newMessage]);
    setInputValue('');
    setIsLoading(true);

    try {
      const res = await fetch(`${API_BASE}/chat/${sessionId}/complete`, {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
          'Authorization': `Bearer ${token}`,
        },
        body: JSON.stringify({ content: newMessage.content }),
      });

      const data = await res.json();
      
      if (data.data?.assistantMessage) {
        const aiResponse: Message = {
          id: (Date.now() + 1).toString(),
          role: 'assistant',
          content: data.data.assistantMessage.content,
          createdAt: new Date().toISOString(),
        };
        setMessages(prev => [...prev, aiResponse]);
      }
    } catch (err) {
      console.error('Failed to send message:', err);
      const errorMessage: Message = {
        id: (Date.now() + 1).toString(),
        role: 'assistant',
        content: '抱歉，发生了错误。请稍后重试。',
        createdAt: new Date().toISOString(),
      };
      setMessages(prev => [...prev, errorMessage]);
    } finally {
      setIsLoading(false);
    }
  };

  const handleCopy = (content: string) => {
    navigator.clipboard.writeText(content);
  };

  const handleSwitchRole = async (newRoleId: string) => {
    if (!sessionId || newRoleId === currentRoleId) return;

    const token = localStorage.getItem('token');
    if (!token) {
      alert('请先登录');
      return;
    }

    setIsSwitching(true);
    setShowRoleSelector(false);

    try {
      const result = await chatApi.switchRole(sessionId, newRoleId);
      
      // 更新当前角色
      setCurrentRoleId(result.newRoleId);
      setCurrentRoleName(result.newRoleName);
      
      // 添加系统消息
      const systemMessage: Message = {
        id: `system-${Date.now()}`,
        role: 'system',
        content: `已切换到角色：${result.newRoleName}`,
        createdAt: new Date().toISOString(),
      };
      setMessages(prev => [...prev, systemMessage]);
      
    } catch (err) {
      console.error('Failed to switch role:', err);
      alert('切换角色失败，请重试');
    } finally {
      setIsSwitching(false);
    }
  };

  return (
    <div className="h-[calc(100vh-8rem)] flex flex-col">
      {/* Chat Header */}
      <div className="flex items-center justify-between pb-4 border-b border-slate-200">
        <div className="flex items-center gap-3">
          <div className={`w-10 h-10 rounded-full bg-gradient-to-br from-primary to-primary-dark flex items-center justify-center text-white font-semibold transition-all duration-300 role-avatar ${isSwitching ? 'switching' : ''}`}>
            {currentRoleName[0]}
          </div>
          <div>
            <div className="flex items-center gap-2">
              <h2 className="font-semibold text-slate-900">{currentRoleName}</h2>
              {sessionId && (
                <div className="relative" ref={roleSelectorRef}>
                  <button
                    onClick={() => setShowRoleSelector(!showRoleSelector)}
                    disabled={isSwitching}
                    className="flex items-center gap-1 px-2 py-1 text-xs bg-slate-100 hover:bg-slate-200 rounded-md transition-colors disabled:opacity-50"
                  >
                    <Sparkles className="w-3 h-3 text-primary" />
                    <span>切换</span>
                    <ChevronDown className={`w-3 h-3 transition-transform ${showRoleSelector ? 'rotate-180' : ''}`} />
                  </button>
                  
                  {/* Role Selector Dropdown */}
                  {showRoleSelector && (
                    <div className="absolute top-full left-0 mt-1 w-48 bg-white rounded-lg shadow-xl border border-slate-200 py-1 z-50 role-selector-dropdown">
                      <div className="px-3 py-2 text-xs font-medium text-slate-500 border-b border-slate-100">
                        选择角色
                      </div>
                      <div className="max-h-64 overflow-y-auto">
                        {availableRoles.map((role) => (
                          <button
                            key={role.id}
                            onClick={() => handleSwitchRole(role.id)}
                            disabled={isSwitching || role.id === currentRoleId}
                            className={`w-full text-left px-3 py-2.5 text-sm hover:bg-slate-50 transition-colors flex items-center gap-2 disabled:opacity-50 ${
                              role.id === currentRoleId ? 'bg-primary/5 text-primary' : ''
                            }`}
                          >
                            <div className="w-6 h-6 rounded-full bg-gradient-to-br from-primary to-primary-dark flex items-center justify-center text-white text-xs flex-shrink-0">
                              {role.name[0]}
                            </div>
                            <div className="flex-1 min-w-0">
                              <div className="font-medium truncate">{role.name}</div>
                              <div className="text-xs text-slate-500 truncate">{role.category}</div>
                            </div>
                            {role.id === currentRoleId && (
                              <div className="w-2 h-2 rounded-full bg-primary flex-shrink-0" />
                            )}
                          </button>
                        ))}
                      </div>
                    </div>
                  )}
                </div>
              )}
            </div>
            <p className="text-xs text-slate-500">
              {sessionId ? '🟢 在线' : '🟡 连接中...'}
            </p>
          </div>
        </div>
        <div className="flex items-center gap-2">
          <button className="p-2 text-slate-500 hover:bg-slate-100 rounded-lg transition-colors">
            历史记录
          </button>
          <button className="p-2 text-slate-500 hover:bg-slate-100 rounded-lg transition-colors">
            <MoreVertical className="w-5 h-5" />
          </button>
        </div>
      </div>

      {/* Messages */}
      <div className="flex-1 overflow-y-auto py-6 space-y-6">
        {messages.length === 0 && !isLoading && (
          <div className="flex flex-col items-center justify-center h-full text-slate-400">
            <p className="text-lg">👋 开始和 {currentRoleName} 对话吧！</p>
            <p className="text-sm mt-2">输入问题或需求，AI 将为你提供帮助</p>
          </div>
        )}

        {messages.map((message) => (
          <div
            key={message.id}
            className={`flex gap-4 ${
              message.role === 'user' ? 'flex-row-reverse' : 
              message.role === 'system' ? 'justify-center' : ''
            }`}
          >
            {message.role === 'system' ? (
              <div className="flex items-center gap-2 py-2 system-message">
                <div className="h-px w-12 bg-slate-200" />
                <div className="px-3 py-1 bg-slate-100 rounded-full text-xs text-slate-500 font-medium">
                  {message.content}
                </div>
                <div className="h-px w-12 bg-slate-200" />
              </div>
            ) : (
              <>
                <div className={`w-10 h-10 rounded-full flex items-center justify-center flex-shrink-0 ${
                  message.role === 'user'
                    ? 'bg-slate-200 text-slate-600'
                    : 'bg-gradient-to-br from-primary to-primary-dark text-white'
                }`}>
                  {message.role === 'user' ? 'U' : 'AI'}
                </div>

                <div className={`max-w-[70%] ${message.role === 'user' ? 'items-end' : 'items-start'}`}>
                  <div className={`p-4 rounded-2xl ${
                    message.role === 'user'
                      ? 'bg-slate-900 text-white rounded-tr-none'
                      : 'bg-white border border-slate-200 rounded-tl-none shadow-sm'
                  }`}>
                    <div className={`prose prose-sm max-w-none ${
                      message.role === 'user' ? 'prose-invert' : ''
                    }`}>
                      {message.content.split('\n').map((line, i) => (
                        <p key={i} className={line.trim() === '' ? 'h-2' : ''}>
                          {line}
                        </p>
                      ))}
                    </div>

                    {message.sources && message.sources.length > 0 && (
                      <div className="mt-4 pt-3 border-t border-slate-200/20">
                        <p className="text-xs opacity-70 mb-2">📚 参考来源：</p>
                        <div className="flex flex-wrap gap-2">
                          {message.sources.map((source, i) => (
                            <span
                              key={i}
                              className="text-xs px-2 py-1 bg-white/10 rounded cursor-pointer hover:bg-white/20 transition-colors"
                            >
                              {source}
                            </span>
                          ))}
                        </div>
                      </div>
                    )}
                  </div>

                  {message.role === 'assistant' && (
                    <div className="flex items-center gap-1 mt-2 opacity-0 hover:opacity-100 transition-opacity">
                      <button 
                        onClick={() => handleCopy(message.content)}
                        className="p-1.5 text-slate-400 hover:text-slate-600 hover:bg-slate-100 rounded transition-colors" 
                        title="复制"
                      >
                        <Copy className="w-4 h-4" />
                      </button>
                      <button className="p-1.5 text-slate-400 hover:text-slate-600 hover:bg-slate-100 rounded transition-colors" title="重新生成">
                        <RotateCcw className="w-4 h-4" />
                      </button>
                      <button className="p-1.5 text-slate-400 hover:text-green-500 hover:bg-green-50 rounded transition-colors" title="有用">
                        <ThumbsUp className="w-4 h-4" />
                      </button>
                      <button className="p-1.5 text-slate-400 hover:text-red-500 hover:bg-red-50 rounded transition-colors" title="无用">
                        <ThumbsDown className="w-4 h-4" />
                      </button>
                    </div>
                  )}
                </div>
              </>
            )}
          </div>
        ))}

        {/* Loading Indicator */}
        {isLoading && (
          <div className="flex gap-4">
            <div className="w-10 h-10 rounded-full bg-gradient-to-br from-primary to-primary-dark flex items-center justify-center text-white">
              AI
            </div>
            <div className="bg-white border border-slate-200 rounded-2xl rounded-tl-none p-4 shadow-sm">
              <div className="flex gap-1">
                <span className="w-2 h-2 bg-slate-400 rounded-full animate-bounce" style={{ animationDelay: '0ms' }} />
                <span className="w-2 h-2 bg-slate-400 rounded-full animate-bounce" style={{ animationDelay: '150ms' }} />
                <span className="w-2 h-2 bg-slate-400 rounded-full animate-bounce" style={{ animationDelay: '300ms' }} />
              </div>
            </div>
          </div>
        )}

        <div ref={messagesEndRef} />
      </div>

      {/* Input Area */}
      <div className="border-t border-slate-200 pt-4">
        {/* Input */}
        <div className="flex items-end gap-2 bg-white border border-slate-200 rounded-2xl p-2 shadow-sm">
          <button className="p-3 text-slate-400 hover:text-slate-600 hover:bg-slate-100 rounded-xl transition-colors">
            <Paperclip className="w-5 h-5" />
          </button>
          <textarea
            value={inputValue}
            onChange={(e) => setInputValue(e.target.value)}
            onKeyDown={(e) => {
              if (e.key === 'Enter' && !e.shiftKey) {
                e.preventDefault();
                handleSend();
              }
            }}
            placeholder={sessionId ? "输入消息..." : "正在连接..."}
            disabled={!sessionId || isLoading}
            rows={1}
            className="flex-1 py-3 px-2 outline-none resize-none max-h-32 disabled:opacity-50"
            style={{ minHeight: '48px' }}
          />
          <button
            onClick={handleSend}
            disabled={!inputValue.trim() || isLoading || !sessionId}
            className="p-3 bg-primary text-white rounded-xl hover:bg-primary-dark transition-colors disabled:opacity-50 disabled:cursor-not-allowed"
          >
            <Send className="w-5 h-5" />
          </button>
        </div>
        <p className="text-xs text-slate-400 mt-2 text-center">按 Enter 发送，Shift + Enter 换行</p>
      </div>
    </div>
  );
};
