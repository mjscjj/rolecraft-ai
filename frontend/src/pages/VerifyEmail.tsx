import { useState } from 'react';
import { useNavigate } from 'react-router-dom';
import { Mail, ArrowLeft, CheckCircle } from 'lucide-react';
import client from '../api/client';

export const VerifyEmail = () => {
  const navigate = useNavigate();
  const [email, setEmail] = useState('');
  const [loading, setLoading] = useState(false);
  const [sent, setSent] = useState(false);
  const [error, setError] = useState('');

  const handleResend = async () => {
    if (!email) {
      setError('请输入邮箱地址');
      return;
    }

    setLoading(true);
    setError('');

    try {
      await client.post('/auth/resend-verification', { email });
      setSent(true);
    } catch (err: any) {
      setError(err.response?.data?.message || '发送失败，请重试');
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
            <Mail className="w-8 h-8 text-indigo-600" />
          </div>

          {/* 标题 */}
          <h1 className="text-2xl font-bold text-gray-900 text-center mb-2">
            验证邮箱
          </h1>

          <p className="text-gray-600 text-center mb-6">
            请输入您的邮箱地址，我们将发送验证链接
          </p>

          {/* 成功提示 */}
          {sent ? (
            <div className="mb-6 p-4 bg-green-50 border border-green-200 rounded-lg">
              <div className="flex items-start gap-3">
                <CheckCircle className="w-5 h-5 text-green-600 flex-shrink-0 mt-0.5" />
                <div>
                  <p className="text-green-800 font-medium">验证邮件已发送</p>
                  <p className="text-green-700 text-sm mt-1">
                    请检查邮箱 {email}，点击验证链接完成验证
                  </p>
                </div>
              </div>
            </div>
          ) : (
            /* 输入框 */
            <div className="mb-6">
              <label className="block text-sm font-medium text-gray-700 mb-2">
                邮箱地址
              </label>
              <input
                type="email"
                value={email}
                onChange={(e) => setEmail(e.target.value)}
                placeholder="your@email.com"
                className="w-full px-4 py-3 border border-gray-300 rounded-lg focus:ring-2 focus:ring-indigo-500 focus:border-transparent"
              />
              {error && (
                <p className="mt-2 text-sm text-red-600">{error}</p>
              )}
            </div>
          )}

          {/* 按钮 */}
          <div className="space-y-3">
            {!sent ? (
              <button
                onClick={handleResend}
                disabled={loading}
                className="w-full py-3 bg-indigo-600 text-white rounded-lg hover:bg-indigo-700 transition-colors font-medium disabled:opacity-50"
              >
                {loading ? '发送中...' : '发送验证邮件'}
              </button>
            ) : (
              <button
                onClick={() => setSent(false)}
                className="w-full py-3 bg-gray-200 text-gray-700 rounded-lg hover:bg-gray-300 transition-colors font-medium"
              >
                重新输入邮箱
              </button>
            )}

            <button
              onClick={() => navigate('/login')}
              className="w-full py-3 bg-white border border-gray-300 text-gray-700 rounded-lg hover:bg-gray-50 transition-colors font-medium flex items-center justify-center gap-2"
            >
              <ArrowLeft className="w-4 h-4" />
              返回登录
            </button>
          </div>
        </div>

        {/* 提示信息 */}
        <div className="mt-6 text-center text-sm text-gray-600">
          <p>提示：验证邮件可能在 5 分钟内到达，请检查垃圾邮件箱</p>
        </div>
      </div>
    </div>
  );
};
