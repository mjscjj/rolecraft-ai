import { useEffect, useState } from 'react';
import companyApi, {
  type Company,
  type CompanyDelivery,
  type CompanyExportFormat,
  type CompanyExportRecord,
  type CompanyOutcome,
  type CompanyStats,
} from '../api/company';
import roleApi, { type Role } from '../api/role';

const formatTime = (value?: string) => {
  if (!value) return '-';
  const date = new Date(value);
  if (Number.isNaN(date.getTime())) return value;
  return date.toLocaleString('zh-CN', { hour12: false });
};

export const CompanyPage = () => {
  const [rows, setRows] = useState<Company[]>([]);
  const [companyDetails, setCompanyDetails] = useState<
    Record<string, { stats: CompanyStats; recentOutcomes: CompanyOutcome[]; deliveryBoard: CompanyDelivery[] }>
  >({});
  const [companyRoles, setCompanyRoles] = useState<Role[]>([]);
  const [selectedCompanyId, setSelectedCompanyId] = useState('');
  const [name, setName] = useState('');
  const [description, setDescription] = useState('');
  const [roleName, setRoleName] = useState('');
  const [roleDescription, setRoleDescription] = useState('');
  const [roleCategory, setRoleCategory] = useState('通用');
  const [roleSystemPrompt, setRoleSystemPrompt] = useState('');
  const [roleWelcomeMessage, setRoleWelcomeMessage] = useState('');
  const [roleModelConfigText, setRoleModelConfigText] = useState('{\n  "temperature": 0.7\n}');
  const [editingRoleId, setEditingRoleId] = useState('');
  const [selectedDelivery, setSelectedDelivery] = useState<CompanyDelivery | null>(null);
  const [deliveryQuery, setDeliveryQuery] = useState('');
  const [minConfidence, setMinConfidence] = useState('0');
  const [deliveryFrom, setDeliveryFrom] = useState('');
  const [deliveryTo, setDeliveryTo] = useState('');
  const [exportHistory, setExportHistory] = useState<CompanyExportRecord[]>([]);
  const [exportLoadingKey, setExportLoadingKey] = useState('');
  const [historyDownloadingId, setHistoryDownloadingId] = useState('');
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState('');

  const load = async () => {
    setLoading(true);
    setError('');
    try {
      setRows(await companyApi.list());
    } catch (err: any) {
      setError(err?.message || '加载公司失败');
    } finally {
      setLoading(false);
    }
  };

  const loadCompanyDetails = async (companyIds: string[]) => {
    if (companyIds.length === 0) {
      setCompanyDetails({});
      return;
    }
    try {
      const detailRows = await Promise.all(companyIds.map((id) => companyApi.get(id)));
      const next: Record<string, { stats: CompanyStats; recentOutcomes: CompanyOutcome[]; deliveryBoard: CompanyDelivery[] }> = {};
      detailRows.forEach((item) => {
        next[item.company.id] = {
          stats: item.stats,
          recentOutcomes: item.recentOutcomes || [],
          deliveryBoard: item.deliveryBoard || [],
        };
      });
      setCompanyDetails(next);
    } catch {
      // ignore detail failure, primary list remains available
    }
  };

  const loadExportHistory = async (companyIds: string[]) => {
    if (companyIds.length === 0) {
      setExportHistory([]);
      return;
    }
    try {
      const rows = await Promise.all(
        companyIds.map(async (id) => {
          try {
            return await companyApi.listExports(id, 20);
          } catch {
            return [] as CompanyExportRecord[];
          }
        })
      );
      const merged = rows
        .flat()
        .sort((a, b) => new Date(b.createdAt).getTime() - new Date(a.createdAt).getTime())
        .slice(0, 100);
      setExportHistory(merged);
    } catch {
      // ignore export history load failures
    }
  };

  useEffect(() => {
    load();
  }, []);

  useEffect(() => {
    const ids = rows.map((item) => item.id);
    loadCompanyDetails(ids);
    loadExportHistory(ids);
  }, [rows]);

  const loadCompanyRoles = async (companyId: string) => {
    if (!companyId) {
      setCompanyRoles([]);
      return;
    }
    try {
      const roles = await roleApi.list({ companyId });
      setCompanyRoles(roles);
    } catch (err: any) {
      setError(err?.message || '加载公司角色失败');
    }
  };

  useEffect(() => {
    if (!selectedCompanyId && rows.length > 0) {
      setSelectedCompanyId(rows[0].id);
      return;
    }
    loadCompanyRoles(selectedCompanyId);
  }, [rows, selectedCompanyId]);

  const createCompany = async () => {
    if (!name.trim()) return;
    try {
      await companyApi.create({ name: name.trim(), description });
      setName('');
      setDescription('');
      await load();
    } catch (err: any) {
      setError(err?.message || '创建公司失败');
    }
  };

  const resetRoleForm = () => {
    setEditingRoleId('');
    setRoleName('');
    setRoleDescription('');
    setRoleCategory('通用');
    setRoleSystemPrompt('');
    setRoleWelcomeMessage('');
    setRoleModelConfigText('{\n  "temperature": 0.7\n}');
  };

  const createOrUpdateRole = async () => {
    if (!selectedCompanyId) {
      setError('请先选择公司');
      return;
    }
    if (!roleName.trim() || !roleSystemPrompt.trim()) {
      setError('角色名称和系统提示词必填');
      return;
    }
    try {
      let modelConfig: Record<string, any> | undefined = undefined;
      const rawConfig = roleModelConfigText.trim();
      if (rawConfig) {
        modelConfig = JSON.parse(rawConfig);
      }
      const payload = {
        name: roleName.trim(),
        description: roleDescription.trim(),
        category: roleCategory.trim() || '通用',
        companyId: selectedCompanyId,
        systemPrompt: roleSystemPrompt.trim(),
        welcomeMessage: roleWelcomeMessage.trim(),
        modelConfig,
      };
      if (editingRoleId) {
        await roleApi.update(editingRoleId, payload);
      } else {
        await roleApi.create(payload);
      }
      resetRoleForm();
      await loadCompanyRoles(selectedCompanyId);
      setError('');
    } catch (err: any) {
      if (err instanceof SyntaxError) {
        setError('modelConfig JSON 格式错误，请修正后再保存');
        return;
      }
      setError(err?.message || '保存角色失败');
    }
  };

  const startEditRole = (role: Role) => {
    setEditingRoleId(role.id);
    setRoleName(role.name || '');
    setRoleDescription(role.description || '');
    setRoleCategory(role.category || '通用');
    setRoleSystemPrompt(role.systemPrompt || '');
    setRoleWelcomeMessage(role.welcomeMessage || '');
    setRoleModelConfigText(
      role.modelConfig ? JSON.stringify(role.modelConfig, null, 2) : '{\n  "temperature": 0.7\n}'
    );
  };

  const deleteRole = async (role: Role) => {
    const confirmed = window.confirm(`确认删除角色「${role.name}」吗？`);
    if (!confirmed) return;
    try {
      await roleApi.delete(role.id);
      if (editingRoleId === role.id) {
        resetRoleForm();
      }
      await loadCompanyRoles(selectedCompanyId);
      setError('');
    } catch (err: any) {
      setError(err?.message || '删除角色失败');
    }
  };

  const downloadContent = (content: string, fileName: string, format: CompanyExportFormat | string) => {
    const mimeType =
      format === 'json' ? 'application/json;charset=utf-8' : 'text/markdown;charset=utf-8';
    const blob = new Blob([content], { type: mimeType });
    const url = URL.createObjectURL(blob);
    const link = document.createElement('a');
    link.href = url;
    link.download = fileName;
    document.body.appendChild(link);
    link.click();
    document.body.removeChild(link);
    URL.revokeObjectURL(url);
  };

  const exportDeliveryBoard = async (company: Company, format: CompanyExportFormat) => {
    try {
      setExportLoadingKey(`${company.id}:${format}`);
      const threshold = Number(minConfidence || '0');
      const payload = {
        format,
        keyword: deliveryQuery.trim() || undefined,
        minConfidence: Number.isNaN(threshold) ? undefined : threshold,
        from: deliveryFrom || undefined,
        to: deliveryTo || undefined,
      };
      const record = await companyApi.createExport(company.id, payload);
      downloadContent(record.content, record.fileName, record.format);
      setExportHistory((prev) =>
        [record, ...prev.filter((item) => item.id !== record.id)]
          .sort((a, b) => new Date(b.createdAt).getTime() - new Date(a.createdAt).getTime())
          .slice(0, 100)
      );
      setError('');
    } catch (err: any) {
      setError(err?.message || '导出失败');
    } finally {
      setExportLoadingKey('');
    }
  };

  const downloadExportRecord = async (item: CompanyExportRecord) => {
    try {
      setHistoryDownloadingId(item.id);
      const detail = await companyApi.getExport(item.companyId, item.id);
      downloadContent(detail.content, detail.fileName, detail.format);
      setError('');
    } catch (err: any) {
      setError(err?.message || '下载导出记录失败');
    } finally {
      setHistoryDownloadingId('');
    }
  };

  const filterDeliveries = (deliveries: CompanyDelivery[]) => {
    const keyword = deliveryQuery.trim().toLowerCase();
    const threshold = Number(minConfidence || '0');
    const parsedFrom = deliveryFrom ? new Date(deliveryFrom) : null;
    const parsedTo = deliveryTo ? new Date(`${deliveryTo}T23:59:59`) : null;
    const from = parsedFrom && !Number.isNaN(parsedFrom.getTime()) ? parsedFrom : null;
    const to = parsedTo && !Number.isNaN(parsedTo.getTime()) ? parsedTo : null;
    return deliveries.filter((item) => {
      const byQuery =
        keyword === '' ||
        (item.workName || '').toLowerCase().includes(keyword) ||
        (item.summary || '').toLowerCase().includes(keyword) ||
        (item.finalAnswer || '').toLowerCase().includes(keyword);
      const byConfidence = Number.isNaN(threshold) ? true : (item.confidence || 0) >= threshold;
      const updatedAt = new Date(item.updatedAt);
      const byTimeRange =
        (!from || updatedAt >= from) &&
        (!to || updatedAt <= to);
      return byQuery && byConfidence && byTimeRange;
    });
  };

  const refreshExportHistory = async () => {
    await loadExportHistory(rows.map((item) => item.id));
  };

  return (
    <div className="space-y-6">
      <div>
        <h1 className="text-2xl font-bold text-slate-900">我的公司</h1>
        <p className="mt-1 text-slate-500">公司是总交付区：集中管理角色、项目产出和知识成果。</p>
      </div>

      <div className="rounded-xl border border-slate-200 bg-white p-4">
        <h2 className="mb-3 text-base font-semibold text-slate-800">创建公司</h2>
        <div className="grid gap-3 md:grid-cols-2">
          <input
            value={name}
            onChange={(e) => setName(e.target.value)}
            className="rounded-lg border border-slate-300 px-3 py-2"
            placeholder="公司名称"
          />
          <input
            value={description}
            onChange={(e) => setDescription(e.target.value)}
            className="rounded-lg border border-slate-300 px-3 py-2"
            placeholder="公司描述（可选）"
          />
        </div>
        <button
          onClick={createCompany}
          className="mt-3 rounded-lg bg-slate-900 px-4 py-2 text-white hover:bg-slate-800"
        >
          创建
        </button>
      </div>

      {loading && <div className="text-slate-500">加载中...</div>}
      {error && <div className="rounded-lg bg-red-50 px-3 py-2 text-sm text-red-600">{error}</div>}

      <div className="rounded-xl border border-slate-200 bg-white p-4">
        <h2 className="mb-3 text-base font-semibold text-slate-800">交付筛选</h2>
        <div className="grid gap-2 md:grid-cols-2">
          <input
            value={deliveryQuery}
            onChange={(e) => setDeliveryQuery(e.target.value)}
            className="rounded-lg border border-slate-300 px-3 py-2"
            placeholder="按任务名、摘要、最终答案搜索"
          />
          <input
            value={minConfidence}
            onChange={(e) => setMinConfidence(e.target.value)}
            className="rounded-lg border border-slate-300 px-3 py-2"
            placeholder="最低置信度（0-1）"
          />
          <input
            type="date"
            value={deliveryFrom}
            onChange={(e) => setDeliveryFrom(e.target.value)}
            className="rounded-lg border border-slate-300 px-3 py-2"
          />
          <input
            type="date"
            value={deliveryTo}
            onChange={(e) => setDeliveryTo(e.target.value)}
            className="rounded-lg border border-slate-300 px-3 py-2"
          />
        </div>
      </div>

      <div className="grid gap-4 md:grid-cols-2">
        {rows.map((item) => (
          <div key={item.id} className="rounded-xl border border-slate-200 bg-white p-4">
            <p className="text-lg font-semibold text-slate-900">{item.name}</p>
            <p className="mt-1 text-sm text-slate-500">{item.description || '暂无描述'}</p>
            <div className="mt-3 grid grid-cols-3 gap-2 text-xs text-slate-500">
              <div className="rounded-md bg-slate-50 px-2 py-1">
                角色 {companyDetails[item.id]?.stats?.roleCount ?? 0}
              </div>
              <div className="rounded-md bg-slate-50 px-2 py-1">
                工作区 {companyDetails[item.id]?.stats?.workspaceCount ?? 0}
              </div>
              <div className="rounded-md bg-slate-50 px-2 py-1">
                成果 {companyDetails[item.id]?.stats?.outcomeCount ?? 0}
              </div>
            </div>
            {(companyDetails[item.id]?.recentOutcomes || []).length > 0 && (
              <div className="mt-3 rounded-lg border border-slate-200 bg-slate-50 p-2">
                <p className="text-xs font-medium text-slate-700">最近交付成果</p>
                <div className="mt-2 space-y-1">
                  {companyDetails[item.id].recentOutcomes.slice(0, 3).map((outcome) => (
                    <p key={outcome.id} className="text-xs text-slate-600">
                      {(outcome.workName || `任务 ${outcome.workId?.slice(0, 8) || '-'}`)}（置信度 {outcome.confidence?.toFixed?.(2) || '0.00'}）：{outcome.resultSummary}
                    </p>
                  ))}
                </div>
              </div>
            )}
            {filterDeliveries(companyDetails[item.id]?.deliveryBoard || []).length > 0 && (
              <div className="mt-3 rounded-lg border border-slate-200 bg-white p-2">
                <div className="flex items-center justify-between gap-2">
                  <p className="text-xs font-medium text-slate-700">结构化交付看板</p>
                  <div className="flex gap-1">
                    <button
                      onClick={() => exportDeliveryBoard(item, 'markdown')}
                      disabled={exportLoadingKey === `${item.id}:markdown` || exportLoadingKey === `${item.id}:json`}
                      className="rounded border border-slate-300 px-2 py-1 text-[11px] text-slate-600 hover:bg-slate-50"
                    >
                      {exportLoadingKey === `${item.id}:markdown` ? '导出中...' : '导出 MD'}
                    </button>
                    <button
                      onClick={() => exportDeliveryBoard(item, 'json')}
                      disabled={exportLoadingKey === `${item.id}:markdown` || exportLoadingKey === `${item.id}:json`}
                      className="rounded border border-slate-300 px-2 py-1 text-[11px] text-slate-600 hover:bg-slate-50"
                    >
                      {exportLoadingKey === `${item.id}:json` ? '导出中...' : '导出 JSON'}
                    </button>
                  </div>
                </div>
                <div className="mt-2 space-y-2">
                  {filterDeliveries(companyDetails[item.id].deliveryBoard).slice(0, 3).map((delivery) => (
                    <div key={delivery.id} className="rounded border border-slate-200 bg-slate-50 p-2">
                      <p className="text-xs font-medium text-slate-700">
                        {delivery.workName || `任务 ${delivery.workId?.slice(0, 8) || '-'}`} · 步骤 {delivery.stepCount}
                      </p>
                      <p className="mt-1 text-xs text-slate-600">{delivery.summary || '暂无摘要'}</p>
                      {(delivery.nextActions || []).length > 0 && (
                        <p className="mt-1 text-xs text-slate-500">下一步：{delivery.nextActions.slice(0, 2).join('；')}</p>
                      )}
                      <div className="mt-1">
                        <button
                          onClick={() => setSelectedDelivery(delivery)}
                          className="rounded border border-slate-300 px-2 py-1 text-[11px] text-slate-600 hover:bg-slate-100"
                        >
                          查看详情
                        </button>
                      </div>
                    </div>
                  ))}
                </div>
              </div>
            )}
            <p className="mt-2 text-xs text-slate-400">ID: {item.id}</p>
          </div>
        ))}
      </div>

      <div className="rounded-xl border border-slate-200 bg-white p-4">
        <h2 className="mb-3 text-base font-semibold text-slate-800">公司职能角色管理</h2>
        <div className="mb-3">
          <select
            value={selectedCompanyId}
            onChange={(e) => setSelectedCompanyId(e.target.value)}
            className="w-full rounded-lg border border-slate-300 px-3 py-2 md:w-96"
          >
            <option value="">请选择公司</option>
            {rows.map((item) => (
              <option key={item.id} value={item.id}>
                {item.name}
              </option>
            ))}
          </select>
        </div>

        <div className="grid gap-3 md:grid-cols-2">
          <input
            value={roleName}
            onChange={(e) => setRoleName(e.target.value)}
            className="rounded-lg border border-slate-300 px-3 py-2"
            placeholder="角色名称"
          />
          <input
            value={roleCategory}
            onChange={(e) => setRoleCategory(e.target.value)}
            className="rounded-lg border border-slate-300 px-3 py-2"
            placeholder="角色分类"
          />
          <input
            value={roleDescription}
            onChange={(e) => setRoleDescription(e.target.value)}
            className="rounded-lg border border-slate-300 px-3 py-2 md:col-span-2"
            placeholder="角色描述"
          />
          <textarea
            value={roleSystemPrompt}
            onChange={(e) => setRoleSystemPrompt(e.target.value)}
            className="min-h-28 rounded-lg border border-slate-300 px-3 py-2 md:col-span-2"
            placeholder="系统提示词（核心自定义）"
          />
          <input
            value={roleWelcomeMessage}
            onChange={(e) => setRoleWelcomeMessage(e.target.value)}
            className="rounded-lg border border-slate-300 px-3 py-2 md:col-span-2"
            placeholder="欢迎语"
          />
          <textarea
            value={roleModelConfigText}
            onChange={(e) => setRoleModelConfigText(e.target.value)}
            className="min-h-28 rounded-lg border border-slate-300 px-3 py-2 font-mono text-xs md:col-span-2"
            placeholder='modelConfig JSON，例如 {"temperature":0.7}'
          />
        </div>
        <div className="mt-3 flex gap-2">
          <button
            onClick={createOrUpdateRole}
            className="rounded-lg bg-slate-900 px-4 py-2 text-white hover:bg-slate-800"
          >
            {editingRoleId ? '保存角色修改' : '创建公司角色'}
          </button>
          {editingRoleId && (
            <button
              onClick={resetRoleForm}
              className="rounded-lg border border-slate-300 px-4 py-2 text-slate-700 hover:bg-slate-50"
            >
              取消编辑
            </button>
          )}
        </div>
      </div>

      <div className="rounded-xl border border-slate-200 bg-white p-4">
        <h3 className="mb-3 text-base font-semibold text-slate-800">当前公司角色</h3>
        <div className="space-y-3">
          {companyRoles.map((role) => (
            <div
              key={role.id}
              className="flex items-center justify-between rounded-lg border border-slate-200 px-3 py-2"
            >
              <div>
                <p className="font-medium text-slate-900">{role.name}</p>
                <p className="text-xs text-slate-500">{role.category || '通用'}</p>
              </div>
              <button
                onClick={() => startEditRole(role)}
                className="rounded-md border border-slate-300 px-3 py-1.5 text-sm text-slate-700 hover:bg-slate-50"
              >
                编辑自定义
              </button>
              <button
                onClick={() => deleteRole(role)}
                className="ml-2 rounded-md border border-red-300 px-3 py-1.5 text-sm text-red-600 hover:bg-red-50"
              >
                删除
              </button>
            </div>
          ))}
          {selectedCompanyId && companyRoles.length === 0 && (
            <p className="text-sm text-slate-500">该公司还没有角色，先创建一个职能角色。</p>
          )}
        </div>
      </div>

      <div className="rounded-xl border border-slate-200 bg-white p-4">
        <div className="mb-3 flex items-center justify-between gap-2">
          <h3 className="text-base font-semibold text-slate-800">导出历史</h3>
          <button
            onClick={refreshExportHistory}
            className="rounded border border-slate-300 px-3 py-1.5 text-xs text-slate-600 hover:bg-slate-50"
          >
            刷新历史
          </button>
        </div>
        <div className="space-y-2">
          {exportHistory.map((item) => (
            <div key={item.id} className="rounded border border-slate-200 px-3 py-2 text-xs text-slate-600">
              <p>
                {item.companyName} · {item.format.toUpperCase()} · {item.deliveryCount} 条
              </p>
              <p className="mt-1 text-slate-500">{item.fileName}</p>
              <p className="mt-1 text-slate-400">{formatTime(item.createdAt)}</p>
              <div className="mt-1">
                <button
                  onClick={() => downloadExportRecord(item)}
                  className="rounded border border-slate-300 px-2 py-1 text-[11px] text-slate-600 hover:bg-slate-50"
                >
                  {historyDownloadingId === item.id ? '下载中...' : '下载'}
                </button>
              </div>
            </div>
          ))}
          {exportHistory.length === 0 && (
            <p className="text-sm text-slate-500">暂无导出记录。</p>
          )}
        </div>
      </div>

      {selectedDelivery && (
        <div className="fixed inset-0 z-50 flex items-center justify-center bg-slate-900/40 p-4">
          <div className="max-h-[85vh] w-full max-w-3xl overflow-auto rounded-xl bg-white p-5 shadow-xl">
            <div className="mb-3 flex items-center justify-between">
              <h3 className="text-lg font-semibold text-slate-900">
                {selectedDelivery.workName || `任务 ${selectedDelivery.workId.slice(0, 8)}`} 交付详情
              </h3>
              <button
                onClick={() => setSelectedDelivery(null)}
                className="rounded border border-slate-300 px-2 py-1 text-sm text-slate-600 hover:bg-slate-50"
              >
                关闭
              </button>
            </div>
            <div className="grid gap-2 text-sm text-slate-600 md:grid-cols-2">
              <p>交付ID：{selectedDelivery.id}</p>
              <p>更新时间：{formatTime(selectedDelivery.updatedAt)}</p>
              <p>置信度：{selectedDelivery.confidence?.toFixed?.(2) || '0.00'}</p>
              <p>步骤数：{selectedDelivery.stepCount || 0}</p>
            </div>
            <div className="mt-3 rounded border border-slate-200 bg-slate-50 p-3">
              <p className="text-xs font-semibold text-slate-700">摘要</p>
              <p className="mt-1 whitespace-pre-line text-sm text-slate-700">{selectedDelivery.summary || '暂无摘要'}</p>
            </div>
            <div className="mt-3 rounded border border-slate-200 bg-white p-3">
              <p className="text-xs font-semibold text-slate-700">最终答案</p>
              <p className="mt-1 whitespace-pre-line text-sm text-slate-700">{selectedDelivery.finalAnswer || '暂无最终答案'}</p>
            </div>
            {(selectedDelivery.nextActions || []).length > 0 && (
              <div className="mt-3 rounded border border-slate-200 bg-white p-3">
                <p className="text-xs font-semibold text-slate-700">下一步建议</p>
                <ul className="mt-1 list-disc pl-5 text-sm text-slate-700">
                  {selectedDelivery.nextActions.map((item, idx) => (
                    <li key={`${selectedDelivery.id}-next-${idx}`}>{item}</li>
                  ))}
                </ul>
              </div>
            )}
            {(selectedDelivery.evidence || []).length > 0 && (
              <div className="mt-3 rounded border border-slate-200 bg-white p-3">
                <p className="text-xs font-semibold text-slate-700">证据要点</p>
                <ul className="mt-1 list-disc pl-5 text-sm text-slate-700">
                  {selectedDelivery.evidence.map((item, idx) => (
                    <li key={`${selectedDelivery.id}-evidence-${idx}`}>{item}</li>
                  ))}
                </ul>
              </div>
            )}
          </div>
        </div>
      )}
    </div>
  );
};

export default CompanyPage;
