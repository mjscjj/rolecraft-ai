import type { FC } from 'react';
import { useState, useRef, useEffect } from 'react';
import { Send, Paperclip, MoreVertical, Copy, RotateCcw, ThumbsUp, ThumbsDown } from 'lucide-react';
import type { Message } from '../types';

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
  const messagesEndRef = useRef<HTMLDivElement>(null);

  const scrollToBottom = () => {
    messagesEndRef.current?.scrollIntoView({ behavior: 'smooth' });
  };

  useEffect(() => {
    scrollToBottom();
  }, [messages]);

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

  return (
    <div className="h-[calc(100vh-8rem)] flex flex-col">
      {/* Chat Header */}
      <div className="flex items-center justify-between pb-4 border-b border-slate-200">
        <div className="flex items-center gap-3">
          <div className="w-10 h-10 rounded-full bg-gradient-to-br from-primary to-primary-dark flex items-center justify-center text-white font-semibold">
            {roleName[0]}
          </div>
          <div>
            <h2 className="font-semibold text-slate-900">{roleName}</h2>
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
            <p className="text-lg">👋 开始和 {roleName} 对话吧！</p>
            <p className="text-sm mt-2">输入问题或需求，AI 将为你提供帮助</p>
          </div>
        )}

        {messages.map((message) => (
          <div
            key={message.id}
            className={`flex gap-4 ${message.role === 'user' ? 'flex-row-reverse' : ''}`}
          >
            {/* Avatar */}
            <div className={`w-10 h-10 rounded-full flex items-center justify-center flex-shrink-0 ${
              message.role === 'user'
                ? 'bg-slate-200 text-slate-600'
                : 'bg-gradient-to-br from-primary to-primary-dark text-white'
            }`}>
              {message.role === 'user' ? 'U' : 'AI'}
            </div>

            {/* Message Content */}
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

                {/* Sources */}
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

              {/* Actions */}
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
