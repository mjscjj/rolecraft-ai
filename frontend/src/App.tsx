import { FC } from 'react';
import { BrowserRouter, Routes, Route } from 'react-router-dom';
import { Layout } from './components/Layout';
import { Dashboard } from './pages/Dashboard';
import { RoleMarket } from './pages/RoleMarket';
import { RoleEditor } from './pages/RoleEditor';
import { Chat } from './pages/Chat';
import { KnowledgeBase } from './pages/KnowledgeBase';

const App: FC = () => {
  return (
    <BrowserRouter>
      <Layout>
        <Routes>
          <Route path="/" element={<Dashboard />} />
          <Route path="/roles" element={<RoleMarket />} />
          <Route path="/roles/create" element={<RoleEditor />} />
          <Route path="/chat" element={<Chat />} />
          <Route path="/documents" element={<KnowledgeBase />} />
          <Route path="/settings" element={<div className="text-center py-20 text-slate-500">设置页面开发中...</div>} />
        </Routes>
      </Layout>
    </BrowserRouter>
  );
};

export default App;