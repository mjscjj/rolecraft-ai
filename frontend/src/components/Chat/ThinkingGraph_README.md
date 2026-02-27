# ThinkingGraph - 思维导图式 AI 思考路径可视化

## 📖 概述

ThinkingGraph 是一个基于 React Flow 的可交互思维导图组件，用于可视化展示 AI 的思考过程。它支持多种布局算法、丰富的节点类型和完整的交互功能。

## ✨ 特性

### 1. 数据结构
- **ThinkingNode**: 支持 8 种节点类型 (ROOT, ANALYSIS, PLANNING, EXECUTION, DECISION, TOOL, REFLECTION, CONCLUSION)
- **ThinkingEdge**: 灵活的边配置，支持多种样式和动画
- **ThinkingGraph**: 完整的图结构，包含配置和元数据

### 2. 可视化组件
- ✅ 思维导图渲染器 (基于 React Flow)
- ✅ 节点组件 (带类型图标和状态指示器)
- ✅ 连线组件 (贝塞尔曲线)
- ✅ 缩放/平移功能
- ✅ 小地图导航
- ✅ 控制面板

### 3. 布局算法
- ✅ 树状布局 (从上到下)
- ✅ 力导向布局
- ✅ 层次布局
- ✅ 树状布局 (从左到右)

### 4. 交互功能
- ✅ 节点点击展开/折叠
- ✅ 双击切换详细视图
- ✅ 拖拽调整位置
- ✅ 节点选择高亮
- ✅ 边点击交互

### 5. 样式设计
- ✅ 节点类型颜色编码
- ✅ 节点状态样式 (pending, active, completed, skipped, error)
- ✅ 连线样式配置
- ✅ 动画效果 (脉冲、滑入、绘制)
- ✅ 主题系统 (浅色/深色)

## 📦 安装

```bash
cd rolecraft-ai/frontend
npm install @xyflow/react
```

## 🚀 快速开始

### 基础用法

```tsx
import { ThinkingGraph } from './components/Chat/ThinkingGraph';
import { createEmptyGraph, createRootNode, addNode, addEdge } from './utils/thinkingGraph';

// 创建图
let graph = createEmptyGraph('session_1', 'msg_1', 'qwen3.5-plus');

// 添加节点
const rootNode = createRootNode('问题', '如何学习编程？');
graph = addNode(graph, rootNode);

// 渲染
<ThinkingGraph
  graph={graph}
  layout={ThinkingGraphLayout.TREE}
  theme={ThinkingGraphTheme.LIGHT}
  showControls={true}
  showMinimap={true}
/>
```

### 完整示例

```tsx
import React, { useState } from 'react';
import { ThinkingGraph } from './components/Chat/ThinkingGraph';
import {
  createEmptyGraph,
  createRootNode,
  createAnalysisNode,
  createConclusionNode,
  addNode,
  addEdge,
  completeNode,
} from './utils/thinkingGraph';
import { ThinkingGraphLayout, ThinkingNodeStatus } from './types/thinking';

const MyComponent = () => {
  const [graph, setGraph] = useState(() => {
    let g = createEmptyGraph();
    
    // 创建节点
    const root = createRootNode('用户问题', '如何实现排序算法？');
    const analysis = createAnalysisNode('分析', '需要分析复杂度');
    const conclusion = createConclusionNode('结论', '推荐使用快速排序');
    
    // 添加节点
    g = addNode(g, root);
    g = addNode(g, analysis);
    g = addNode(g, conclusion);
    
    // 添加边
    g = addEdge(g, { id: 'e1', source: root.id, target: analysis.id });
    g = addEdge(g, { id: 'e2', source: analysis.id, target: conclusion.id });
    
    // 完成节点
    g = completeNode(g, root.id);
    g = completeNode(g, analysis.id);
    
    return g;
  });

  return (
    <div style={{ height: '500px' }}>
      <ThinkingGraph
        graph={graph}
        layout={ThinkingGraphLayout.TREE}
        theme={ThinkingGraphTheme.LIGHT}
      />
    </div>
  );
};
```

## 📚 API 文档

### ThinkingGraph 组件

```tsx
interface ThinkingGraphProps {
  /** 思考图数据 */
  graph: ThinkingGraph;
  
  /** 布局类型 */
  layout?: ThinkingGraphLayout;
  
  /** 主题 */
  theme?: ThinkingGraphTheme;
  
  /** 是否显示控制面板 */
  showControls?: boolean;
  
  /** 是否显示小地图 */
  showMinimap?: boolean;
  
  /** 是否显示背景 */
  showBackground?: boolean;
  
  /** 节点点击回调 */
  onNodeClick?: (node: ThinkingNode) => void;
  
  /** 节点双击回调 */
  onNodeDoubleClick?: (node: ThinkingNode) => void;
  
  /** 边点击回调 */
  onEdgeClick?: (edge: ThinkingEdge) => void;
  
  /** 类名 */
  className?: string;
  
  /** 样式 */
  style?: React.CSSProperties;
}
```

### 节点类型

```typescript
enum ThinkingNodeType {
  ROOT = 'root',         // 🎯 根节点
  ANALYSIS = 'analysis', // 🔍 分析节点
  PLANNING = 'planning', // 📋 计划节点
  EXECUTION = 'execution', // ⚡ 执行节点
  DECISION = 'decision', // 🤔 决策节点
  TOOL = 'tool',         // 🛠️ 工具节点
  REFLECTION = 'reflection', // 💭 反思节点
  CONCLUSION = 'conclusion', // ✅ 结论节点
}
```

### 节点状态

```typescript
enum ThinkingNodeStatus {
  PENDING = 'pending',     // 待处理
  ACTIVE = 'active',       // 进行中
  COMPLETED = 'completed', // 已完成
  SKIPPED = 'skipped',     // 已跳过
  ERROR = 'error',         // 错误
}
```

### 布局类型

```typescript
enum ThinkingGraphLayout {
  TREE = 'tree',           // 树状布局 (从上到下)
  TREE_LR = 'tree-lr',     // 树状布局 (从左到右)
  FORCE = 'force',         // 力导向布局
  HIERARCHICAL = 'hierarchical', // 层次布局
}
```

### 主题

```typescript
enum ThinkingGraphTheme {
  LIGHT = 'light',         // 浅色主题
  DARK = 'dark',           // 深色主题
  AUTO = 'auto',           // 自动 (跟随系统)
}
```

## 🛠️ 工具函数

### 创建图

```typescript
import { createEmptyGraph } from './utils/thinkingGraph';

const graph = createEmptyGraph(
  'session_id',      // 会话 ID
  'message_id',      // 消息 ID
  'model_name'       // 模型名称
);
```

### 创建节点

```typescript
import {
  createRootNode,
  createAnalysisNode,
  createPlanningNode,
  createExecutionNode,
  createDecisionNode,
  createToolNode,
  createReflectionNode,
  createConclusionNode,
} from './utils/thinkingGraph';

// 根节点
const root = createRootNode('标题', '内容', { confidence: 0.95 });

// 分析节点
const analysis = createAnalysisNode('分析', '详细内容', {
  confidence: 0.92,
  tokenUsage: { input: 100, output: 200 }
});

// 决策节点 (带选项)
const decision = createDecisionNode(
  '选择算法',
  '应该使用哪种排序？',
  ['快速排序', '归并排序', '堆排序'],
  { confidence: 0.88 }
);

// 工具节点
const tool = createToolNode(
  'CodeSearch',           // 工具名称
  '搜索代码',             // 标题
  '搜索最佳实践',         // 内容
  { query: 'sort' },      // 工具参数
  { confidence: 0.90 }    // 元数据
);
```

### 操作图

```typescript
import {
  addNode,
  addEdge,
  updateNode,
  deleteNode,
  completeNode,
  activateNode,
  toggleNodeExpand,
  expandAllNodes,
  collapseAllNodes,
} from './utils/thinkingGraph';

// 添加节点
graph = addNode(graph, node);

// 添加边
graph = addEdge(graph, { 
  id: 'e1', 
  source: node1.id, 
  target: node2.id,
  label: '导致'
});

// 更新节点
graph = updateNode(graph, nodeId, (node) => ({
  ...node,
  content: '新内容',
}));

// 完成节点
graph = completeNode(graph, nodeId, 500); // 500ms 耗时

// 激活节点 (其他自动设为非激活)
graph = activateNode(graph, nodeId);

// 切换展开状态
graph = toggleNodeExpand(graph, nodeId);

// 展开/折叠所有
graph = expandAllNodes(graph);
graph = collapseAllNodes(graph);
```

### 序列化/反序列化

```typescript
import {
  serializeGraph,
  deserializeGraph,
  graphToJson,
  graphFromJson,
} from './utils/thinkingGraph';

// 序列化为对象
const serialized = serializeGraph(graph);

// 序列化为 JSON 字符串
const json = graphToJson(graph);

// 从 JSON 加载
const loadedGraph = graphFromJson(json);
```

### 查询和分析

```typescript
import {
  getGraphStats,
  getNodeChildren,
  getNodeParent,
  getNodePath,
  getGraphDepth,
  validateGraph,
  exportToMermaid,
  exportToDot,
} from './utils/thinkingGraph';

// 统计信息
const stats = getGraphStats(graph);
// { totalNodes: 8, totalEdges: 7, nodesByType: {...}, ... }

// 获取子节点
const children = getNodeChildren(graph, nodeId);

// 获取父节点
const parent = getNodeParent(graph, nodeId);

// 获取从根到该节点的路径
const path = getNodePath(graph, nodeId);

// 获取图的深度
const depth = getGraphDepth(graph);

// 验证图完整性
const { valid, errors } = validateGraph(graph);

// 导出为 Mermaid
const mermaid = exportToMermaid(graph);

// 导出为 DOT (Graphviz)
const dot = exportToDot(graph);
```

## 🎨 样式定制

### CSS 变量

```css
.thinking-graph-container {
  --tg-bg: #ffffff;              /* 背景色 */
  --tg-node-bg: #ffffff;         /* 节点背景 */
  --tg-node-border: #e2e8f0;     /* 节点边框 */
  --tg-text: #1a202c;            /* 文字颜色 */
  --tg-text-muted: #718096;      /* 次要文字 */
  --tg-edge-color: #cbd5e0;      /* 边颜色 */
}
```

### 节点类型颜色

- **ROOT**: 🎯 蓝色 (#3b82f6)
- **ANALYSIS**: 🔍 紫色 (#8b5cf6)
- **PLANNING**: 📋 青色 (#06b6d4)
- **EXECUTION**: ⚡ 绿色 (#10b981)
- **DECISION**: 🤔 橙色 (#f59e0b)
- **TOOL**: 🛠️ 红色 (#ef4444)
- **REFLECTION**: 💭 粉色 (#ec4899)
- **CONCLUSION**: ✅ 深绿 (#22c55e)

## 📝 使用场景

### 1. AI 对话思考过程展示

```tsx
// 在聊天组件中
const [thinkingGraph, setThinkingGraph] = useState(null);

// 当 AI 开始思考时
useEffect(() => {
  let graph = createEmptyGraph(sessionId, messageId, modelName);
  
  // 逐步添加思考步骤
  aiStream.on('thinking', (step) => {
    const node = createNodeFromStep(step);
    graph = addNode(graph, node);
    if (graph.nodes.length > 1) {
      graph = addEdge(graph, {
        id: `e${graph.nodes.length}`,
        source: graph.nodes[graph.nodes.length - 2].id,
        target: node.id,
      });
    }
    setThinkingGraph({ ...graph });
  });
}, [aiStream]);

// 渲染
{thinkingGraph && (
  <ThinkingGraph graph={thinkingGraph} />
)}
```

### 2. 问题解决流程可视化

```tsx
// 展示问题解决的完整流程
const problemSolvingGraph = () => {
  let graph = createEmptyGraph();
  
  const root = createRootNode('问题', '系统性能下降');
  const analysis = createAnalysisNode('分析', 'CPU 使用率过高');
  const decision = createDecisionNode('方案选择', '', ['优化代码', '增加资源']);
  const execution = createExecutionNode('实施', '代码优化完成');
  const reflection = createReflectionNode('验证', '性能提升 50%');
  const conclusion = createConclusionNode('结论', '优化成功');
  
  // 添加所有节点和边...
  
  return graph;
};
```

### 3. 学习路径展示

```tsx
// 展示学习路线
const learningPathGraph = () => {
  let graph = createEmptyGraph();
  
  const start = createRootNode('学习目标', '成为前端工程师');
  const html = createExecutionNode('HTML', '学习 HTML5 语义化');
  const css = createExecutionNode('CSS', '掌握 CSS3 和响应式');
  const js = createExecutionNode('JavaScript', '深入理解 JS');
  const framework = createExecutionNode('框架', 'React/Vue');
  const end = createConclusionNode('完成', '可以开始找工作了');
  
  // 添加节点和边...
  
  return graph;
};
```

## 🔧 高级功能

### 自定义节点组件

```tsx
import { ThinkingNodeComponent } from './ThinkingGraph';

const CustomNodeComponent = (props) => {
  return (
    <div className="custom-node">
      {/* 自定义渲染逻辑 */}
      <ThinkingNodeComponent {...props} />
      <div className="custom-footer">
        自定义内容
      </div>
    </div>
  );
};
```

### 自定义布局算法

```typescript
import { applyTreeLayout } from './ThinkingGraph';

const customLayout = (nodes, edges) => {
  // 实现自定义布局逻辑
  return { nodes: layoutNodes, edges: layoutEdges };
};
```

### 实时流式更新

```typescript
// 配合 AI 流式响应
aiStream.on('thinking-step', (step) => {
  setGraph(prevGraph => {
    const newNode = createNodeFromStep(step);
    let updated = addNode(prevGraph, newNode);
    
    if (prevGraph.nodes.length > 0) {
      const lastNode = prevGraph.nodes[prevGraph.nodes.length - 1];
      updated = addEdge(updated, {
        id: `e${Date.now()}`,
        source: lastNode.id,
        target: newNode.id,
      });
    }
    
    return updated;
  });
});
```

## 📊 性能优化

1. **节点折叠**: 默认折叠非关键节点
2. **懒加载**: 只在展开时加载详细内容
3. **虚拟滚动**: 大型图使用虚拟滚动
4. **防抖更新**: 流式更新时使用防抖

```typescript
// 防抖示例
const debouncedUpdate = useMemo(
  () => debounce((newGraph) => setGraph(newGraph), 100),
  []
);
```

## 🐛 常见问题

### Q: 节点重叠怎么办？
A: 调整布局参数或切换到力导向布局
```typescript
<ThinkingGraph
  graph={graph}
  layout={ThinkingGraphLayout.FORCE}
/>
```

### Q: 如何自定义节点样式？
A: 通过 CSS 类名或内联样式
```typescript
const node = createRootNode('标题', '内容', {
  className: 'custom-node-style',
});
```

### Q: 支持移动端吗？
A: 支持，组件已做响应式适配

### Q: 如何导出为图片？
A: 使用 React Flow 的截图功能
```typescript
import { useReactFlow } from '@xyflow/react';

const { getViewport } = useReactFlow();
// 实现截图逻辑
```

## 📄 许可证

MIT License

## 🤝 贡献

欢迎提交 Issue 和 Pull Request！

---

**作者**: RoleCraft AI Team  
**版本**: 1.0.0  
**最后更新**: 2026-02-27
