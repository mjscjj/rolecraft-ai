/**
 * Thinking Graph Types
 * 用于思维导图式 AI 思考路径展示的数据结构
 */

/**
 * 思考节点类型
 */
export const ThinkingNodeType = {
  /** 根节点 - 初始问题/任务 */
  ROOT: 'root',
  /** 分析节点 - 问题分析/理解 */
  ANALYSIS: 'analysis',
  /** 计划节点 - 策略规划 */
  PLANNING: 'planning',
  /** 执行节点 - 具体执行步骤 */
  EXECUTION: 'execution',
  /** 决策节点 - 分支决策点 */
  DECISION: 'decision',
  /** 工具节点 - 工具调用 */
  TOOL: 'tool',
  /** 反思节点 - 自我反思/检查 */
  REFLECTION: 'reflection',
  /** 结论节点 - 最终结论 */
  CONCLUSION: 'conclusion',
} as const;

export type ThinkingNodeTypeType = typeof ThinkingNodeType[keyof typeof ThinkingNodeType];

/**
 * 节点状态
 */
export const ThinkingNodeStatus = {
  /** 待处理 */
  PENDING: 'pending',
  /** 进行中 */
  ACTIVE: 'active',
  /** 已完成 */
  COMPLETED: 'completed',
  /** 已跳过 */
  SKIPPED: 'skipped',
  /** 错误 */
  ERROR: 'error',
} as const;

export type ThinkingNodeStatusType = typeof ThinkingNodeStatus[keyof typeof ThinkingNodeStatus];

/**
 * 思考节点
 */
export interface ThinkingNode {
  /** 节点唯一标识 */
  id: string;
  /** 节点类型 */
  type: ThinkingNodeTypeType;
  /** 节点标题/简短描述 */
  title: string;
  /** 详细内容 */
  content: string;
  /** 节点状态 */
  status: ThinkingNodeStatusType;
  /** 创建时间戳 */
  timestamp: number;
  /** 执行耗时 (毫秒) */
  duration?: number;
  /** 子节点 ID 列表 */
  children?: string[];
  /** 元数据 */
  metadata?: {
    /** 工具名称 (tool 类型) */
    toolName?: string;
    /** 工具参数 */
    toolParams?: Record<string, any>;
    /** 决策选项 */
    options?: string[];
    /** 选择的结果 */
    selectedOption?: string;
    /** 错误信息 */
    error?: string;
    /** 置信度 (0-1) */
    confidence?: number;
    /** Token 使用量 */
    tokenUsage?: {
      input: number;
      output: number;
    };
  };
  /** 是否展开 (UI 状态) */
  expanded?: boolean;
  /** 自定义样式类名 */
  className?: string;
}

/**
 * 思考边 (节点间连接)
 */
export interface ThinkingEdge {
  /** 边唯一标识 */
  id: string;
  /** 源节点 ID */
  source: string;
  /** 目标节点 ID */
  target: string;
  /** 边类型 */
  type?: 'default' | 'smoothstep' | 'step' | 'straight';
  /** 边标签 */
  label?: string;
  /** 是否动画 */
  animated?: boolean;
  /** 样式 */
  style?: {
    stroke?: string;
    strokeWidth?: number;
    strokeDasharray?: string;
  };
  /** 元数据 */
  metadata?: {
    /** 转换原因 */
    reason?: string;
    /** 条件 */
    condition?: string;
  };
}

/**
 * 思考图布局类型
 */
export const ThinkingGraphLayout = {
  /** 树状布局 (从上到下) */
  TREE: 'tree',
  /** 力导向布局 */
  FORCE: 'force',
  /** 层次布局 */
  HIERARCHICAL: 'hierarchical',
  /** 从左到右树状 */
  TREE_LR: 'tree-lr',
} as const;

export type ThinkingGraphLayoutType = typeof ThinkingGraphLayout[keyof typeof ThinkingGraphLayout];

/**
 * 思考图主题
 */
export const ThinkingGraphTheme = {
  /** 浅色主题 */
  LIGHT: 'light',
  /** 深色主题 */
  DARK: 'dark',
  /** 自动 (跟随系统) */
  AUTO: 'auto',
} as const;

export type ThinkingGraphThemeType = typeof ThinkingGraphTheme[keyof typeof ThinkingGraphTheme];

/**
 * 思考图配置
 */
export interface ThinkingGraphConfig {
  /** 布局类型 */
  layout: ThinkingGraphLayoutType;
  /** 主题 */
  theme: ThinkingGraphThemeType;
  /** 是否自动布局 */
  autoLayout: boolean;
  /** 节点间距 */
  nodeSpacing: number;
  /** 层级间距 */
  levelSpacing: number;
  /** 是否显示时间戳 */
  showTimestamp: boolean;
  /** 是否显示耗时 */
  showDuration: boolean;
  /** 是否显示置信度 */
  showConfidence: boolean;
  /** 是否动画 */
  animated: boolean;
  /** 最小缩放 */
  minZoom: number;
  /** 最大缩放 */
  maxZoom: number;
  /** 默认缩放 */
  defaultZoom: number;
}

/**
 * 思考图数据结构
 */
export interface ThinkingGraph {
  /** 图唯一标识 */
  id: string;
  /** 根节点 ID */
  rootId: string;
  /** 所有节点 */
  nodes: ThinkingNode[];
  /** 所有边 */
  edges: ThinkingEdge[];
  /** 配置 */
  config: ThinkingGraphConfig;
  /** 创建时间 */
  createdAt: number;
  /** 更新时间 */
  updatedAt: number;
  /** 元数据 */
  metadata?: {
    /** 会话 ID */
    sessionId?: string;
    /** 消息 ID */
    messageId?: string;
    /** 模型名称 */
    modelName?: string;
    /** 总耗时 */
    totalDuration?: number;
    /** 总 token 使用 */
    totalTokenUsage?: {
      input: number;
      output: number;
    };
  };
}

/**
 * React Flow 节点数据类型
 */
export interface ThinkingFlowNodeData {
  /** 节点类型 */
  type: ThinkingNodeTypeType;
  /** 标题 */
  title: string;
  /** 内容 */
  content: string;
  /** 状态 */
  status: ThinkingNodeStatusType;
  /** 时间戳 */
  timestamp: number;
  /** 耗时 */
  duration?: number;
  /** 置信度 */
  confidence?: number;
  /** 是否展开 */
  expanded: boolean;
  /** 元数据 */
  metadata?: ThinkingNode['metadata'];
  /** 切换展开状态回调 */
  onToggleExpand?: (id: string) => void;
  /** 编辑内容回调 */
  onEdit?: (id: string, content: string) => void;
}

/**
 * 思考图序列化数据 (用于存储/传输)
 */
export interface ThinkingGraphSerialized {
  version: string;
  graph: ThinkingGraph;
}

/**
 * 布局选项
 */
export interface LayoutOptions {
  /** 布局方向 */
  direction: 'TB' | 'LR' | 'RL' | 'BT';
  /** 节点宽度 */
  nodeWidth?: number;
  /** 节点高度 */
  nodeHeight?: number;
  /** 水平间距 */
  spacingX?: number;
  /** 垂直间距 */
  spacingY?: number;
  /** 对齐方式 */
  align?: 'left' | 'center' | 'right';
}

/**
 * 工具函数类型
 */
export type NodeUpdater = (node: ThinkingNode) => ThinkingNode;
export type GraphUpdater = (graph: ThinkingGraph) => ThinkingGraph;
