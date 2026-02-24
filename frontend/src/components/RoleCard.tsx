import type { FC } from 'react';
import type { Role } from '../types';

interface RoleCardProps {
  role: Role;
  onClick?: () => void;
  onUse?: () => void;
}

export const RoleCard: FC<RoleCardProps> = ({ role, onClick, onUse }) => {
  return (
    <div 
      className="bg-white rounded-xl p-6 shadow-md hover:shadow-lg transition-all duration-200 hover:-translate-y-1 cursor-pointer group"
      onClick={onClick}
    >
      <div className="flex items-start gap-4">
        <div className="w-16 h-16 rounded-full bg-gradient-to-br from-primary to-primary-dark flex items-center justify-center text-white text-2xl font-bold flex-shrink-0">
          {role.avatar || role.name.charAt(0)}
        </div>
        <div className="flex-1 min-w-0">
          <div className="flex items-center justify-between">
            <h3 className="font-semibold text-slate-900 truncate">{role.name}</h3>
            {role.isTemplate && (
              <span className="text-xs px-2 py-0.5 bg-primary/10 text-primary rounded-full">
                模板
              </span>
            )}
          </div>
          <span className="text-xs text-slate-500 mt-1 inline-block">{role.category}</span>
          <p className="text-sm text-slate-600 mt-2 line-clamp-2">{role.description}</p>
        </div>
      </div>
      
      {role.skills && role.skills.length > 0 && (
        <div className="flex flex-wrap gap-2 mt-4">
          {role.skills.slice(0, 3).map(skill => (
            <span key={skill.id} className="text-xs px-2 py-1 bg-slate-100 text-slate-600 rounded-full">
              {skill.name}
            </span>
          ))}
        </div>
      )}
      
      <div className="flex gap-2 mt-4 opacity-0 group-hover:opacity-100 transition-opacity">
        <button 
          className="flex-1 py-2 text-sm text-slate-600 hover:bg-slate-100 rounded-lg transition-colors"
          onClick={(e) => { e.stopPropagation(); }}
        >
          预览
        </button>
        <button 
          className="flex-1 py-2 text-sm bg-primary text-white rounded-lg hover:bg-primary-dark transition-colors"
          onClick={(e) => { e.stopPropagation(); onUse?.(); }}
        >
          使用
        </button>
      </div>
    </div>
  );
};