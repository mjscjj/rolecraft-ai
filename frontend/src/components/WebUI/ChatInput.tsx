import React, { useState, useRef, useEffect, KeyboardEvent } from 'react';
import ReactMarkdown from 'react-markdown';
import remarkGfm from 'remark-gfm';
import remarkMath from 'remark-math';
import rehypeKatex from 'rehype-katex';
import 'katex/dist/katex.min.css';

interface ChatInputProps {
  onSend: (content: string) => void;
  onAttach?: (files: FileList) => void;
  onVoiceInput?: () => void;
  disabled?: boolean;
  placeholder?: string;
  isStreaming?: boolean;
}

const ChatInput: React.FC<ChatInputProps> = ({
  onSend,
  onAttach,
  onVoiceInput,
  disabled = false,
  placeholder = '输入消息... (Shift+Enter 换行)',
  isStreaming = false,
}) => {
  const [value, setValue] = useState('');
  const [showPreview, setShowPreview] = useState(false);
  const [height, setHeight] = useState(56);
  const textareaRef = useRef<HTMLTextAreaElement>(null);
  const fileInputRef = useRef<HTMLInputElement>(null);

  const MIN_HEIGHT = 56;
  const MAX_HEIGHT = 200;

  // Auto-resize textarea
  useEffect(() => {
    if (textareaRef.current) {
      textareaRef.current.style.height = 'auto';
      const scrollHeight = textareaRef.current.scrollHeight;
      const newHeight = Math.min(Math.max(scrollHeight, MIN_HEIGHT), MAX_HEIGHT);
      textareaRef.current.style.height = `${newHeight}px`;
      setHeight(newHeight);
    }
  }, [value]);

  const handleSend = () => {
    if (value.trim() && !disabled && !isStreaming) {
      onSend(value.trim());
      setValue('');
      if (textareaRef.current) {
        textareaRef.current.style.height = `${MIN_HEIGHT}px`;
        setHeight(MIN_HEIGHT);
      }
    }
  };

  const handleKeyDown = (e: KeyboardEvent<HTMLTextAreaElement>) => {
    if (e.key === 'Enter' && !e.shiftKey) {
      e.preventDefault();
      handleSend();
    }
  };

  const handleAttachClick = () => {
    fileInputRef.current?.click();
  };

  const handleFileChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    if (e.target.files && e.target.files.length > 0 && onAttach) {
      onAttach(e.target.files);
      e.target.value = ''; // Reset for re-selection
    }
  };

  const hasContent = value.trim().length > 0;

  return (
    <div className="input-area">
      <div className="input-wrapper">
        <div className="input-container">
          {/* Markdown Preview */}
          {showPreview && value.trim() && (
            <div className="markdown-preview">
              <div
                className="markdown-preview-content"
                style={{
                  padding: '8px',
                  background: 'var(--bg-secondary)',
                  borderRadius: '6px',
                  border: '1px solid var(--border-color)',
                }}
              >
                <ReactMarkdown
                  remarkPlugins={[remarkGfm, remarkMath]}
                  rehypePlugins={[rehypeKatex]}
                  components={{
                    h1: ({ children }) => <h1 style={{ fontSize: '1.5em', margin: '8px 0' }}>{children}</h1>,
                    h2: ({ children }) => <h2 style={{ fontSize: '1.3em', margin: '8px 0' }}>{children}</h2>,
                    h3: ({ children }) => <h3 style={{ fontSize: '1.1em', margin: '8px 0' }}>{children}</h3>,
                    p: ({ children }) => <p style={{ margin: '4px 0' }}>{children}</p>,
                    code: ({ node, inline, className, children, ...props }: any) => {
                      if (inline) {
                        return (
                          <code
                            style={{
                              padding: '2px 6px',
                              background: 'var(--bg-tertiary)',
                              borderRadius: '4px',
                              fontFamily: 'monospace',
                              fontSize: '0.9em',
                            }}
                            {...props}
                          >
                            {children}
                          </code>
                        );
                      }
                      return (
                        <pre
                          style={{
                            background: 'var(--code-bg)',
                            color: 'var(--code-text)',
                            padding: '12px',
                            borderRadius: '6px',
                            overflowX: 'auto',
                            margin: '8px 0',
                          }}
                        >
                          <code {...props}>{children}</code>
                        </pre>
                      );
                    },
                  }}
                >
                  {value}
                </ReactMarkdown>
              </div>
            </div>
          )}

          {/* Textarea */}
          <textarea
            ref={textareaRef}
            className="chat-input"
            value={value}
            onChange={(e) => setValue(e.target.value)}
            onKeyDown={handleKeyDown}
            placeholder={placeholder}
            disabled={disabled || isStreaming}
            style={{
              minHeight: `${height}px`,
              maxHeight: `${MAX_HEIGHT}px`,
            }}
          />

          {/* Input Actions */}
          <div className="input-actions">
            <div className="input-left-actions">
              {/* Attach File */}
              {onAttach && (
                <>
                  <input
                    ref={fileInputRef}
                    type="file"
                    multiple
                    onChange={handleFileChange}
                    style={{ display: 'none' }}
                  />
                  <button
                    className="attach-btn"
                    onClick={handleAttachClick}
                    title="上传附件"
                    disabled={disabled || isStreaming}
                  >
                    <svg
                      width="20"
                      height="20"
                      viewBox="0 0 24 24"
                      fill="none"
                      stroke="currentColor"
                      strokeWidth="2"
                      strokeLinecap="round"
                      strokeLinejoin="round"
                    >
                      <path d="M21.44 11.05l-9.19 9.19a6 6 0 0 1-8.49-8.49l9.19-9.19a4 4 0 0 1 5.66 5.66l-9.2 9.19a2 2 0 0 1-2.83-2.83l8.49-8.48" />
                    </svg>
                  </button>
                </>
              )}

              {/* Voice Input */}
              {onVoiceInput && (
                <button
                  className="voice-btn"
                  onClick={onVoiceInput}
                  title="语音输入"
                  disabled={disabled || isStreaming}
                >
                  <svg
                    width="20"
                    height="20"
                    viewBox="0 0 24 24"
                    fill="none"
                    stroke="currentColor"
                    strokeWidth="2"
                    strokeLinecap="round"
                    strokeLinejoin="round"
                  >
                    <path d="M12 1a3 3 0 0 0-3 3v8a3 3 0 0 0 6 0V4a3 3 0 0 0-3-3z" />
                    <path d="M19 10v2a7 7 0 0 1-14 0v-2" />
                    <line x1="12" y1="19" x2="12" y2="23" />
                    <line x1="8" y1="23" x2="16" y2="23" />
                  </svg>
                </button>
              )}

              {/* Markdown Preview Toggle */}
              <button
                className="markdown-toggle"
                onClick={() => setShowPreview(!showPreview)}
                title={showPreview ? '隐藏预览' : 'Markdown 预览'}
                disabled={!value.trim()}
              >
                <svg
                  width="20"
                  height="20"
                  viewBox="0 0 24 24"
                  fill="none"
                  stroke="currentColor"
                  strokeWidth="2"
                  strokeLinecap="round"
                  strokeLinejoin="round"
                >
                  <path d="M14.5 2H6a2 2 0 0 0-2 2v16a2 2 0 0 0 2 2h12a2 2 0 0 0 2-2V7.5L14.5 2z" />
                  <polyline points="14 2 14 8 20 8" />
                  <path d="M8 13h2" />
                  <path d="M8 17h2" />
                  <path d="M14 13h2" />
                  <path d="M14 17h2" />
                </svg>
              </button>
            </div>

            <div className="input-right-actions">
              {/* Character Count */}
              {value.length > 0 && (
                <span
                  style={{
                    fontSize: '12px',
                    color: value.length > 2000 ? 'var(--error-color)' : 'var(--text-tertiary)',
                  }}
                >
                  {value.length}/2000
                </span>
              )}

              {/* Send Button */}
              <button
                className={`send-btn ${isStreaming ? 'sending' : ''}`}
                onClick={handleSend}
                disabled={!hasContent || disabled || isStreaming}
                title="发送 (Enter)"
              >
                {isStreaming ? (
                  <svg
                    width="20"
                    height="20"
                    viewBox="0 0 24 24"
                    fill="none"
                    stroke="currentColor"
                    strokeWidth="2"
                    strokeLinecap="round"
                    strokeLinejoin="round"
                    className="animate-spin"
                  >
                    <path d="M21 12a9 9 0 1 1-6.219-8.56" />
                  </svg>
                ) : (
                  <svg
                    width="20"
                    height="20"
                    viewBox="0 0 24 24"
                    fill="currentColor"
                  >
                    <path d="M2.01 21L23 12 2.01 3 2 10l15 2-15 2z" />
                  </svg>
                )}
              </button>
            </div>
          </div>

          {/* Input Hints */}
          <div className="input-hints">
            <span>支持 Markdown 和 LaTeX 公式</span>
            <span>Enter 发送 · Shift+Enter 换行</span>
          </div>
        </div>
      </div>
    </div>
  );
};

export default ChatInput;
