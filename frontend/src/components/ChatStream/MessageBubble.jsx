import { Copy, RotateCcw, ThumbsUp, ThumbsDown } from 'lucide-react';
import ReactMarkdown from 'react-markdown';
import remarkGfm from 'remark-gfm';
import { Prism as SyntaxHighlighter } from 'react-syntax-highlighter';
import { vscDarkPlus, oneLight } from 'react-syntax-highlighter/dist/esm/styles/prism';
import './styles.css';

/**
 * @typedef {Object} Message
 * @property {string} id
 * @property {'user' | 'assistant' | 'system'} role
 * @property {string} content
 * @property {string[]} [sources]
 */

/**
 * @typedef {Object} MessageBubbleProps
 * @property {Message} message
 * @property {(content: string) => void} [onCopy]
 * @property {(messageId: string) => void} [onRegenerate]
 * @property {(messageId: string, feedback: 'up' | 'down') => void} [onFeedback]
 */

/**
 * @param {MessageBubbleProps} props
 */
export const MessageBubble = ({
  message,
  onCopy,
  onRegenerate,
  onFeedback,
}) => {
  const isUser = message.role === 'user';
  const isAssistant = message.role === 'assistant';

  const handleCopy = () => {
    if (onCopy) {
      onCopy(message.content);
    } else {
      navigator.clipboard.writeText(message.content);
    }
  };

  return (
    <div className={`chat-stream-message ${isUser ? 'user' : ''}`}>
      <div className={`chat-stream-message-avatar ${isUser ? 'user' : 'assistant'}`}>
        {isUser ? 'U' : 'AI'}
      </div>
      
      <div className={`chat-stream-message-content ${isUser ? 'user' : 'assistant'}`}>
        <div className={`chat-stream-bubble ${isUser ? 'user' : 'assistant'}`}>
          <div className="chat-stream-bubble-content markdown-content">
            <ReactMarkdown
              remarkPlugins={[remarkGfm]}
              components={{
                code({ node, inline, className, children, ...props }) {
                  const match = /language-(\w+)/.exec(className || '');
                  const language = match ? match[1] : 'text';
                  const content = String(children).replace(/\n$/, '');
                  
                  if (!inline) {
                    return (
                      <SyntaxHighlighter
                        style={isUser ? vscDarkPlus : oneLight}
                        language={language}
                        PreTag="div"
                        {...props}
                      >
                        {content}
                      </SyntaxHighlighter>
                    );
                  }
                  
                  return (
                    <code className={className} {...props}>
                      {children}
                    </code>
                  );
                },
                p({ children }) {
                  return <p>{children}</p>;
                },
              }}
            >
              {message.content}
            </ReactMarkdown>
          </div>

          {/* Sources */}
          {message.sources && message.sources.length > 0 && (
            <div className="chat-stream-sources">
              <p className="chat-stream-sources-title">📚 参考来源：</p>
              <div className="chat-stream-sources-list">
                {message.sources.map((source, i) => (
                  <span
                    key={i}
                    className="chat-stream-source-tag"
                    title={source}
                  >
                    {source.length > 30 ? source.substring(0, 30) + '...' : source}
                  </span>
                ))}
              </div>
            </div>
          )}
        </div>

        {/* Actions */}
        {isAssistant && (
          <div className="chat-stream-actions">
            <button
              onClick={handleCopy}
              className="chat-stream-action-btn"
              title="复制"
            >
              <Copy className="w-4 h-4" />
            </button>
            {onRegenerate && (
              <button
                onClick={() => onRegenerate(message.id)}
                className="chat-stream-action-btn"
                title="重新生成"
              >
                <RotateCcw className="w-4 h-4" />
              </button>
            )}
            {onFeedback && (
              <>
                <button
                  onClick={() => onFeedback(message.id, 'up')}
                  className="chat-stream-action-btn"
                  title="有用"
                >
                  <ThumbsUp className="w-4 h-4" />
                </button>
                <button
                  onClick={() => onFeedback(message.id, 'down')}
                  className="chat-stream-action-btn"
                  title="无用"
                >
                  <ThumbsDown className="w-4 h-4" />
                </button>
              </>
            )}
          </div>
        )}
      </div>
    </div>
  );
};
