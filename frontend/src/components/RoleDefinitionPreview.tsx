import type { FC } from 'react';
import type { Role } from '../types';

interface RoleDefinitionPreviewProps {
  role: Role | null;
  onClose: () => void;
  onUse?: () => void;
  onEdit?: () => void;
}

export const RoleDefinitionPreview: FC<RoleDefinitionPreviewProps> = ({ role, onClose, onUse, onEdit }) => {
  if (!role) return null;

  const modelConfigText = role.modelConfig ? JSON.stringify(role.modelConfig, null, 2) : '';

  return (
    <div
      className="fixed inset-0 z-50 flex items-center justify-center bg-black/40 px-4"
      onClick={onClose}
    >
      <div
        className="max-h-[85vh] w-full max-w-3xl overflow-auto rounded-xl bg-white p-6 shadow-2xl"
        onClick={(e) => e.stopPropagation()}
      >
        <div className="mb-4 flex items-start justify-between">
          <div>
            <h3 className="text-xl font-bold text-slate-900">{role.name}</h3>
            <p className="mt-1 text-sm text-slate-500">{role.category || '通用'}</p>
          </div>
          <button
            className="rounded-lg px-3 py-1 text-sm text-slate-500 hover:bg-slate-100"
            onClick={onClose}
          >
            关闭
          </button>
        </div>

        <div className="space-y-4">
          <section>
            <h4 className="mb-1 text-sm font-semibold text-slate-700">角色描述</h4>
            <p className="rounded-lg bg-slate-50 p-3 text-sm text-slate-700">
              {role.description || '暂无描述'}
            </p>
          </section>

          <section>
            <h4 className="mb-1 text-sm font-semibold text-slate-700">系统提示词</h4>
            <pre className="overflow-auto whitespace-pre-wrap rounded-lg bg-slate-900 p-3 text-xs text-slate-100">
              {role.systemPrompt || '暂无系统提示词'}
            </pre>
          </section>

          <section>
            <h4 className="mb-1 text-sm font-semibold text-slate-700">欢迎语</h4>
            <p className="rounded-lg bg-slate-50 p-3 text-sm text-slate-700">
              {role.welcomeMessage || '暂无欢迎语'}
            </p>
          </section>

          {modelConfigText && (
            <section>
              <h4 className="mb-1 text-sm font-semibold text-slate-700">模型配置</h4>
              <pre className="overflow-auto whitespace-pre-wrap rounded-lg bg-slate-900 p-3 text-xs text-slate-100">
                {modelConfigText}
              </pre>
            </section>
          )}
        </div>

        <div className="mt-6 flex justify-end gap-2">
          <button
            className="rounded-lg border border-slate-300 px-4 py-2 text-sm text-slate-700 hover:bg-slate-50"
            onClick={onClose}
          >
            取消
          </button>
          {onEdit && (
            <button
              className="rounded-lg border border-amber-300 px-4 py-2 text-sm text-amber-700 hover:bg-amber-50"
              onClick={onEdit}
            >
              编辑角色
            </button>
          )}
          {onUse && (
            <button
              className="rounded-lg bg-primary px-4 py-2 text-sm text-white hover:bg-primary-dark"
              onClick={onUse}
            >
              使用这个角色
            </button>
          )}
        </div>
      </div>
    </div>
  );
};

export default RoleDefinitionPreview;
