import { useEffect, useMemo, useState } from 'react';
import companyApi, { type Company } from '../api/company';
import roleApi, { type Role } from '../api/role';
import workApi, { type WorkspaceTask } from '../api/work';

const formatTime = (value?: string) => {
  if (!value) return '未设置';
  const date = new Date(value);
  if (Number.isNaN(date.getTime())) return value;
  return date.toLocaleString('zh-CN', { hour12: false });
};

const triggerValueHint = (triggerType: string) => {
  if (triggerType === 'daily') return '例如 09:00';
  if (triggerType === 'interval_hours') return '例如 4（每 4 小时）';
  if (triggerType === 'once') return '例如 2026-03-01T09:00:00+08:00';
  return 'manual 模式可留空';
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

  const [filterCompanyId, setFilterCompanyId] = useState('');
  const [filterAsyncStatus, setFilterAsyncStatus] = useState('');

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
      await load();
    } catch (err: any) {
      setError(err?.message || '创建工作区任务失败');
    }
  };

  const runWorkspaceTask = async (id: string) => {
    try {
      await workApi.run(id);
      await load();
    } catch (err: any) {
      setError(err?.message || '执行失败');
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

        <div className="space-y-3">
          {rows.map((item) => (
            <div key={item.id} className="rounded-lg border border-slate-200 p-4">
              <div className="flex flex-wrap items-center justify-between gap-3">
                <div>
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
                    className="rounded-lg bg-slate-900 px-3 py-1.5 text-sm text-white hover:bg-slate-800"
                  >
                    立即执行
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
              </div>
              {item.resultSummary && (
                <div className="mt-3 rounded-md bg-slate-50 px-3 py-2 text-sm text-slate-700">
                  结果摘要：{item.resultSummary}
                </div>
              )}
            </div>
          ))}
          {!loading && rows.length === 0 && (
            <p className="text-sm text-slate-500">暂无工作区任务，先创建一个定时或异步任务。</p>
          )}
        </div>
      </div>
    </div>
  );
};

export default WorkspacePage;
