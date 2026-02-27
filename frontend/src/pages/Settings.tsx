import type { FC } from 'react';
import { useEffect, useState } from 'react';
import userApi from '../api/user';

export const Settings: FC = () => {
  const [name, setName] = useState('');
  const [avatar, setAvatar] = useState('');
  const [model, setModel] = useState(localStorage.getItem('preferredModel') || 'qwen-plus');
  const [temperature, setTemperature] = useState(localStorage.getItem('preferredTemperature') || '0.7');
  const [customAPIKey, setCustomAPIKey] = useState(localStorage.getItem('customAPIKey') || '');
  const [newPassword, setNewPassword] = useState('');
  const [saving, setSaving] = useState(false);
  const [message, setMessage] = useState('');

  useEffect(() => {
    const loadUser = async () => {
      try {
        const me = await userApi.getMe();
        setName(me.name || '');
        setAvatar(me.avatar || '');
      } catch {
        setMessage('加载用户信息失败');
      }
    };
    loadUser();
  }, []);

  const handleSaveProfile = async () => {
    setSaving(true);
    setMessage('');
    try {
      const payload: { name: string; avatar: string; password?: string } = { name, avatar };
      if (newPassword.trim()) payload.password = newPassword.trim();
      const updated = await userApi.updateMe(payload);
      localStorage.setItem('user', JSON.stringify(updated));
      setNewPassword('');
      setMessage('个人资料已保存');
    } catch (err: any) {
      setMessage(err?.message || '保存失败');
    } finally {
      setSaving(false);
    }
  };

  const handleSavePreference = () => {
    localStorage.setItem('preferredModel', model);
    localStorage.setItem('preferredTemperature', temperature);
    localStorage.setItem('customAPIKey', customAPIKey);
    setMessage('偏好设置已保存');
  };

  return (
    <div className="space-y-6">
      <div>
        <h1 className="text-2xl font-bold text-slate-900">设置</h1>
        <p className="mt-1 text-slate-500">管理个人信息与模型偏好</p>
      </div>

      {message && (
        <div className="rounded-lg border border-slate-200 bg-white px-4 py-3 text-sm text-slate-700">
          {message}
        </div>
      )}

      <section className="rounded-xl border border-slate-200 bg-white p-6">
        <h2 className="mb-4 text-lg font-semibold text-slate-900">个人资料</h2>
        <div className="space-y-4">
          <div>
            <label className="mb-1 block text-sm text-slate-600">昵称</label>
            <input
              value={name}
              onChange={(e) => setName(e.target.value)}
              className="w-full rounded-lg border border-slate-200 px-3 py-2 focus:outline-none focus:ring-2 focus:ring-primary/30"
            />
          </div>
          <div>
            <label className="mb-1 block text-sm text-slate-600">头像 URL</label>
            <input
              value={avatar}
              onChange={(e) => setAvatar(e.target.value)}
              className="w-full rounded-lg border border-slate-200 px-3 py-2 focus:outline-none focus:ring-2 focus:ring-primary/30"
            />
          </div>
          <div>
            <label className="mb-1 block text-sm text-slate-600">新密码（可选）</label>
            <input
              type="password"
              value={newPassword}
              onChange={(e) => setNewPassword(e.target.value)}
              placeholder="留空则不修改密码"
              className="w-full rounded-lg border border-slate-200 px-3 py-2 focus:outline-none focus:ring-2 focus:ring-primary/30"
            />
          </div>
          <button
            onClick={handleSaveProfile}
            disabled={saving}
            className="rounded-lg bg-indigo-600 px-4 py-2 text-white hover:bg-indigo-700 disabled:cursor-not-allowed disabled:opacity-60"
          >
            {saving ? '保存中...' : '保存资料'}
          </button>
        </div>
      </section>

      <section className="rounded-xl border border-slate-200 bg-white p-6">
        <h2 className="mb-4 text-lg font-semibold text-slate-900">模型偏好</h2>
        <div className="space-y-4">
          <div className="flex items-center gap-3">
            <select
              value={model}
              onChange={(e) => setModel(e.target.value)}
              className="rounded-lg border border-slate-200 px-3 py-2 focus:outline-none focus:ring-2 focus:ring-primary/30"
            >
              <option value="qwen-plus">Qwen Plus</option>
              <option value="qwen-max">Qwen Max</option>
              <option value="gpt-4">GPT-4</option>
              <option value="gpt-3.5-turbo">GPT-3.5 Turbo</option>
            </select>
            <input
              type="number"
              min="0"
              max="2"
              step="0.1"
              value={temperature}
              onChange={(e) => setTemperature(e.target.value)}
              className="w-24 rounded-lg border border-slate-200 px-3 py-2 focus:outline-none focus:ring-2 focus:ring-primary/30"
            />
          </div>
          <div>
            <label className="mb-1 block text-sm text-slate-600">自定义 API Key（MVP：会话级覆盖）</label>
            <input
              type="password"
              value={customAPIKey}
              onChange={(e) => setCustomAPIKey(e.target.value)}
              placeholder="留空则使用服务端默认 Key"
              className="w-full rounded-lg border border-slate-200 px-3 py-2 focus:outline-none focus:ring-2 focus:ring-primary/30"
            />
          </div>
          <button
            onClick={handleSavePreference}
            className="rounded-lg bg-slate-900 px-4 py-2 text-white hover:bg-slate-800"
          >
            保存偏好
          </button>
        </div>
      </section>
    </div>
  );
};
