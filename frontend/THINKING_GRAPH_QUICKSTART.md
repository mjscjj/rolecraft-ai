# ThinkingGraph 快速开始指南

## 🚀 5 分钟上手

### 1. 安装依赖 (已完成)

```bash
cd rolecraft-ai/frontend
npm install @xyflow/react
```

### 2. 最简示例

创建 `src/App.tsx`:

```tsx
import { ThinkingGraph } from './components/Chat/ThinkingGraph';
import { createEmptyGraph, createRootNode, addNode } from './utils/thinkingGraph';

// 创建图
let graph = createEmptyGraph();
const root = createRootNode('我的问题', '如何实现一个待办事项应用？');
graph = addNode(graph, root);

function App() {
  return (
    <div style={{ height: '500px' }}>
      <ThinkingGraph graph={graph} />
    </div>
  );
}

export default App;
```

### 3. 查看演示

运行演示组件:

```tsx
import { ThinkingGraphDemo } from './components/Chat/ThinkingGraphDemo';

function App() {
  return <ThinkingGraphDemo />;
}
```

启动开发服务器:

```bash
npm run dev
```

## 📖 核心 API

### 创建图

```typescript
import { createEmptyGraph } from './utils/thinkingGraph';

const graph = createEmptyGraph(
  'session_123',  // 会话 ID (可选)
  'msg_456',      // 消息 ID (可选)
  'qwen3.5-plus'  // 模型名称 (可选)
);
```

### 创建节点

```typescript
import {
  createRootNode,       // 🎯 根节点
  createAnalysisNode,   // 🔍 分析节点
  createPlanningNode,   // 📋 计划节点
  createExecutionNode,  // ⚡ 执行节点
  createDecisionNode,   // 🤔 决策节点
  createToolNode,       // 🛠️ 工具节点
  createReflectionNode, // 💭 反思节点
  createConclusionNode, // ✅ 结论节点
} from './utils/thinkingGraph';

// 示例
const root = createRootNode(
  '问题标题',
  '详细内容...',
  { confidence: 0.95 }  // 元数据 (可选)
);
```

### 添加节点和边

```typescript
import { addNode, addEdge } from './utils/thinkingGraph';

// 添加节点
graph = addNode(graph, node);

// 添加边
graph = addEdge(graph, {
  id: 'edge-1',
  source: node1.id,
  target: node2.id,
  label: '导致',  // 可选
});
```

### 更新节点状态

```typescript
import {
  activateNode,    // 激活节点
  completeNode,    // 完成节点
  toggleNodeExpand // 切换展开状态
} from './utils/thinkingGraph';

// 激活节点 (其他自动设为非激活)
graph = activateNode(graph, nodeId);

// 完成节点
graph = completeNode(graph, nodeId, 500); // 500ms 耗时

// 切换展开
graph = toggleNodeExpand(graph, nodeId);
```

### 渲染图

```tsx
import { ThinkingGraph } from './components/Chat/ThinkingGraph';

<ThinkingGraph
  graph={graph}
  layout="tree"           // 布局：tree | force | hierarchical
  theme="light"           // 主题：light | dark
  showControls={true}     // 显示控制面板
  showMinimap={true}      // 显示小地图
  showBackground={true}   // 显示背景
  onNodeClick={handleNodeClick}
  onNodeDoubleClick={handleNodeDoubleClick}
/>
```

## 🎯 完整示例：AI 思考流程

```tsx
import { useState, useEffect } from 'react';
import { ThinkingGraph } from './components/Chat/ThinkingGraph';
import {
  createEmptyGraph,
  createRootNode,
  createAnalysisNode,
  createConclusionNode,
  addNode,
  addEdge,
  activateNode,
  completeNode,
} from './utils/thinkingGraph';

function AIThinkingVisualization({ aiStream }) {
  const [graph, setGraph] = useState(() => createEmptyGraph());

  useEffect(() => {
    // 监听 AI 思考步骤
    aiStream.on('thinking-start', () => {
      setGraph(createEmptyGraph());
    });

    aiStream.on('thinking-step', async (step) => {
      setGraph(prevGraph => {
        let updated = prevGraph;

        // 创建节点
        let node;
        switch (step.type) {
          case 'root':
            node = createRootNode(step.title, step.content);
            break;
          case 'analysis':
            node = createAnalysisNode(step.title, step.content);
            break;
          case 'conclusion':
            node = createConclusionNode(step.title, step.content);
            break;
          default:
            node = createExecutionNode(step.title, step.content);
        }

        // 添加节点
        updated = addNode(updated, node);

        // 添加边 (连接到上一个节点)
        if (prevGraph.nodes.length > 0) {
          const lastNode = prevGraph.nodes[prevGraph.nodes.length - 1];
          updated = addEdge(updated, {
            id: `e${Date.now()}`,
            source: lastNode.id,
            target: node.id,
          });
        }

        // 激活新节点
        updated = activateNode(updated, node.id);

        return updated;
      });

      // 模拟耗时后完成节点
      setTimeout(() => {
        setGraph(prev => completeNode(prev, step.id, step.duration));
      }, step.duration);
    });
  }, [aiStream]);

  return (
    <div style={{ height: '600px', border: '1px solid #ddd', borderRadius: '8px' }}>
      <ThinkingGraph
        graph={graph}
        layout="tree"
        theme="light"
      />
    </div>
  );
}
```

## 🎨 样式定制

### 修改主题颜色

在 CSS 中覆盖变量:

```css
.my-custom-theme {
  --tg-bg: #f0f0f0;
  --tg-node-bg: #ffffff;
  --tg-node-border: #3b82f6;
  --tg-text: #1a202c;
  --tg-edge-color: #3b82f6;
}
```

```tsx
<ThinkingGraph
  graph={graph}
  className="my-custom-theme"
/>
```

### 自定义节点样式

```typescript
const node = createRootNode('标题', '内容', {
  className: 'my-custom-node',
});
```

然后在 CSS 中:

```css
.my-custom-node {
  background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
  color: white;
}
```

## 📤 导出功能

### 导出为 JSON

```typescript
import { graphToJson, graphFromJson } from './utils/thinkingGraph';

// 导出
const json = graphToJson(graph);
localStorage.setItem('myGraph', json);

// 导入
const savedJson = localStorage.getItem('myGraph');
const loadedGraph = graphFromJson(savedJson);
```

### 导出为 Mermaid

```typescript
import { exportToMermaid } from './utils/thinkingGraph';

const mermaid = exportToMermaid(graph);
console.log(mermaid);
// 可以在 Markdown 中渲染
```

### 导出为 DOT (Graphviz)

```typescript
import { exportToDot } from './utils/thinkingGraph';

const dot = exportToDot(graph);
// 可以用 Graphviz 渲染为图片
```

## 🔧 常见问题

### Q: 如何实时更新图？

使用 React 的 `useState` 和 AI 流式事件:

```typescript
aiStream.on('thinking-step', (step) => {
  setGraph(prev => {
    const node = createNodeFromStep(step);
    let updated = addNode(prev, node);
    if (prev.nodes.length > 0) {
      updated = addEdge(updated, {
        id: `e${Date.now()}`,
        source: prev.nodes[prev.nodes.length - 1].id,
        target: node.id,
      });
    }
    return updated;
  });
});
```

### Q: 节点太多性能差怎么办？

1. 默认折叠节点: `graph = collapseAllNodes(graph)`
2. 使用力导向布局: `<ThinkingGraph layout="force" />`
3. 减少显示细节: 只在展开时加载详细内容

### Q: 如何自定义布局？

修改 `ThinkingGraph.tsx` 中的布局算法，或传入自定义布局函数。

### Q: 支持移动端吗？

支持！组件已做响应式适配，支持触摸缩放和拖拽。

## 📚 更多资源

- **完整文档**: `src/components/Chat/ThinkingGraph_README.md`
- **实现细节**: `THINKING_GRAPH_IMPLEMENTATION.md`
- **演示代码**: `src/components/Chat/ThinkingGraphDemo.tsx`
- **类型定义**: `src/types/thinking.ts`
- **工具函数**: `src/utils/thinkingGraph.ts`

---

**开始构建你的 AI 思考可视化吧!** 🚀
