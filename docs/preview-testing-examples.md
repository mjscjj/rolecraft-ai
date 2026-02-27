# RoleCraft AI 预览和测试功能 - 使用示例

## 示例 1: 基础预览功能

```tsx
import React from 'react';
import { RolePreview } from './components/RolePreview';

// 最简单的使用方式
function App() {
  return (
    <div className="app">
      <RolePreview
        role={{
          id: '1',
          name: '智能助理',
          description: '全能型办公助手',
          category: '通用',
          avatar: '智',
          systemPrompt: '你是一位智能助理，擅长帮助用户处理各种办公任务...',
        }}
      />
    </div>
  );
}
```

## 示例 2: 带测试功能的预览

```tsx
import React from 'react';
import { RolePreview } from './components/RolePreview';
import testApi from './api/test';

function RoleCreator() {
  const [systemPrompt, setSystemPrompt] = React.useState('');

  const handleTestChat = async (message: string) => {
    try {
      const response = await testApi.sendMessage({
        content: message,
        systemPrompt: systemPrompt,
        roleName: '营销专家',
      });
      return response.content;
    } catch (error) {
      console.error('测试失败:', error);
      return '测试失败，请稍后重试';
    }
  };

  return (
    <div className="role-creator">
      <div className="editor">
        <textarea
          value={systemPrompt}
          onChange={(e) => setSystemPrompt(e.target.value)}
          placeholder="输入系统提示词..."
        />
      </div>
      
      <div className="preview">
        <RolePreview
          systemPrompt={systemPrompt}
          modelName="营销专家"
          category="营销"
          onTestChat={handleTestChat}
        />
      </div>
    </div>
  );
}
```

## 示例 3: A/B 测试完整流程

```tsx
import React, { useState } from 'react';
import { TestDialog } from './components/TestDialog';
import testApi from './api/test';

function ABTestPage() {
  const [versions, setVersions] = useState([
    {
      versionId: 'v1',
      versionName: '版本 A - 正式版',
      systemPrompt: '你是一位专业的营销专家，提供严谨、专业的建议。',
    },
    {
      versionId: 'v2',
      versionName: '版本 B - 友好版',
      systemPrompt: '你是一位亲切的营销顾问，用轻松友好的方式提供建议。',
    },
  ]);

  const handleTest = async (versions, question) => {
    const result = await testApi.runABTest({
      versions,
      question,
    });
    return result.results;
  };

  const handleSaveTest = async (results, question) => {
    // 保存测试结果到数据库
    for (const result of results) {
      await testApi.saveTestResult({
        roleId: 'current-role-id',
        roleName: '营销专家',
        testType: 'ab',
        question: question,
        response: result.response,
        rating: result.rating,
        feedback: result.feedback,
      });
    }
    alert('测试结果已保存！');
  };

  return (
    <div className="ab-test-page">
      <h1>A/B 测试</h1>
      <TestDialog
        versions={versions}
        onTest={handleTest}
        onSaveTest={handleSaveTest}
      />
    </div>
  );
}
```

## 示例 4: 测试报告页面

```tsx
import React, { useEffect, useState } from 'react';
import testApi, { TestReport } from './api/test';

function TestReportDashboard({ roleId }) {
  const [report, setReport] = useState(null);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    async function loadReport() {
      try {
        const data = await testApi.getTestReport(roleId);
        setReport(data);
      } catch (error) {
        console.error('加载报告失败:', error);
      } finally {
        setLoading(false);
      }
    }
    loadReport();
  }, [roleId]);

  if (loading) return <div>加载中...</div>;
  if (!report) return <div>加载失败</div>;

  return (
    <div className="test-report">
      <h1>测试报告 - {report.roleName}</h1>
      
      {/* 统计卡片 */}
      <div className="stats-grid">
        <div className="stat-card">
          <h3>总测试次数</h3>
          <p className="stat-value">{report.totalTests}</p>
        </div>
        <div className="stat-card">
          <h3>平均评分</h3>
          <p className="stat-value">{report.averageRating.toFixed(1)}⭐</p>
        </div>
        <div className="stat-card">
          <h3>通过率</h3>
          <p className="stat-value">{report.passRate.toFixed(1)}%</p>
        </div>
      </div>

      {/* 评分分布 */}
      <div className="rating-distribution">
        <h3>评分分布</h3>
        {Object.entries(report.testsByRating).map(([rating, count]) => (
          <div key={rating} className="rating-bar">
            <span>{rating}星</span>
            <div className="bar">
              <div 
                className="fill" 
                style={{ width: `${(count / report.totalTests) * 100}%` }}
              />
            </div>
            <span>{count}次</span>
          </div>
        ))}
      </div>

      {/* 改进建议 */}
      <div className="suggestions">
        <h3>改进建议</h3>
        <ul>
          {report.suggestions.map((suggestion, idx) => (
            <li key={idx}>{suggestion}</li>
          ))}
        </ul>
      </div>

      {/* 导出按钮 */}
      <button onClick={() => testApi.exportTestReport(roleId, 'pdf')}>
        导出 PDF 报告
      </button>
    </div>
  );
}
```

## 示例 5: 在角色编辑器中集成预览

```tsx
import React, { useState } from 'react';
import { RolePreview } from './components/RolePreview';
import testApi from './api/test';

function EnhancedRoleEditor() {
  const [formData, setFormData] = useState({
    name: '',
    description: '',
    category: '通用',
    systemPrompt: '',
    welcomeMessage: '',
  });

  const handleTestChat = async (message: string) => {
    const response = await testApi.sendMessage({
      content: message,
      systemPrompt: formData.systemPrompt,
      roleName: formData.name || '未命名角色',
    });
    return response.content;
  };

  return (
    <div className="enhanced-editor">
      <div className="editor-panel">
        <h2>角色配置</h2>
        
        <div className="form-group">
          <label>角色名称</label>
          <input
            type="text"
            value={formData.name}
            onChange={(e) => setFormData({ ...formData, name: e.target.value })}
            placeholder="输入角色名称"
          />
        </div>

        <div className="form-group">
          <label>系统提示词</label>
          <textarea
            value={formData.systemPrompt}
            onChange={(e) => setFormData({ ...formData, systemPrompt: e.target.value })}
            placeholder="定义角色的身份和行为..."
            rows={8}
          />
        </div>

        {/* 更多表单字段... */}
      </div>

      <div className="preview-panel">
        <h2>实时预览</h2>
        <RolePreview
          role={formData}
          onTestChat={handleTestChat}
        />
      </div>
    </div>
  );
}
```

## 示例 6: 批量测试多个问题

```tsx
import React, { useState } from 'react';
import testApi from './api/test';

function BatchTestTool({ roleId, systemPrompt }) {
  const [questions] = useState([
    '你好，请介绍一下你自己',
    '我在工作中遇到一个难题...',
    '如何用更专业的方式表达这个观点？',
    '这个方案有什么潜在风险？',
    '请帮我优化这段文字',
  ]);
  const [results, setResults] = useState([]);
  const [testing, setTesting] = useState(false);

  const runBatchTest = async () => {
    setTesting(true);
    const testResults = [];

    for (const question of questions) {
      try {
        const response = await testApi.sendMessage({
          content: question,
          systemPrompt: systemPrompt,
          roleName: '测试角色',
        });

        testResults.push({
          question,
          response: response.content,
          responseTime: response.responseTime,
        });
      } catch (error) {
        testResults.push({
          question,
          response: '测试失败',
          responseTime: 0,
        });
      }
    }

    setResults(testResults);
    setTesting(false);
  };

  return (
    <div className="batch-test">
      <h3>批量测试 ({questions.length}个问题)</h3>
      
      <button 
        onClick={runBatchTest} 
        disabled={testing}
      >
        {testing ? '测试中...' : '开始批量测试'}
      </button>

      <div className="results">
        {results.map((result, idx) => (
          <div key={idx} className="result-item">
            <div className="question">
              <strong>Q:</strong> {result.question}
            </div>
            <div className="answer">
              <strong>A:</strong> {result.response}
            </div>
            <div className="meta">
              响应时间：{result.responseTime.toFixed(2)}s
            </div>
          </div>
        ))}
      </div>
    </div>
  );
}
```

## 示例 7: 测试历史对比

```tsx
import React, { useEffect, useState } from 'react';
import testApi, { TestHistory } from './api/test';

function TestHistoryComparison({ roleId }) {
  const [history, setHistory] = useState([]);

  useEffect(() => {
    async function loadHistory() {
      const data = await testApi.getTestHistory(roleId);
      setHistory(data.history);
    }
    loadHistory();
  }, [roleId]);

  // 按评分分组
  const groupedByRating = history.reduce((acc, item) => {
    const rating = item.rating;
    if (!acc[rating]) acc[rating] = [];
    acc[rating].push(item);
    return acc;
  }, {});

  return (
    <div className="history-comparison">
      <h3>测试历史对比</h3>
      
      {Object.entries(groupedByRating).map(([rating, items]) => (
        <div key={rating} className="rating-group">
          <h4>{rating}星测试 ({items.length}次)</h4>
          <div className="test-items">
            {items.slice(0, 3).map((item) => (
              <div key={item.testId} className="test-item">
                <p className="question">{item.question}</p>
                <p className="response">{item.response}</p>
                <p className="feedback">{item.feedback}</p>
                <span className="date">
                  {new Date(item.createdAt).toLocaleDateString()}
                </span>
              </div>
            ))}
          </div>
        </div>
      ))}
    </div>
  );
}
```

## 示例 8: 实时能力雷达图更新

```tsx
import React, { useState, useEffect } from 'react';

function LiveCapabilityRadar({ systemPrompt }) {
  const [capabilities, setCapabilities] = useState([]);

  useEffect(() => {
    // 分析能力
    const analyze = () => {
      const prompt = systemPrompt.toLowerCase();
      
      const calcScore = (keywords) => {
        let score = 50;
        keywords.forEach(keyword => {
          if (prompt.includes(keyword)) score += 10;
        });
        return Math.min(100, Math.max(0, score));
      };

      const caps = [
        { name: '创造性', value: calcScore(['创意', '创新', '设计']), color: '#FF6B6B' },
        { name: '逻辑性', value: calcScore(['逻辑', '分析', '推理']), color: '#4ECDC4' },
        { name: '专业性', value: calcScore(['专业', '专家', '资深']), color: '#45B7D1' },
        { name: '共情力', value: calcScore(['理解', '关心', '支持']), color: '#FFA07A' },
        { name: '效率', value: calcScore(['快速', '高效', '简洁']), color: '#98D8C8' },
      ];

      setCapabilities(caps);
    };

    analyze();
  }, [systemPrompt]);

  return (
    <div className="capability-radar">
      <h4>能力雷达图</h4>
      <div className="capabilities">
        {capabilities.map((cap) => (
          <div key={cap.name} className="capability">
            <div className="capability-header">
              <span>{cap.name}</span>
              <span>{cap.value}</span>
            </div>
            <div className="progress-bar">
              <div 
                className="progress" 
                style={{ 
                  width: `${cap.value}%`,
                  backgroundColor: cap.color 
                }}
              />
            </div>
          </div>
        ))}
      </div>
    </div>
  );
}
```

## 使用场景

### 场景 1: 角色创建前的效果验证

1. 用户在角色编辑器中输入基本信息
2. 配置系统提示词
3. 实时预览自动展示能力雷达图
4. 使用预设问题或自定义问题进行测试
5. 根据测试反馈调整提示词
6. 满意后保存角色

### 场景 2: 提示词优化迭代

1. 选择已有的角色
2. 创建提示词的多个版本
3. 运行 A/B 测试对比效果
4. 查看测试报告和评分
5. 选择最优版本
6. 保存优化结果

### 场景 3: 团队协作评审

1. 创建角色后运行批量测试
2. 导出测试报告
3. 分享给团队成员
4. 收集反馈意见
5. 根据反馈进一步优化

## 注意事项

1. **测试环境** - 确保 API 端点配置正确
2. **网络延迟** - 测试响应时间可能受网络影响
3. **Mock 数据** - 当前使用模拟数据，实际需接入 AI API
4. **权限控制** - 测试功能需要用户认证
5. **数据保存** - 测试结果会保存到数据库

## 下一步

- 接入真实的大语言模型 API
- 实现智能评估算法
- 添加更多测试模板
- 支持团队协作功能
- 优化可视化图表
