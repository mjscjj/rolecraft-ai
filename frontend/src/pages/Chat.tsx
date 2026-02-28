import { useState, useEffect, useRef } from 'react';
import { useParams, useNavigate } from 'react-router-dom';
import { Send, Bot, User, ArrowLeft, Brain, Eye, EyeOff, Edit2, RotateCcw, ThumbsUp, ThumbsDown, Copy, Download, Trash2, Archive, History } from 'lucide-react';
import ReactMarkdown from 'react-markdown';
import remarkGfm from 'remark-gfm';
import client from '../api/client';
import { ThinkingDisplay } from '../components/Thinking/ThinkingDisplay';
import { ChatHistory } from '../components/ChatHistory';
import type { ThinkingProcess } from '../components/Thinking/ThinkingDisplay';

interface Message {
  id: string;
  role: 'user' | 'assistant';
  content: string;
  timestamp: string;
  likes?: number;
  dislikes?: number;
  isEdited?: boolean;
}

export const Chat = () => {
  const { roleId } = useParams();
  const navigate = useNavigate();
  const [messages, setMessages] = useState<Message[]>([]);
  const [input, setInput] = useState('');
  const [loading, setLoading] = useState(false);
  const [role, setRole] = useState<any>(null);
  const [sessionId, setSessionId] = useState<string>('');
  const [showThinking, setShowThinking] = useState(true);
  const [thinkingProcess, setThinkingProcess] = useState<ThinkingProcess | null>(null);
  const [editingId, setEditingId] = useState<string | null>(null);
  const [editContent, setEditContent] = useState('');
  const [showActions, setShowActions] = useState<string | null>(null);
  const [showHistory, setShowHistory] = useState(false);
  const messagesEndRef = useRef<HTMLDivElement>(null);

  useEffect(() => {
    loadRole();
  }, [roleId]);

  useEffect(() => {
    scrollToBottom();
  }, [messages]);

  const loadRole = async () => {
    try {
      setSessionId('');
      setMessages([]);
      const response = await client.get(`/roles/${roleId}`);
      if (response.data.code === 200 || response.data.code === 0) {
        setRole(response.data.data);
        // 添加欢迎消息
        if (response.data.data.welcomeMessage) {
          setMessages([{
            id: 'welcome',
            role: 'assistant',
            content: response.data.data.welcomeMessage,
            timestamp: new Date().toLocaleString()
          }]);
        }
      }
    } catch (error) {
      console.error('加载角色失败:', error);
    }
  };

  const scrollToBottom = () => {
    messagesEndRef.current?.scrollIntoView({ behavior: 'smooth' });
  };

  const ensureSession = async (): Promise<string> => {
    if (sessionId) return sessionId;

    const modelConfig = {
      preferredModel: localStorage.getItem('preferredModel') || 'qwen-plus',
      preferredTemperature: Number(localStorage.getItem('preferredTemperature') || '0.7'),
      knowledgeScope: 'none',
      customAPIKey: localStorage.getItem('customAPIKey') || '',
    };

    const sessionResponse = await client.post('/chat-sessions', {
      roleId,
      mode: 'quick',
      modelConfig,
    });

    if (sessionResponse.data.code !== 200 && sessionResponse.data.code !== 0) {
      throw new Error('创建会话失败');
    }

    const id = sessionResponse.data.data.id;
    setSessionId(id);
    return id;
  };

  const handleSend = async () => {
    if (!input.trim() || loading) return;
    const content = input.trim();

    const userMessage: Message = {
      id: Date.now().toString(),
      role: 'user',
      content,
      timestamp: new Date().toLocaleString()
    };

    setMessages(prev => [...prev, userMessage]);
    setInput('');
    setLoading(true);
    setThinkingProcess(null);

    try {
      const currentSessionId = await ensureSession();

      // 发送消息
      const chatResponse = await client.post(`/chat/${currentSessionId}/complete`, {
        content
      });

      if (chatResponse.data.code === 200 || chatResponse.data.code === 0) {
        const assistantMessage: Message = {
          id: (Date.now() + 1).toString(),
          role: 'assistant',
          content: chatResponse.data.data.assistantMessage.content,
          timestamp: new Date().toLocaleString()
        };
        setMessages(prev => [...prev, assistantMessage]);
      }
    } catch (error: any) {
      console.error('发送消息失败:', error);
      const errorMessage: Message = {
        id: (Date.now() + 1).toString(),
        role: 'assistant',
        content: '抱歉，出现错误：' + (error.response?.data?.message || error.message),
        timestamp: new Date().toLocaleString()
      };
      setMessages(prev => [...prev, errorMessage]);
    } finally {
      setLoading(false);
    }
  };

  const handleKeyPress = (e: React.KeyboardEvent) => {
    if (e.key === 'Enter' && !e.shiftKey) {
      e.preventDefault();
      handleSend();
    }
  };

  // 编辑消息
  const handleEdit = (message: Message) => {
    setEditingId(message.id);
    setEditContent(message.content);
  };

  const handleSaveEdit = async (messageId: string) => {
    try {
      await client.put(`/messages/${messageId}`, { content: editContent });
      setMessages(prev => prev.map(msg => 
        msg.id === messageId ? { ...msg, content: editContent, isEdited: true } : msg
      ));
      setEditingId(null);
    } catch (error) {
      console.error('编辑消息失败:', error);
    }
  };

  // 重新生成回复
  const handleRegenerate = async (messageId: string) => {
    try {
      setLoading(true);
      const response = await client.post(`/chat/${sessionId}/regenerate`, { messageId });
      if (response.data.code === 0) {
        setMessages(prev => prev.map(msg => 
          msg.id === messageId ? { ...msg, content: response.data.data.content } : msg
        ));
      }
    } catch (error) {
      console.error('重新生成失败:', error);
    } finally {
      setLoading(false);
    }
  };

  // 点赞/点踩
  const handleFeedback = async (messageId: string, type: 'like' | 'dislike') => {
    try {
      await client.post(`/messages/${messageId}/feedback`, { type });
      setMessages(prev => prev.map(msg => {
        if (msg.id === messageId) {
          return {
            ...msg,
            likes: type === 'like' ? (msg.likes || 0) + 1 : msg.likes,
            dislikes: type === 'dislike' ? (msg.dislikes || 0) + 1 : msg.dislikes
          };
        }
        return msg;
      }));
    } catch (error) {
      console.error('反馈失败:', error);
    }
  };

  // 复制消息
  const handleCopy = (content: string) => {
    navigator.clipboard.writeText(content);
  };

  // 导出对话
  const handleExport = async (format: 'md' | 'json') => {
    try {
      const response = await client.get(`/chat-sessions/${sessionId}/export?format=${format}`);
      if (response.data.code === 0) {
        const blob = new Blob([response.data.data], { type: format === 'md' ? 'text/markdown' : 'application/json' });
        const url = URL.createObjectURL(blob);
        const a = document.createElement('a');
        a.href = url;
        a.download = `chat-${sessionId}.${format}`;
        a.click();
        URL.revokeObjectURL(url);
      }
    } catch (error) {
      console.error('导出失败:', error);
    }
  };

  // 删除会话
  const handleDeleteSession = async () => {
    if (!confirm('确定要删除这个会话吗？')) return;
    try {
      await client.delete(`/chat-sessions/${sessionId}`);
      navigate('/');
    } catch (error) {
      console.error('删除会话失败:', error);
    }
  };

  // 归档会话
  const handleArchiveSession = async () => {
    try {
      await client.put(`/chat-sessions/${sessionId}/archive`, { isArchived: true });
      navigate('/');
    } catch (error) {
      console.error('归档会话失败:', error);
    }
  };

  // 选择会话
  const handleSelectSession = (newSessionId: string) => {
    navigate(`/chat/${roleId}?session=${newSessionId}`);
    window.location.reload(); // 重新加载以刷新会话
  };

  return (
    <>
      <div className="h-[calc(100vh-8rem)] flex flex-col bg-white rounded-xl shadow-sm border border-slate-100">
      {/* Header */}
      <div className="flex items-center justify-between p-4 border-b border-slate-100">
        <div className="flex items-center gap-4">
          <button
            onClick={() => navigate('/')}
            className="p-2 hover:bg-slate-100 rounded-lg"
          >
            <ArrowLeft className="w-5 h-5" />
          </button>
          <div className="flex items-center gap-3">
            <div className="w-10 h-10 bg-gradient-to-br from-indigo-600 to-purple-600 rounded-xl flex items-center justify-center">
              <Bot className="w-6 h-6 text-white" />
            </div>
            <div>
              <h2 className="font-bold text-slate-900">{role?.name || 'AI 对话'}</h2>
              <p className="text-sm text-slate-500">使用 OpenRouter AI</p>
            </div>
          </div>
        </div>
        <div className="flex items-center gap-2">
          {/* 历史按钮 */}
          <button
            onClick={() => setShowHistory(true)}
            className="p-2 hover:bg-slate-100 rounded-lg text-slate-500"
            title="对话历史"
          >
            <History className="w-5 h-5" />
          </button>
          {/* 深度思考开关 */}
          <button
            onClick={() => setShowThinking(!showThinking)}
            className={`p-2 rounded-lg transition-colors ${
              showThinking ? 'bg-indigo-100 text-indigo-600' : 'hover:bg-slate-100 text-slate-500'
            }`}
            title="深度思考"
          >
            <Brain className="w-5 h-5" />
          </button>
          {/* 导出菜单 */}
          <div className="relative">
            <button
              onClick={() => handleExport('md')}
              className="p-2 hover:bg-slate-100 rounded-lg text-slate-500"
              title="导出 Markdown"
            >
              <Download className="w-5 h-5" />
            </button>
          </div>
          {/* 归档 */}
          <button
            onClick={handleArchiveSession}
            className="p-2 hover:bg-slate-100 rounded-lg text-slate-500"
            title="归档会话"
          >
            <Archive className="w-5 h-5" />
          </button>
          {/* 删除 */}
          <button
            onClick={handleDeleteSession}
            className="p-2 hover:bg-red-50 rounded-lg text-red-500"
            title="删除会话"
          >
            <Trash2 className="w-5 h-5" />
          </button>
        </div>
      </div>

      {/* Messages */}
      <div className="flex-1 overflow-y-auto p-6 space-y-4">
        {messages.map((message, index) => (
          <div
            key={message.id}
            className={`flex items-start gap-3 group ${message.role === 'user' ? 'flex-row-reverse' : ''}`}
          >
            <div className={`w-8 h-8 rounded-full flex items-center justify-center flex-shrink-0 ${
              message.role === 'user' 
                ? 'bg-blue-600' 
                : 'bg-gradient-to-br from-indigo-600 to-purple-600'
            }`}>
              {message.role === 'user' ? (
                <User className="w-5 h-5 text-white" />
              ) : (
                <Bot className="w-5 h-5 text-white" />
              )}
            </div>
            <div className="max-w-[70%]">
              <div className={`rounded-2xl p-4 ${
                message.role === 'user'
                  ? 'bg-blue-600 text-white'
                  : 'bg-slate-100 text-slate-900'
              }`}>
                {editingId === message.id ? (
                  <div className="space-y-2">
                    <textarea
                      value={editContent}
                      onChange={(e) => setEditContent(e.target.value)}
                      className="w-full p-2 text-sm bg-white text-slate-900 rounded border border-slate-300"
                      rows={3}
                    />
                    <div className="flex gap-2">
                      <button
                        onClick={() => handleSaveEdit(message.id)}
                        className="px-3 py-1 text-xs bg-green-600 text-white rounded hover:bg-green-700"
                      >
                        保存
                      </button>
                      <button
                        onClick={() => setEditingId(null)}
                        className="px-3 py-1 text-xs bg-slate-300 text-slate-700 rounded hover:bg-slate-400"
                      >
                        取消
                      </button>
                    </div>
                  </div>
                ) : (
                  <>
                    <div className="text-sm whitespace-pre-wrap">
                      {message.role === 'assistant' ? (
                        <ReactMarkdown remarkPlugins={[remarkGfm]}>
                          {message.content}
                        </ReactMarkdown>
                      ) : (
                        message.content
                      )}
                    </div>
                    {message.isEdited && (
                      <div className="text-xs text-slate-400 mt-1">已编辑</div>
                    )}
                  </>
                )}
              </div>
              <div className={`text-xs mt-1 ${
                message.role === 'user' ? 'text-blue-300' : 'text-slate-400'
              }`}>
                {message.timestamp}
              </div>
              
              {/* 消息操作按钮 */}
              {message.role === 'assistant' && !editingId && (
                <div className={`flex items-center gap-1 mt-2 opacity-0 group-hover:opacity-100 transition-opacity ${
                  showActions === message.id ? 'opacity-100' : ''
                }`}>
                  <button
                    onClick={() => handleCopy(message.content)}
                    className="p-1 hover:bg-slate-200 rounded"
                    title="复制"
                  >
                    <Copy className="w-3 h-3 text-slate-500" />
                  </button>
                  <button
                    onClick={() => handleEdit(message)}
                    className="p-1 hover:bg-slate-200 rounded"
                    title="编辑"
                  >
                    <Edit2 className="w-3 h-3 text-slate-500" />
                  </button>
                  <button
                    onClick={() => handleRegenerate(message.id)}
                    className="p-1 hover:bg-slate-200 rounded"
                    title="重新生成"
                    disabled={loading}
                  >
                    <RotateCcw className={`w-3 h-3 text-slate-500 ${loading ? 'animate-spin' : ''}`} />
                  </button>
                  <button
                    onClick={() => handleFeedback(message.id, 'like')}
                    className="p-1 hover:bg-green-100 rounded"
                    title="点赞"
                  >
                    <ThumbsUp className="w-3 h-3 text-slate-500 hover:text-green-600" />
                  </button>
                  <button
                    onClick={() => handleFeedback(message.id, 'dislike')}
                    className="p-1 hover:bg-red-100 rounded"
                    title="点踩"
                  >
                    <ThumbsDown className="w-3 h-3 text-slate-500 hover:text-red-600" />
                  </button>
                  {message.likes !== undefined && message.likes > 0 && (
                    <span className="text-xs text-green-600 px-1">{message.likes}</span>
                  )}
                  {message.dislikes !== undefined && message.dislikes > 0 && (
                    <span className="text-xs text-red-600 px-1">{message.dislikes}</span>
                  )}
                </div>
              )}
            </div>
          </div>
        ))}
        {loading && (
          <div className="flex items-center gap-3">
            <div className="w-8 h-8 bg-gradient-to-br from-indigo-600 to-purple-600 rounded-full flex items-center justify-center">
              <Bot className="w-5 h-5 text-white" />
            </div>
            <div className="bg-slate-100 rounded-2xl p-4">
              <div className="flex gap-1">
                <div className="w-2 h-2 bg-slate-400 rounded-full animate-bounce" style={{ animationDelay: '0ms' }}></div>
                <div className="w-2 h-2 bg-slate-400 rounded-full animate-bounce" style={{ animationDelay: '150ms' }}></div>
                <div className="w-2 h-2 bg-slate-400 rounded-full animate-bounce" style={{ animationDelay: '300ms' }}></div>
              </div>
            </div>
          </div>
        )}
        <div ref={messagesEndRef} />
      </div>

      {/* Input */}
      <div className="p-4 border-t border-slate-100">
        <div className="flex items-end gap-3">
          <textarea
            value={input}
            onChange={(e) => setInput(e.target.value)}
            onKeyPress={handleKeyPress}
            placeholder="输入消息..."
            className="flex-1 resize-none bg-slate-100 rounded-xl px-4 py-3 focus:outline-none focus:ring-2 focus:ring-primary min-h-[60px] max-h-[200px]"
            rows={1}
          />
          <button
            onClick={handleSend}
            disabled={loading || !input.trim()}
            className="p-3 bg-gradient-to-r from-indigo-600 to-purple-600 text-white rounded-xl hover:shadow-lg transition-all disabled:opacity-50 disabled:cursor-not-allowed"
          >
            <Send className="w-5 h-5" />
          </button>
        </div>
        <p className="text-xs text-slate-500 mt-2 text-center">
          按 Enter 发送，Shift+Enter 换行
        </p>
      </div>
    </div>
  );
};
