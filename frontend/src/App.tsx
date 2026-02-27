import { BrowserRouter, Routes, Route, Navigate } from 'react-router-dom';
import { Dashboard } from './pages/Dashboard';
import { Login } from './pages/Login';
import { Chat } from './pages/Chat';
import { RoleEditor } from './pages/RoleEditor';

const App = () => {
  return (
    <BrowserRouter>
      <Routes>
        <Route path="/login" element={<Login />} />
        <Route path="/" element={<Dashboard />} />
        <Route path="/chat/:roleId" element={<Chat />} />
        <Route path="/roles/create" element={<RoleEditor />} />
        <Route path="*" element={<Navigate to="/" />} />
      </Routes>
    </BrowserRouter>
  );
};

export default App;