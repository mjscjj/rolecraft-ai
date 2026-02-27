import React, { useState, useEffect } from 'react';
import { useParams, useNavigate } from 'react-router-dom';
import testApi, { TestReport, TestHistory } from '../api/test';
import { RolePreview } from '../components/RolePreview';
import { TestDialog } from '../components/TestDialog';

const TestReportPage: React.FC = () => {
  const { roleId } = useParams<{ roleId: string }>();
  const navigate = useNavigate();
  
  const [report, setReport] = useState<TestReport | null>(null);
  const [history, setHistory] = useState<TestHistory[]>([]);
  const [isLoading, setIsLoading] = useState(true);
  const [activeTab, setActiveTab] = useState<'report' | 'history' | 'preview'>('report');

  useEffect(() => {
    if (roleId) {
      loadTestData();
    }
  }, [roleId]);

  const loadTestData = async () => {
    setIsLoading(true);
    try {
      const [reportData, historyData] = await Promise.all([
        testApi.getTestReport(roleId!),
        testApi.getTestHistory(roleId!),
      ]);
      setReport(reportData);
      setHistory(historyData.history);
    } catch (error) {
      console.error('加载测试数据失败:', error);
    } finally {
      setIsLoading(false);
    }
  };

  const handleExportReport = async (format: string = 'pdf') => {
    if (!roleId) return;
    
    try {
      const blob = await testApi.exportTestReport(roleId, format);
      const url = window.URL.createObjectURL(blob);
      const link = document.createElement('a');
      link.href = url;
      link.download = `test_report_${roleId}.${format}`;
      document.body.appendChild(link);
      link.click();
      document.body.removeChild(link);
      window.URL.revokeObjectURL(url);
    } catch (error) {
      console.error('导出失败:', error);
      alert('导出失败，请稍后重试');
    }
  };

  if (isLoading) {
    return (
      <div className="min-h-screen bg-slate-50 flex items-center justify-center">
        <div className="text-center">
          <div className="animate-spin rounded-full h-12 w-12 border-b-2 border-primary mx-auto"></div>
          <p className="mt-4 text-slate-600">加载测试数据...</p>
        </div>
      </div>
    );
  }

  return (
    <div className="min-h-screen bg-slate-50 py-8">
      <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
        {/* 头部 */}
        <div className="mb-8">
          <button
            onClick={() => navigate('/roles')}
            className="text-sm text-slate-600 hover:text-slate-900 mb-4"
          >
            ← 返回角色列表
          </button>
          <h1 className="text-3xl font-bold text-slate-900">测试报告</h1>
          {report && (
            <p className="text-slate-600 mt-2">角色：{report.roleName}</p>
          )}
        </div>

        {/* 标签页 */}
        <div className="mb-6">
          <div className="border-b border-slate-200">
            <nav className="-mb-px flex space-x-8">
              {[
                { id: 'report', label: '测试报告' },
                { id: 'history', label: '测试历史' },
                { id: 'preview', label: '角色预览' },
              ].map((tab) => (
                <button
                  key={tab.id}
                  onClick={() => setActiveTab(tab.id as any)}
                  className={`py-4 px-1 border-b-2 font-medium text-sm ${
                    activeTab === tab.id
                      ? 'border-primary text-primary'
                      : 'border-transparent text-slate-500 hover:text-slate-700 hover:border-slate-300'
                  }`}
                >
                  {tab.label}
                </button>
              ))}
            </nav>
          </div>
        </div>

        {/* 测试报告 */}
        {activeTab === 'report' && report && (
          <div className="space-y-6">
            {/* 概览统计 */}
            <div className="grid grid-cols-1 md:grid-cols-4 gap-4">
              <div className="bg-white rounded-xl p-6 shadow">
                <p className="text-sm text-slate-600">总测试次数</p>
                <p className="text-3xl font-bold text-slate-900 mt-2">
                  {report.totalTests}
                </p>
              </div>
              <div className="bg-white rounded-xl p-6 shadow">
                <p className="text-sm text-slate-600">平均评分</p>
                <p className="text-3xl font-bold text-primary mt-2">
                  {report.averageRating.toFixed(1)}⭐
                </p>
              </div>
              <div className="bg-white rounded-xl p-6 shadow">
                <p className="text-sm text-slate-600">通过率</p>
                <p className="text-3xl font-bold text-green-600 mt-2">
                  {report.passRate.toFixed(1)}%
                </p>
              </div>
              <div className="bg-white rounded-xl p-6 shadow flex items-center justify-between">
                <div>
                  <p className="text-sm text-slate-600">导出报告</p>
                  <p className="text-xs text-slate-500 mt-1">PDF / Markdown</p>
                </div>
                <button
                  onClick={() => handleExportReport('pdf')}
                  className="px-4 py-2 bg-primary text-white rounded-lg hover:bg-primary-dark text-sm"
                >
                  导出
                </button>
              </div>
            </div>

            {/* 评分分布 */}
            <div className="bg-white rounded-xl p-6 shadow">
              <h2 className="text-xl font-semibold text-slate-900 mb-4">
                评分分布
              </h2>
              <div className="space-y-3">
                {[5, 4, 3, 2, 1].map((rating) => {
                  const count = report.testsByRating[rating] || 0;
                  const percentage = report.totalTests > 0 
                    ? (count / report.totalTests) * 100 
                    : 0;
                  
                  return (
                    <div key={rating} className="flex items-center gap-3">
                      <span className="text-sm text-slate-600 w-12">
                        {rating}星
                      </span>
                      <div className="flex-1 bg-slate-100 rounded-full h-4 overflow-hidden">
                        <div
                          className="bg-yellow-400 h-full rounded-full transition-all"
                          style={{ width: `${percentage}%` }}
                        />
                      </div>
                      <span className="text-sm text-slate-600 w-12 text-right">
                        {count}
                      </span>
                    </div>
                  );
                })}
              </div>
            </div>

            {/* 改进趋势 */}
            <div className="bg-white rounded-xl p-6 shadow">
              <h2 className="text-xl font-semibold text-slate-900 mb-4">
                改进趋势
              </h2>
              <div className="h-64 flex items-end justify-between gap-2">
                {report.improvementTrend.map((item, idx) => (
                  <div key={idx} className="flex-1 flex flex-col items-center">
                    <div
                      className="w-full bg-primary/80 rounded-t transition-all hover:bg-primary"
                      style={{ 
                        height: `${(item.avgRating / 5) * 200}px`,
                        minHeight: '20px'
                      }}
                    />
                    <p className="text-xs text-slate-600 mt-2">
                      {item.date.slice(5)}
                    </p>
                    <p className="text-xs text-slate-500">
                      {item.avgRating.toFixed(1)}⭐
                    </p>
                  </div>
                ))}
              </div>
            </div>

            {/* 改进建议 */}
            <div className="bg-white rounded-xl p-6 shadow">
              <h2 className="text-xl font-semibold text-slate-900 mb-4">
                改进建议
              </h2>
              <ul className="space-y-3">
                {report.suggestions.map((suggestion, idx) => (
                  <li
                    key={idx}
                    className="flex items-start gap-3 p-3 bg-blue-50 rounded-lg"
                  >
                    <span className="text-blue-500 text-lg">💡</span>
                    <span className="text-slate-700">{suggestion}</span>
                  </li>
                ))}
              </ul>
            </div>
          </div>
        )}

        {/* 测试历史 */}
        {activeTab === 'history' && (
          <div className="bg-white rounded-xl shadow">
            <div className="p-6 border-b border-slate-200">
              <h2 className="text-xl font-semibold text-slate-900">
                测试历史 ({history.length})
              </h2>
            </div>
            <div className="divide-y divide-slate-200">
              {history.length === 0 ? (
                <div className="p-8 text-center text-slate-500">
                  暂无测试记录
                </div>
              ) : (
                history.map((item) => (
                  <div key={item.testId} className="p-4 hover:bg-slate-50">
                    <div className="flex items-start justify-between">
                      <div className="flex-1">
                        <p className="text-slate-900 font-medium">
                          {item.question}
                        </p>
                        <p className="text-sm text-slate-600 mt-1 line-clamp-2">
                          {item.response}
                        </p>
                        <div className="flex items-center gap-4 mt-2 text-xs text-slate-500">
                          <span>{item.testType}</span>
                          <span>
                            {new Date(item.createdAt).toLocaleString('zh-CN')}
                          </span>
                        </div>
                      </div>
                      <div className="flex items-center gap-1 ml-4">
                        {[1, 2, 3, 4, 5].map((star) => (
                          <span
                            key={star}
                            className={
                              star <= item.rating ? 'text-yellow-400' : 'text-slate-300'
                            }
                          >
                            ★
                          </span>
                        ))}
                      </div>
                    </div>
                  </div>
                ))
              )}
            </div>
          </div>
        )}

        {/* 角色预览 */}
        {activeTab === 'preview' && roleId && (
          <div className="max-w-3xl">
            <RolePreview
              role={{ id: roleId }}
              onTestChat={async (message) => {
                // 这里可以调用实际的测试 API
                return `这是测试回复：${message}`;
              }}
            />
          </div>
        )}
      </div>
    </div>
  );
};

export default TestReportPage;
