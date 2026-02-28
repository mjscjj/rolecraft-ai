import { useState } from 'react';
import { useNavigate } from 'react-router-dom';
import { Lock, Mail, ArrowLeft, CheckCircle } from 'lucide-react';
import client from '../api/client';

export const ForgotPassword = () => {
  const navigate = useNavigate();
  const [email, setEmail] = useState('');
  const [step, setStep] = useState<'email' | 'reset'>('email');
  const [token, setToken] = useState('');
  const [newPassword, setNewPassword] = useState('');
  const [confirmPassword, setConfirmPassword] = useState('');
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState('');
  const [success, setSuccess] = useState(false);

  const handleSendResetLink = async () => {
    if (!email) {
      setError('请输入邮箱地址');
      return;
    }

    setLoading(true);
    setError('');

    try {
      await client.post('/auth/forgot-password', { email });
      setStep('reset');
    } catch (err: any) {
      setError(err.response?.data?.message || '发送失败，请重试');
    } finally {
      setLoading(false);
    }
  };

  const handleResetPassword = async () => {
    if (!token) {
      setError('请输入重置令牌');
      return;
    }

    if (!newPassword || newPassword.length < 6) {
      setError('密码至少 6 位');
      return;
    }

    if (newPassword !== confirmPassword) {
      setError('两次输入的密码不一致');
      return;
    }

    setLoading(true);
    setError('');

    try {
      await client.post('/auth/reset-password', {
        token,
        newPassword,
      });
      setSuccess(true);
    } catch (err: any) {
      setError(err.response?.data?.message || '重置失败，请重试');
    } finally {
      setLoading(false);
    }
  };

  return (
    <div className="min-h-screen bg-gray-50 flex items-center justify-center p-4">
      <div className="max-w-md w-full">
        {/* 卡片 */}
        <div className="bg-white rounded-2xl shadow-sm border border-gray-200 p-8">
          {/* 图标 */}
          <div className="w-16 h-16 mx-auto mb-6 bg-indigo-100 rounded-full flex items-center justify-center">
            <Lock className="w-8 h-8 text-indigo-600" />
          </div>

          {/* 标题 */}
          <h1 className="text-2xl font-bold text-gray-900 text-center mb-2">
            忘记密码
          </h1>

          <p className="text-gray-600 text-center mb-6">
            别担心，我们将帮助您重置密码
          </p>

          {/* 成功提示 */}
          {success ? (
            <div className="mb-6 p-4 bg-green-50 border border-green-200 rounded-lg">
              <div className="flex items-start gap-3">
                <CheckCircle className="w-5 h-5 text-green-600 flex-shrink-0 mt-0.5" />
                <div>
                  <p className="text-green-800 font-medium">密码重置成功</p>
                  <p className="text-green-700 text-sm mt-1">
                    请使用新密码重新登录
                  </p>
                </div>
              </div>
            </div>
          ) : (
            <>
              {/* Step 1: 输入邮箱 */}
              {step === 'email' && (
                <div className="mb-6">
                  <label className="block text-sm font-medium text-gray-700 mb-2">
                    邮箱地址
                  </label>
                  <div className="relative">
                    <Mail className="absolute left-3 top-1/2 -translate-y-1/2 w-5 h-5 text-gray-400" />
                    <input
                      type="email"
                      value={email}
                      onChange={(e) => setEmail(e.target.value)}
                      placeholder="your@email.com"
                      className="w-full pl-10 pr-4 py-3 border border-gray-300 rounded-lg focus:ring-2 focus:ring-indigo-500 focus:border-transparent"
                    />
                  </div>
                  {error && (
                    <p className="mt-2 text-sm text-red-600">{error}</p>
                  )}
                </div>
              )}

              {/* Step 2: 输入令牌和新密码 */}
              {step === 'reset' && (
                <div className="space-y-4 mb-6">
                  <div>
                    <label className="block text-sm font-medium text-gray-700 mb-2">
                      重置令牌
                    </label>
                    <input
                      type="text"
                      value={token}
                      onChange={(e) => setToken(e.target.value)}
                      placeholder="输入邮件中的令牌"
                      className="w-full px-4 py-3 border border-gray-300 rounded-lg focus:ring-2 focus:ring-indigo-500 focus:border-transparent"
                    />
                  </div>

                  <div>
                    <label className="block text-sm font-medium text-gray-700 mb-2">
                      新密码
                    </label>
                    <input
                      type="password"
                      value={newPassword}
                      onChange={(e) => setNewPassword(e.target.value)}
                      placeholder="至少 6 位"
                      className="w-full px-4 py-3 border border-gray-300 rounded-lg focus:ring-2 focus:ring-indigo-500 focus:border-transparent"
                    />
                  </div>

                  <div>
                    <label className="block text-sm font-medium text-gray-700 mb-2">
                      确认新密码
                    </label>
                    <input
                      type="password"
                      value={confirmPassword}
                      onChange={(e) => setConfirmPassword(e.target.value)}
                      placeholder="再次输入新密码"
                      className="w-full px-4 py-3 border border-gray-300 rounded-lg focus:ring-2 focus:ring-indigo-500 focus:border-transparent"
                    />
                  </div>

                  {error && (
                    <p className="text-sm text-red-600">{error}</p>
                  )}
                </div>
              )}
            </>
          )}

          {/* 按钮 */}
          <div className="space-y-3">
            {!success ? (
              <>
                {step === 'email' ? (
                  <button
                    onClick={handleSendResetLink}
                    disabled={loading}
                    className="w-full py-3 bg-indigo-600 text-white rounded-lg hover:bg-indigo-700 transition-colors font-medium disabled:opacity-50"
                  >
                    {loading ? '发送中...' : '发送重置链接'}
                  </button>
                ) : (
                  <button
                    onClick={handleResetPassword}
                    disabled={loading}
                    className="w-full py-3 bg-indigo-600 text-white rounded-lg hover:bg-indigo-700 transition-colors font-medium disabled:opacity-50"
                  >
                    {loading ? '重置中...' : '重置密码'}
                  </button>
                )}

                <button
                  onClick={() => {
                    setStep('email');
                    navigate('/login');
                  }}
                  className="w-full py-3 bg-white border border-gray-300 text-gray-700 rounded-lg hover:bg-gray-50 transition-colors font-medium flex items-center justify-center gap-2"
                >
                  <ArrowLeft className="w-4 h-4" />
                  返回登录
                </button>
              </>
            ) : (
              <button
                onClick={() => navigate('/login')}
                className="w-full py-3 bg-indigo-600 text-white rounded-lg hover:bg-indigo-700 transition-colors font-medium"
              >
                返回登录
              </button>
            )}
          </div>
        </div>

        {/* 提示信息 */}
        <div className="mt-6 text-center text-sm text-gray-600">
          {step === 'email' && (
            <p>提示：重置链接将发送到您的邮箱，有效期 30 分钟</p>
          )}
          {step === 'reset' && (
            <p>提示：请检查邮箱获取重置令牌</p>
          )}
        </div>
      </div>
    </div>
  );
};
