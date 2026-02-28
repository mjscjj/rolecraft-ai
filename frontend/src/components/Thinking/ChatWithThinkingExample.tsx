import React, { useState, useEffect } from 'react';
import ThinkingDisplay, { ThinkingProcess, ThinkingStep } from './ThinkingDisplay';
import { API_BASE_URL as API_BASE } from '../../api/client';

// 示例：如何在 Chat 页面中集成 ThinkingDisplay

interface Message {
  id: string;
  role: 'user' | 'assistant' | 'system';
  content: string;
  thinkingProcess?: ThinkingProcess;
  isStreaming?: boolean;
}

const ChatWithThinkingExample: React.FC = () => {
  const [messages, setMessages] = useState<Message[]>([]);
  const [inputValue, setInputValue] = useState('');
  const [isLoading, setIsLoading] = useState(false);
  const [sessionId, setSessionId] = useState<string | null>(null);

  // 处理流式消息（带思考过程）
  const handleStreamChatWithThinking = async (message: string) => {
    if (!sessionId) return;

    const token = localStorage.getItem('token');
    if (!token) {
      alert('请先登录');
      return;
    }

    // 创建用户消息
    const userMessage: Message = {
      id: Date.now().toString(),
      role: 'user',
      content: message,
    };

    // 创建 AI 消息占位符（带思考过程）
    const aiMessageId = (Date.now() + 1).toString();
    const aiMessage: Message = {
      id: aiMessageId,
      role: 'assistant',
      content: '',
      thinkingProcess: undefined,
      isStreaming: true,
    };

    setMessages(prev => [...prev, userMessage, aiMessage]);
    setIsLoading(true);

    try {
      // 调用带思考过程的流式 API
      const response = await fetch(`${API_BASE}/chat/${sessionId}/stream-with-thinking`, {
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
      let thinkingSteps: ThinkingStep[] = [];
      let thinkingStartTime = Date.now();

      while (true) {
        const { done, value } = await reader.read();
        if (done) break;

        const chunk = decoder.decode(value, { stream: true });
        const lines = chunk.split('\n');

        for (const line of lines) {
          if (line.startsWith('data: ')) {
            try {
              const data = JSON.parse(line.slice(6));
              
              // 处理思考步骤
              if (data.type === 'thinking') {
                const newStep: ThinkingStep = data.data;
                thinkingSteps = [...thinkingSteps, newStep];
                
                // 更新消息的思考过程
                setMessages(prev => 
                  prev.map(msg => 
                    msg.id === aiMessageId 
                      ? { 
                          ...msg, 
                          thinkingProcess: {
                            steps: thinkingSteps,
                            startTime: thinkingStartTime,
                            duration: (Date.now() - thinkingStartTime) / 1000,
                            isComplete: false,
                          }
                        }
                      : msg
                  )
                );
              }
              
              // 处理思考步骤更新
              if (data.type === 'thinking_update') {
                const updatedStep: ThinkingStep = data.data;
                thinkingSteps = thinkingSteps.map(step => 
                  step.id === updatedStep.id ? updatedStep : step
                );
                
                setMessages(prev => 
                  prev.map(msg => 
                    msg.id === aiMessageId 
                      ? { 
                          ...msg, 
                          thinkingProcess: {
                            ...msg.thinkingProcess!,
                            steps: thinkingSteps,
                          }
                        }
                      : msg
                  )
                );
              }
              
              // 处理思考完成
              if (data.type === 'thinking_done') {
                const process: ThinkingProcess = data.data;
                setMessages(prev => 
                  prev.map(msg => 
                    msg.id === aiMessageId 
                      ? { 
                          ...msg, 
                          thinkingProcess: process
                        }
                      : msg
                  )
                );
              }
              
              // 处理最终答案
              if (data.type === 'answer') {
                accumulatedContent = data.data.content;
                
                setMessages(prev => 
                  prev.map(msg => 
                    msg.id === aiMessageId 
                      ? { ...msg, content: accumulatedContent }
                      : msg
                  )
                );
              }
              
              // 处理完成
              if (data.type === 'done' || data.done) {
                setMessages(prev => 
                  prev.map(msg => 
                    msg.id === aiMessageId 
                      ? { 
                          ...msg, 
                          isStreaming: false,
                          thinkingProcess: msg.thinkingProcess ? {
                            ...msg.thinkingProcess,
                            isComplete: true,
                          } : undefined
                        }
                      : msg
                  )
                );
              }
              
              // 处理错误
              if (data.type === 'error') {
                throw new Error(data.data.message);
              }
            } catch (e) {
              console.warn('Failed to parse SSE data:', e);
            }
          }
        }
      }
    } catch (err) {
      console.error('Failed to stream message:', err);
      setMessages(prev => 
        prev.map(msg => 
          msg.id === aiMessageId 
            ? { 
                ...msg, 
                content: msg.content || '抱歉，发生了错误。请稍后重试。',
                isStreaming: false,
              }
            : msg
        )
      );
    } finally {
      setIsLoading(false);
    }
  };

  const handleSend = () => {
    if (!inputValue.trim() || !sessionId || isLoading) return;
    
    handleStreamChatWithThinking(inputValue.trim());
    setInputValue('');
  };

  return (
    <div className="chat-container">
      {/* 消息列表 */}
      <div className="messages-container">
        {messages.map((message) => (
          <div key={message.id} className={`message ${message.role}`}>
            {/* 消息内容 */}
            <div className="message-content">
              {message.content}
            </div>
            
            {/* 思考过程展示 */}
            {message.thinkingProcess && (
              <ThinkingDisplay
                thinkingProcess={message.thinkingProcess}
                isStreaming={message.isStreaming}
                defaultExpanded={true}
                onToggle={(expanded) => {
                  console.log('Thinking display toggled:', expanded);
                }}
              />
            )}
          </div>
        ))}
      </div>

      {/* 输入区域 */}
      <div className="input-area">
        <textarea
          value={inputValue}
          onChange={(e) => setInputValue(e.target.value)}
          placeholder="输入消息..."
          disabled={isLoading || !sessionId}
        />
        <button onClick={handleSend} disabled={isLoading || !sessionId}>
          发送
        </button>
      </div>
    </div>
  );
};

export default ChatWithThinkingExample;
