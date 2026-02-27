/**
 * ThinkingGraph 工具函数
 * 用于创建、操作和序列化思考图数据
 */

import type {
  ThinkingGraph,
  ThinkingNode,
  ThinkingEdge,
  ThinkingGraphConfig,
  ThinkingGraphSerialized,
  NodeUpdater,
  ThinkingNodeStatusType,
} from '../types/thinking';
import {
  ThinkingNodeType,
  ThinkingNodeStatus,
  ThinkingGraphLayout,
  ThinkingGraphTheme,
} from '../types/thinking';

/**
 * 生成唯一 ID
 */
export const generateId = (): string => {
  return `node_${Date.now()}_${Math.random().toString(36).substr(2, 9)}`;
};

/**
 * 获取当前时间戳
 */
export const now = (): number => Date.now();

/**
 * 创建默认配置
 */
export const createDefaultConfig = (): ThinkingGraphConfig => ({
  layout: ThinkingGraphLayout.TREE,
  theme: ThinkingGraphTheme.LIGHT,
  autoLayout: true,
  nodeSpacing: 50,
  levelSpacing: 100,
  showTimestamp: true,
  showDuration: true,
  showConfidence: true,
  animated: true,
  minZoom: 0.1,
  maxZoom: 2,
  defaultZoom: 1,
});

/**
 * 创建空思考图
 */
export const createEmptyGraph = (
  sessionId?: string,
  messageId?: string,
  modelName?: string
): ThinkingGraph => {
  const config = createDefaultConfig();
  const timestamp = now();

  return {
    id: generateId(),
    rootId: '',
    nodes: [],
    edges: [],
    config,
    createdAt: timestamp,
    updatedAt: timestamp,
    metadata: {
      sessionId,
      messageId,
      modelName,
      totalDuration: 0,
      totalTokenUsage: {
        input: 0,
        output: 0,
      },
    },
  };
};

/**
 * 创建根节点
 */
export const createRootNode = (
  title: string,
  content: string,
  metadata?: ThinkingNode['metadata']
): ThinkingNode => {
  return {
    id: generateId(),
    type: ThinkingNodeType.ROOT,
    title,
    content,
    status: ThinkingNodeStatus.ACTIVE,
    timestamp: now(),
    expanded: true,
    metadata,
  };
};

/**
 * 创建分析节点
 */
export const createAnalysisNode = (
  title: string,
  content: string,
  metadata?: ThinkingNode['metadata']
): ThinkingNode => {
  return {
    id: generateId(),
    type: ThinkingNodeType.ANALYSIS,
    title,
    content,
    status: ThinkingNodeStatus.PENDING,
    timestamp: now(),
    metadata,
  };
};

/**
 * 创建计划节点
 */
export const createPlanningNode = (
  title: string,
  content: string,
  metadata?: ThinkingNode['metadata']
): ThinkingNode => {
  return {
    id: generateId(),
    type: ThinkingNodeType.PLANNING,
    title,
    content,
    status: ThinkingNodeStatus.PENDING,
    timestamp: now(),
    metadata,
  };
};

/**
 * 创建执行节点
 */
export const createExecutionNode = (
  title: string,
  content: string,
  metadata?: ThinkingNode['metadata']
): ThinkingNode => {
  return {
    id: generateId(),
    type: ThinkingNodeType.EXECUTION,
    title,
    content,
    status: ThinkingNodeStatus.PENDING,
    timestamp: now(),
    metadata,
  };
};

/**
 * 创建决策节点
 */
export const createDecisionNode = (
  title: string,
  content: string,
  options: string[],
  metadata?: ThinkingNode['metadata']
): ThinkingNode => {
  return {
    id: generateId(),
    type: ThinkingNodeType.DECISION,
    title,
    content,
    status: ThinkingNodeStatus.PENDING,
    timestamp: now(),
    metadata: {
      ...metadata,
      options,
    },
  };
};

/**
 * 创建工具节点
 */
export const createToolNode = (
  toolName: string,
  title: string,
  content: string,
  params?: Record<string, any>,
  metadata?: ThinkingNode['metadata']
): ThinkingNode => {
  return {
    id: generateId(),
    type: ThinkingNodeType.TOOL,
    title,
    content,
    status: ThinkingNodeStatus.PENDING,
    timestamp: now(),
    metadata: {
      ...metadata,
      toolName,
      toolParams: params,
    },
  };
};

/**
 * 创建反思节点
 */
export const createReflectionNode = (
  title: string,
  content: string,
  metadata?: ThinkingNode['metadata']
): ThinkingNode => {
  return {
    id: generateId(),
    type: ThinkingNodeType.REFLECTION,
    title,
    content,
    status: ThinkingNodeStatus.PENDING,
    timestamp: now(),
    metadata,
  };
};

/**
 * 创建结论节点
 */
export const createConclusionNode = (
  title: string,
  content: string,
  metadata?: ThinkingNode['metadata']
): ThinkingNode => {
  return {
    id: generateId(),
    type: ThinkingNodeType.CONCLUSION,
    title,
    content,
    status: ThinkingNodeStatus.PENDING,
    timestamp: now(),
    metadata,
  };
};

/**
 * 创建边
 */
export const createEdge = (
  source: string,
  target: string,
  label?: string,
  type: ThinkingEdge['type'] = 'smoothstep',
  metadata?: ThinkingEdge['metadata']
): ThinkingEdge => {
  return {
    id: `edge_${source}_${target}`,
    source,
    target,
    type,
    label,
    animated: false,
    metadata,
  };
};

/**
 * 向图中添加节点
 */
export const addNode = (graph: ThinkingGraph, node: ThinkingNode): ThinkingGraph => {
  const updatedGraph = { ...graph };
  updatedGraph.nodes = [...graph.nodes, node];
  updatedGraph.updatedAt = now();

  // 如果是第一个节点，设为根节点
  if (graph.nodes.length === 0) {
    updatedGraph.rootId = node.id;
  }

  return updatedGraph;
};

/**
 * 向图中添加边
 */
export const addEdge = (
  graph: ThinkingGraph,
  edge: ThinkingEdge
): ThinkingGraph => {
  const updatedGraph = { ...graph };
  updatedGraph.edges = [...graph.edges, edge];
  updatedGraph.updatedAt = now();

  return updatedGraph;
};

/**
 * 更新节点
 */
export const updateNode = (
  graph: ThinkingGraph,
  nodeId: string,
  updater: NodeUpdater
): ThinkingGraph => {
  const updatedGraph = { ...graph };
  updatedGraph.nodes = graph.nodes.map((node) =>
    node.id === nodeId ? updater(node) : node
  );
  updatedGraph.updatedAt = now();

  return updatedGraph;
};

/**
 * 更新节点状态
 */
export const updateNodeStatus = (
  graph: ThinkingGraph,
  nodeId: string,
  status: ThinkingNodeStatusType,
  duration?: number
): ThinkingGraph => {
  return updateNode(graph, nodeId, (node) => ({
    ...node,
    status,
    duration: duration !== undefined ? duration : node.duration,
  }));
};

/**
 * 完成节点
 */
export const completeNode = (
  graph: ThinkingGraph,
  nodeId: string,
  duration?: number
): ThinkingGraph => {
  return updateNodeStatus(graph, nodeId, ThinkingNodeStatus.COMPLETED, duration);
};

/**
 * 激活节点
 */
export const activateNode = (graph: ThinkingGraph, nodeId: string): ThinkingGraph => {
  // 先将所有节点设为非激活
  const deactivatedGraph = { ...graph };
  deactivatedGraph.nodes = graph.nodes.map((node) => ({
    ...node,
    status: node.status === ThinkingNodeStatus.ACTIVE ? ThinkingNodeStatus.PENDING : node.status,
  }));

  // 然后激活指定节点
  return updateNodeStatus(deactivatedGraph, nodeId, ThinkingNodeStatus.ACTIVE);
};

/**
 * 切换节点展开状态
 */
export const toggleNodeExpand = (
  graph: ThinkingGraph,
  nodeId: string
): ThinkingGraph => {
  return updateNode(graph, nodeId, (node) => ({
    ...node,
    expanded: !node.expanded,
  }));
};

/**
 * 展开所有节点
 */
export const expandAllNodes = (graph: ThinkingGraph): ThinkingGraph => {
  const updatedGraph = { ...graph };
  updatedGraph.nodes = graph.nodes.map((node) => ({
    ...node,
    expanded: true,
  }));
  updatedGraph.updatedAt = now();
  return updatedGraph;
};

/**
 * 折叠所有节点
 */
export const collapseAllNodes = (graph: ThinkingGraph): ThinkingGraph => {
  const updatedGraph = { ...graph };
  updatedGraph.nodes = graph.nodes.map((node) => ({
    ...node,
    expanded: node.type === ThinkingNodeType.ROOT, // 根节点保持展开
  }));
  updatedGraph.updatedAt = now();
  return updatedGraph;
};

/**
 * 删除节点
 */
export const deleteNode = (
  graph: ThinkingGraph,
  nodeId: string
): ThinkingGraph => {
  const updatedGraph = { ...graph };
  updatedGraph.nodes = graph.nodes.filter((node) => node.id !== nodeId);
  updatedGraph.edges = graph.edges.filter(
    (edge) => edge.source !== nodeId && edge.target !== nodeId
  );
  updatedGraph.updatedAt = now();

  // 如果删除的是根节点，重新设置根节点
  if (graph.rootId === nodeId) {
    updatedGraph.rootId = updatedGraph.nodes[0]?.id || '';
  }

  return updatedGraph;
};

/**
 * 序列化思考图
 */
export const serializeGraph = (graph: ThinkingGraph): ThinkingGraphSerialized => {
  return {
    version: '1.0.0',
    graph,
  };
};

/**
 * 反序列化思考图
 */
export const deserializeGraph = (data: ThinkingGraphSerialized): ThinkingGraph => {
  // 可以添加版本迁移逻辑
  return data.graph;
};

/**
 * 将思考图转换为 JSON 字符串
 */
export const graphToJson = (graph: ThinkingGraph): string => {
  return JSON.stringify(serializeGraph(graph), null, 2);
};

/**
 * 从 JSON 字符串加载思考图
 */
export const graphFromJson = (json: string): ThinkingGraph => {
  const data = JSON.parse(json) as ThinkingGraphSerialized;
  return deserializeGraph(data);
};

/**
 * 计算图的统计信息
 */
export const getGraphStats = (graph: ThinkingGraph) => {
  const nodesByType = graph.nodes.reduce((acc, node) => {
    acc[node.type] = (acc[node.type] || 0) + 1;
    return acc;
  }, {} as Record<string, number>);

  const nodesByStatus = graph.nodes.reduce((acc, node) => {
    acc[node.status] = (acc[node.status] || 0) + 1;
    return acc;
  }, {} as Record<string, number>);

  const totalDuration = graph.nodes.reduce((sum, node) => sum + (node.duration || 0), 0);
  const totalTokens = graph.nodes.reduce(
    (sum, node) => ({
      input: sum.input + (node.metadata?.tokenUsage?.input || 0),
      output: sum.output + (node.metadata?.tokenUsage?.output || 0),
    }),
    { input: 0, output: 0 }
  );

  return {
    totalNodes: graph.nodes.length,
    totalEdges: graph.edges.length,
    nodesByType,
    nodesByStatus,
    totalDuration,
    totalTokens,
    avgDuration: graph.nodes.length > 0 ? totalDuration / graph.nodes.length : 0,
  };
};

/**
 * 获取节点的子节点
 */
export const getNodeChildren = (
  graph: ThinkingGraph,
  nodeId: string
): ThinkingNode[] => {
  const edgeTargets = graph.edges
    .filter((edge) => edge.source === nodeId)
    .map((edge) => edge.target);

  return graph.nodes.filter((node) => edgeTargets.includes(node.id));
};

/**
 * 获取节点的父节点
 */
export const getNodeParent = (
  graph: ThinkingGraph,
  nodeId: string
): ThinkingNode | undefined => {
  const edgeSource = graph.edges
    .find((edge) => edge.target === nodeId)
    ?.source;

  if (!edgeSource) return undefined;

  return graph.nodes.find((node) => node.id === edgeSource);
};

/**
 * 获取节点的路径 (从根节点到该节点)
 */
export const getNodePath = (
  graph: ThinkingGraph,
  nodeId: string
): ThinkingNode[] => {
  const path: ThinkingNode[] = [];
  let currentNode = graph.nodes.find((n) => n.id === nodeId);

  while (currentNode) {
    path.unshift(currentNode);
    const parent = getNodeParent(graph, currentNode.id);
    currentNode = parent;
  }

  return path;
};

/**
 * 获取图的深度
 */
export const getGraphDepth = (graph: ThinkingGraph): number => {
  if (!graph.rootId || graph.nodes.length === 0) return 0;

  const visited = new Set<string>();
  let maxDepth = 0;

  const dfs = (nodeId: string, depth: number) => {
    if (visited.has(nodeId)) return;
    visited.add(nodeId);

    maxDepth = Math.max(maxDepth, depth);

    const children = getNodeChildren(graph, nodeId);
    children.forEach((child) => dfs(child.id, depth + 1));
  };

  dfs(graph.rootId, 1);
  return maxDepth;
};

/**
 * 验证图的完整性
 */
export const validateGraph = (graph: ThinkingGraph): { valid: boolean; errors: string[] } => {
  const errors: string[] = [];

  // 检查根节点
  if (!graph.rootId) {
    errors.push('缺少根节点 ID');
  } else if (!graph.nodes.find((n) => n.id === graph.rootId)) {
    errors.push('根节点不存在');
  }

  // 检查边的引用
  graph.edges.forEach((edge) => {
    if (!graph.nodes.find((n) => n.id === edge.source)) {
      errors.push(`边的源节点不存在：${edge.source}`);
    }
    if (!graph.nodes.find((n) => n.id === edge.target)) {
      errors.push(`边的目标节点不存在：${edge.target}`);
    }
  });

  // 检查孤立节点 (除了根节点)
  graph.nodes.forEach((node) => {
    if (node.id === graph.rootId) return;

    const hasParent = graph.edges.some((e) => e.target === node.id);
    const hasChild = graph.edges.some((e) => e.source === node.id);

    if (!hasParent && !hasChild) {
      errors.push(`孤立节点：${node.id}`);
    }
  });

  return {
    valid: errors.length === 0,
    errors,
  };
};

/**
 * 克隆图
 */
export const cloneGraph = (graph: ThinkingGraph): ThinkingGraph => {
  return JSON.parse(graphToJson(graph)) as ThinkingGraph;
};

/**
 * 导出为 DOT 格式 (用于 Graphviz)
 */
export const exportToDot = (graph: ThinkingGraph): string => {
  let dot = 'digraph ThinkingGraph {\n';
  dot += '  rankdir=TB;\n';
  dot += '  node [shape=box, style=rounded];\n\n';

  // 节点
  graph.nodes.forEach((node) => {
    const label = `${node.title}\\n${node.type}`;
    dot += `  "${node.id}" [label="${label}"];\n`;
  });

  dot += '\n';

  // 边
  graph.edges.forEach((edge) => {
    const label = edge.label ? ` [label="${edge.label}"]` : '';
    dot += `  "${edge.source}" -> "${edge.target}"${label};\n`;
  });

  dot += '}';
  return dot;
};

/**
 * 导出为 Mermaid 格式
 */
export const exportToMermaid = (graph: ThinkingGraph): string => {
  let mermaid = 'graph TD\n';

  // 节点
  graph.nodes.forEach((node) => {
    const id = node.id.replace(/[^a-zA-Z0-9]/g, '_');
    const label = node.title.replace(/"/g, "'");
    mermaid += `  ${id}["${label}"]\n`;
  });

  // 边
  graph.edges.forEach((edge) => {
    const source = edge.source.replace(/[^a-zA-Z0-9]/g, '_');
    const target = edge.target.replace(/[^a-zA-Z0-9]/g, '_');
    const label = edge.label ? `|${edge.label}|` : '';
    mermaid += `  ${source} -->${label} ${target}\n`;
  });

  return mermaid;
};
