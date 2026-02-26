import './styles.css';

/**
 * @returns {JSX.Element}
 */
export const TypingIndicator = () => {
  return (
    <div className="chat-stream-message">
      <div className="chat-stream-message-avatar assistant">
        AI
      </div>
      <div className="chat-stream-message-content assistant">
        <div className="chat-stream-bubble assistant">
          <div className="typing-indicator">
            <span className="typing-indicator-dot" />
            <span className="typing-indicator-dot" />
            <span className="typing-indicator-dot" />
          </div>
        </div>
      </div>
    </div>
  );
};
