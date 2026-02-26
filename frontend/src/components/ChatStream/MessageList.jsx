import { useEffect, useState, useRef } from 'react';
import { MessageBubble } from './MessageBubble';
import { TypingIndicator } from './TypingIndicator';
import './styles.css';

/**
 * @typedef {Object} Message
 * @property {string} id
 * @property {'user' | 'assistant' | 'system'} role
 * @property {string} content
 * @property {string[]} [sources]
 */

/**
 * @typedef {Object} MessageListProps
 * @property {Message[]} messages
 * @property {boolean} isLoading
 * @property {string | null} [streamingMessageId]
 * @property {(content: string) => void} [onCopy]
 * @property {(messageId: string) => void} [onRegenerate]
 * @property {(messageId: string, feedback: 'up' | 'down') => void} [onFeedback]
 */

/**
 * @param {MessageListProps} props
 * @returns {JSX.Element}
 */
export const MessageList = ({
  messages,
  isLoading,
  streamingMessageId,
  onCopy,
  onRegenerate,
  onFeedback,
}) => {
  const messagesEndRef = useRef<HTMLDivElement>(null);
  const containerRef = useRef<HTMLDivElement>(null);
  const [showScrollButton, setShowScrollButton] = useState(false);
  const [userHasScrolled, setUserHasScrolled] = useState(false);
  const lastMessageCount = useRef(messages.length);

  // Check if user is near bottom
  const isNearBottom = () => {
    if (!containerRef.current) return true;
    
    const { scrollTop, scrollHeight, clientHeight } = containerRef.current;
    const threshold = 100; // pixels from bottom
    return scrollHeight - scrollTop - clientHeight < threshold;
  };

  // Scroll to bottom
  const scrollToBottom = (smooth = true) => {
    if (messagesEndRef.current) {
      messagesEndRef.current.scrollIntoView({ 
        behavior: smooth ? 'smooth' : 'auto',
        block: 'end'
      });
    }
  };

  // Handle scroll events
  useEffect(() => {
    const container = containerRef.current;
    if (!container) return;

    const handleScroll = () => {
      const nearBottom = isNearBottom();
      setShowScrollButton(!nearBottom);
      
      // Track if user has manually scrolled up
      if (!nearBottom) {
        setUserHasScrolled(true);
      } else {
        setUserHasScrolled(false);
      }
    };

    container.addEventListener('scroll', handleScroll);
    return () => container.removeEventListener('scroll', handleScroll);
  }, []);

  // Auto-scroll when new messages arrive
  useEffect(() => {
    const hasNewMessage = messages.length > lastMessageCount.current;
    lastMessageCount.current = messages.length;

    if (hasNewMessage && !userHasScrolled) {
      // Only auto-scroll if user hasn't scrolled up
      scrollToBottom(true);
    }
  }, [messages, userHasScrolled]);

  // Auto-scroll when streaming completes
  useEffect(() => {
    if (!isLoading && streamingMessageId && messages.length > 0) {
      // Streaming just completed, scroll to bottom
      setTimeout(() => scrollToBottom(true), 100);
    }
  }, [isLoading, streamingMessageId]);

  const handleScrollButtonClick = () => {
    scrollToBottom(true);
    setUserHasScrolled(false);
  };

  return (
    <div 
      ref={containerRef}
      className="chat-stream-messages"
      style={{ position: 'relative' }}
    >
      {messages.length === 0 && !isLoading && (
        <div className="chat-stream-empty">
          <p className="chat-stream-empty-title">👋 开始对话吧！</p>
          <p className="chat-stream-empty-subtitle">输入问题或需求，AI 将为你提供帮助</p>
        </div>
      )}

      {messages.map((message) => (
        <MessageBubble
          key={message.id}
          message={message}
          onCopy={onCopy}
          onRegenerate={onRegenerate}
          onFeedback={onFeedback}
        />
      ))}

      {isLoading && !streamingMessageId && <TypingIndicator />}

      <div ref={messagesEndRef} />

      {/* Scroll to Bottom Button */}
      <button
        className={`scroll-to-bottom-btn ${showScrollButton ? '' : 'hidden'}`}
        onClick={handleScrollButtonClick}
        aria-label="滚动到底部"
      >
        <svg 
          width="16" 
          height="16" 
          viewBox="0 0 24 24" 
          fill="none" 
          stroke="currentColor" 
          strokeWidth="2"
        >
          <path d="M12 5v14M19 12l-7 7-7-7" />
        </svg>
        新消息
      </button>
    </div>
  );
};
