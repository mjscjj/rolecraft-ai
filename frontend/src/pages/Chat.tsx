import { useState, useEffect, useRef } from 'react';
import { useParams, useNavigate } from 'react-router-dom';
import { Send, Bot, User, Sparkles, ArrowLeft } from 'lucide-react';
import ReactMarkdown from 'react-markdown';
import remarkGfm from 'remark-gfm';
import client from '../api/client';

interface Message {
  id: string;
  role: 'user' | 'assistant';
  content: string;
  timestamp: string;
}

export const Chat = () => {
  const { roleId } = useParams();
  const navigate = useNavigate();
  const [messages, setMessages] = useState<Message[]>([]);
  const [input, setInput] = useState('');
  const [loading, setLoading] = useState(false);
  const [role, setRole] = useState<any>(null);
  const messagesEndRef = useRef<HTMLDivElement>(null);

  useEffect(() => {
    loadRole();
  }, [roleId]);

  useEffect(() => {
    scrollToBottom();
  }, [messages]);

  const loadRole = async () => {
    try {
      const response = await client.get(`/roles/${roleId}`);
      if (response.data.code === 0) {
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

  const handleSend = async () => {
    if (!input.trim() || loading) return;

    const userMessage: Message = {
      id: Date.now().toString(),
      role: 'user',
      content: input,
      timestamp: new Date().toLocaleString()
    };

    setMessages(prev => [...prev, userMessage]);
    setInput('');
    setLoading(true);

    try {
      // 创建会话
      const sessionResponse = await client.post('/chat-sessions', {
        roleId,
        mode: 'quick'
      });

      if (sessionResponse.data.code === 0) {
        const sessionId = sessionResponse.data.data.id;

        // 发送消息
        const chatResponse = await client.post(`/chat/${sessionId}/complete`, {
          content: input
        });

        if (chatResponse.data.code === 0) {
          const assistantMessage: Message = {
            id: (Date.now() + 1).toString(),
            role: 'assistant',
            content: chatResponse.data.data.assistantMessage.content,
            timestamp: new Date().toLocaleString()
          };
          setMessages(prev => [...prev, assistantMessage]);
        }
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

  return (
    <div className="h-[calc(100vh-8rem)] flex flex-col bg-white rounded-xl shadow-sm border border-slate-100">
      {/* Header */}
      <div className="flex items-center gap-4 p-4 border-b border-slate-100">
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

      {/* Messages */}
      <div className="flex-1 overflow-y-auto p-6 space-y-6">
        {messages.map(message => (
          <div
            key={message.id}
            className={`flex items-start gap-3 ${message.role === 'user' ? 'flex-row-reverse' : ''}`}
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
            <div className={`max-w-[70%] rounded-2xl p-4 ${
              message.role === 'user'
                ? 'bg-blue-600 text-white'
                : 'bg-slate-100 text-slate-900'
            }`}>
              <div className="text-sm whitespace-pre-wrap">
                {message.role === 'assistant' ? (
                  <ReactMarkdown remarkPlugins={[remarkGfm]}>
                    {message.content}
                  </ReactMarkdown>
                ) : (
                  message.content
                )}
              </div>
              <div className={`text-xs mt-2 ${
                message.role === 'user' ? 'text-blue-100' : 'text-slate-500'
              }`}>
                {message.timestamp}
              </div>
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
