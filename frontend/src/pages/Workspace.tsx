import { useEffect, useMemo, useState } from 'react';
import companyApi, { type Company } from '../api/company';
import roleApi, { type Role } from '../api/role';
import workApi, { type AgentRun, type WorkspaceTask } from '../api/work';

const formatTime = (value?: string) => {
  if (!value) return '未设置';
  const date = new Date(value);
  if (Number.isNaN(date.getTime())) return value;
  return date.toLocaleString('zh-CN', { hour12: false });
};

const clipText = (value?: string, limit = 200) => {
  if (!value) return '';
  const text = value.trim();
  if (!text) return '';
  return text.length > limit ? `${text.slice(0, limit)}...` : text;
};

const triggerValueHint = (triggerType: string) => {
  if (triggerType === 'daily') return '例如 09:00';
  if (triggerType === 'interval_hours') return '例如 4（每 4 小时）';
  if (triggerType === 'once') return '例如 2026-03-01T09:00:00+08:00';
  return 'manual 模式可留空';
};

type ExecutionConfig = {
  executionMode: 'serial' | 'parallel' | string;
  timeoutSeconds: number;
  maxRetries: number;
  retryDelaySeconds: number;
  archiveToCompany: boolean;
  queueRetryOnFailure: boolean;
  retryWindowMinutes: number;
  maxFailureCycles: number;
};

const toInt = (value: unknown, fallback: number) => {
  const num = Number(value);
  if (Number.isNaN(num) || !Number.isFinite(num)) return fallback;
  return Math.trunc(num);
};

const defaultExecutionConfig = (archiveToCompany: boolean): ExecutionConfig => ({
  executionMode: 'serial',
  timeoutSeconds: 180,
  maxRetries: 1,
  retryDelaySeconds: 3,
  archiveToCompany,
  queueRetryOnFailure: true,
  retryWindowMinutes: 60,
  maxFailureCycles: 3,
});

const readExecutionConfig = (task: WorkspaceTask): ExecutionConfig => {
  const fallback = defaultExecutionConfig(Boolean(task.companyId));
  const config = (task.config || {}) as Record<string, any>;
  const mode = String(config.executionMode || fallback.executionMode).toLowerCase();
  return {
    executionMode: mode === 'parallel' ? 'parallel' : 'serial',
    timeoutSeconds: toInt(config.timeoutSeconds, fallback.timeoutSeconds),
    maxRetries: toInt(config.maxRetries, fallback.maxRetries),
    retryDelaySeconds: toInt(config.retryDelaySeconds, fallback.retryDelaySeconds),
    archiveToCompany: Boolean(
      typeof config.archiveToCompany === 'boolean' ? config.archiveToCompany : fallback.archiveToCompany
    ),
    queueRetryOnFailure: Boolean(
      typeof config.queueRetryOnFailure === 'boolean' ? config.queueRetryOnFailure : fallback.queueRetryOnFailure
    ),
    retryWindowMinutes: toInt(config.retryWindowMinutes, fallback.retryWindowMinutes),
    maxFailureCycles: toInt(config.maxFailureCycles, fallback.maxFailureCycles),
  };
};

export const WorkspacePage = () => {
  const [rows, setRows] = useState<WorkspaceTask[]>([]);
  const [companies, setCompanies] = useState<Company[]>([]);
  const [roles, setRoles] = useState<Role[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState('');

  const [name, setName] = useState('');
  const [description, setDescription] = useState('');
  const [companyId, setCompanyId] = useState('');
  const [roleId, setRoleId] = useState('');
  const [type, setType] = useState('general');
  const [status, setStatus] = useState('todo');
  const [priority, setPriority] = useState('medium');
  const [triggerType, setTriggerType] = useState('manual');
  const [triggerValue, setTriggerValue] = useState('');
  const [timezone, setTimezone] = useState('Asia/Shanghai');
  const [inputSource, setInputSource] = useState('');
  const [reportRule, setReportRule] = useState('');
  const [executionMode, setExecutionMode] = useState('serial');
  const [timeoutSeconds, setTimeoutSeconds] = useState('180');
  const [maxRetries, setMaxRetries] = useState('1');
  const [retryDelaySeconds, setRetryDelaySeconds] = useState('3');
  const [queueRetryOnFailure, setQueueRetryOnFailure] = useState(true);
  const [retryWindowMinutes, setRetryWindowMinutes] = useState('60');
  const [maxFailureCycles, setMaxFailureCycles] = useState('3');

  const [filterCompanyId, setFilterCompanyId] = useState('');
  const [filterAsyncStatus, setFilterAsyncStatus] = useState('');
  const [selectedTaskIds, setSelectedTaskIds] = useState<string[]>([]);
  const [runDetails, setRunDetails] = useState<Record<string, AgentRun[]>>({});
  const [openRuns, setOpenRuns] = useState<Record<string, boolean>>({});
  const [selectedRun, setSelectedRun] = useState<AgentRun | null>(null);
  const [runLoadingId, setRunLoadingId] = useState('');
  const [runningWorkId, setRunningWorkId] = useState('');
  const [batchRunning, setBatchRunning] = useState(false);
  const [editingStrategyTask, setEditingStrategyTask] = useState<WorkspaceTask | null>(null);
  const [strategyConfig, setStrategyConfig] = useState<ExecutionConfig>(defaultExecutionConfig(false));
  const [batchStrategyOpen, setBatchStrategyOpen] = useState(false);
  const [batchStrategyConfig, setBatchStrategyConfig] = useState<ExecutionConfig>(defaultExecutionConfig(false));

  const load = async () => {
    setLoading(true);
    setError('');
    try {
      const [workspaceRows, companyRows, roleRows] = await Promise.all([
        workApi.list({
          companyId: filterCompanyId || undefined,
          asyncStatus: filterAsyncStatus || undefined,
        }),
        companyApi.list(),
        roleApi.list(),
      ]);
      setRows(workspaceRows);
      setCompanies(companyRows);
      setRoles(roleRows);
      setSelectedTaskIds((prev) => prev.filter((id) => workspaceRows.some((row) => row.id === id)));
    } catch (err: any) {
      setError(err?.message || '加载工作区失败');
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    load();
  }, [filterCompanyId, filterAsyncStatus]);

  const roleOptions = useMemo(() => {
    if (!companyId) return roles.filter((item) => !item.companyId);
    return roles.filter((item) => item.companyId === companyId);
  }, [roles, companyId]);

  const createWorkspaceTask = async () => {
    if (!name.trim()) {
      setError('请输入工作区任务名称');
      return;
    }
    try {
      await workApi.create({
        name: name.trim(),
        description: description.trim(),
        companyId: companyId || undefined,
        roleId: roleId || undefined,
        type,
        status,
        priority,
        triggerType,
        triggerValue: triggerValue.trim() || undefined,
        timezone: timezone.trim() || 'Asia/Shanghai',
        inputSource: inputSource.trim() || undefined,
        reportRule: reportRule.trim() || undefined,
        config: {
          executionMode,
          timeoutSeconds: toInt(timeoutSeconds, 180),
          maxRetries: toInt(maxRetries, 1),
          retryDelaySeconds: toInt(retryDelaySeconds, 3),
          archiveToCompany: Boolean(companyId),
          queueRetryOnFailure,
          retryWindowMinutes: toInt(retryWindowMinutes, 60),
          maxFailureCycles: toInt(maxFailureCycles, 3),
        },
      });
      setName('');
      setDescription('');
      setRoleId('');
      setType('general');
      setStatus('todo');
      setPriority('medium');
      setTriggerType('manual');
      setTriggerValue('');
      setInputSource('');
      setReportRule('');
      setExecutionMode('serial');
      setTimeoutSeconds('180');
      setMaxRetries('1');
      setRetryDelaySeconds('3');
      setQueueRetryOnFailure(true);
      setRetryWindowMinutes('60');
      setMaxFailureCycles('3');
      await load();
    } catch (err: any) {
      setError(err?.message || '创建工作区任务失败');
    }
  };

  const runWorkspaceTask = async (id: string) => {
    try {
      setRunningWorkId(id);
      const result = await workApi.run(id);
      setRunDetails((prev) => ({
        ...prev,
        [id]: [result.run, ...(prev[id] || []).filter((item) => item.id !== result.run.id)].slice(0, 10),
      }));
      await load();
    } catch (err: any) {
      setError(err?.message || '执行失败');
    } finally {
      setRunningWorkId('');
    }
  };

  const updateStatus = async (item: WorkspaceTask, nextStatus: string) => {
    try {
      await workApi.update(item.id, {
        name: item.name,
        description: item.description,
        companyId: item.companyId,
        roleId: item.roleId,
        type: item.type,
        status: nextStatus,
        priority: item.priority,
        triggerType: item.triggerType,
        triggerValue: item.triggerValue,
        timezone: item.timezone,
        inputSource: item.inputSource,
        reportRule: item.reportRule,
      });
      await load();
    } catch (err: any) {
      setError(err?.message || '更新状态失败');
    }
  };

  const deleteTask = async (item: WorkspaceTask) => {
    if (!window.confirm(`确认删除「${item.name}」吗？`)) return;
    try {
      await workApi.delete(item.id);
      await load();
    } catch (err: any) {
      setError(err?.message || '删除失败');
    }
  };

  const toggleRuns = async (id: string) => {
    const shouldOpen = !openRuns[id];
    setOpenRuns((prev) => ({ ...prev, [id]: shouldOpen }));
    if (!shouldOpen || runDetails[id]) {
      return;
    }
    try {
      const runs = await workApi.listRuns(id, 10);
      setRunDetails((prev) => ({ ...prev, [id]: runs }));
    } catch (err: any) {
      setError(err?.message || '加载协商记录失败');
    }
  };

  const openRunDetail = async (workId: string, runId: string) => {
    try {
      setRunLoadingId(runId);
      const run = await workApi.getRun(workId, runId);
      setSelectedRun(run);
    } catch (err: any) {
      setError(err?.message || '加载执行详情失败');
    } finally {
      setRunLoadingId('');
    }
  };

  const openStrategyEditor = (item: WorkspaceTask) => {
    setEditingStrategyTask(item);
    setStrategyConfig(readExecutionConfig(item));
  };

  const saveStrategyConfig = async () => {
    if (!editingStrategyTask) return;
    try {
      await workApi.update(editingStrategyTask.id, {
        name: editingStrategyTask.name,
        description: editingStrategyTask.description,
        companyId: editingStrategyTask.companyId,
        roleId: editingStrategyTask.roleId,
        type: editingStrategyTask.type,
        status: editingStrategyTask.status,
        priority: editingStrategyTask.priority,
        triggerType: editingStrategyTask.triggerType,
        triggerValue: editingStrategyTask.triggerValue,
        timezone: editingStrategyTask.timezone,
        inputSource: editingStrategyTask.inputSource,
        reportRule: editingStrategyTask.reportRule,
        config: strategyConfig,
      });
      setEditingStrategyTask(null);
      await load();
    } catch (err: any) {
      setError(err?.message || '保存策略失败');
    }
  };

  const toggleTaskSelected = (id: string) => {
    setSelectedTaskIds((prev) => (prev.includes(id) ? prev.filter((item) => item !== id) : [...prev, id]));
  };

  const selectAllVisible = () => {
    setSelectedTaskIds(rows.map((item) => item.id));
  };

  const clearSelected = () => {
    setSelectedTaskIds([]);
  };

  const getSelectedTasks = () => rows.filter((item) => selectedTaskIds.includes(item.id));

  const updateTaskWithConfig = async (item: WorkspaceTask, config: ExecutionConfig, asyncStatus?: string) => {
    await workApi.update(item.id, {
      name: item.name,
      description: item.description,
      companyId: item.companyId,
      roleId: item.roleId,
      type: item.type,
      status: item.status,
      priority: item.priority,
      triggerType: item.triggerType,
      triggerValue: item.triggerValue,
      timezone: item.timezone,
      inputSource: item.inputSource,
      reportRule: item.reportRule,
      asyncStatus,
      config,
    });
  };

  const openBatchStrategyEditor = () => {
    const selected = getSelectedTasks();
    if (selected.length === 0) {
      setError('请先选择至少一个任务');
      return;
    }
    setBatchStrategyConfig(readExecutionConfig(selected[0]));
    setBatchStrategyOpen(true);
  };

  const saveBatchStrategy = async () => {
    const selected = getSelectedTasks();
    if (selected.length === 0) {
      setBatchStrategyOpen(false);
      return;
    }
    try {
      await Promise.all(selected.map((item) => updateTaskWithConfig(item, batchStrategyConfig)));
      setBatchStrategyOpen(false);
      await load();
    } catch (err: any) {
      setError(err?.message || '批量保存策略失败');
    }
  };

  const batchSetAsyncStatus = async (mode: 'pause' | 'resume') => {
    const selected = getSelectedTasks();
    if (selected.length === 0) {
      setError('请先选择至少一个任务');
      return;
    }
    try {
      await Promise.all(
        selected.map((item) => {
          const config = readExecutionConfig(item);
          const nextStatus =
            mode === 'pause' ? 'paused' : item.triggerType === 'manual' ? 'idle' : 'scheduled';
          return updateTaskWithConfig(item, config, nextStatus);
        })
      );
      await load();
    } catch (err: any) {
      setError(err?.message || '批量更新状态失败');
    }
  };

  const batchRunSelected = async () => {
    if (selectedTaskIds.length === 0) {
      setError('请先选择至少一个任务');
      return;
    }
    try {
      setBatchRunning(true);
      const result = await workApi.batchRun(selectedTaskIds, 3);
      setRunDetails((prev) => {
        const next = { ...prev };
        result.items.forEach((item) => {
          if (!item.run) return;
          const history = next[item.workId] || [];
          next[item.workId] = [item.run, ...history.filter((run) => run.id !== item.run?.id)].slice(0, 10);
        });
        return next;
      });
      await load();
      if (result.failedCount > 0) {
        setError(
          `批量执行完成：成功 ${result.successCount}，失败 ${result.failedCount}（忙碌 ${result.busyCount}，未找到 ${result.notFoundCount}）`
        );
      } else {
        setError('');
      }
    } catch (err: any) {
      setError(err?.message || '批量执行失败');
    } finally {
      setBatchRunning(false);
    }
  };

  const ignoreFailure = async (item: WorkspaceTask) => {
    try {
      await updateTaskWithConfig(item, readExecutionConfig(item), 'ignored');
      await load();
    } catch (err: any) {
      setError(err?.message || '忽略失败任务失败');
    }
  };

  const failedTasks = rows.filter((item) => item.asyncStatus === 'failed');

  return (
    <div className="space-y-6">
      <div>
        <h1 className="text-2xl font-bold text-slate-900">工作区</h1>
        <p className="mt-1 text-slate-500">
          异步任务中心：定义在什么时间做什么工作、何时汇报、何时分析文件，并自动沉淀结果。
        </p>
      </div>

      <div className="rounded-xl border border-slate-200 bg-white p-4">
        <h2 className="mb-3 text-base font-semibold text-slate-800">新建工作区任务</h2>
        <div className="grid gap-3 md:grid-cols-2">
          <input
            value={name}
            onChange={(e) => setName(e.target.value)}
            className="rounded-lg border border-slate-300 px-3 py-2"
            placeholder="任务名称（例如：09:00 营销日报汇报）"
          />
          <select
            value={companyId}
            onChange={(e) => {
              setCompanyId(e.target.value);
              setRoleId('');
            }}
            className="rounded-lg border border-slate-300 px-3 py-2"
          >
            <option value="">个人工作区</option>
            {companies.map((item) => (
              <option key={item.id} value={item.id}>
                {item.name}
              </option>
            ))}
          </select>
          <textarea
            value={description}
            onChange={(e) => setDescription(e.target.value)}
            className="min-h-24 rounded-lg border border-slate-300 px-3 py-2 md:col-span-2"
            placeholder="任务说明（目标、约束、交付格式）"
          />
          <select
            value={roleId}
            onChange={(e) => setRoleId(e.target.value)}
            className="rounded-lg border border-slate-300 px-3 py-2"
          >
            <option value="">不绑定角色</option>
            {roleOptions.map((item) => (
              <option key={item.id} value={item.id}>
                {item.name}
              </option>
            ))}
          </select>
          <select
            value={type}
            onChange={(e) => setType(e.target.value)}
            className="rounded-lg border border-slate-300 px-3 py-2"
          >
            <option value="general">通用执行</option>
            <option value="report">信息汇报</option>
            <option value="analyze">文件分析</option>
          </select>
          <div className="grid grid-cols-2 gap-2">
            <select
              value={status}
              onChange={(e) => setStatus(e.target.value)}
              className="rounded-lg border border-slate-300 px-3 py-2"
            >
              <option value="todo">待办</option>
              <option value="in_progress">进行中</option>
              <option value="done">完成</option>
            </select>
            <select
              value={priority}
              onChange={(e) => setPriority(e.target.value)}
              className="rounded-lg border border-slate-300 px-3 py-2"
            >
              <option value="low">低优先级</option>
              <option value="medium">中优先级</option>
              <option value="high">高优先级</option>
            </select>
          </div>
          <div className="grid grid-cols-2 gap-2">
            <select
              value={triggerType}
              onChange={(e) => setTriggerType(e.target.value)}
              className="rounded-lg border border-slate-300 px-3 py-2"
            >
              <option value="manual">手动执行</option>
              <option value="once">定时一次</option>
              <option value="daily">每日定时</option>
              <option value="interval_hours">每 N 小时</option>
            </select>
            <input
              value={triggerValue}
              onChange={(e) => setTriggerValue(e.target.value)}
              className="rounded-lg border border-slate-300 px-3 py-2"
              placeholder={triggerValueHint(triggerType)}
            />
          </div>
          <input
            value={timezone}
            onChange={(e) => setTimezone(e.target.value)}
            className="rounded-lg border border-slate-300 px-3 py-2"
            placeholder="时区，例如 Asia/Shanghai"
          />
          <input
            value={inputSource}
            onChange={(e) => setInputSource(e.target.value)}
            className="rounded-lg border border-slate-300 px-3 py-2 md:col-span-2"
            placeholder="输入来源（文档ID、文件夹、链接等）"
          />
          <input
            value={reportRule}
            onChange={(e) => setReportRule(e.target.value)}
            className="rounded-lg border border-slate-300 px-3 py-2 md:col-span-2"
            placeholder="汇报规则（例如：每天 18:00 汇总发送到公司交付区）"
          />
          <div className="grid grid-cols-2 gap-2 md:col-span-2">
            <select
              value={executionMode}
              onChange={(e) => setExecutionMode(e.target.value)}
              className="rounded-lg border border-slate-300 px-3 py-2"
            >
              <option value="serial">协商模式：串行</option>
              <option value="parallel">协商模式：并行</option>
            </select>
            <input
              value={timeoutSeconds}
              onChange={(e) => setTimeoutSeconds(e.target.value)}
              className="rounded-lg border border-slate-300 px-3 py-2"
              placeholder="超时秒数（30-1800）"
            />
          </div>
          <div className="grid grid-cols-2 gap-2 md:col-span-2">
            <input
              value={maxRetries}
              onChange={(e) => setMaxRetries(e.target.value)}
              className="rounded-lg border border-slate-300 px-3 py-2"
              placeholder="失败重试次数（0-5）"
            />
            <input
              value={retryDelaySeconds}
              onChange={(e) => setRetryDelaySeconds(e.target.value)}
              className="rounded-lg border border-slate-300 px-3 py-2"
              placeholder="重试间隔秒数（0-120）"
            />
          </div>
          <div className="grid grid-cols-2 gap-2 md:col-span-2">
            <label className="flex items-center gap-2 rounded-lg border border-slate-300 px-3 py-2 text-sm text-slate-700">
              <input
                type="checkbox"
                checked={queueRetryOnFailure}
                onChange={(e) => setQueueRetryOnFailure(e.target.checked)}
              />
              调度级失败重试队列
            </label>
            <input
              value={retryWindowMinutes}
              onChange={(e) => setRetryWindowMinutes(e.target.value)}
              className="rounded-lg border border-slate-300 px-3 py-2"
              placeholder="重试窗口分钟（5-1440）"
            />
          </div>
          <input
            value={maxFailureCycles}
            onChange={(e) => setMaxFailureCycles(e.target.value)}
            className="rounded-lg border border-slate-300 px-3 py-2 md:col-span-2"
            placeholder="窗口内最大失败周期（1-20）"
          />
        </div>
        <button
          onClick={createWorkspaceTask}
          className="mt-3 rounded-lg bg-slate-900 px-4 py-2 text-white hover:bg-slate-800"
        >
          创建工作区任务
        </button>
      </div>

      <div className="rounded-xl border border-slate-200 bg-white p-4">
        <div className="mb-3 flex flex-wrap items-center gap-2">
          <select
            value={filterCompanyId}
            onChange={(e) => setFilterCompanyId(e.target.value)}
            className="rounded-lg border border-slate-300 px-3 py-2"
          >
            <option value="">全部范围</option>
            {companies.map((item) => (
              <option key={item.id} value={item.id}>
                公司：{item.name}
              </option>
            ))}
          </select>
          <select
            value={filterAsyncStatus}
            onChange={(e) => setFilterAsyncStatus(e.target.value)}
            className="rounded-lg border border-slate-300 px-3 py-2"
          >
            <option value="">全部异步状态</option>
            <option value="idle">idle</option>
            <option value="scheduled">scheduled</option>
            <option value="running">running</option>
            <option value="completed">completed</option>
            <option value="failed">failed</option>
            <option value="paused">paused</option>
            <option value="ignored">ignored</option>
          </select>
          <button
            onClick={load}
            className="rounded-lg border border-slate-300 px-3 py-2 text-sm text-slate-700 hover:bg-slate-50"
          >
            刷新
          </button>
        </div>

        {loading && <div className="text-slate-500">加载中...</div>}
        {error && <div className="mb-3 rounded-lg bg-red-50 px-3 py-2 text-sm text-red-600">{error}</div>}

        <div className="mb-3 flex flex-wrap items-center gap-2 rounded-lg border border-slate-200 bg-slate-50 px-3 py-2">
          <span className="text-xs text-slate-600">已选 {selectedTaskIds.length} 项</span>
          <button
            onClick={selectAllVisible}
            className="rounded border border-slate-300 px-2 py-1 text-xs text-slate-600 hover:bg-white"
          >
            全选当前列表
          </button>
          <button
            onClick={clearSelected}
            className="rounded border border-slate-300 px-2 py-1 text-xs text-slate-600 hover:bg-white"
          >
            清空选择
          </button>
          <button
            onClick={batchRunSelected}
            disabled={batchRunning}
            className="rounded border border-slate-300 px-2 py-1 text-xs text-slate-600 hover:bg-white disabled:cursor-not-allowed disabled:opacity-60"
          >
            {batchRunning ? '批量执行中...' : '批量执行'}
          </button>
          <button
            onClick={openBatchStrategyEditor}
            className="rounded border border-slate-300 px-2 py-1 text-xs text-slate-600 hover:bg-white"
          >
            批量策略编辑
          </button>
          <button
            onClick={() => batchSetAsyncStatus('pause')}
            className="rounded border border-slate-300 px-2 py-1 text-xs text-slate-600 hover:bg-white"
          >
            批量暂停
          </button>
          <button
            onClick={() => batchSetAsyncStatus('resume')}
            className="rounded border border-slate-300 px-2 py-1 text-xs text-slate-600 hover:bg-white"
          >
            批量恢复
          </button>
        </div>

        {failedTasks.length > 0 && (
          <div className="mb-3 rounded-lg border border-red-200 bg-red-50 p-3">
            <p className="text-sm font-medium text-red-700">失败任务面板（{failedTasks.length}）</p>
            <div className="mt-2 space-y-2">
              {failedTasks.slice(0, 5).map((item) => (
                <div key={`failed-${item.id}`} className="flex items-center justify-between rounded border border-red-200 bg-white px-2 py-2 text-xs">
                  <span className="text-slate-700">{item.name}</span>
                  <div className="flex gap-2">
                    <button
                      onClick={() => runWorkspaceTask(item.id)}
                      className="rounded border border-slate-300 px-2 py-1 text-slate-600 hover:bg-slate-50"
                    >
                      手动重试
                    </button>
                    <button
                      onClick={() => ignoreFailure(item)}
                      className="rounded border border-red-300 px-2 py-1 text-red-600 hover:bg-red-50"
                    >
                      忽略失败
                    </button>
                  </div>
                </div>
              ))}
            </div>
          </div>
        )}

        <div className="space-y-3">
          {rows.map((item) => (
            <div key={item.id} className="rounded-lg border border-slate-200 p-4" data-testid={`workspace-item-${item.id}`}>
              <div className="flex flex-wrap items-center justify-between gap-3">
                <div>
                  <label className="mb-1 flex items-center gap-2 text-xs text-slate-500">
                    <input
                      type="checkbox"
                      checked={selectedTaskIds.includes(item.id)}
                      onChange={() => toggleTaskSelected(item.id)}
                    />
                    选择任务
                  </label>
                  <p className="font-semibold text-slate-900">{item.name}</p>
                  <p className="text-sm text-slate-500">{item.description || '暂无描述'}</p>
                </div>
                <div className="flex items-center gap-2">
                  <select
                    value={item.status}
                    onChange={(e) => updateStatus(item, e.target.value)}
                    className="rounded-lg border border-slate-300 px-3 py-1.5 text-sm"
                  >
                    <option value="todo">待办</option>
                    <option value="in_progress">进行中</option>
                    <option value="done">完成</option>
                  </select>
                  <button
                    onClick={() => runWorkspaceTask(item.id)}
                    disabled={runningWorkId === item.id}
                    className="rounded-lg bg-slate-900 px-3 py-1.5 text-sm text-white hover:bg-slate-800"
                    data-testid={`workspace-run-${item.id}`}
                  >
                    {runningWorkId === item.id ? '执行中...' : '立即执行'}
                  </button>
                  <button
                    onClick={() => toggleRuns(item.id)}
                    className="rounded-lg border border-slate-300 px-3 py-1.5 text-sm text-slate-700 hover:bg-slate-50"
                    data-testid={`workspace-toggle-runs-${item.id}`}
                  >
                    {openRuns[item.id] ? '收起协商记录' : '查看协商记录'}
                  </button>
                  <button
                    onClick={() => openStrategyEditor(item)}
                    className="rounded-lg border border-slate-300 px-3 py-1.5 text-sm text-slate-700 hover:bg-slate-50"
                  >
                    编辑策略
                  </button>
                  <button
                    onClick={() => deleteTask(item)}
                    className="rounded-lg border border-red-300 px-3 py-1.5 text-sm text-red-600 hover:bg-red-50"
                  >
                    删除
                  </button>
                </div>
              </div>
              <div className="mt-3 grid gap-2 text-xs text-slate-500 md:grid-cols-2">
                <p>类型：{item.type || 'general'}</p>
                <p>异步状态：{item.asyncStatus || 'idle'}</p>
                <p>调度：{item.triggerType || 'manual'} {item.triggerValue ? `(${item.triggerValue})` : ''}</p>
                <p>下次执行：{formatTime(item.nextRunAt)}</p>
                <p>最近执行：{formatTime(item.lastRunAt)}</p>
                <p>时区：{item.timezone || 'Asia/Shanghai'}</p>
                <p>
                  策略：
                  {(item.config as any)?.executionMode || 'serial'} /
                  超时 {(item.config as any)?.timeoutSeconds ?? 180}s /
                  重试 {(item.config as any)?.maxRetries ?? 1} /
                  队列 {(item.config as any)?.queueRetryOnFailure ? '开' : '关'}
                </p>
              </div>
              {item.resultSummary && (
                <div className="mt-3 rounded-md bg-slate-50 px-3 py-2 text-sm text-slate-700">
                  结果摘要：{item.resultSummary}
                </div>
              )}
              {openRuns[item.id] && (
                <div className="mt-3 rounded-md border border-slate-200 bg-slate-50 px-3 py-3" data-testid={`workspace-runs-${item.id}`}>
                  <p className="text-xs font-semibold text-slate-700">多 Agent 协商记录</p>
                  <div className="mt-2 space-y-2">
                    {(runDetails[item.id] || []).map((run) => (
                      <div key={run.id} className="rounded border border-slate-200 bg-white p-2" data-testid={`workspace-run-item-${run.id}`}>
                        <div className="flex flex-wrap items-center gap-2 text-xs text-slate-500">
                          <span>状态：{run.status}</span>
                          <span>触发：{run.triggerSource}</span>
                          <span>置信度：{run.confidence?.toFixed?.(2) ?? run.confidence}</span>
                          <span>时间：{formatTime(run.finishedAt || run.createdAt)}</span>
                        </div>
                        <p className="mt-1 whitespace-pre-line text-sm text-slate-700">{clipText(run.summary || run.errorMessage || '无摘要', 180)}</p>
                        {(run.trace?.steps || []).length > 0 && (
                          <div className="mt-2 space-y-1 text-xs text-slate-600">
                            {run.trace?.steps?.map((step, index) => (
                              <div key={`${run.id}-${step.agent}-${index}`}>
                                <span className="font-medium">{step.agent}</span>（{step.durationMs}ms）：{clipText(step.output, 120)}
                              </div>
                            ))}
                          </div>
                        )}
                        <div className="mt-2">
                          <button
                            onClick={() => openRunDetail(item.id, run.id)}
                            className="rounded border border-slate-300 px-2 py-1 text-xs text-slate-600 hover:bg-slate-50"
                            data-testid={`workspace-run-detail-${run.id}`}
                          >
                            {runLoadingId === run.id ? '加载中...' : '查看详情'}
                          </button>
                        </div>
                      </div>
                    ))}
                    {(runDetails[item.id] || []).length === 0 && (
                      <p className="text-xs text-slate-500">暂无协商记录</p>
                    )}
                  </div>
                </div>
              )}
            </div>
          ))}
          {!loading && rows.length === 0 && (
            <p className="text-sm text-slate-500">暂无工作区任务，先创建一个定时或异步任务。</p>
          )}
        </div>
      </div>

      {batchStrategyOpen && (
        <div className="fixed inset-0 z-50 flex items-center justify-center bg-slate-900/40 p-4">
          <div className="w-full max-w-2xl rounded-xl bg-white p-5 shadow-xl">
            <div className="mb-3 flex items-center justify-between">
              <h3 className="text-lg font-semibold text-slate-900">批量编辑策略（{selectedTaskIds.length} 项）</h3>
              <button
                onClick={() => setBatchStrategyOpen(false)}
                className="rounded border border-slate-300 px-2 py-1 text-sm text-slate-600 hover:bg-slate-50"
              >
                关闭
              </button>
            </div>
            <div className="grid gap-2 md:grid-cols-2">
              <select
                value={batchStrategyConfig.executionMode}
                onChange={(e) => setBatchStrategyConfig((prev) => ({ ...prev, executionMode: e.target.value }))}
                className="rounded border border-slate-300 px-3 py-2"
              >
                <option value="serial">协商模式：串行</option>
                <option value="parallel">协商模式：并行</option>
              </select>
              <input
                value={String(batchStrategyConfig.timeoutSeconds)}
                onChange={(e) =>
                  setBatchStrategyConfig((prev) => ({ ...prev, timeoutSeconds: toInt(e.target.value, prev.timeoutSeconds) }))
                }
                className="rounded border border-slate-300 px-3 py-2"
                placeholder="超时秒数"
              />
              <input
                value={String(batchStrategyConfig.maxRetries)}
                onChange={(e) =>
                  setBatchStrategyConfig((prev) => ({ ...prev, maxRetries: toInt(e.target.value, prev.maxRetries) }))
                }
                className="rounded border border-slate-300 px-3 py-2"
                placeholder="LLM重试次数"
              />
              <input
                value={String(batchStrategyConfig.retryDelaySeconds)}
                onChange={(e) =>
                  setBatchStrategyConfig((prev) => ({ ...prev, retryDelaySeconds: toInt(e.target.value, prev.retryDelaySeconds) }))
                }
                className="rounded border border-slate-300 px-3 py-2"
                placeholder="重试间隔秒数"
              />
              <label className="flex items-center gap-2 rounded border border-slate-300 px-3 py-2 text-sm text-slate-700 md:col-span-2">
                <input
                  type="checkbox"
                  checked={batchStrategyConfig.queueRetryOnFailure}
                  onChange={(e) =>
                    setBatchStrategyConfig((prev) => ({ ...prev, queueRetryOnFailure: e.target.checked }))
                  }
                />
                启用调度级失败重试队列
              </label>
              <input
                value={String(batchStrategyConfig.retryWindowMinutes)}
                onChange={(e) =>
                  setBatchStrategyConfig((prev) => ({ ...prev, retryWindowMinutes: toInt(e.target.value, prev.retryWindowMinutes) }))
                }
                className="rounded border border-slate-300 px-3 py-2"
                placeholder="失败窗口（分钟）"
              />
              <input
                value={String(batchStrategyConfig.maxFailureCycles)}
                onChange={(e) =>
                  setBatchStrategyConfig((prev) => ({ ...prev, maxFailureCycles: toInt(e.target.value, prev.maxFailureCycles) }))
                }
                className="rounded border border-slate-300 px-3 py-2"
                placeholder="窗口内最大失败周期"
              />
            </div>
            <div className="mt-4 flex justify-end gap-2">
              <button
                onClick={() => setBatchStrategyOpen(false)}
                className="rounded border border-slate-300 px-4 py-2 text-sm text-slate-700 hover:bg-slate-50"
              >
                取消
              </button>
              <button
                onClick={saveBatchStrategy}
                className="rounded bg-slate-900 px-4 py-2 text-sm text-white hover:bg-slate-800"
              >
                批量保存
              </button>
            </div>
          </div>
        </div>
      )}

      {editingStrategyTask && (
        <div className="fixed inset-0 z-50 flex items-center justify-center bg-slate-900/40 p-4">
          <div className="w-full max-w-2xl rounded-xl bg-white p-5 shadow-xl">
            <div className="mb-3 flex items-center justify-between">
              <h3 className="text-lg font-semibold text-slate-900">编辑执行策略</h3>
              <button
                onClick={() => setEditingStrategyTask(null)}
                className="rounded border border-slate-300 px-2 py-1 text-sm text-slate-600 hover:bg-slate-50"
              >
                关闭
              </button>
            </div>
            <p className="mb-3 text-sm text-slate-500">{editingStrategyTask.name}</p>
            <div className="grid gap-2 md:grid-cols-2">
              <select
                value={strategyConfig.executionMode}
                onChange={(e) => setStrategyConfig((prev) => ({ ...prev, executionMode: e.target.value }))}
                className="rounded border border-slate-300 px-3 py-2"
              >
                <option value="serial">协商模式：串行</option>
                <option value="parallel">协商模式：并行</option>
              </select>
              <input
                value={String(strategyConfig.timeoutSeconds)}
                onChange={(e) =>
                  setStrategyConfig((prev) => ({ ...prev, timeoutSeconds: toInt(e.target.value, prev.timeoutSeconds) }))
                }
                className="rounded border border-slate-300 px-3 py-2"
                placeholder="超时秒数"
              />
              <input
                value={String(strategyConfig.maxRetries)}
                onChange={(e) =>
                  setStrategyConfig((prev) => ({ ...prev, maxRetries: toInt(e.target.value, prev.maxRetries) }))
                }
                className="rounded border border-slate-300 px-3 py-2"
                placeholder="LLM重试次数"
              />
              <input
                value={String(strategyConfig.retryDelaySeconds)}
                onChange={(e) =>
                  setStrategyConfig((prev) => ({ ...prev, retryDelaySeconds: toInt(e.target.value, prev.retryDelaySeconds) }))
                }
                className="rounded border border-slate-300 px-3 py-2"
                placeholder="重试间隔秒数"
              />
              <label className="flex items-center gap-2 rounded border border-slate-300 px-3 py-2 text-sm text-slate-700 md:col-span-2">
                <input
                  type="checkbox"
                  checked={strategyConfig.queueRetryOnFailure}
                  onChange={(e) =>
                    setStrategyConfig((prev) => ({ ...prev, queueRetryOnFailure: e.target.checked }))
                  }
                />
                启用调度级失败重试队列
              </label>
              <input
                value={String(strategyConfig.retryWindowMinutes)}
                onChange={(e) =>
                  setStrategyConfig((prev) => ({ ...prev, retryWindowMinutes: toInt(e.target.value, prev.retryWindowMinutes) }))
                }
                className="rounded border border-slate-300 px-3 py-2"
                placeholder="失败窗口（分钟）"
              />
              <input
                value={String(strategyConfig.maxFailureCycles)}
                onChange={(e) =>
                  setStrategyConfig((prev) => ({ ...prev, maxFailureCycles: toInt(e.target.value, prev.maxFailureCycles) }))
                }
                className="rounded border border-slate-300 px-3 py-2"
                placeholder="窗口内最大失败周期"
              />
            </div>
            <div className="mt-4 flex justify-end gap-2">
              <button
                onClick={() => setEditingStrategyTask(null)}
                className="rounded border border-slate-300 px-4 py-2 text-sm text-slate-700 hover:bg-slate-50"
              >
                取消
              </button>
              <button
                onClick={saveStrategyConfig}
                className="rounded bg-slate-900 px-4 py-2 text-sm text-white hover:bg-slate-800"
              >
                保存策略
              </button>
            </div>
          </div>
        </div>
      )}

      {selectedRun && (
        <div className="fixed inset-0 z-50 flex items-center justify-center bg-slate-900/40 p-4">
          <div className="max-h-[85vh] w-full max-w-4xl overflow-auto rounded-xl bg-white p-5 shadow-xl">
            <div className="mb-3 flex items-center justify-between">
              <h3 className="text-lg font-semibold text-slate-900">执行详情</h3>
              <button
                onClick={() => setSelectedRun(null)}
                className="rounded border border-slate-300 px-2 py-1 text-sm text-slate-600 hover:bg-slate-50"
              >
                关闭
              </button>
            </div>
            <div className="grid gap-2 text-sm text-slate-600 md:grid-cols-2">
              <p>状态：{selectedRun.status}</p>
              <p>触发来源：{selectedRun.triggerSource}</p>
              <p>开始时间：{formatTime(selectedRun.startedAt || selectedRun.createdAt)}</p>
              <p>结束时间：{formatTime(selectedRun.finishedAt)}</p>
              <p>置信度：{selectedRun.confidence?.toFixed?.(2) ?? selectedRun.confidence}</p>
            </div>
            {(selectedRun.trace as any)?.policy && (
              <div className="mt-3 rounded-md border border-slate-200 bg-slate-50 p-3">
                <p className="text-xs font-semibold text-slate-700">执行策略</p>
                <div className="mt-1 grid gap-1 text-xs text-slate-600 md:grid-cols-2">
                  <p>协商模式：{(selectedRun.trace as any).policy?.executionMode || 'serial'}</p>
                  <p>超时：{(selectedRun.trace as any).policy?.timeoutSeconds ?? 180}s</p>
                  <p>LLM重试：{(selectedRun.trace as any).policy?.maxRetries ?? 1}</p>
                  <p>重试间隔：{(selectedRun.trace as any).policy?.retryDelaySeconds ?? 3}s</p>
                  <p>失败队列：{(selectedRun.trace as any).policy?.queueRetryOnFailure ? '开启' : '关闭'}</p>
                  <p>失败窗口：{(selectedRun.trace as any).policy?.retryWindowMinutes ?? 60} 分钟</p>
                </div>
              </div>
            )}
            {((selectedRun.trace as any)?.attempts || []).length > 0 && (
              <div className="mt-3 rounded-md border border-slate-200 bg-white p-3">
                <p className="text-xs font-semibold text-slate-700">执行尝试</p>
                <div className="mt-1 space-y-1 text-xs text-slate-600">
                  {((selectedRun.trace as any).attempts || []).map((attempt: any, idx: number) => (
                    <p key={`${selectedRun.id}-attempt-${idx}`}>
                      第 {attempt.attempt || idx + 1} 次：{attempt.status || '-'}，耗时 {attempt.durationMs ?? '-'}ms
                      {attempt.error ? `，错误：${attempt.error}` : ''}
                    </p>
                  ))}
                </div>
              </div>
            )}
            {(selectedRun.trace as any)?.retryQueue && (
              <div className="mt-3 rounded-md border border-slate-200 bg-white p-3">
                <p className="text-xs font-semibold text-slate-700">失败重试队列</p>
                <div className="mt-1 grid gap-1 text-xs text-slate-600 md:grid-cols-2">
                  <p>是否入队：{(selectedRun.trace as any).retryQueue?.queued ? '是' : '否'}</p>
                  <p>重试时间：{formatTime((selectedRun.trace as any).retryQueue?.retryAt)}</p>
                  <p>当前失败周期：{(selectedRun.trace as any).retryQueue?.currentCycle ?? '-'}</p>
                  <p>最大失败周期：{(selectedRun.trace as any).retryQueue?.maxFailureCycles ?? '-'}</p>
                  {(selectedRun.trace as any).retryQueue?.reason && (
                    <p className="md:col-span-2">原因：{(selectedRun.trace as any).retryQueue?.reason}</p>
                  )}
                </div>
              </div>
            )}
            <div className="mt-4 rounded-md border border-slate-200 bg-slate-50 p-3">
              <p className="text-xs font-semibold text-slate-700">摘要</p>
              <p className="mt-1 whitespace-pre-line text-sm text-slate-700">{selectedRun.summary || '无摘要'}</p>
            </div>
            <div className="mt-3 rounded-md border border-slate-200 bg-white p-3">
              <p className="text-xs font-semibold text-slate-700">最终答案</p>
              <p className="mt-1 whitespace-pre-line text-sm text-slate-700">{selectedRun.finalAnswer || '无最终答案'}</p>
            </div>
            {(selectedRun.trace?.nextActions || []).length > 0 && (
              <div className="mt-3 rounded-md border border-slate-200 bg-white p-3">
                <p className="text-xs font-semibold text-slate-700">下一步建议</p>
                <ul className="mt-1 list-disc pl-5 text-sm text-slate-700">
                  {selectedRun.trace?.nextActions?.map((action, idx) => (
                    <li key={`${selectedRun.id}-next-${idx}`}>{action}</li>
                  ))}
                </ul>
              </div>
            )}
            {(selectedRun.trace?.evidence || []).length > 0 && (
              <div className="mt-3 rounded-md border border-slate-200 bg-white p-3">
                <p className="text-xs font-semibold text-slate-700">证据要点</p>
                <ul className="mt-1 list-disc pl-5 text-sm text-slate-700">
                  {selectedRun.trace?.evidence?.map((evidence, idx) => (
                    <li key={`${selectedRun.id}-evidence-${idx}`}>{evidence}</li>
                  ))}
                </ul>
              </div>
            )}
            {(selectedRun.trace?.steps || []).length > 0 && (
              <div className="mt-3 space-y-2 rounded-md border border-slate-200 bg-slate-50 p-3">
                <p className="text-xs font-semibold text-slate-700">多 Agent 协商过程</p>
                {selectedRun.trace?.steps?.map((step, index) => (
                  <div key={`${selectedRun.id}-step-${index}`} className="rounded border border-slate-200 bg-white p-2">
                    <p className="text-xs text-slate-500">
                      {step.agent} · {step.purpose} · {step.durationMs}ms
                    </p>
                    <p className="mt-1 whitespace-pre-line text-sm text-slate-700">{step.output}</p>
                  </div>
                ))}
              </div>
            )}
            {selectedRun.errorMessage && (
              <div className="mt-3 rounded-md border border-red-200 bg-red-50 p-3 text-sm text-red-700">
                错误信息：{selectedRun.errorMessage}
              </div>
            )}
          </div>
        </div>
      )}
    </div>
  );
};

export default WorkspacePage;
