// @ts-ignore - JSX component
import { ChatStream } from '../components/ChatStream/index.jsx';

const ChatStreamDemo = () => {
  return (
    <div className="p-6">
      <div className="mb-4">
        <h1 className="text-2xl font-bold text-slate-900">流式聊天组件演示</h1>
        <p className="text-slate-600 mt-2">
          这是一个支持流式输出的聊天组件，参考了 AnythingLLM 的设计。
        </p>
      </div>
      
      <div className="bg-white rounded-xl shadow-sm border border-slate-200 overflow-hidden">
        <ChatStream 
          roleId="demo-role" 
          roleName="AI 助手" 
        />
      </div>
    </div>
  );
};

export default ChatStreamDemo;
