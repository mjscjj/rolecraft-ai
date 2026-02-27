# RoleCraft AI 数据分析平台

## 📊 功能概述

数据分析平台为 RoleCraft AI 提供全面的数据驱动洞察和决策支持，包含以下核心模块：

### 1. 用户行为分析
- **DAU/WAU/MAU 统计**: 日活/周活/月活用户数统计
- **功能使用率分析**: 各功能模块的使用情况
- **用户留存率**: 1 日/7 日/30 日留存分析
- **用户流失预警**: 识别有流失风险的用户

### 2. 对话质量评估
- **对话满意度评分**: 用户评分统计和分析
- **回复质量分析**: 响应时间、Token 使用、回复长度
- **常见问题统计**: 用户高频问题 TOP 榜
- **敏感词检测**: 敏感词拦截和统计

### 3. 成本统计
- **Token 使用统计**: 总 Token 消耗和分类统计
- **API 调用成本**: 实时成本计算
- **按角色/用户分类**: 成本分摊和对比
- **成本趋势预测**: 基于历史数据的成本预测

### 4. 效果报告
- **自动生成周报/月报**: 定期生成数据报告
- **关键指标趋势图**: 可视化趋势分析
- **对比分析**: 环比/同比数据对比
- **报告导出**: 支持 PDF 格式导出

### 5. 可视化 Dashboard
- **核心指标概览**: 关键数据一目了然
- **实时数据更新**: 数据实时刷新
- **自定义图表**: 灵活的图表展示
- **数据筛选和钻取**: 深入分析数据细节

---

## 🚀 API 接口

### Dashboard 核心指标
```http
GET /api/v1/analytics/dashboard
```

**响应示例:**
```json
{
  "code": 200,
  "message": "success",
  "data": {
    "totalUsers": 892,
    "activeUsers": 428,
    "totalRoles": 45,
    "totalSessions": 1256,
    "totalMessages": 8945,
    "totalDocuments": 320,
    "totalCost": 49.18,
    "averageRating": 4.5,
    "userActivity": {
      "dau": 156,
      "wau": 428,
      "mau": 892
    },
    "costStats": {
      "totalTokens": 2458920,
      "totalCost": 49.18,
      "averageCostPerDay": 7.03
    },
    "qualityStats": {
      "averageRating": 4.5,
      "satisfactionRate": 92.5
    },
    "topRoles": [...],
    "recentTrends": [...]
  }
}
```

### 用户活跃度
```http
GET /api/v1/analytics/user-activity
```

### 功能使用率
```http
GET /api/v1/analytics/feature-usage
```

### 留存率
```http
GET /api/v1/analytics/retention
```

### 流失风险用户
```http
GET /api/v1/analytics/churn-risk
```

### 对话质量
```http
GET /api/v1/analytics/conversation-quality
```

### 回复质量
```http
GET /api/v1/analytics/reply-quality
```

### 常见问题
```http
GET /api/v1/analytics/faq
```

### 敏感词检测
```http
GET /api/v1/analytics/sensitive-words
```

### 成本统计
```http
GET /api/v1/analytics/cost
```

### 按角色分类成本
```http
GET /api/v1/analytics/cost/by-role
```

### 按用户分类成本
```http
GET /api/v1/analytics/cost/by-user
```

### 成本趋势
```http
GET /api/v1/analytics/cost/trend?days=30
```

### 成本预测
```http
GET /api/v1/analytics/cost/prediction?period=month
```

### 生成报告
```http
GET /api/v1/analytics/report?type=weekly
```
参数: `type` - `weekly` (周报) 或 `monthly` (月报)

### 导出报告
```http
GET /api/v1/analytics/report/export?type=weekly
```
返回: PDF 文件 (开发中) 或 JSON 格式

---

## 🖥️ 前端页面

### 访问地址
```
http://localhost:5173/analytics
```

### 页面结构

1. **顶部导航栏**
   - 报告类型切换 (周报/月报)
   - 导出报告按钮

2. **核心指标卡片** (4 个)
   - 活跃用户 (DAU/WAU/MAU)
   - 对话次数
   - 总成本
   - 平均评分

3. **用户活跃度分析**
   - 趋势折线图
   - DAU/WAU/MAU 统计

4. **成本分析**
   - 成本趋势图 (近 7 天)
   - Top 5 角色使用排行

5. **对话质量评估**
   - 平均评分
   - 满意度
   - 高质量对话占比

6. **报告摘要**
   - 关键指标
   - 环比/同比对比
   - 优化建议

---

## 📁 文件结构

```
rolecraft-ai/
├── backend/
│   ├── cmd/server/
│   │   └── main.go                    # 主入口 (已添加 analytics 路由)
│   ├── internal/api/handler/
│   │   └── analytics.go               # 数据分析处理器 (新建)
│   └── ...
├── frontend/
│   ├── src/
│   │   ├── pages/
│   │   │   └── Analytics.tsx          # 数据分析页面 (新建)
│   │   ├── api/
│   │   │   └── analytics.ts           # API 客户端 (新建)
│   │   ├── components/
│   │   │   └── Layout.tsx             # (已更新导航)
│   │   └── App.tsx                    # (已添加路由)
│   └── ...
└── docs/
    └── analytics-sample-report.md     # 示例报告 (新建)
```

---

## 🔧 部署说明

### 后端部署

1. **编译后端**
```bash
cd backend
go build -o ../bin/server ./cmd/server/main.go
```

2. **配置环境变量**
```bash
export DATABASE_URL="your-database-url"
export ANYTHINGLLM_URL="http://your-anythingllm-url"
export ANYTHINGLLM_KEY="your-api-key"
```

3. **运行服务**
```bash
./bin/server
```

### 前端部署

1. **安装依赖**
```bash
cd frontend
npm install
```

2. **开发模式**
```bash
npm run dev
```

3. **生产构建**
```bash
npm run build
```

---

## 📊 数据计算说明

### 成本计算
- **Token 单价**: ¥0.00002/Token (可根据实际调整)
- **总成本**: 总 Token 数 × 单价
- **日均成本**: 总成本 / 运营天数

### 活跃度统计
- **DAU**: 当日有对话会话的用户数
- **WAU**: 近 7 天有对话会话的用户数
- **MAU**: 近 30 天有对话会话的用户数

### 留存率计算
- **N 日留存**: 在第 N 天仍有活跃的用户比例

### 流失风险
- **高风险**: 90 天未活跃
- **中风险**: 60-90 天未活跃
- **低风险**: 30-60 天未活跃

---

## 🎯 使用场景

### 产品经理
- 监控核心指标变化
- 识别用户行为模式
- 制定产品优化策略

### 运营团队
- 追踪用户活跃度
- 识别流失风险用户
- 制定召回策略

### 技术团队
- 监控 API 使用成本
- 优化 Token 使用效率
- 性能瓶颈分析

### 管理层
- 查看周报/月报
- 了解业务发展趋势
- 数据驱动决策

---

## 📝 示例报告

查看完整示例报告：[analytics-sample-report.md](./analytics-sample-report.md)

---

## 🔮 后续优化

### 短期 (1-2 周)
- [ ] 实现 PDF 报告导出功能
- [ ] 添加数据导出 (CSV/Excel)
- [ ] 优化移动端展示
- [ ] 添加数据刷新机制

### 中期 (1 个月)
- [ ] 集成 NLP 分析用户问题
- [ ] 实现自定义 Dashboard
- [ ] 添加告警通知功能
- [ ] 支持数据对比 (多周期)

### 长期 (3 个月)
- [ ] AI 智能洞察和建议
- [ ] 预测性分析模型
- [ ] 实时数据流处理
- [ ] 多维度数据钻取

---

## 📞 技术支持

如有问题或建议，请联系：
- 产品团队：product@rolecraft.ai
- 技术支持：support@rolecraft.ai

---

*最后更新：2026-02-27*
