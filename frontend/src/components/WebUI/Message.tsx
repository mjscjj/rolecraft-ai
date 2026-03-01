import React, { useState } from 'react';
import ReactMarkdown from 'react-markdown';
import remarkGfm from 'remark-gfm';
import remarkMath from 'remark-math';
import rehypeKatex from 'rehype-katex';
import rehypeHighlight from 'rehype-highlight';
import type { MessageMeta } from '../../api/chat';
import 'katex/dist/katex.min.css';
import 'highlight.js/styles/github-dark.css';

interface MessageProps {
  id: string;
  role: 'user' | 'assistant' | 'system';
  content: string;
  sources?: MessageMeta;
  createdAt: string;
  isStreaming?: boolean;
  onEdit?: (messageId: string, newContent: string) => void;
  onRegenerate?: (messageId: string) => void;
  onCopy?: (content: string) => void;
  onRate?: (messageId: string, rating: 'up' | 'down') => void;
}

const Message: React.FC<MessageProps> = ({
  id,
  role,
  content,
  sources,
  createdAt,
  isStreaming = false,
  onEdit,
  onRegenerate,
  onCopy,
  onRate,
}) => {
  const [isEditing, setIsEditing] = useState(false);
  const [editContent, setEditContent] = useState(content);

  const isUser = role === 'user';
  const isAI = role === 'assistant';
  const isAgentMode = isAI && sources?.mode === 'agent';
  const sourceItems = Array.isArray(sources?.sources) ? sources.sources : [];
  const thoughtItems = Array.isArray(sources?.thoughts) ? sources.thoughts : [];

  const handleCopy = () => {
    navigator.clipboard.writeText(content);
    onCopy?.(content);
  };

  const handleSaveEdit = () => {
    onEdit?.(id, editContent);
    setIsEditing(false);
  };

  const handleCancelEdit = () => {
    setEditContent(content);
    setIsEditing(false);
  };

  const formatTime = (dateString: string) => {
    const date = new Date(dateString);
    return date.toLocaleTimeString('zh-CN', { hour: '2-digit', minute: '2-digit' });
  };

  return (
    <div className={`message ${role}`}>
      {/* Avatar */}
      <div className="message-avatar">
        {isAI ? '🤖' : isUser ? '👤' : '⚙️'}
      </div>

      {/* Content */}
      <div className="message-content">
        <div className="message-bubble">
          {isEditing ? (
            <div>
              <textarea
                value={editContent}
                onChange={(e) => setEditContent(e.target.value)}
                style={{
                  width: '100%',
                  minHeight: '100px',
                  padding: '8px',
                  borderRadius: '6px',
                  border: '1px solid var(--border-color)',
                  fontFamily: 'inherit',
                  fontSize: '14px',
                  resize: 'vertical',
                }}
              />
              <div style={{ display: 'flex', gap: '8px', marginTop: '8px' }}>
                <button
                  onClick={handleSaveEdit}
                  style={{
                    padding: '6px 12px',
                    background: 'var(--accent-color)',
                    color: 'white',
                    border: 'none',
                    borderRadius: '4px',
                    cursor: 'pointer',
                    fontSize: '12px',
                  }}
                >
                  保存
                </button>
                <button
                  onClick={handleCancelEdit}
                  style={{
                    padding: '6px 12px',
                    background: 'var(--bg-tertiary)',
                    color: 'var(--text-primary)',
                    border: 'none',
                    borderRadius: '4px',
                    cursor: 'pointer',
                    fontSize: '12px',
                  }}
                >
                  取消
                </button>
              </div>
            </div>
          ) : (
            <>
              {isAgentMode && (
                <div className="agent-mode-header">
                  <span className="agent-mode-badge">Agent</span>
                  <span className="agent-mode-label">深度执行模式</span>
                </div>
              )}
              <ReactMarkdown
                remarkPlugins={[remarkGfm, remarkMath]}
                rehypePlugins={[rehypeKatex, rehypeHighlight]}
                components={{
                  // Custom code block rendering
                  code({ node, inline, className, children, ...props }: any) {
                    const match = /language-(\w+)/.exec(className || '');
                    const language = match ? match[1] : '';
                    const codeText = String(children).replace(/\n$/, '');

                    if (!inline) {
                      return (
                        <div style={{ position: 'relative', margin: '12px 0' }}>
                          {language && (
                            <div
                              style={{
                                position: 'absolute',
                                top: '8px',
                                right: '12px',
                                fontSize: '11px',
                                color: 'var(--code-comment)',
                                fontFamily: 'monospace',
                              }}
                            >
                              {language}
                            </div>
                          )}
                          <pre style={{ margin: 0 }}>
                            <code className={className} {...props}>
                              {children}
                            </code>
                          </pre>
                        </div>
                      );
                    }

                    return (
                      <code
                        style={{
                          padding: '2px 6px',
                          background: 'var(--bg-tertiary)',
                          borderRadius: '4px',
                          fontFamily: "'SF Mono', 'Consolas', monospace",
                          fontSize: '0.9em',
                        }}
                        {...props}
                      >
                        {children}
                      </code>
                    );
                  },
                  // Custom table rendering
                  table({ children }: any) {
                    return (
                      <div style={{ overflowX: 'auto', margin: '12px 0' }}>
                        <table style={{ width: '100%', borderCollapse: 'collapse' }}>
                          {children}
                        </table>
                      </div>
                    );
                  },
                  th({ children }: any) {
                    return (
                      <th
                        style={{
                          padding: '8px 12px',
                          border: '1px solid var(--border-color)',
                          background: 'var(--bg-secondary)',
                          fontWeight: 600,
                          textAlign: 'left',
                        }}
                      >
                        {children}
                      </th>
                    );
                  },
                  td({ children }: any) {
                    return (
                      <td
                        style={{
                          padding: '8px 12px',
                          border: '1px solid var(--border-color)',
                        }}
                      >
                        {children}
                      </td>
                    );
                  },
                  // Custom blockquote
                  blockquote({ children }: any) {
                    return (
                      <blockquote
                        style={{
                          margin: '12px 0',
                          padding: '8px 16px',
                          borderLeft: '4px solid var(--accent-color)',
                          background: 'var(--bg-secondary)',
                          borderRadius: '4px',
                        }}
                      >
                        {children}
                      </blockquote>
                    );
                  },
                  a({ href, children }: any) {
                    return (
                      <a
                        href={href}
                        target="_blank"
                        rel="noopener noreferrer"
                        style={{
                          color: 'var(--accent-color)',
                          textDecoration: 'underline',
                          wordBreak: 'break-all',
                        }}
                      >
                        {children}
                      </a>
                    );
                  },
                }}
              >
                {content}
              </ReactMarkdown>
              {thoughtItems.length > 0 && (
                <div className="agent-thoughts">
                  <div className="agent-thoughts-title">思考要点</div>
                  <ul>
                    {thoughtItems.slice(0, 4).map((thought, index) => (
                      <li key={`${thought}-${index}`}>{thought}</li>
                    ))}
                  </ul>
                </div>
              )}
              {sourceItems.length > 0 && (
                <div className="agent-sources">
                  <div className="agent-sources-title">检索来源</div>
                  <ol>
                    {sourceItems.map((item, index) => {
                      const title =
                        (typeof item?.title === 'string' && item.title) ||
                        (typeof item?.name === 'string' && item.name) ||
                        `来源 ${index + 1}`;
                      const url =
                        (typeof item?.url === 'string' && item.url) ||
                        (typeof item?.link === 'string' && item.link) ||
                        '';
                      return (
                        <li key={`${title}-${index}`}>
                          {url ? (
                            <a href={url} target="_blank" rel="noopener noreferrer">
                              {title}
                            </a>
                          ) : (
                            <span>{title}</span>
                          )}
                        </li>
                      );
                    })}
                  </ol>
                </div>
              )}
            </>
          )}

          {/* Typing Indicator */}
          {isStreaming && (
            <div className="typing-indicator">
              <div className="typing-dot" />
              <div className="typing-dot" />
              <div className="typing-dot" />
            </div>
          )}
        </div>

        {/* Message Metadata */}
        <div
          style={{
            display: 'flex',
            alignItems: 'center',
            gap: '12px',
            fontSize: '12px',
            color: 'var(--text-tertiary)',
          }}
        >
          <span>{formatTime(createdAt)}</span>
        </div>

        {/* Hover Actions */}
        {!isStreaming && !isEditing && (
          <div className="message-actions">
            {isUser && onEdit && (
              <button
                className="message-action"
                onClick={() => setIsEditing(true)}
                title="编辑"
              >
                ✏️ 编辑
              </button>
            )}
            {isAI && onRegenerate && (
              <button
                className="message-action"
                onClick={() => onRegenerate(id)}
                title="重新生成"
              >
                🔄 重新生成
              </button>
            )}
            <button
              className="message-action"
              onClick={handleCopy}
              title="复制"
            >
              📋 复制
            </button>
            {isAI && onRate && (
              <>
                <button
                  className="message-action"
                  onClick={() => onRate(id, 'up')}
                  title="有用"
                >
                  👍
                </button>
                <button
                  className="message-action"
                  onClick={() => onRate(id, 'down')}
                  title="无用"
                >
                  👎
                </button>
              </>
            )}
          </div>
        )}
      </div>
    </div>
  );
};

export default Message;
