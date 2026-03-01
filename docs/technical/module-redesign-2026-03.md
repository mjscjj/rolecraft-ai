# RoleCraft 模块重设计（2026-03）

## 目标

围绕 4 个核心模块重构产品信息架构，形成可快速落地的 MVP：

1. 对话
2. 我的公司
3. 角色市场
4. 工作区（由“工作”升级）

---

## 1. 模块定义

### 对话（同步协作入口）

- 用途：即时问答、方案讨论、角色协同输出。
- 能力：
  - `chat` 普通对话模式
  - `agent` 深度执行模式（默认可走联网搜索能力）
  - 历史会话与消息元数据（来源、思考过程）
- 结果：产生可追溯的对话结论，可沉淀到工作区/公司成果。

### 我的公司（总结果交付区）

- 用途：统一查看公司级角色、工作区任务、成果交付。
- 能力：
  - 公司角色管理（安装、编辑、自定义）
  - 汇总指标（角色数、工作区数、成果数、知识文档数）
  - 最近成果清单（来自工作区任务执行结果）
- 结果：形成可管理、可审阅的组织级交付视图。

### 角色市场（能力供给中心）

- 用途：发现并安装角色模板。
- 能力：
  - 模板检索与预览
  - 安装到个人或公司
  - 安装后立即进入对话或绑定到工作区任务
- 结果：实现角色能力复用与快速部署。

### 工作区（异步执行中心）

- 用途：管理“什么时候做什么事”，并异步沉淀结果。
- 能力：
  - 任务类型：`general` / `report` / `analyze`
  - 调度类型：`manual` / `once` / `daily` / `interval_hours`
  - 调度参数：如 `09:00`、`每 4 小时`、`一次性时间点`
  - 绑定角色、输入来源、汇报规则
  - 立即执行（MVP 模拟异步执行并写回结果摘要）
- 结果：把 AI 工作从“即时对话”升级到“可计划、可追踪、可复盘”的异步流程。

---

## 2. 核心实体（MVP）

### Work（工作区任务）

新增/关键字段：

- `type`: `general|report|analyze`
- `triggerType`: `manual|once|daily|interval_hours`
- `triggerValue`: 调度值（如 `09:00`、`4`、RFC3339 时间）
- `timezone`: 时区
- `nextRunAt`, `lastRunAt`: 调度执行时间
- `asyncStatus`: `idle|scheduled|running|completed|failed`
- `inputSource`: 输入源定义
- `reportRule`: 汇报规则定义
- `resultSummary`: 最近执行结果摘要
- `config`: 扩展配置 JSON

### Company（公司）

聚合输出增强：

- `stats.workspaceCount`
- `stats.outcomeCount`
- `recentOutcomes[]`（最近交付成果）

---

## 3. API 设计（MVP）

### 工作区

- `GET /api/v1/workspaces`
- `POST /api/v1/workspaces`
- `PUT /api/v1/workspaces/:id`
- `DELETE /api/v1/workspaces/:id`
- `POST /api/v1/workspaces/:id/run`

兼容保留旧路径：

- `/api/v1/works`（等价映射）

### 公司详情增强

- `GET /api/v1/companies/:id`
  - 返回 `stats.workspaceCount / stats.outcomeCount / recentOutcomes`

---

## 4. 前端路由与导航

- 新主路由：`/workspaces`
- 兼容路由：`/works`（同页面）
- 左侧导航命名统一为“工作区”

---

## 5. 交互闭环

1. 从角色市场安装角色到个人或公司
2. 在工作区创建异步任务并绑定角色
3. 任务定时/间隔执行，产出结果摘要
4. 在我的公司查看集中成果交付
5. 在对话中继续追问和细化

---

## 6. 后续迭代建议

1. 引入真实任务调度器（cron/queue/worker）替代当前 MVP 执行模拟
2. 工作区结果支持结构化产物（文档、图表、JSON 报告）
3. 公司交付区增加“项目维度”与成果审批流
4. 工作区支持依赖编排（任务 A 成功后触发任务 B）
