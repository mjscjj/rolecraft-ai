/**
 * ThinkingGraph - 思维导图式 AI 思考路径展示组件
 * 
 * 使用 React Flow 实现可交互的思维导图，展示 AI 的思考过程
 */

import React, { useCallback, useMemo, useState, useEffect } from 'react';
import {
  ReactFlow,
  Background,
  Controls,
  MiniMap,
  useNodesState,
  useEdgesState,
  Handle,
  Position,
  ReactFlowProvider,
  Panel,
} from '@xyflow/react';
import '@xyflow/react/dist/style.css';

import type {
  ThinkingNode,
  ThinkingEdge,
  ThinkingGraph as ThinkingGraphType,
  ThinkingFlowNodeData,
  LayoutOptions,
  ThinkingGraphLayoutType,
  ThinkingGraphThemeType,
} from '../../types/thinking';
import {
  ThinkingNodeType,
  ThinkingNodeStatus,
  ThinkingGraphLayout,
  ThinkingGraphTheme,
} from '../../types/thinking';

import './ThinkingGraph.css';

// ============== 节点组件 ==============

/**
 * 思考节点渲染组件
 */
const ThinkingNodeComponent: React.FC<{
  id: string;
  data: ThinkingFlowNodeData;
  selected?: boolean;
}> = ({ id, data, selected }) => {
  const [isExpanded, setIsExpanded] = useState(data.expanded);

  useEffect(() => {
    setIsExpanded(data.expanded);
  }, [data.expanded]);

  const handleToggle = useCallback(() => {
    setIsExpanded(!isExpanded);
    data.onToggleExpand?.(id);
  }, [id, isExpanded, data.onToggleExpand]);

  const getNodeIcon = () => {
    switch (data.type) {
      case ThinkingNodeType.ROOT: return '🎯';
      case ThinkingNodeType.ANALYSIS: return '🔍';
      case ThinkingNodeType.PLANNING: return '📋';
      case ThinkingNodeType.EXECUTION: return '⚡';
      case ThinkingNodeType.DECISION: return '🤔';
      case ThinkingNodeType.TOOL: return '🛠️';
      case ThinkingNodeType.REFLECTION: return '💭';
      case ThinkingNodeType.CONCLUSION: return '✅';
      default: return '📌';
    }
  };

  const getStatusClass = () => {
    switch (data.status) {
      case ThinkingNodeStatus.PENDING: return 'node-pending';
      case ThinkingNodeStatus.ACTIVE: return 'node-active';
      case ThinkingNodeStatus.COMPLETED: return 'node-completed';
      case ThinkingNodeStatus.SKIPPED: return 'node-skipped';
      case ThinkingNodeStatus.ERROR: return 'node-error';
      default: return '';
    }
  };

  const getTypeClass = () => {
    return `node-type-${data.type.toLowerCase()}`;
  };

  const formatDuration = (ms?: number) => {
    if (!ms) return '';
    if (ms < 1000) return `${ms}ms`;
    return `${(ms / 1000).toFixed(1)}s`;
  };

  const formatTime = (timestamp: number) => {
    return new Date(timestamp).toLocaleTimeString('zh-CN', {
      hour: '2-digit',
      minute: '2-digit',
      second: '2-digit',
    });
  };

  return (
    <div className={`thinking-node ${getStatusClass()} ${getTypeClass()} ${selected ? 'selected' : ''}`}>
      <Handle type="target" position={Position.Top} className="handle target" />
      
      {/* 节点头部 */}
      <div className="node-header" onDoubleClick={handleToggle}>
        <span className="node-icon">{getNodeIcon()}</span>
        <span className="node-title">{data.title}</span>
        {data.confidence !== undefined && (
          <span className="node-confidence" title="置信度">
            {(data.confidence * 100).toFixed(0)}%
          </span>
        )}
      </div>

      {/* 节点状态指示器 */}
      <div className="node-status-indicator">
        {data.status === ThinkingNodeStatus.ACTIVE && (
          <span className="status-dot active"></span>
        )}
        {data.status === ThinkingNodeStatus.ERROR && (
          <span className="status-dot error"></span>
        )}
      </div>

      {/* 展开内容 */}
      {isExpanded && (
        <div className="node-content">
          <div className="node-text">{data.content}</div>
          
          {/* 元数据信息 */}
          {data.metadata && (
            <div className="node-metadata">
              {data.metadata.toolName && (
                <div className="meta-item">
                  <strong>工具:</strong> {data.metadata.toolName}
                </div>
              )}
              {data.metadata.selectedOption && (
                <div className="meta-item">
                  <strong>选择:</strong> {data.metadata.selectedOption}
                </div>
              )}
              {data.metadata.tokenUsage && (
                <div className="meta-item">
                  <strong>Tokens:</strong> 
                  ↑{data.metadata.tokenUsage.input} ↓{data.metadata.tokenUsage.output}
                </div>
              )}
              {data.metadata.error && (
                <div className="meta-item error">
                  <strong>错误:</strong> {data.metadata.error}
                </div>
              )}
            </div>
          )}

          {/* 时间和耗时 */}
          <div className="node-footer">
            <span className="node-time">{formatTime(data.timestamp)}</span>
            {data.duration !== undefined && (
              <span className="node-duration">{formatDuration(data.duration)}</span>
            )}
          </div>
        </div>
      )}

      <Handle type="source" position={Position.Bottom} className="handle source" />
    </div>
  );
};

// ============== 自定义边组件 ==============

const ThinkingEdgeComponent: React.FC<{
  id: string;
  sourceX: number;
  sourceY: number;
  targetX: number;
  targetY: number;
  sourcePosition: Position;
  targetPosition: Position;
  style?: React.CSSProperties;
  markerEnd?: string;
  data?: {
    label?: string;
    reason?: string;
    condition?: string;
  };
}> = ({ 
  sourceX, 
  sourceY, 
  targetX, 
  targetY, 
  sourcePosition, 
  targetPosition,
  style = {},
  markerEnd,
  data,
}) => {
  // 计算贝塞尔曲线路径
  const [edgePath, labelX, labelY] = React.useMemo(() => {
    const deltaX = Math.abs(targetX - sourceX);
    const deltaY = Math.abs(targetY - sourceY);
    
    // 控制点偏移
    const controlPointOffset = Math.max(deltaX * 0.5, deltaY * 0.5, 50);
    
    let path = '';
    let lx = (sourceX + targetX) / 2;
    let ly = (sourceY + targetY) / 2;

    if (sourcePosition === Position.Bottom && targetPosition === Position.Top) {
      // 垂直流向
      path = `M${sourceX},${sourceY} C${sourceX},${sourceY + controlPointOffset} ${targetX},${targetY - controlPointOffset} ${targetX},${targetY}`;
    } else if (sourcePosition === Position.Right && targetPosition === Position.Left) {
      // 水平流向
      path = `M${sourceX},${sourceY} C${sourceX + controlPointOffset},${sourceY} ${targetX - controlPointOffset},${targetY} ${targetX},${targetY}`;
    } else {
      // 默认直线
      path = `M${sourceX},${sourceY} L${targetX},${targetY}`;
    }

    return [path, lx, ly];
  }, [sourceX, sourceY, targetX, targetY, sourcePosition, targetPosition]);

  return (
    <g className="thinking-edge">
      <path
        className="react-flow__edge-path"
        d={edgePath}
        style={style}
        markerEnd={markerEnd}
      />
      {data?.label && (
        <foreignObject
          x={labelX - 50}
          y={labelY - 15}
          width={100}
          height={30}
          className="edge-label-fo"
          style={{ overflow: 'visible' }}
        >
          <div className="edge-label">{data.label}</div>
        </foreignObject>
      )}
    </g>
  );
};

// ============== 布局算法 ==============

/**
 * 树状布局算法
 */
const applyTreeLayout = (
  nodes: ThinkingNode[],
  edges: ThinkingEdge[],
  options: LayoutOptions = { direction: 'TB' }
): { nodes: any[]; edges: any[] } => {
  const nodeMap = new Map<string, ThinkingNode>();
  const childrenMap = new Map<string, string[]>();
  const parentMap = new Map<string, string>();

  // 构建父子关系
  nodes.forEach(node => {
    nodeMap.set(node.id, node);
    childrenMap.set(node.id, []);
  });

  edges.forEach(edge => {
    const children = childrenMap.get(edge.source) || [];
    children.push(edge.target);
    childrenMap.set(edge.source, children);
    parentMap.set(edge.target, edge.source);
  });

  // 找到根节点
  const rootId = nodes.find(n => !parentMap.has(n.id))?.id || nodes[0]?.id;
  if (!rootId) return { nodes: [], edges: [] };

  const layoutNodes: any[] = [];
  const layoutEdges: any[] = [];
  
  const nodeWidth = options.nodeWidth || 200;
  const nodeHeight = options.nodeHeight || 100;
  const spacingX = options.spacingX || 50;
  const spacingY = options.spacingY || 100;

  // 递归布局
  const layoutNode = (nodeId: string, x: number, y: number, level: number) => {
    const node = nodeMap.get(nodeId);
    if (!node) return;

    layoutNodes.push({
      id: node.id,
      type: 'thinkingNode',
      position: { x, y },
      data: {
        type: node.type,
        title: node.title,
        content: node.content,
        status: node.status,
        timestamp: node.timestamp,
        duration: node.duration,
        confidence: node.metadata?.confidence,
        expanded: node.expanded ?? false,
        metadata: node.metadata,
      },
      style: {
        width: nodeWidth,
      },
    });

    const children = childrenMap.get(nodeId) || [];
    const totalWidth = children.length * (nodeWidth + spacingX) - spacingX;
    let childX = x - totalWidth / 2;

    children.forEach(childId => {
      const edge = edges.find(e => e.source === nodeId && e.target === childId);
      if (edge) {
        layoutEdges.push({
          id: edge.id,
          source: edge.source,
          target: edge.target,
          type: edge.type || 'smoothstep',
          label: edge.label,
          animated: edge.animated,
          style: edge.style,
        });
      }

      layoutNode(childId, childX, y + nodeHeight + spacingY, level + 1);
      childX += nodeWidth + spacingX;
    });
  };

  layoutNode(rootId, 0, 0, 0);

  return { nodes: layoutNodes, edges: layoutEdges };
};

/**
 * 力导向布局算法 (简化版)
 */
const applyForceLayout = (
  nodes: ThinkingNode[],
  edges: ThinkingEdge[],
  iterations: number = 50
): { nodes: any[]; edges: any[] } => {
  const nodeMap = new Map<string, { x: number; y: number; vx: number; vy: number }>();
  const layoutNodes: any[] = [];
  const layoutEdges: any[] = [];

  // 初始化位置
  nodes.forEach((node, i) => {
    const angle = (2 * Math.PI * i) / nodes.length;
    const radius = 200;
    nodeMap.set(node.id, {
      x: Math.cos(angle) * radius,
      y: Math.sin(angle) * radius,
      vx: 0,
      vy: 0,
    });
  });

  // 力导向迭代
  for (let iter = 0; iter < iterations; iter++) {
    const forces = new Map<string, { fx: number; fy: number }>();
    nodes.forEach(n => forces.set(n.id, { fx: 0, fy: 0 }));

    // 斥力 (节点间)
    nodes.forEach((node1, i) => {
      nodes.slice(i + 1).forEach(node2 => {
        const pos1 = nodeMap.get(node1.id)!;
        const pos2 = nodeMap.get(node2.id)!;
        
        const dx = pos2.x - pos1.x;
        const dy = pos2.y - pos1.y;
        const dist = Math.sqrt(dx * dx + dy * dy) || 1;
        
        const force = 5000 / (dist * dist);
        const fx = (dx / dist) * force;
        const fy = (dy / dist) * force;

        const f1 = forces.get(node1.id)!;
        const f2 = forces.get(node2.id)!;
        
        f1.fx -= fx;
        f1.fy -= fy;
        f2.fx += fx;
        f2.fy += fy;
      });
    });

    // 引力 (边连接)
    edges.forEach(edge => {
      const pos1 = nodeMap.get(edge.source)!;
      const pos2 = nodeMap.get(edge.target)!;
      
      const dx = pos2.x - pos1.x;
      const dy = pos2.y - pos1.y;
      const dist = Math.sqrt(dx * dx + dy * dy) || 1;
      
      const force = dist * 0.01;
      const fx = (dx / dist) * force;
      const fy = (dy / dist) * force;

      const f1 = forces.get(edge.source)!;
      const f2 = forces.get(edge.target)!;
      
      f1.fx += fx;
      f1.fy += fy;
      f2.fx -= fx;
      f2.fy -= fy;
    });

    // 更新位置
    nodes.forEach(node => {
      const pos = nodeMap.get(node.id)!;
      const force = forces.get(node.id)!;
      
      pos.x += force.fx;
      pos.y += force.fy;
      pos.x *= 0.95; // 阻尼
      pos.y *= 0.95;
    });
  }

  // 生成结果
  nodes.forEach(node => {
    const pos = nodeMap.get(node.id)!;
    layoutNodes.push({
      id: node.id,
      type: 'thinkingNode',
      position: { x: pos.x, y: pos.y },
      data: {
        type: node.type,
        title: node.title,
        content: node.content,
        status: node.status,
        timestamp: node.timestamp,
        duration: node.duration,
        confidence: node.metadata?.confidence,
        expanded: node.expanded ?? false,
        metadata: node.metadata,
      },
    });
  });

  edges.forEach(edge => {
    layoutEdges.push({
      id: edge.id,
      source: edge.source,
      target: edge.target,
      type: edge.type || 'smoothstep',
      label: edge.label,
      animated: edge.animated,
      style: edge.style,
    });
  });

  return { nodes: layoutNodes, edges: layoutEdges };
};

// ============== 主组件 ==============

interface ThinkingGraphProps {
  /** 思考图数据 */
  graph: ThinkingGraphType;
  /** 布局类型 */
  layout?: ThinkingGraphLayoutType;
  /** 主题 */
  theme?: ThinkingGraphThemeType;
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

const ThinkingGraphInner: React.FC<ThinkingGraphProps> = ({
  graph,
  layout = ThinkingGraphLayout.TREE,
  theme = ThinkingGraphTheme.LIGHT,
  showControls = true,
  showMinimap = true,
  showBackground = true,
  onNodeClick,
  onNodeDoubleClick,
  onEdgeClick,
  className = '',
  style,
}) => {
  const [nodes, setNodes, onNodesChange] = useNodesState<any[]>([]);
  const [edges, setEdges, onEdgesChange] = useEdgesState<any[]>([]);
  const [currentLayout, setCurrentLayout] = useState<ThinkingGraphLayoutType>(layout);

  // 节点类型映射
  const nodeTypes = useMemo(() => ({
    thinkingNode: ThinkingNodeComponent,
  }), []);

  // 边类型映射
  const edgeTypes = useMemo(() => ({
    thinkingEdge: ThinkingEdgeComponent,
  }), []);

  // 应用布局
  useEffect(() => {
    let layoutResult: { nodes: any[]; edges: any[] };

    if (currentLayout === 'tree' || currentLayout === 'tree-lr') {
      layoutResult = applyTreeLayout(graph.nodes, graph.edges, {
        direction: currentLayout === 'tree-lr' ? 'LR' : 'TB',
        nodeWidth: 220,
        nodeHeight: 80,
        spacingX: 60,
        spacingY: 120,
      });
    } else if (currentLayout === 'force') {
      layoutResult = applyForceLayout(graph.nodes, graph.edges, 50);
    } else if (currentLayout === 'hierarchical') {
      layoutResult = applyTreeLayout(graph.nodes, graph.edges, {
        direction: 'TB',
        nodeWidth: 200,
        nodeHeight: 100,
        spacingX: 40,
        spacingY: 150,
        align: 'center',
      });
    } else {
      layoutResult = applyTreeLayout(graph.nodes, graph.edges);
    }

    setNodes(layoutResult.nodes);
    setEdges(layoutResult.edges);
  }, [graph, currentLayout, setNodes, setEdges]);

  // 处理节点点击
  const handleNodeClick = useCallback(
    (_: React.MouseEvent, node: any) => {
      const thinkingNode = graph.nodes.find(n => n.id === node.id);
      if (thinkingNode && onNodeClick) {
        onNodeClick(thinkingNode);
      }
    },
    [graph.nodes, onNodeClick]
  );

  // 处理节点双击
  const handleNodeDoubleClick = useCallback(
    (_: React.MouseEvent, node: any) => {
      const thinkingNode = graph.nodes.find(n => n.id === node.id);
      if (thinkingNode && onNodeDoubleClick) {
        onNodeDoubleClick(thinkingNode);
      }
    },
    [graph.nodes, onNodeDoubleClick]
  );

  // 处理边点击
  const handleEdgeClick = useCallback(
    (_: React.MouseEvent, edge: any) => {
      const thinkingEdge = graph.edges.find(e => e.id === edge.id);
      if (thinkingEdge && onEdgeClick) {
        onEdgeClick(thinkingEdge);
      }
    },
    [graph.edges, onEdgeClick]
  );

  // 切换布局
  const handleLayoutChange = useCallback((newLayout: ThinkingGraphLayoutType) => {
    setCurrentLayout(newLayout);
  }, []);

  // 主题类名
  const themeClass = theme === 'dark' ? 'theme-dark' : 'theme-light';

  return (
    <div className={`thinking-graph-container ${themeClass} ${className}`} style={style}>
      <ReactFlow
        nodes={nodes}
        edges={edges}
        onNodesChange={onNodesChange}
        onEdgesChange={onEdgesChange}
        onNodeClick={handleNodeClick}
        onNodeDoubleClick={handleNodeDoubleClick}
        onEdgeClick={handleEdgeClick}
        nodeTypes={nodeTypes}
        edgeTypes={edgeTypes}
        fitView
        snapToGrid
        snapGrid={[15, 15]}
        minZoom={graph.config?.minZoom || 0.1}
        maxZoom={graph.config?.maxZoom || 2}
        defaultZoom={graph.config?.defaultZoom || 1}
        className={`thinking-flow ${themeClass}`}
      >
        {showBackground && <Background variant="dots" gap={20} size={1} />}
        {showControls && <Controls />}
        {showMinimap && (
          <MiniMap
            nodeColor={(node) => {
              const type = (node.data as ThinkingFlowNodeData)?.type;
              switch (type) {
                case ThinkingNodeType.ROOT: return '#3b82f6';
                case ThinkingNodeType.ANALYSIS: return '#8b5cf6';
                case ThinkingNodeType.PLANNING: return '#06b6d4';
                case ThinkingNodeType.EXECUTION: return '#10b981';
                case ThinkingNodeType.DECISION: return '#f59e0b';
                case ThinkingNodeType.TOOL: return '#ef4444';
                case ThinkingNodeType.REFLECTION: return '#ec4899';
                case ThinkingNodeType.CONCLUSION: return '#22c55e';
                default: return '#6b7280';
              }
            }}
            maskColor="rgb(240, 240, 240, 0.8)"
          />
        )}

        {/* 布局切换面板 */}
        <Panel position="top-right" className="layout-panel">
          <div className="layout-buttons">
            <button
              className={currentLayout === ThinkingGraphLayout.TREE ? 'active' : ''}
              onClick={() => handleLayoutChange(ThinkingGraphLayout.TREE)}
              title="树状布局"
            >
              🌳 树状
            </button>
            <button
              className={currentLayout === ThinkingGraphLayout.FORCE ? 'active' : ''}
              onClick={() => handleLayoutChange(ThinkingGraphLayout.FORCE)}
              title="力导向布局"
            >
              🧲 力导向
            </button>
            <button
              className={currentLayout === ThinkingGraphLayout.HIERARCHICAL ? 'active' : ''}
              onClick={() => handleLayoutChange(ThinkingGraphLayout.HIERARCHICAL)}
              title="层次布局"
            >
              📊 层次
            </button>
          </div>
        </Panel>

        {/* 统计信息面板 */}
        <Panel position="top-left" className="stats-panel">
          <div className="stats-info">
            <span>📊 节点：{graph.nodes.length}</span>
            <span>🔗 边：{graph.edges.length}</span>
            {graph.metadata?.totalDuration && (
              <span>⏱️ 耗时：{(graph.metadata.totalDuration / 1000).toFixed(1)}s</span>
            )}
          </div>
        </Panel>
      </ReactFlow>
    </div>
  );
};

/**
 * ThinkingGraph 组件 (带 Provider)
 */
export const ThinkingGraph: React.FC<ThinkingGraphProps> = (props) => {
  return (
    <ReactFlowProvider>
      <ThinkingGraphInner {...props} />
    </ReactFlowProvider>
  );
};

export default ThinkingGraph;
