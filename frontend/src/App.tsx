import { Suspense, lazy } from 'react';
import { BrowserRouter, Routes, Route, Navigate, Outlet } from 'react-router-dom';
import { Layout } from './components/Layout';

const Dashboard = lazy(() => import('./pages/Dashboard').then((m) => ({ default: m.Dashboard })));
const Login = lazy(() => import('./pages/Login').then((m) => ({ default: m.Login })));
const ChatWebUI = lazy(() => import('./pages/ChatWebUI'));
const Chat = lazy(() => import('./pages/Chat').then((m) => ({ default: m.Chat })));
const RoleEditor = lazy(() => import('./pages/RoleEditor').then((m) => ({ default: m.RoleEditor })));
const RoleMarket = lazy(() => import('./pages/RoleMarket').then((m) => ({ default: m.RoleMarket })));
const KnowledgeBase = lazy(() => import('./pages/KnowledgeBase').then((m) => ({ default: m.KnowledgeBase })));
const Analytics = lazy(() => import('./pages/Analytics').then((m) => ({ default: m.Analytics })));
const Settings = lazy(() => import('./pages/Settings').then((m) => ({ default: m.Settings })));

const AppLoading = () => (
  <div style={{ minHeight: '100vh', display: 'grid', placeItems: 'center', color: '#64748b' }}>
    加载中...
  </div>
);

const ProtectedLayout = () => {
  const token = typeof window !== 'undefined' ? localStorage.getItem('token') : null;
  if (!token) return <Navigate to="/login" replace />;

  return (
    <Layout>
      <Outlet />
    </Layout>
  );
};

const ProtectedOnly = () => {
  const token = typeof window !== 'undefined' ? localStorage.getItem('token') : null;
  if (!token) return <Navigate to="/login" replace />;
  return <Outlet />;
};

const App = () => {
  return (
    <BrowserRouter>
      <Suspense fallback={<AppLoading />}>
        <Routes>
          <Route path="/login" element={<Login />} />
          <Route element={<ProtectedLayout />}>
            <Route path="/" element={<Dashboard />} />
            <Route path="/roles" element={<RoleMarket />} />
            <Route path="/roles/create" element={<RoleEditor />} />
            <Route path="/documents" element={<KnowledgeBase />} />
            <Route path="/analytics" element={<Analytics />} />
            <Route path="/settings" element={<Settings />} />
          </Route>
          <Route element={<ProtectedOnly />}>
            <Route path="/chat" element={<ChatWebUI />} />
            <Route path="/chat/:roleId" element={<ChatWebUI />} />
            <Route path="/chat-legacy/:roleId" element={<Chat />} />
          </Route>
          <Route path="*" element={<Navigate to="/" />} />
        </Routes>
      </Suspense>
    </BrowserRouter>
  );
};

export default App;
