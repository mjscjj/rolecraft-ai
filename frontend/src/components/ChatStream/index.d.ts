// Type declarations for ChatStream components

export interface Message {
  id: string;
  role: 'user' | 'assistant' | 'system';
  content: string;
  sources?: string[];
  createdAt?: string;
}

export interface ChatStreamProps {
  roleId?: string;
  roleName?: string;
}

export interface MessageListProps {
  messages: Message[];
  isLoading: boolean;
  streamingMessageId?: string | null;
  onCopy?: (content: string) => void;
  onRegenerate?: (messageId: string) => void;
  onFeedback?: (messageId: string, feedback: 'up' | 'down') => void;
}

export interface MessageBubbleProps {
  message: Message;
  onCopy?: (content: string) => void;
  onRegenerate?: (messageId: string) => void;
  onFeedback?: (messageId: string, feedback: 'up' | 'down') => void;
}

export declare const ChatStream: React.FC<ChatStreamProps>;
export declare const MessageList: React.FC<MessageListProps>;
export declare const MessageBubble: React.FC<MessageBubbleProps>;
export declare const TypingIndicator: React.FC;
