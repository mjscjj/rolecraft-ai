import React, { useState } from 'react';

interface TestVersion {
  versionId: string;
  versionName: string;
  systemPrompt: string;
  modelConfig?: Record<string, any>;
}

interface TestResult {
  versionId: string;
  versionName: string;
  response: string;
  responseTime: number;
  score: number;
  rating: number;
  feedback: string;
}

interface TestDialogProps {
  versions: TestVersion[];
  onTest?: (versions: TestVersion[], question: string) => Promise<TestResult[]>;
  onSaveTest?: (results: TestResult[], question: string) => void;
}

export const TestDialog: React.FC<TestDialogProps> = ({
  versions = [],
  onTest,
  onSaveTest,
}) => {
  const [question, setQuestion] = useState('');
  const [isTesting, setIsTesting] = useState(false);
  const [results, setResults] = useState<TestResult[]>([]);
  const [selectedWinner, setSelectedWinner] = useState<string | null>(null);
  const [testHistory, setTestHistory] = useState<Array<{
    question: string;
    results: TestResult[];
    timestamp: Date;
  }>>([]);

  const handleRunTest = async () => {
    if (!question.trim() || versions.length < 2) return;

    setIsTesting(true);
    try {
      let testResults: TestResult[];
      
      if (onTest) {
        testResults = await onTest(versions, question);
      } else {
        // Mock 测试
        testResults = versions.map((version) => ({
          versionId: version.versionId,
          versionName: version.versionName,
          response: `【${version.versionName}】基于版本 "${version.versionId}" 的回复：${question}`,
          responseTime: Math.random() * 2,
          score: 70 + Math.random() * 25,
          rating: 4,
          feedback: '回复质量良好',
        }));
      }

      setResults(testResults);
      setTestHistory((prev) => [
        ...prev,
        { question, results: testResults, timestamp: new Date() },
      ]);
    } catch (error) {
      console.error('测试失败:', error);
    } finally {
      setIsTesting(false);
    }
  };

  const handleSaveTest = () => {
    if (onSaveTest && results.length > 0) {
      onSaveTest(results, question);
    }
  };

  const handleSelectWinner = (versionId: string) => {
    setSelectedWinner(versionId);
  };

  const presetQuestions = [
    '你好，请介绍一下你自己',
    '我在工作中遇到一个难题，能给我一些建议吗？',
    '如何用更专业的方式表达这个观点？',
    '这个方案有什么潜在风险？',
    '请帮我优化这段文字',
  ];

  return (
    <div className="bg-white rounded-xl shadow-lg p-6 space-y-6">
      {/* 版本信息 */}
      <div>
        <h3 className="text-lg font-semibold text-slate-900 mb-3">
          A/B 测试版本 ({versions.length}个)
        </h3>
        <div className="grid grid-cols-1 md:grid-cols-2 gap-3">
          {versions.map((version) => (
            <div
              key={version.versionId}
              className={`p-3 rounded-lg border-2 transition-all ${
                selectedWinner === version.versionId
                  ? 'border-green-500 bg-green-50'
                  : 'border-slate-200 bg-slate-50'
              }`}
            >
              <div className="flex items-center justify-between mb-2">
                <span className="font-medium text-slate-900">
                  {version.versionName}
                </span>
                {selectedWinner === version.versionId && (
                  <span className="text-xs px-2 py-1 bg-green-500 text-white rounded-full">
                    优胜者
                  </span>
                )}
              </div>
              <p className="text-xs text-slate-600 line-clamp-2">
                {version.systemPrompt}
              </p>
            </div>
          ))}
        </div>
      </div>

      {/* 测试问题输入 */}
      <div>
        <h3 className="text-lg font-semibold text-slate-900 mb-3">测试问题</h3>
        
        {/* 预设问题 */}
        <div className="mb-3">
          <p className="text-sm text-slate-600 mb-2">快速选择：</p>
          <div className="flex flex-wrap gap-2">
            {presetQuestions.map((preset, idx) => (
              <button
                key={idx}
                onClick={() => setQuestion(preset)}
                className="text-xs px-3 py-1.5 bg-slate-100 hover:bg-slate-200 text-slate-700 rounded-full transition-colors"
              >
                {preset.length > 15 ? preset.substring(0, 15) + '...' : preset}
              </button>
            ))}
          </div>
        </div>

        {/* 自定义问题 */}
        <textarea
          value={question}
          onChange={(e) => setQuestion(e.target.value)}
          placeholder="输入自定义测试问题..."
          className="w-full px-4 py-2 border border-slate-300 rounded-lg focus:ring-2 focus:ring-primary focus:border-transparent resize-none"
          rows={3}
        />

        <button
          onClick={handleRunTest}
          disabled={isTesting || !question.trim() || versions.length < 2}
          className="mt-3 w-full py-2.5 bg-primary text-white rounded-lg hover:bg-primary-dark disabled:opacity-50 disabled:cursor-not-allowed transition-colors font-medium"
        >
          {isTesting ? '测试中...' : '开始并排测试'}
        </button>
      </div>

      {/* 测试结果对比 */}
      {results.length > 0 && (
        <div>
          <h3 className="text-lg font-semibold text-slate-900 mb-3">
            测试结果对比
          </h3>
          <div className="space-y-4">
            {results.map((result) => (
              <div
                key={result.versionId}
                className={`border rounded-lg p-4 transition-all ${
                  selectedWinner === result.versionId
                    ? 'border-green-500 bg-green-50'
                    : 'border-slate-200 bg-white'
                }`}
              >
                <div className="flex items-start justify-between mb-3">
                  <div>
                    <h4 className="font-semibold text-slate-900">
                      {result.versionName}
                    </h4>
                    <div className="flex items-center gap-3 mt-1 text-xs text-slate-600">
                      <span>回复时间：{result.responseTime.toFixed(2)}s</span>
                      <span>评分：{result.score.toFixed(1)}</span>
                      <span>{result.rating}星</span>
                    </div>
                  </div>
                  <button
                    onClick={() => handleSelectWinner(result.versionId)}
                    className={`px-3 py-1.5 text-xs rounded-lg transition-colors ${
                      selectedWinner === result.versionId
                        ? 'bg-green-500 text-white'
                        : 'bg-slate-100 text-slate-700 hover:bg-slate-200'
                    }`}
                  >
                    {selectedWinner === result.versionId ? '已选择' : '选择此版本'}
                  </button>
                </div>

                <div className="bg-slate-50 rounded-lg p-3 mb-3">
                  <p className="text-sm text-slate-800">{result.response}</p>
                </div>

                <div className="flex items-center justify-between">
                  <span className="text-xs text-slate-600">{result.feedback}</span>
                  <div className="flex gap-1">
                    {[1, 2, 3, 4, 5].map((star) => (
                      <span
                        key={star}
                        className={star <= result.rating ? 'text-yellow-400' : 'text-slate-300'}
                      >
                        ★
                      </span>
                    ))}
                  </div>
                </div>
              </div>
            ))}
          </div>

          {/* 操作按钮 */}
          <div className="flex gap-3 mt-4">
            <button
              onClick={handleSaveTest}
              className="flex-1 py-2.5 bg-primary text-white rounded-lg hover:bg-primary-dark transition-colors font-medium"
            >
              保存测试结果
            </button>
            <button
              onClick={() => {
                setResults([]);
                setSelectedWinner(null);
              }}
              className="flex-1 py-2.5 border border-slate-300 text-slate-700 rounded-lg hover:bg-slate-50 transition-colors font-medium"
            >
              重新测试
            </button>
          </div>
        </div>
      )}

      {/* 测试历史 */}
      {testHistory.length > 0 && (
        <div>
          <h3 className="text-lg font-semibold text-slate-900 mb-3">
            测试历史 ({testHistory.length})
          </h3>
          <div className="max-h-64 overflow-y-auto space-y-2">
            {testHistory.slice(-5).reverse().map((history, idx) => (
              <div
                key={idx}
                className="p-3 bg-slate-50 rounded-lg border border-slate-200"
              >
                <p className="text-sm text-slate-800 line-clamp-1">
                  {history.question}
                </p>
                <div className="flex items-center justify-between mt-2 text-xs text-slate-600">
                  <span>{history.results.length}个版本对比</span>
                  <span>
                    {history.timestamp.toLocaleTimeString('zh-CN', {
                      hour: '2-digit',
                      minute: '2-digit',
                    })}
                  </span>
                </div>
              </div>
            ))}
          </div>
        </div>
      )}
    </div>
  );
};

export default TestDialog;
