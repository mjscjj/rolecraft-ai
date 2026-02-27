/**
 * ThinkingGraph 使用示例
 * 演示如何创建和展示思考图
 */

import React, { useState, useEffect } from 'react';
import { ThinkingGraph } from './ThinkingGraph';
import type {
  ThinkingGraph as ThinkingGraphType,
  ThinkingGraphLayoutType,
  ThinkingGraphThemeType,
} from '../../types/thinking';
import {
  ThinkingNodeType,
  ThinkingNodeStatus,
  ThinkingGraphLayout,
  ThinkingGraphTheme,
} from '../../types/thinking';
import {
  createEmptyGraph,
  createRootNode,
  createAnalysisNode,
  createPlanningNode,
  createExecutionNode,
  createDecisionNode,
  createToolNode,
  createReflectionNode,
  createConclusionNode,
  addNode,
  addEdge,
  activateNode,
  completeNode,
  expandAllNodes,
  collapseAllNodes,
  updateNode,
} from '../../utils/thinkingGraph';

/**
 * 模拟 AI 思考过程
 */
const simulateThinkingProcess = async (): Promise<ThinkingGraphType> => {
  // 创建空图
  let graph = createEmptyGraph(
    'session_123',
    'msg_456',
    'qwen3.5-plus'
  );

  // 1. 创建根节点 (用户问题)
  const rootNode = createRootNode(
    '用户问题：如何实现一个高效的排序算法？',
    '需要分析不同排序算法的优缺点，并给出实现建议'
  );
  graph = addNode(graph, rootNode);

  // 等待一下模拟思考
  await new Promise(resolve => setTimeout(resolve, 500));

  // 2. 分析问题
  const analysisNode = createAnalysisNode(
    '问题分析',
    '这是一个关于算法选择和实现的问题。需要考虑：\n1. 数据规模\n2. 时间复杂度要求\n3. 空间复杂度要求\n4. 稳定性要求',
    { confidence: 0.95 }
  );
  graph = addNode(graph, analysisNode);
  graph = addEdge(graph, { id: 'e1', source: rootNode.id, target: analysisNode.id });
  graph = activateNode(graph, analysisNode.id);

  await new Promise(resolve => setTimeout(resolve, 600));
  graph = completeNode(graph, analysisNode.id, 600);

  // 3. 规划解决步骤
  const planningNode = createPlanningNode(
    '解决规划',
    '1. 介绍常见排序算法\n2. 对比时间和空间复杂度\n3. 分析适用场景\n4. 提供代码实现示例',
    { confidence: 0.92 }
  );
  graph = addNode(graph, planningNode);
  graph = addEdge(graph, { id: 'e2', source: analysisNode.id, target: planningNode.id });
  graph = activateNode(graph, planningNode.id);

  await new Promise(resolve => setTimeout(resolve, 700));
  graph = completeNode(graph, planningNode.id, 700);

  // 4. 决策点 - 选择重点讲解的算法
  const decisionNode = createDecisionNode(
    '算法选择',
    '应该重点讲解哪些排序算法？',
    ['快速排序', '归并排序', '堆排序', 'TimSort'],
    { confidence: 0.88 }
  );
  graph = addNode(graph, decisionNode);
  graph = addEdge(graph, { id: 'e3', source: planningNode.id, target: decisionNode.id });
  graph = activateNode(graph, decisionNode.id);

  await new Promise(resolve => setTimeout(resolve, 500));

  // 更新决策结果
  graph = updateNode(graph, decisionNode.id, (node) => ({
    ...node,
    status: ThinkingNodeStatus.COMPLETED,
    metadata: {
      ...node.metadata,
      selectedOption: '快速排序 + 归并排序',
    },
    duration: 500,
  }));

  // 5. 工具调用 - 搜索代码示例
  const toolNode = createToolNode(
    'CodeSearch',
    '搜索代码示例',
    '在代码库中搜索快速排序和归并排序的最佳实践实现',
    { query: 'quicksort mergesort best practice', language: 'typescript' },
    { confidence: 0.90 }
  );
  graph = addNode(graph, toolNode);
  graph = addEdge(graph, { id: 'e4', source: decisionNode.id, target: toolNode.id, label: '需要示例' });
  graph = activateNode(graph, toolNode.id);

  await new Promise(resolve => setTimeout(resolve, 800));
  graph = completeNode(graph, toolNode.id, 800);
  graph = updateNode(graph, toolNode.id, (node) => ({
    ...node,
    metadata: {
      ...node.metadata,
      tokenUsage: { input: 150, output: 450 },
    },
  }));

  // 6. 执行 - 提供详细信息
  const executionNode1 = createExecutionNode(
    '快速排序详解',
    '快速排序采用分治策略：\n1. 选择基准值\n2. 分区操作\n3. 递归排序\n\n时间复杂度：O(n log n)\n空间复杂度：O(log n)',
    { confidence: 0.96 }
  );
  graph = addNode(graph, executionNode1);
  graph = addEdge(graph, { id: 'e5', source: toolNode.id, target: executionNode1.id });

  await new Promise(resolve => setTimeout(resolve, 600));
  graph = completeNode(graph, executionNode1.id, 600);

  const executionNode2 = createExecutionNode(
    '归并排序详解',
    '归并排序采用分治策略：\n1. 分解数组\n2. 递归排序\n3. 合并结果\n\n时间复杂度：O(n log n)\n空间复杂度：O(n)\n稳定性：稳定',
    { confidence: 0.96 }
  );
  graph = addNode(graph, executionNode2);
  graph = addEdge(graph, { id: 'e6', source: toolNode.id, target: executionNode2.id });

  await new Promise(resolve => setTimeout(resolve, 600));
  graph = completeNode(graph, executionNode2.id, 600);

  // 7. 反思
  const reflectionNode = createReflectionNode(
    '质量检查',
    '检查内容是否完整：\n✓ 算法原理清晰\n✓ 复杂度分析准确\n✓ 代码示例完整\n✓ 适用场景明确\n\n可以进一步提供性能对比数据',
    { confidence: 0.93 }
  );
  graph = addNode(graph, reflectionNode);
  graph = addEdge(graph, { id: 'e7', source: executionNode2.id, target: reflectionNode.id });
  graph = activateNode(graph, reflectionNode.id);

  await new Promise(resolve => setTimeout(resolve, 400));
  graph = completeNode(graph, reflectionNode.id, 400);

  // 8. 结论
  const conclusionNode = createConclusionNode(
    '总结建议',
    '推荐方案：\n1. 一般场景：使用内置排序 (TimSort)\n2. 内存受限：快速排序\n3. 需要稳定：归并排序\n4. 大数据：外部排序\n\n已提供 TypeScript 实现示例',
    { 
      confidence: 0.97,
      tokenUsage: { input: 2800, output: 650 }
    }
  );
  graph = addNode(graph, conclusionNode);
  graph = addEdge(graph, { id: 'e8', source: reflectionNode.id, target: conclusionNode.id });
  graph = activateNode(graph, conclusionNode.id);

  await new Promise(resolve => setTimeout(resolve, 300));
  graph = completeNode(graph, conclusionNode.id, 300);

  // 更新元数据
  graph.metadata = {
    ...graph.metadata,
    totalDuration: graph.nodes.reduce((sum, n) => sum + (n.duration || 0), 0),
    totalTokenUsage: {
      input: 2950,
      output: 1100,
    },
  };

  return graph;
};

/**
 * 演示组件
 */
export const ThinkingGraphDemo: React.FC = () => {
  const [graph, setGraph] = useState<ThinkingGraphType | null>(null);
  const [layout, setLayout] = useState<ThinkingGraphLayoutType>(ThinkingGraphLayout.TREE);
  const [theme, setTheme] = useState<ThinkingGraphThemeType>(ThinkingGraphTheme.LIGHT);
  const [isSimulating, setIsSimulating] = useState(false);

  const handleSimulate = async () => {
    setIsSimulating(true);
    try {
      const result = await simulateThinkingProcess();
      setGraph(result);
    } catch (error) {
      console.error('Simulation failed:', error);
    } finally {
      setIsSimulating(false);
    }
  };

  const handleExpandAll = () => {
    if (graph) {
      setGraph(expandAllNodes(graph));
    }
  };

  const handleCollapseAll = () => {
    if (graph) {
      setGraph(collapseAllNodes(graph));
    }
  };

  const handleNodeClick = (node: any) => {
    console.log('Node clicked:', node);
  };

  const handleNodeDoubleClick = (node: any) => {
    console.log('Node double clicked:', node);
  };

  return (
    <div style={{ padding: '20px', fontFamily: 'sans-serif' }}>
      <div style={{ marginBottom: '20px' }}>
        <h1>🧠 ThinkingGraph 演示</h1>
        <p>思维导图式 AI 思考路径可视化</p>

        <div style={{ marginBottom: '10px' }}>
          <button
            onClick={handleSimulate}
            disabled={isSimulating}
            style={{
              padding: '10px 20px',
              marginRight: '10px',
              backgroundColor: isSimulating ? '#ccc' : '#3b82f6',
              color: 'white',
              border: 'none',
              borderRadius: '6px',
              cursor: isSimulating ? 'not-allowed' : 'pointer',
            }}
          >
            {isSimulating ? '⏳ 模拟中...' : '▶️ 开始模拟思考过程'}
          </button>

          <button
            onClick={handleExpandAll}
            disabled={!graph}
            style={{
              padding: '10px 20px',
              marginRight: '10px',
              backgroundColor: !graph ? '#ccc' : '#10b981',
              color: 'white',
              border: 'none',
              borderRadius: '6px',
              cursor: !graph ? 'not-allowed' : 'pointer',
            }}
          >
            📖 展开全部
          </button>

          <button
            onClick={handleCollapseAll}
            disabled={!graph}
            style={{
              padding: '10px 20px',
              marginRight: '10px',
              backgroundColor: !graph ? '#ccc' : '#f59e0b',
              color: 'white',
              border: 'none',
              borderRadius: '6px',
              cursor: !graph ? 'not-allowed' : 'pointer',
            }}
          >
            📕 折叠全部
          </button>
        </div>

        <div style={{ marginBottom: '10px' }}>
          <label style={{ marginRight: '10px' }}>
            布局：
            <select
              value={layout}
              onChange={(e) => setLayout(e.target.value as ThinkingGraphLayoutType)}
              style={{ marginLeft: '5px', padding: '5px' }}
            >
              <option value={ThinkingGraphLayout.TREE}>树状布局</option>
              <option value={ThinkingGraphLayout.FORCE}>力导向布局</option>
              <option value={ThinkingGraphLayout.HIERARCHICAL}>层次布局</option>
            </select>
          </label>

          <label>
            主题：
            <select
              value={theme}
              onChange={(e) => setTheme(e.target.value as ThinkingGraphThemeType)}
              style={{ marginLeft: '5px', padding: '5px' }}
            >
              <option value={ThinkingGraphTheme.LIGHT}>浅色</option>
              <option value={ThinkingGraphTheme.DARK}>深色</option>
            </select>
          </label>
        </div>
      </div>

      {graph ? (
        <div style={{ height: '600px', border: '1px solid #e2e8f0', borderRadius: '8px' }}>
          <ThinkingGraph
            graph={graph}
            layout={layout}
            theme={theme}
            showControls={true}
            showMinimap={true}
            showBackground={true}
            onNodeClick={handleNodeClick}
            onNodeDoubleClick={handleNodeDoubleClick}
          />
        </div>
      ) : (
        <div
          style={{
            height: '600px',
            border: '2px dashed #e2e8f0',
            borderRadius: '8px',
            display: 'flex',
            alignItems: 'center',
            justifyContent: 'center',
            color: '#718096',
          }}
        >
          <div style={{ textAlign: 'center' }}>
            <p style={{ fontSize: '48px', marginBottom: '20px' }}>🎯</p>
            <p>点击"开始模拟思考过程"查看演示</p>
            <p style={{ fontSize: '14px', marginTop: '10px' }}>
              这将模拟一个完整的 AI 思考流程，包括分析、规划、决策、工具调用等步骤
            </p>
          </div>
        </div>
      )}

      {graph && (
        <div style={{ marginTop: '20px', padding: '15px', background: '#f7fafc', borderRadius: '8px' }}>
          <h3>📊 统计信息</h3>
          <ul style={{ listStyle: 'none', padding: 0 }}>
            <li>节点数：{graph.nodes.length}</li>
            <li>边数：{graph.edges.length}</li>
            <li>总耗时：{((graph.metadata?.totalDuration || 0) / 1000).toFixed(1)} 秒</li>
            <li>
              Token 使用：
              ↑{graph.metadata?.totalTokenUsage?.input || 0} 
              ↓{graph.metadata?.totalTokenUsage?.output || 0}
            </li>
            <li>模型：{graph.metadata?.modelName || 'N/A'}</li>
          </ul>
        </div>
      )}
    </div>
  );
};

export default ThinkingGraphDemo;
