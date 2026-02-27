# RoleCraft AI 零提示词体验 - 预览和测试功能文档

## 功能概述

预览和测试功能允许用户在创建 AI 角色之前实时看到效果，增强用户信心。主要包含以下功能：

1. **实时预览** - 展示 AI 角色形象、能力雷达图、预计效果
2. **测试对话框** - 内置测试聊天界面，支持满意度评分
3. **A/B 测试** - 创建多个版本并排对比
4. **测试报告** - 历史记录、效果对比、改进建议

## 文件结构

```
backend/
├── internal/
│   └── api/
│       └── handler/
│           └── test.go              # 测试 API 处理器

frontend/
├── src/
│   ├── components/
│   │   ├── RolePreview.tsx          # 角色预览组件
│   │   └── TestDialog.tsx           # A/B 测试对话框组件
│   ├── pages/
│   │   └── TestReport.tsx           # 测试报告页面
│   ├── api/
│   │   └── test.ts                  # 测试 API 客户端
│   └── types/
│       └── index.ts                 # 类型定义（已更新）
```

## API 接口

### 1. 发送测试消息

**POST** `/api/v1/test/message`

**请求体：**
```json
{
  "content": "测试问题内容",
  "systemPrompt": "系统提示词",
  "modelConfig": {},
  "roleName": "角色名称"
}
```

**响应：**
```json
{
  "code": 200,
  "message": "success",
  "data": {
    "content": "AI 回复内容",
    "responseTime": 0.5,
    "tokens": 100,
    "model": "mock-model",
    "metadata": {}
  }
}
```

### 2. 运行 A/B 测试

**POST** `/api/v1/test/ab`

**请求体：**
```json
{
  "versions": [
    {
      "versionId": "v1",
      "versionName": "版本 A",
      "systemPrompt": "提示词版本 A"
    },
    {
      "versionId": "v2",
      "versionName": "版本 B",
      "systemPrompt": "提示词版本 B"
    }
  ],
  "question": "测试问题"
}
```

**响应：**
```json
{
  "code": 200,
  "message": "success",
  "data": {
    "testId": "uuid",
    "question": "测试问题",
    "results": [
      {
        "versionId": "v1",
        "versionName": "版本 A",
        "response": "回复内容",
        "responseTime": 0.5,
        "score": 85.5,
        "rating": 4,
        "feedback": "回复质量良好"
      }
    ],
    "winnerId": "v1",
    "createdAt": "2026-02-27T08:00:00Z"
  }
}
```

### 3. 获取测试历史

**GET** `/api/v1/test/history?roleId={roleId}`

**响应：**
```json
{
  "code": 200,
  "message": "success",
  "data": {
    "history": [
      {
        "testId": "uuid",
        "roleId": "role-uuid",
        "roleName": "角色名称",
        "testType": "single",
        "question": "测试问题",
        "response": "回复内容",
        "rating": 4,
        "feedback": "反馈",
        "createdAt": "2026-02-27T08:00:00Z"
      }
    ],
    "total": 10
  }
}
```

### 4. 获取测试报告

**GET** `/api/v1/test/report?roleId={roleId}`

**响应：**
```json
{
  "code": 200,
  "message": "success",
  "data": {
    "roleId": "uuid",
    "roleName": "角色名称",
    "totalTests": 50,
    "averageRating": 4.2,
    "passRate": 86.5,
    "testsByRating": {
      "5": 15,
      "4": 25,
      "3": 8,
      "2": 2,
      "1": 0
    },
    "improvementTrend": [
      {
        "date": "2026-02-27",
        "avgRating": 4.2,
        "testCount": 10
      }
    ],
    "suggestions": [
      "系统提示词可以更具体一些",
      "建议增加示例对话"
    ],
    "exportUrl": "/api/v1/test/export/uuid"
  }
}
```

### 5. 导出测试报告

**GET** `/api/v1/test/export/{roleId}?format=pdf`

**响应：** 文件下载（PDF/Markdown/JSON）

### 6. 评分测试回复

**POST** `/api/v1/test/rate`

**请求体：**
```json
{
  "testId": "uuid",
  "rating": 5,
  "feedback": "非常好的回复"
}
```

## 组件使用

### RolePreview 组件

```tsx
import { RolePreview } from './components/RolePreview';

<RolePreview
  role={{
    id: '1',
    name: '营销专家',
    description: '专业的营销策划助手',
    category: '营销',
    systemPrompt: '你是一位资深的营销专家...',
  }}
  onTestChat={async (message) => {
    // 调用测试 API
    const response = await testApi.sendMessage({
      content: message,
      systemPrompt: '你是一位资深的营销专家...',
      roleName: '营销专家',
    });
    return response.content;
  }}
/>
```

### TestDialog 组件

```tsx
import { TestDialog } from './components/TestDialog';

<TestDialog
  versions={[
    {
      versionId: 'v1',
      versionName: '版本 A',
      systemPrompt: '提示词版本 A',
    },
    {
      versionId: 'v2',
      versionName: '版本 B',
      systemPrompt: '提示词版本 B',
    },
  ]}
  onTest={async (versions, question) => {
    // 调用 A/B 测试 API
    const result = await testApi.runABTest({
      versions,
      question,
    });
    return result.results;
  }}
  onSaveTest={(results, question) => {
    // 保存测试结果
    console.log('保存测试结果:', results);
  }}
/>
```

## 能力雷达图

系统会自动分析系统提示词，评估以下六个维度的能力：

1. **创造性** - 创意、创新、设计相关能力
2. **逻辑性** - 分析、推理、结构化思维能力
3. **专业性** - 专业知识、经验、资质
4. **共情力** - 理解、关心、支持能力
5. **效率** - 快速、高效、简洁程度
6. **适应性** - 灵活、多场景适应能力

评分基于提示词中的关键词匹配度，范围 0-100。

## 测试报告

测试报告包含以下信息：

### 概览统计
- 总测试次数
- 平均评分
- 通过率

### 评分分布
可视化展示 1-5 星的分布情况

### 改进趋势
显示最近 7 天的评分变化趋势

### 改进建议
基于测试结果自动生成优化建议：
- 提示词优化建议
- 技能配置建议
- 模型参数调整建议

## 使用流程

### 1. 创建角色时预览

1. 进入角色创建页面
2. 填写基础信息
3. 配置系统提示词
4. 实时预览自动展示
5. 使用测试对话框验证效果
6. 调整配置直到满意

### 2. A/B 测试

1. 创建多个提示词版本
2. 输入相同的测试问题
3. 并排查看不同版本的回复
4. 对比回复质量、响应时间
5. 选择优胜版本
6. 保存测试结果

### 3. 查看测试报告

1. 进入测试报告页面
2. 查看整体统计数据
3. 分析评分分布和趋势
4. 阅读改进建议
5. 导出报告分享

## 最佳实践

### 提示词优化

1. **明确角色定位** - 清晰描述身份、专业领域
2. **提供示例对话** - 帮助 AI 理解期望的回复风格
3. **设定行为准则** - 明确什么该做，什么不该做
4. **逐步迭代** - 基于测试反馈持续优化

### A/B 测试策略

1. **控制变量** - 每次只测试一个变量的变化
2. **多样化问题** - 使用不同类型的问题测试
3. **足够样本** - 至少测试 5-10 个问题
4. **记录反馈** - 详细记录每次测试的反馈

### 评分标准

- **5 星** - 超出预期，完美回复
- **4 星** - 符合预期，质量良好
- **3 星** - 基本满意，有小问题
- **2 星** - 不太满意，需要改进
- **1 星** - 完全不符合预期

## 技术实现

### 后端

- **语言**: Go (Gin 框架)
- **数据库**: SQLite (GORM ORM)
- **测试模拟**: Mock AI 回复生成
- **评估算法**: 基于关键词和回复质量评分

### 前端

- **框架**: React + TypeScript
- **路由**: React Router
- **样式**: Tailwind CSS
- **图表**: 自定义 SVG 雷达图

## 未来改进

1. **真实 AI 集成** - 接入实际的大语言模型 API
2. **智能评估** - 使用 AI 自动评估回复质量
3. **批量测试** - 一次性运行多个测试用例
4. **测试模板** - 预定义的测试用例库
5. **团队协作** - 多人协作测试和评审
6. **自动化建议** - 基于测试结果自动生成优化建议

## 故障排除

### 测试失败

- 检查网络连接
- 验证 API 端点配置
- 查看浏览器控制台错误日志

### 评分不保存

- 确认用户已登录
- 检查数据库连接
- 验证 API 权限

### 预览不更新

- 清除浏览器缓存
- 检查组件状态更新
- 确认 props 正确传递

## 联系方式

如有问题或建议，请联系：
- 技术支持：support@rolecraft.ai
- 文档：https://docs.rolecraft.ai
