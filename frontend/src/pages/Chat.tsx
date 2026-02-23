import { FC, useState, useRef, useEffect } from 'react';
import { Send, Paperclip, MoreVertical, Copy, RotateCcw, ThumbsUp, ThumbsDown } from 'lucide-react';
import { Message } from '../types';

const mockMessages: Message[] = [
  {
    id: '1',
    role: 'assistant',
    content: '你好！我是你的营销助手，可以帮助你制定营销策略、撰写文案、分析市场趋势。请问有什么可以帮你的？',
    createdAt: '2024-01-20 10:00:00',
  },
  {
    id: '2',
    role: 'user',
    content: '帮我写一条关于新品发布的朋友圈文案',
    createdAt: '2024-01-20 10:01:00',
  },
  {
    id: '3',
    role: 'assistant',
    content: `好的！根据新品发布的需求，我为你准备了几个版本：

**版本一（正式商务风）**
🎉 重磅发布！我们全新产品正式上线！

经过数月的精心打磨，我们终于迎来了这款划时代的产品。它不仅代表了技术的突破，更承载着我们对品质的执着追求。

✨ 核心亮点：
• 领先技术，性能卓越
• 匠心设计，品质保证  
• 贴心服务，全程无忧

现在就体验，开启全新篇章！

#新品发布 #科技创新 #品质生活

---

**版本二（轻松活泼风）**
姐妹们！好消息来啦～🎊

我们的新品终于和大家见面啦！这次真的准备了超级久，就为了给宝贝们最好的体验！

💖 为什么值得期待：
✅ 颜值高到没朋友
✅ 好用程度五颗星
✅ 价格美丽到哭泣

第一批还有限时优惠哦，手慢无！💨

点击下方链接，抢先体验 👇

#新品上市 #种草好物 #限时优惠`,
    sources: ['产品手册.pdf', '营销策略指南.pdf'],
    createdAt: '2024-01-20 10:02:00',
  },
];

const quickCommands = ['/总结', '/翻译', '/扩展', '/精炼', '/润色'];

export const Chat: FC = () => {
  const [messages, setMessages] = useState<Message[]>(mockMessages);
  const [inputValue, setInputValue] = useState('');
  const [isLoading, setIsLoading] = useState(false);
  const messagesEndRef = useRef<HTMLDivElement>(null);

  const scrollToBottom = () => {
    messagesEndRef.current?.scrollIntoView({ behavior: 'smooth' });
  };

  useEffect(() => {
    scrollToBottom();
  }, [messages]);

  const handleSend = () => {
    if (!inputValue.trim()) return;

    const newMessage: Message = {
      id: Date.now().toString(),
      role: 'user',
      content: inputValue,
      createdAt: new Date().toISOString(),
    };

    setMessages([...messages, newMessage]);
    setInputValue('');
    setIsLoading(true);

    // 模拟 AI 回复
    setTimeout(() => {
      const aiResponse: Message = {
        id: (Date.now() + 1).toString(),
        role: 'assistant',
        content: '收到！我来帮你优化这段文案。基于你提供的背景信息，我建议从以下几个角度进行调整...',
        createdAt: new Date().toISOString(),
      };
      setMessages(prev => [...prev, aiResponse]);
      setIsLoading(false);
    }, 1500);
  };

  return (
    <div className="h-[calc(100vh-8rem)] flex flex-col">
      {/* Chat Header */}
      <div className="flex items-center justify-between pb-4 border-b border-slate-200">
        <div className="flex items-center gap-3">
          <div className="w-10 h-10 rounded-full bg-gradient-to-br from-primary to-primary-dark flex items-center justify-center text-white font-semibold">
            营
          </div>
          <div>
            <h2 className="font-semibold text-slate-900">营销专家</h2>
            <p className="text-xs text-slate-500">在线</p>
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
      <div className="flex-1 overflow-y-auto py-6 space-y-6">
        {messages.map((message) => (
          <div
            key={message.id}
            className={`flex gap-4 ${message.role === 'user' ? 'flex-row-reverse' : ''}`}
          >
            {/* Avatar */}
            <div className={`w-10 h-10 rounded-full flex items-center justify-center flex-shrink-0 ${
              message.role === 'user'
                ? 'bg-slate-200 text-slate-600'
                : 'bg-gradient-to-br from-primary to-primary-dark text-white'
            }`}>
              {message.role === 'user' ? 'U' : 'AI'}
            </div>

            {/* Message Content */}
            <div className={`max-w-[70%] ${message.role === 'user' ? 'items-end' : 'items-start'}`}>
              <div className={`p-4 rounded-2xl ${
                message.role === 'user'
                  ? 'bg-slate-900 text-white rounded-tr-none'
                  : 'bg-white border border-slate-200 rounded-tl-none shadow-sm'
              }`}>
                <div className={`prose prose-sm max-w-none ${
                  message.role === 'user' ? 'prose-invert' : ''
                }`}>
                  {message.content.split('\n').map((line, i) => (
                    <p key={i} className={line.trim() === '' ? 'h-2' : ''}>
                      {line}
                    </p>
                  ))}
                </div>

                {/* Sources */}
                {message.sources && message.sources.length > 0 && (
                  <div className="mt-4 pt-3 border-t border-slate-200/20">
                    <p className="text-xs opacity-70 mb-2">📚 参考来源：</p>
                    <div className="flex flex-wrap gap-2">
                      {message.sources.map((source, i) => (
                        <span
                          key={i}
                          className="text-xs px-2 py-1 bg-white/10 rounded cursor-pointer hover:bg-white/20 transition-colors"
                        >
                          {source}
                        </span>
                      ))}
                    </div>
                  </div>
                )}
              </div>

              {/* Actions */}
              {message.role === 'assistant' && (
                <div className="flex items-center gap-1 mt-2 opacity-0 hover:opacity-100 transition-opacity">
                  <button className="p-1.5 text-slate-400 hover:text-slate-600 hover:bg-slate-100 rounded transition-colors" title="复制">
                    <Copy className="w-4 h-4" />
                  </button>
                  <button className="p-1.5 text-slate-400 hover:text-slate-600 hover:bg-slate-100 rounded transition-colors" title="重新生成">
                    <RotateCcw className="w-4 h-4" />
                  </button>
                  <button className="p-1.5 text-slate-400 hover:text-green-500 hover:bg-green-50 rounded transition-colors" title="有用">
                    <ThumbsUp className="w-4 h-4" />
                  </button>
                  <button className="p-1.5 text-slate-400 hover:text-red-500 hover:bg-red-50 rounded transition-colors" title="无用">
                    <ThumbsDown className="w-4 h-4" />
                  </button>
                </div>
              )}
            </div>
          </div>
        ))}

        {/* Loading Indicator */}
        {isLoading && (
          <div className="flex gap-4">
            <div className="w-10 h-10 rounded-full bg-gradient-to-br from-primary to-primary-dark flex items-center justify-center text-white">
              AI
            </div>
            <div className="bg-white border border-slate-200 rounded-2xl rounded-tl-none p-4 shadow-sm">
              <div className="flex gap-1">
                <span className="w-2 h-2 bg-slate-400 rounded-full animate-bounce" style={{ animationDelay: '0ms' }} />
                <span className="w-2 h-2 bg-slate-400 rounded-full animate-bounce" style={{ animationDelay: '150ms' }} />
                <span className="w-2 h-2 bg-slate-400 rounded-full animate-bounce" style={{ animationDelay: '300ms' }} />
              </div>
            </div>
          </div>
        )}

        <div ref={messagesEndRef} />
      </div>

      {/* Input Area */}
      <div className="border-t border-slate-200 pt-4">
        {/* Quick Commands */}
        <div className="flex items-center gap-2 mb-3 overflow-x-auto pb-1">
          {quickCommands.map(cmd => (
            <button
              key={cmd}
              onClick={() => setInputValue(cmd + ' ')}
              className="text-xs px-3 py-1.5 bg-slate-100 text-slate-600 rounded-full hover:bg-slate-200 transition-colors whitespace-nowrap"
            >
              {cmd}
            </button>
          ))}
        </div>

        {/* Input */}
        <div className="flex items-end gap-2 bg-white border border-slate-200 rounded-2xl p-2 shadow-sm">
          <button className="p-3 text-slate-400 hover:text-slate-600 hover:bg-slate-100 rounded-xl transition-colors">
            <Paperclip className="w-5 h-5" />
          </button>
          <textarea
            value={inputValue}
            onChange={(e) => setInputValue(e.target.value)}
            onKeyDown={(e) => {
              if (e.key === 'Enter' && !e.shiftKey) {
                e.preventDefault();
                handleSend();
              }
            }}
            placeholder="输入消息..."
            rows={1}
            className="flex-1 py-3 px-2 outline-none resize-none max-h-32"
            style={{ minHeight: '48px' }}
          />
          <button
            onClick={handleSend}
            disabled={!inputValue.trim() || isLoading}
            className="p-3 bg-primary text-white rounded-xl hover:bg-primary-dark transition-colors disabled:opacity-50 disabled:cursor-not-allowed"
          >
            <Send className="w-5 h-5" />
          </button>
        </div>
        <p className="text-xs text-slate-400 mt-2 text-center">按 Enter 发送，Shift + Enter 换行</p>
      </div>
    </div>
  );
};