import type { FC } from 'react'; import { useState } from 'react';
import { useNavigate } from 'react-router-dom';
import { 
  ChevronRight, 
  ChevronLeft, 
  User, 
  MessageSquare, 
  Zap, 
  FileText, 
  Play,
  Save,
  Upload
} from 'lucide-react';
import client from '../api/client';

const steps = [
  { id: 1, label: '基础信息', icon: User },
  { id: 2, label: '提示词配置', icon: MessageSquare },
  { id: 3, label: '技能配置', icon: Zap },
  { id: 4, label: '知识绑定', icon: FileText },
  { id: 5, label: '测试发布', icon: Play },
];

const skillsList = [
  { id: '1', name: '邮件撰写', description: '撰写专业的商务邮件' },
  { id: '2', name: '日程管理', description: '安排和管理日程' },
  { id: '3', name: '资料整理', description: '整理和归纳资料' },
  { id: '4', name: '文案撰写', description: '撰写营销文案' },
  { id: '5', name: '数据分析', description: '分析数据并生成报告' },
  { id: '6', name: '合同审核', description: '审查合同条款' },
];

const documentsList = [
  { id: '1', name: '产品手册.pdf', size: '2.5 MB', status: 'completed' },
  { id: '2', name: '公司制度.docx', size: '1.2 MB', status: 'completed' },
  { id: '3', name: '竞品分析.pdf', size: '5.1 MB', status: 'completed' },
];

export const RoleEditor: FC = () => {
  const navigate = useNavigate();
  const [currentStep, setCurrentStep] = useState(1);
  const [publishing, setPublishing] = useState(false);
  const [error, setError] = useState('');
  const [isPublic, setIsPublic] = useState(false);
  const [formData, setFormData] = useState({
    name: '',
    description: '',
    category: '通用',
    systemPrompt: '',
    welcomeMessage: '',
    temperature: 0.7,
  });
  const [selectedSkills, setSelectedSkills] = useState<string[]>([]);
  const [selectedDocs, setSelectedDocs] = useState<string[]>([]);

  const handleNext = () => {
    if (currentStep < 5) setCurrentStep(currentStep + 1);
  };

  const handlePrev = () => {
    if (currentStep > 1) setCurrentStep(currentStep - 1);
  };

  const toggleSkill = (skillId: string) => {
    setSelectedSkills(prev => 
      prev.includes(skillId) 
        ? prev.filter(id => id !== skillId)
        : [...prev, skillId]
    );
  };

  const toggleDoc = (docId: string) => {
    setSelectedDocs(prev => 
      prev.includes(docId) 
        ? prev.filter(id => id !== docId)
        : [...prev, docId]
    );
  };

  const handleSaveDraft = () => {
    localStorage.setItem('roleEditorDraft', JSON.stringify({ formData, selectedSkills, selectedDocs, isPublic }));
  };

  const handlePublish = async () => {
    if (!formData.name.trim() || !formData.systemPrompt.trim()) {
      setError('请至少填写角色名称和系统提示词');
      return;
    }

    setPublishing(true);
    setError('');

    try {
      const response = await client.post('/roles', {
        name: formData.name.trim(),
        description: formData.description.trim(),
        category: formData.category,
        systemPrompt: formData.systemPrompt.trim(),
        welcomeMessage: formData.welcomeMessage.trim(),
        isPublic,
        modelConfig: {
          temperature: formData.temperature,
          skills: selectedSkills,
          documents: selectedDocs,
        },
      });

      if (response.data.code === 200 || response.data.code === 0) {
        navigate(`/chat/${response.data.data.id}`);
      } else {
        setError('发布失败，请重试');
      }
    } catch (err: any) {
      setError(err.response?.data?.error || '发布失败，请重试');
    } finally {
      setPublishing(false);
    }
  };

  return (
    <div className="max-w-5xl mx-auto">
      {/* Header */}
      <div className="mb-8">
        <h1 className="text-2xl font-bold text-slate-900">创建新角色</h1>
        <p className="text-slate-500 mt-1">配置你的 AI 数字员工</p>
      </div>

      {/* Step Navigation */}
      <div className="mb-8">
        <div className="flex items-center justify-between">
          {steps.map((step, index) => (
            <div key={step.id} className="flex items-center">
              <div className={`flex flex-col items-center ${index < steps.length - 1 ? 'flex-1' : ''}`}>
                <div className={`w-10 h-10 rounded-full flex items-center justify-center font-semibold transition-colors ${
                  currentStep > step.id ? 'bg-primary text-white' :
                  currentStep === step.id ? 'bg-slate-900 text-white' :
                  'bg-slate-100 text-slate-400'
                }`}>
                  {currentStep > step.id ? (
                    <svg className="w-5 h-5" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                      <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M5 13l4 4L19 7" />
                    </svg>
                  ) : (
                    <step.icon className="w-5 h-5" />
                  )}
                </div>
                <span className={`text-xs mt-2 font-medium ${
                  currentStep >= step.id ? 'text-slate-900' : 'text-slate-400'
                }`}>
                  {step.label}
                </span>
              </div>
              {index < steps.length - 1 && (
                <div className={`w-full h-0.5 mx-4 ${
                  currentStep > step.id ? 'bg-primary' : 'bg-slate-200'
                }`} />
              )}
            </div>
          ))}
        </div>
      </div>

      {/* Step Content */}
      <div className="bg-white rounded-xl shadow-sm border border-slate-100 p-8">
        {/* Step 1: Basic Info */}
        {currentStep === 1 && (
          <div className="space-y-6">
            <h2 className="text-lg font-semibold text-slate-900">基础信息</h2>
            
            <div className="flex items-center gap-6">
              <div className="w-24 h-24 rounded-full bg-slate-100 flex items-center justify-center cursor-pointer hover:bg-slate-200 transition-colors border-2 border-dashed border-slate-300">
                <Upload className="w-8 h-8 text-slate-400" />
              </div>
              <div>
                <p className="font-medium text-slate-900">角色头像</p>
                <p className="text-sm text-slate-500 mt-1">建议尺寸 200x200px，支持 JPG、PNG</p>
              </div>
            </div>

            <div>
              <label className="block text-sm font-medium text-slate-700 mb-2">角色名称 *</label>
              <input
                type="text"
                value={formData.name}
                onChange={(e) => setFormData({ ...formData, name: e.target.value })}
                placeholder="例如：营销专家"
                className="w-full px-4 py-2.5 border border-slate-200 rounded-lg outline-none focus:border-primary focus:ring-2 focus:ring-primary/20 transition-all"
              />
            </div>

            <div>
              <label className="block text-sm font-medium text-slate-700 mb-2">角色分类</label>
              <select
                value={formData.category}
                onChange={(e) => setFormData({ ...formData, category: e.target.value })}
                className="w-full px-4 py-2.5 border border-slate-200 rounded-lg outline-none focus:border-primary focus:ring-2 focus:ring-primary/20 transition-all"
              >
                <option>通用</option>
                <option>营销</option>
                <option>法律</option>
                <option>财务</option>
                <option>技术</option>
                <option>人事</option>
                <option>行政</option>
              </select>
            </div>

            <div>
              <label className="block text-sm font-medium text-slate-700 mb-2">角色描述</label>
              <textarea
                value={formData.description}
                onChange={(e) => setFormData({ ...formData, description: e.target.value })}
                placeholder="简要描述这个角色的能力和用途..."
                rows={3}
                className="w-full px-4 py-2.5 border border-slate-200 rounded-lg outline-none focus:border-primary focus:ring-2 focus:ring-primary/20 transition-all resize-none"
              />
            </div>
          </div>
        )}

        {/* Step 2: Prompt Config */}
        {currentStep === 2 && (
          <div className="space-y-6">
            <h2 className="text-lg font-semibold text-slate-900">提示词配置</h2>
            
            <div>
              <label className="block text-sm font-medium text-slate-700 mb-2">系统提示词 *</label>
              <p className="text-xs text-slate-500 mb-2">定义角色的身份、性格和行为方式</p>
              <textarea
                value={formData.systemPrompt}
                onChange={(e) => setFormData({ ...formData, systemPrompt: e.target.value })}
                placeholder="你是一位专业的营销专家，擅长..."
                rows={8}
                className="w-full px-4 py-3 font-mono text-sm border border-slate-200 rounded-lg outline-none focus:border-primary focus:ring-2 focus:ring-primary/20 transition-all resize-none"
              />
            </div>

            <div>
              <label className="block text-sm font-medium text-slate-700 mb-2">欢迎语</label>
              <p className="text-xs text-slate-500 mb-2">用户首次与角色对话时显示</p>
              <input
                type="text"
                value={formData.welcomeMessage}
                onChange={(e) => setFormData({ ...formData, welcomeMessage: e.target.value })}
                placeholder="你好！我是你的营销助手..."
                className="w-full px-4 py-2.5 border border-slate-200 rounded-lg outline-none focus:border-primary focus:ring-2 focus:ring-primary/20 transition-all"
              />
            </div>

            <div>
              <label className="block text-sm font-medium text-slate-700 mb-2">创造性 / 确定性</label>
              <input
                type="range"
                min="0"
                max="1"
                step="0.1"
                value={formData.temperature}
                onChange={(e) => setFormData({ ...formData, temperature: parseFloat(e.target.value) })}
                className="w-full"
              />
              <div className="flex justify-between text-xs text-slate-500 mt-1">
                <span>精确</span>
                <span>平衡</span>
                <span>创造性</span>
              </div>
            </div>
          </div>
        )}

        {/* Step 3: Skills */}
        {currentStep === 3 && (
          <div className="space-y-6">
            <h2 className="text-lg font-semibold text-slate-900">技能配置</h2>
            <p className="text-slate-500">选择该角色具备的技能能力</p>
            
            <div className="grid grid-cols-2 gap-4">
              {skillsList.map(skill => (
                <div
                  key={skill.id}
                  onClick={() => toggleSkill(skill.id)}
                  className={`p-4 rounded-xl border-2 cursor-pointer transition-all ${
                    selectedSkills.includes(skill.id)
                      ? 'border-primary bg-primary/5'
                      : 'border-slate-200 hover:border-slate-300'
                  }`}
                >
                  <div className="flex items-start gap-3">
                    <div className={`w-5 h-5 rounded border-2 flex items-center justify-center mt-0.5 ${
                      selectedSkills.includes(skill.id)
                        ? 'bg-primary border-primary'
                        : 'border-slate-300'
                    }`}>
                      {selectedSkills.includes(skill.id) && (
                        <svg className="w-3 h-3 text-white" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                          <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={3} d="M5 13l4 4L19 7" />
                        </svg>
                      )}
                    </div>
                    <div>
                      <p className="font-medium text-slate-900">{skill.name}</p>
                      <p className="text-sm text-slate-500">{skill.description}</p>
                    </div>
                  </div>
                </div>
              ))}
            </div>
          </div>
        )}

        {/* Step 4: Knowledge */}
        {currentStep === 4 && (
          <div className="space-y-6">
            <h2 className="text-lg font-semibold text-slate-900">知识绑定</h2>
            <p className="text-slate-500">选择该角色可以访问的知识库文档</p>
            
            <div className="space-y-3">
              {documentsList.map(doc => (
                <div
                  key={doc.id}
                  onClick={() => toggleDoc(doc.id)}
                  className={`flex items-center gap-4 p-4 rounded-xl border-2 cursor-pointer transition-all ${
                    selectedDocs.includes(doc.id)
                      ? 'border-primary bg-primary/5'
                      : 'border-slate-200 hover:border-slate-300'
                  }`}
                >
                  <div className={`w-5 h-5 rounded border-2 flex items-center justify-center ${
                    selectedDocs.includes(doc.id)
                      ? 'bg-primary border-primary'
                      : 'border-slate-300'
                  }`}>
                    {selectedDocs.includes(doc.id) && (
                      <svg className="w-3 h-3 text-white" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                        <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={3} d="M5 13l4 4L19 7" />
                      </svg>
                    )}
                  </div>
                  <div className="w-10 h-10 bg-red-100 rounded-lg flex items-center justify-center">
                    <FileText className="w-5 h-5 text-red-500" />
                  </div>
                  <div className="flex-1">
                    <p className="font-medium text-slate-900">{doc.name}</p>
                    <p className="text-sm text-slate-500">{doc.size}</p>
                  </div>
                </div>
              ))}
            </div>

            <button className="flex items-center justify-center gap-2 w-full py-3 border-2 border-dashed border-slate-300 rounded-xl text-slate-500 hover:border-primary hover:text-primary transition-colors">
              <Upload className="w-5 h-5" />
              上传新文档
            </button>
          </div>
        )}

        {/* Step 5: Test & Publish */}
        {currentStep === 5 && (
          <div className="space-y-6">
            <h2 className="text-lg font-semibold text-slate-900">测试与发布</h2>
            
            <div className="bg-slate-50 rounded-xl p-4 h-64 overflow-y-auto">
              <div className="space-y-4">
                <div className="flex gap-3">
                  <div className="w-8 h-8 rounded-full bg-gradient-to-br from-primary to-primary-dark flex items-center justify-center text-white text-sm font-semibold flex-shrink-0">
                    AI
                  </div>
                  <div className="bg-white p-3 rounded-2xl rounded-tl-none shadow-sm max-w-[80%]">
                    <p className="text-slate-700">
                      {formData.welcomeMessage || `你好！我是${formData.name || '你的AI助手'}，${formData.description || '有什么可以帮助你的吗？'}`}
                    </p>
                  </div>
                </div>
              </div>
            </div>

            <div className="flex gap-2">
              <input
                type="text"
                placeholder="输入消息测试角色..."
                className="flex-1 px-4 py-2.5 border border-slate-200 rounded-lg outline-none focus:border-primary focus:ring-2 focus:ring-primary/20 transition-all"
              />
              <button className="px-6 py-2.5 bg-primary text-white rounded-lg hover:bg-primary-dark transition-colors">
                发送
              </button>
            </div>

            <div className="flex items-center gap-4 pt-4 border-t border-slate-100">
              <label className="flex items-center gap-2 cursor-pointer">
                <input
                  type="checkbox"
                  checked={isPublic}
                  onChange={(e) => setIsPublic(e.target.checked)}
                  className="w-4 h-4 rounded border-slate-300 text-primary focus:ring-primary"
                />
                <span className="text-sm text-slate-700">发布到角色市场（公开）</span>
              </label>
            </div>
          </div>
        )}

        {error && (
          <div className="mt-4 rounded-lg bg-red-50 px-4 py-2 text-sm text-red-600">
            {error}
          </div>
        )}

        {/* Navigation Buttons */}
        <div className="flex items-center justify-between mt-8 pt-6 border-t border-slate-100">
          <button
            onClick={handlePrev}
            disabled={currentStep === 1}
            className="flex items-center gap-2 px-6 py-2.5 text-slate-600 hover:bg-slate-100 rounded-lg transition-colors disabled:opacity-50 disabled:cursor-not-allowed"
          >
            <ChevronLeft className="w-5 h-5" />
            上一步
          </button>

          <div className="flex items-center gap-3">
            <button
              onClick={handleSaveDraft}
              className="flex items-center gap-2 px-6 py-2.5 text-slate-600 hover:bg-slate-100 rounded-lg transition-colors"
            >
              <Save className="w-5 h-5" />
              保存草稿
            </button>
            
            {currentStep < 5 ? (
              <button
                onClick={handleNext}
                className="flex items-center gap-2 px-6 py-2.5 bg-slate-900 text-white rounded-lg hover:bg-slate-800 transition-colors"
              >
                下一步
                <ChevronRight className="w-5 h-5" />
              </button>
            ) : (
              <button
                onClick={handlePublish}
                disabled={publishing}
                className="flex items-center gap-2 px-6 py-2.5 bg-primary text-white rounded-lg hover:bg-primary-dark transition-colors disabled:opacity-50 disabled:cursor-not-allowed"
              >
                <Play className="w-5 h-5" />
                {publishing ? '发布中...' : '发布角色'}
              </button>
            )}
          </div>
        </div>
      </div>
    </div>
  );
};
