import { useState, useRef, useEffect, useCallback } from 'react';
import { Send, Paperclip, MoreVertical } from 'lucide-react';
import { MessageList } from './MessageList';
import './styles.css';

const API_BASE = 'http://localhost:8080/api/v1';

/**
 * @typedef {Object} Message
 * @property {string} id
 * @property {'user' | 'assistant' | 'system'} role
 * @property {string} content
 * @property {string[]} [sources]
 */

/**
 * @typedef {Object} ChatStreamProps
 * @property {string} [roleId]
 * @property {string} [roleName]
 */

/**
 * @param {ChatStreamProps} props
 * @returns {JSX.Element}
 */
export const ChatStream = ({ roleId, roleName = 'AI 助手' }) => {
  const [messages, setMessages] = useState([]);
  const [inputValue, setInputValue] = useState('');
  const [isLoading, setIsLoading] = useState(false);
  const [sessionId, setSessionId] = useState(null);
  const [streamingMessageId, setStreamingMessageId] = useState(null);
  const textareaRef = useRef(null);

  // Auto-resize textarea
  useEffect(() => {
    if (textareaRef.current) {
      textareaRef.current.style.height = 'auto';
      textareaRef.current.style.height = `${Math.min(textareaRef.current.scrollHeight, 200)}px`;
    }
  }, [inputValue]);

  // Initialize session
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
          // Load welcome message
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

  // Handle stream chat
  const handleStreamChat = useCallback(async (message) => {
    if (!sessionId) return;

    const token = localStorage.getItem('token');
    if (!token) {
      alert('请先登录');
      return;
    }

    // Create user message
    const userMessage: Message = {
      id: Date.now().toString(),
      role: 'user',
      content: message,
    };

    // Create placeholder for AI response
    const aiMessageId = (Date.now() + 1).toString();
    const aiMessage: Message = {
      id: aiMessageId,
      role: 'assistant',
      content: '',
    };

    setMessages(prev => [...prev, userMessage, aiMessage]);
    setIsLoading(true);
    setStreamingMessageId(aiMessageId);

    try {
      const response = await fetch(`${API_BASE}/chat/${sessionId}/stream`, {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
          'Authorization': `Bearer ${token}`,
        },
        body: JSON.stringify({ content: message }),
      });

      if (!response.ok) {
        throw new Error('Stream request failed');
      }

      const reader = response.body.getReader();
      const decoder = new TextDecoder();
      let accumulatedContent = '';

      while (true) {
        const { done, value } = await reader.read();
        if (done) break;

        const chunk = decoder.decode(value, { stream: true });
        // Parse SSE data: data: {"content": "..."}
        const lines = chunk.split('\n');
        for (const line of lines) {
          if (line.startsWith('data: ')) {
            try {
              const data = JSON.parse(line.slice(6));
              if (data.content) {
                accumulatedContent += data.content;
                
                // Update the streaming message
                setMessages(prev => 
                  prev.map(msg => 
                    msg.id === aiMessageId 
                      ? { ...msg, content: accumulatedContent }
                      : msg
                  )
                );
              }
            } catch (e) {
              console.warn('Failed to parse SSE data:', e);
            }
          }
        }
      }

      // Streaming complete
      setStreamingMessageId(null);
    } catch (err) {
      console.error('Failed to stream message:', err);
      // Update with error message
      setMessages(prev => 
        prev.map(msg => 
          msg.id === aiMessageId 
            ? { 
                ...msg, 
                content: msg.content || '抱歉，发生了错误。请稍后重试。',
                sources: msg.sources || []
              }
            : msg
        )
      );
      setStreamingMessageId(null);
    } finally {
      setIsLoading(false);
    }
  }, [sessionId]);

  const handleSend = () => {
    if (!inputValue.trim() || !sessionId || isLoading) return;
    
    handleStreamChat(inputValue.trim());
    setInputValue('');
    
    // Reset textarea height
    if (textareaRef.current) {
      textareaRef.current.style.height = 'auto';
    }
  };

  const handleKeyDown = (e) => {
    if (e.key === 'Enter' && !e.shiftKey) {
      e.preventDefault();
      handleSend();
    }
  };

  const handleCopy = useCallback((content: string) => {
    navigator.clipboard.writeText(content);
  }, []);

  const handleRegenerate = useCallback(async (messageId) => {
    // Find the message to regenerate
    const messageIndex = messages.findIndex(m => m.id === messageId);
    if (messageIndex === -1) return;

    // Find the previous user message
    let userMessageIndex = messageIndex - 1;
    while (userMessageIndex >= 0 && messages[userMessageIndex].role !== 'user') {
      userMessageIndex--;
    }

    if (userMessageIndex >= 0) {
      const userMessage = messages[userMessageIndex];
      // Remove messages from the point of regeneration
      setMessages(prev => prev.slice(0, messageIndex));
      // Re-send the user message
      await handleStreamChat(userMessage.content);
    }
  }, [messages, handleStreamChat]);

  const handleFeedback = useCallback((messageId, feedback) => {
    console.log(`Feedback for message ${messageId}: ${feedback}`);
    // TODO: Send feedback to backend
  }, []);

  return (
    <div className="chat-stream-container">
      {/* Header */}
      <div className="chat-stream-header">
        <div className="chat-stream-header-info">
          <div className="chat-stream-avatar">
            {roleName[0]}
          </div>
          <div>
            <h2 className="font-semibold text-slate-900">{roleName}</h2>
            <p className="chat-stream-status">
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
      <MessageList
        messages={messages}
        isLoading={isLoading}
        streamingMessageId={streamingMessageId}
        onCopy={handleCopy}
        onRegenerate={handleRegenerate}
        onFeedback={handleFeedback}
      />

      {/* Input Area */}
      <div className="chat-stream-input-area">
        <div className="chat-stream-input-wrapper">
          <button 
            className="chat-stream-input-btn"
            title="附件"
            disabled={!sessionId || isLoading}
          >
            <Paperclip className="w-5 h-5" />
          </button>
          <textarea
            ref={textareaRef}
            value={inputValue}
            onChange={(e) => setInputValue(e.target.value)}
            onKeyDown={handleKeyDown}
            placeholder={sessionId ? "输入消息..." : "正在连接..."}
            disabled={!sessionId || isLoading}
            rows={1}
            className="chat-stream-textarea"
          />
          <button
            onClick={handleSend}
            disabled={!inputValue.trim() || isLoading || !sessionId}
            className="chat-stream-input-btn send"
          >
            <Send className="w-5 h-5" />
          </button>
        </div>
        <p className="chat-stream-input-hint">按 Enter 发送，Shift + Enter 换行</p>
      </div>
    </div>
  );
};
