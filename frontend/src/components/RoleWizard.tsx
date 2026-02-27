import type { FC } from 'react';
import { useState, useEffect } from 'react';
import { 
  Sparkles, 
  Brain, 
  MessageSquare, 
  Check, 
  ChevronRight, 
  ChevronLeft, 
  Send,
  ThumbsUp,
  ThumbsDown,
  RotateCcw,
  Wand2,
  Save,
  X
} from 'lucide-react';

// ============ 数据类型 ============

interface WizardData {
  // 第 1 步：基础信息
  name: string;
  purpose: string;
  style: string;
  
  // 第 2 步：能力配置
  expertise: string[];
  avoidances: string[];
  specialRequirements: string;
  
  // 第 3 步：测试
  testMessage: string;
  testResponse: string;
  satisfaction: number | null;
}

interface Step {
  id: number;
  label: string;
  icon: any;
}

interface Option {
  id: string;
  label: string;
  description: string;
  icon?: string;
}

// ============ 配置数据 ============

const PURPOSES: Option[] = [
  { id: 'assistant', label: '智能助理', description: '处理日常事务、安排日程、撰写邮件', icon: '📋' },
  { id: 'expert', label: '专业顾问', description: '提供专业领域的咨询和建议', icon: '🎯' },
  { id: 'creator', label: '内容创作', description: '撰写文案、故事、营销内容', icon: '✍️' },
  { id: 'teacher', label: '教学辅导', description: '知识讲解、学习辅导、技能培训', icon: '📚' },
  { id: 'companion', label: '情感陪伴', description: '聊天解闷、情感支持、心理疏导', icon: '💙' },
  { id: 'analyst', label: '数据分析', description: '数据处理、报告生成、商业分析', icon: '📊' },
];

const STYLES: Option[] = [
  { id: 'professional', label: '专业严谨', description: '正式、准确、条理清晰', icon: '👔' },
  { id: 'friendly', label: '友好亲切', description: '温暖、耐心、易于接近', icon: '😊' },
  { id: 'humorous', label: '幽默风趣', description: '轻松、有趣、富有创意', icon: '😄' },
  { id: 'concise', label: '简洁直接', description: '高效、直接、不啰嗦', icon: '⚡' },
  { id: 'detailed', label: '详细周全', description: '全面、深入、注重细节', icon: '📝' },
  { id: 'inspirational', label: '激励鼓舞', description: '积极、向上、充满能量', icon: '🌟' },
];

const EXPERTISE_AREAS: Option[] = [
  { id: 'business', label: '商务办公', description: '邮件、文档、会议、项目管理' },
  { id: 'marketing', label: '市场营销', description: '策划、文案、推广、品牌' },
  { id: 'tech', label: '技术编程', description: '开发、调试、架构、算法' },
  { id: 'design', label: '创意设计', description: 'UI/UX、平面、创意构思' },
  { id: 'finance', label: '财务金融', description: '会计、投资、理财、税务' },
  { id: 'legal', label: '法律法务', description: '合同、合规、法律咨询' },
  { id: 'hr', label: '人力资源', description: '招聘、培训、绩效、员工关系' },
  { id: 'health', label: '健康医疗', description: '健身、营养、心理健康' },
  { id: 'education', label: '教育培训', description: '课程、辅导、学习方法' },
  { id: 'lifestyle', label: '生活休闲', description: '旅行、美食、购物、娱乐' },
];

const AVOIDANCES: Option[] = [
  { id: 'speculation', label: '猜测臆断', description: '不确定的信息要明确说明' },
  { id: 'repetition', label: '重复啰嗦', description: '避免重复相同内容' },
  { id: 'jargon', label: '专业术语', description: '少用晦涩难懂的专业词汇' },
  { id: 'controversy', label: '敏感话题', description: '避开政治、宗教等敏感议题' },
  { id: 'overpromise', label: '过度承诺', description: '不夸大能力，诚实告知局限' },
  { id: 'bias', label: '主观偏见', description: '保持客观中立，不带个人偏见' },
];

const STEPS: Step[] = [
  { id: 1, label: '基础信息', icon: Sparkles },
  { id: 2, label: '能力配置', icon: Brain },
  { id: 3, label: '测试优化', icon: MessageSquare },
];

// ============ 智能推荐逻辑 ============

const getRecommendations = (data: Partial<WizardData>) => {
  const recommendations: string[] = [];
  
  // 基于用途推荐
  if (data.purpose === 'assistant') {
    recommendations.push('建议开启「日程管理」和「邮件撰写」技能');
    recommendations.push('说话风格推荐「专业严谨」或「友好亲切」');
  }
  
  if (data.purpose === 'creator') {
    recommendations.push('建议开启「创意思维」和「文案撰写」能力');
    recommendations.push('说话风格推荐「幽默风趣」或「激励鼓舞」');
  }
  
  // 基于风格推荐
  if (data.style === 'professional') {
    recommendations.push('避免使用表情符号和网络用语');
    recommendations.push('回答应结构化，使用清晰的标题和列表');
  }
  
  if (data.style === 'friendly') {
    recommendations.push('可以适当使用表情符号增加亲和力');
    recommendations.push('多用「我们」「一起」等拉近距离的词汇');
  }
  
  // 基于专业领域推荐
  if (data.expertise?.includes('legal')) {
    recommendations.push('重要：添加免责声明「不构成正式法律意见」');
    recommendations.push('建议开启「谨慎准确」避免模式');
  }
  
  if (data.expertise?.includes('health')) {
    recommendations.push('重要：添加医疗免责声明');
    recommendations.push('建议用户咨询专业医师');
  }
  
  return recommendations;
};

// ============ 主组件 ============

export const RoleWizard: FC = () => {
  const [currentStep, setCurrentStep] = useState(1);
  const [isComplete, setIsComplete] = useState(false);
  const [showRecommendations, setShowRecommendations] = useState(true);
  
  const [data, setData] = useState<WizardData>({
    name: '',
    purpose: '',
    style: '',
    expertise: [],
    avoidances: [],
    specialRequirements: '',
    testMessage: '',
    testResponse: '',
    satisfaction: null,
  });

  // 进度保存 (localStorage)
  useEffect(() => {
    const saved = localStorage.getItem('roleWizardData');
    if (saved) {
      try {
        const parsed = JSON.parse(saved);
        setData(prev => ({ ...prev, ...parsed }));
      } catch (e) {
        console.error('Failed to load saved data');
      }
    }
  }, []);

  useEffect(() => {
    localStorage.setItem('roleWizardData', JSON.stringify(data));
  }, [data]);

  // 步骤验证
  const canProceed = () => {
    if (currentStep === 1) {
      return data.name.trim().length > 0 && data.purpose && data.style;
    }
    if (currentStep === 2) {
      return data.expertise.length > 0;
    }
    return true;
  };

  // 导航
  const handleNext = () => {
    if (currentStep < 3 && canProceed()) {
      setCurrentStep(currentStep + 1);
    }
  };

  const handlePrev = () => {
    if (currentStep > 1) {
      setCurrentStep(currentStep - 1);
    }
  };

  // 切换选择
  const toggleExpertise = (id: string) => {
    setData(prev => ({
      ...prev,
      expertise: prev.expertise.includes(id)
        ? prev.expertise.filter(e => e !== id)
        : [...prev.expertise, id]
    }));
  };

  const toggleAvoidance = (id: string) => {
    setData(prev => ({
      ...prev,
      avoidances: prev.avoidances.includes(id)
        ? prev.avoidances.filter(a => a !== id)
        : [...prev.avoidances, id]
    }));
  };

  // 生成提示词
  const generatePrompt = () => {
    const purposeDesc = PURPOSES.find(p => p.id === data.purpose)?.description || '';
    const styleDesc = STYLES.find(s => s.id === data.style)?.description || '';
    const expertiseNames = data.expertise.map(id => 
      EXPERTISE_AREAS.find(e => e.id === id)?.label || ''
    ).filter(Boolean).join('、');
    const avoidanceNames = data.avoidances.map(id =>
      AVOIDANCES.find(a => a.id === id)?.label || ''
    ).filter(Boolean).join('、');

    let prompt = `# 角色设定：${data.name}

## 核心定位
你是一位${purposeDesc}的 AI 助手。你的主要职责是帮助用户${purposeDesc.toLowerCase()}。

## 说话风格
${styleDesc}。在交流中，你应该${styleDesc.toLowerCase()}。

## 专业领域
你擅长以下领域：${expertiseNames || '通用知识'}。在这些领域内，你应该提供专业、准确的建议和信息。
`;

    if (avoidanceNames) {
      prompt += `
## 应避免事项
请注意避免：${avoidanceNames}。在回答时要特别注意这些方面。
`;
    }

    if (data.specialRequirements) {
      prompt += `
## 特殊要求
${data.specialRequirements}
`;
    }

    prompt += `
## 行为准则
1. 始终以帮助用户为首要目标
2. 如遇不确定的信息，诚实告知而非猜测
3. 保持专业且友好的态度
4. 回答应清晰、有条理、实用

## 开始
现在，请以${data.name}的身份，用${styleDesc}的方式，开始为用户提供帮助。`;

    return prompt;
  };

  // 模拟测试对话
  const runTest = () => {
    const prompt = generatePrompt();
    // 这里应该调用后端 API，现在模拟
    const mockResponse = `你好！我是${data.name}，${PURPOSES.find(p => p.id === data.purpose)?.label || '你的 AI 助手'}。${data.style === 'friendly' ? '很高兴见到你！' : '请问有什么可以帮你的？'}`;
    
    setData(prev => ({ ...prev, testResponse: mockResponse }));
  };

  // 完成创建
  const handleComplete = () => {
    const roleData = {
      name: data.name,
      description: `${PURPOSES.find(p => p.id === data.purpose)?.label} - ${STYLES.find(s => s.id === data.style)?.label}`,
      category: EXPERTISE_AREAS.find(e => e.id === data.expertise[0])?.label || '通用',
      systemPrompt: generatePrompt(),
      welcomeMessage: data.testResponse || `你好！我是${data.name}，有什么可以帮助你的吗？`,
      modelConfig: {
        temperature: data.style === 'humorous' || data.style === 'inspirational' ? 0.8 : 0.7,
      },
    };

    console.log('Creating role:', roleData);
    // TODO: 调用后端 API 创建角色
    setIsComplete(true);
    localStorage.removeItem('roleWizardData');
  };

  // AI 优化
  const aiOptimize = () => {
    // 智能推荐优化建议
    const recommendations = getRecommendations(data);
    if (recommendations.length > 0) {
      alert('💡 AI 优化建议：\n\n' + recommendations.join('\n'));
    } else {
      alert('✨ 当前配置已经很完善了！');
    }
  };

  // 重置
  const handleReset = () => {
    if (confirm('确定要重新开始吗？当前进度将丢失。')) {
      setData({
        name: '',
        purpose: '',
        style: '',
        expertise: [],
        avoidances: [],
        specialRequirements: '',
        testMessage: '',
        testResponse: '',
        satisfaction: null,
      });
      setCurrentStep(1);
      setIsComplete(false);
      localStorage.removeItem('roleWizardData');
    }
  };

  // ============ 完成页面 ============
  if (isComplete) {
    return (
      <div className="max-w-3xl mx-auto">
        <div className="bg-gradient-to-br from-primary/10 to-primary-dark/10 rounded-2xl p-12 text-center">
          <div className="w-24 h-24 mx-auto mb-6 bg-gradient-to-br from-primary to-primary-dark rounded-full flex items-center justify-center animate-bounce">
            <Check className="w-12 h-12 text-white" />
          </div>
          
          <h2 className="text-3xl font-bold text-slate-900 mb-4">
            🎉 角色创建成功！
          </h2>
          
          <p className="text-lg text-slate-600 mb-8">
            「{data.name}」已经准备就绪，可以开始使用了
          </p>

          <div className="flex items-center justify-center gap-4">
            <button
              onClick={handleReset}
              className="flex items-center gap-2 px-6 py-3 text-slate-600 hover:bg-slate-100 rounded-xl transition-colors"
            >
              <RotateCcw className="w-5 h-5" />
              创建另一个
            </button>
            
            <button
              onClick={() => window.location.href = '/roles'}
              className="flex items-center gap-2 px-8 py-3 bg-primary text-white rounded-xl hover:bg-primary-dark transition-colors shadow-lg shadow-primary/30"
            >
              <Send className="w-5 h-5" />
              开始对话
            </button>
          </div>
        </div>
      </div>
    );
  }

  // ============ 向导页面 ============
  return (
    <div className="max-w-5xl mx-auto">
      {/* Header */}
      <div className="mb-8 flex items-center justify-between">
        <div>
          <h1 className="text-2xl font-bold text-slate-900 flex items-center gap-3">
            <Sparkles className="w-7 h-7 text-primary" />
            创建 AI 角色
          </h1>
          <p className="text-slate-500 mt-1">通过简单的问答，3 步打造你的专属 AI 助手</p>
        </div>
        
        <button
          onClick={handleReset}
          className="flex items-center gap-2 px-4 py-2 text-slate-500 hover:text-slate-700 hover:bg-slate-100 rounded-lg transition-colors"
        >
          <X className="w-4 h-4" />
          重新开始
        </button>
      </div>

      {/* Step Navigation */}
      <div className="mb-8">
        <div className="flex items-center justify-between">
          {STEPS.map((step, index) => (
            <div key={step.id} className="flex items-center flex-1">
              <div className="flex flex-col items-center">
                <div className={`w-12 h-12 rounded-full flex items-center justify-center font-semibold transition-all ${
                  currentStep > step.id 
                    ? 'bg-primary text-white' 
                    : currentStep === step.id 
                    ? 'bg-slate-900 text-white ring-4 ring-primary/20' 
                    : 'bg-slate-100 text-slate-400'
                }`}>
                  {currentStep > step.id ? (
                    <Check className="w-6 h-6" />
                  ) : (
                    <step.icon className="w-6 h-6" />
                  )}
                </div>
                <span className={`text-sm mt-2 font-medium ${
                  currentStep >= step.id ? 'text-slate-900' : 'text-slate-400'
                }`}>
                  {step.label}
                </span>
              </div>
              {index < STEPS.length - 1 && (
                <div className={`flex-1 h-1 mx-4 rounded ${
                  currentStep > step.id ? 'bg-primary' : 'bg-slate-200'
                }`} />
              )}
            </div>
          ))}
        </div>
      </div>

      {/* Step Content */}
      <div className="bg-white rounded-2xl shadow-sm border border-slate-100 p-8">
        {/* ========== 第 1 步：基础信息 ========== */}
        {currentStep === 1 && (
          <div className="space-y-8">
            <div>
              <h2 className="text-xl font-semibold text-slate-900 mb-2">
                👋 给你的 AI 助手起个名字
              </h2>
              <p className="text-slate-500">这将是用户看到的第一印象</p>
            </div>

            <div>
              <input
                type="text"
                value={data.name}
                onChange={(e) => setData({ ...data, name: e.target.value })}
                placeholder="例如：办公小助手、营销专家、健身教练..."
                className="w-full px-5 py-4 text-lg border-2 border-slate-200 rounded-xl outline-none focus:border-primary focus:ring-4 focus:ring-primary/10 transition-all"
                autoFocus
              />
            </div>

            <div className="pt-6 border-t border-slate-100">
              <h3 className="text-lg font-semibold text-slate-900 mb-4">
                🎯 主要用于什么场景？
              </h3>
              <div className="grid grid-cols-2 md:grid-cols-3 gap-4">
                {PURPOSES.map(purpose => (
                  <button
                    key={purpose.id}
                    onClick={() => setData({ ...data, purpose: purpose.id })}
                    className={`p-5 rounded-xl border-2 text-left transition-all ${
                      data.purpose === purpose.id
                        ? 'border-primary bg-primary/5 ring-2 ring-primary/20'
                        : 'border-slate-200 hover:border-slate-300 hover:bg-slate-50'
                    }`}
                  >
                    <div className="text-2xl mb-2">{purpose.icon}</div>
                    <div className="font-semibold text-slate-900">{purpose.label}</div>
                    <div className="text-sm text-slate-500 mt-1">{purpose.description}</div>
                  </button>
                ))}
              </div>
            </div>

            <div className="pt-6 border-t border-slate-100">
              <h3 className="text-lg font-semibold text-slate-900 mb-4">
                💬 希望它用什么风格说话？
              </h3>
              <div className="grid grid-cols-2 md:grid-cols-3 gap-4">
                {STYLES.map(style => (
                  <button
                    key={style.id}
                    onClick={() => setData({ ...data, style: style.id })}
                    className={`p-5 rounded-xl border-2 text-left transition-all ${
                      data.style === style.id
                        ? 'border-primary bg-primary/5 ring-2 ring-primary/20'
                        : 'border-slate-200 hover:border-slate-300 hover:bg-slate-50'
                    }`}
                  >
                    <div className="text-2xl mb-2">{style.icon}</div>
                    <div className="font-semibold text-slate-900">{style.label}</div>
                    <div className="text-sm text-slate-500 mt-1">{style.description}</div>
                  </button>
                ))}
              </div>
            </div>
          </div>
        )}

        {/* ========== 第 2 步：能力配置 ========== */}
        {currentStep === 2 && (
          <div className="space-y-8">
            <div>
              <h2 className="text-xl font-semibold text-slate-900 mb-2">
                🧠 选择专业领域
              </h2>
              <p className="text-slate-500">你的 AI 助手擅长什么？（可多选）</p>
            </div>

            <div className="grid grid-cols-2 md:grid-cols-3 gap-4">
              {EXPERTISE_AREAS.map(area => (
                <button
                  key={area.id}
                  onClick={() => toggleExpertise(area.id)}
                  className={`p-5 rounded-xl border-2 text-left transition-all ${
                    data.expertise.includes(area.id)
                      ? 'border-primary bg-primary/5 ring-2 ring-primary/20'
                      : 'border-slate-200 hover:border-slate-300 hover:bg-slate-50'
                  }`}
                >
                  <div className="flex items-start gap-3">
                    <div className={`w-5 h-5 rounded border-2 flex items-center justify-center mt-0.5 flex-shrink-0 ${
                      data.expertise.includes(area.id)
                        ? 'bg-primary border-primary'
                        : 'border-slate-300'
                    }`}>
                      {data.expertise.includes(area.id) && (
                        <Check className="w-3 h-3 text-white" />
                      )}
                    </div>
                    <div>
                      <div className="font-semibold text-slate-900">{area.label}</div>
                      <div className="text-sm text-slate-500 mt-1">{area.description}</div>
                    </div>
                  </div>
                </button>
              ))}
            </div>

            <div className="pt-6 border-t border-slate-100">
              <h3 className="text-lg font-semibold text-slate-900 mb-4">
                ⚠️ 应该避免什么？
              </h3>
              <p className="text-slate-500 mb-4">选择你的 AI 助手需要注意的事项（可多选）</p>
              
              <div className="grid grid-cols-2 md:grid-cols-3 gap-4">
                {AVOIDANCES.map(avoid => (
                  <button
                    key={avoid.id}
                    onClick={() => toggleAvoidance(avoid.id)}
                    className={`p-4 rounded-xl border-2 text-left transition-all ${
                      data.avoidances.includes(avoid.id)
                        ? 'border-amber-500 bg-amber-50 ring-2 ring-amber-200'
                        : 'border-slate-200 hover:border-slate-300 hover:bg-slate-50'
                    }`}
                  >
                    <div className="flex items-start gap-3">
                      <div className={`w-5 h-5 rounded border-2 flex items-center justify-center mt-0.5 flex-shrink-0 ${
                        data.avoidances.includes(avoid.id)
                          ? 'bg-amber-500 border-amber-500'
                          : 'border-slate-300'
                      }`}>
                        {data.avoidances.includes(avoid.id) && (
                          <Check className="w-3 h-3 text-white" />
                        )}
                      </div>
                      <div>
                        <div className="font-semibold text-slate-900">{avoid.label}</div>
                        <div className="text-sm text-slate-500 mt-1">{avoid.description}</div>
                      </div>
                    </div>
                  </button>
                ))}
              </div>
            </div>

            <div className="pt-6 border-t border-slate-100">
              <h3 className="text-lg font-semibold text-slate-900 mb-2">
                ✨ 还有其他特殊要求吗？
              </h3>
              <p className="text-slate-500 mb-4">可选，例如特定的工作流程、格式要求等</p>
              
              <textarea
                value={data.specialRequirements}
                onChange={(e) => setData({ ...data, specialRequirements: e.target.value })}
                placeholder="例如：每次回答前先确认理解了我的问题；使用 Markdown 格式输出代码..."
                rows={3}
                className="w-full px-4 py-3 border-2 border-slate-200 rounded-xl outline-none focus:border-primary focus:ring-4 focus:ring-primary/10 transition-all resize-none"
              />
            </div>

            {/* 智能推荐 */}
            {showRecommendations && (
              <div className="bg-gradient-to-r from-primary/5 to-primary-dark/5 rounded-xl p-6 border border-primary/20">
                <div className="flex items-start justify-between mb-4">
                  <h4 className="font-semibold text-slate-900 flex items-center gap-2">
                    <Sparkles className="w-5 h-5 text-primary" />
                    AI 智能推荐
                  </h4>
                  <button
                    onClick={() => setShowRecommendations(false)}
                    className="text-slate-400 hover:text-slate-600"
                  >
                    <X className="w-4 h-4" />
                  </button>
                </div>
                
                <div className="space-y-2">
                  {getRecommendations(data).length > 0 ? (
                    getRecommendations(data).map((rec, i) => (
                      <div key={i} className="flex items-start gap-2 text-sm text-slate-700">
                        <span className="text-primary mt-0.5">•</span>
                        {rec}
                      </div>
                    ))
                  ) : (
                    <p className="text-sm text-slate-500">继续配置以获取个性化推荐...</p>
                  )}
                </div>
              </div>
            )}
          </div>
        )}

        {/* ========== 第 3 步：测试优化 ========== */}
        {currentStep === 3 && (
          <div className="space-y-8">
            <div>
              <h2 className="text-xl font-semibold text-slate-900 mb-2">
                🧪 测试一下效果
              </h2>
              <p className="text-slate-500">发送一条消息，看看你的 AI 助手如何回应</p>
            </div>

            {/* 测试对话框 */}
            <div className="bg-slate-50 rounded-xl p-6 space-y-4">
              <div className="flex gap-3">
                <div className="w-10 h-10 rounded-full bg-gradient-to-br from-primary to-primary-dark flex items-center justify-center text-white font-semibold flex-shrink-0">
                  {data.name.charAt(0) || 'AI'}
                </div>
                <div className="bg-white p-4 rounded-2xl rounded-tl-none shadow-sm flex-1">
                  <p className="text-slate-700">
                    你好！我是{data.name || '你的 AI 助手'}，
                    {data.purpose && `${PURPOSES.find(p => p.id === data.purpose)?.label}，`}
                    有什么可以帮助你的吗？
                  </p>
                </div>
              </div>

              {data.testResponse && (
                <div className="flex gap-3">
                  <div className="w-10 h-10 rounded-full bg-slate-200 flex items-center justify-center text-slate-600 font-semibold flex-shrink-0">
                    我
                  </div>
                  <div className="bg-primary/10 p-4 rounded-2xl rounded-tr-none flex-1">
                    <p className="text-slate-700">{data.testMessage}</p>
                  </div>
                </div>
              )}

              {data.testResponse && (
                <>
                  <div className="flex gap-3">
                    <div className="w-10 h-10 rounded-full bg-gradient-to-br from-primary to-primary-dark flex items-center justify-center text-white font-semibold flex-shrink-0">
                      {data.name.charAt(0) || 'AI'}
                    </div>
                    <div className="bg-white p-4 rounded-2xl rounded-tl-none shadow-sm flex-1">
                      <p className="text-slate-700">{data.testResponse}</p>
                    </div>
                  </div>

                  <div className="flex items-center gap-4 pt-4 border-t border-slate-200">
                    <span className="text-sm text-slate-600">满意这个回答吗？</span>
                    <button
                      onClick={() => setData({ ...data, satisfaction: 1 })}
                      className={`p-2 rounded-lg transition-colors ${
                        data.satisfaction === 1 
                          ? 'bg-green-100 text-green-600' 
                          : 'hover:bg-slate-100 text-slate-400'
                      }`}
                    >
                      <ThumbsUp className="w-5 h-5" />
                    </button>
                    <button
                      onClick={() => setData({ ...data, satisfaction: 0 })}
                      className={`p-2 rounded-lg transition-colors ${
                        data.satisfaction === 0 
                          ? 'bg-red-100 text-red-600' 
                          : 'hover:bg-slate-100 text-slate-400'
                      }`}
                    >
                      <ThumbsDown className="w-5 h-5" />
                    </button>
                  </div>
                </>
              )}

              <div className="flex gap-3 pt-4">
                <input
                  type="text"
                  value={data.testMessage}
                  onChange={(e) => setData({ ...data, testMessage: e.target.value })}
                  onKeyPress={(e) => e.key === 'Enter' && runTest()}
                  placeholder="输入消息测试..."
                  className="flex-1 px-4 py-3 border-2 border-slate-200 rounded-xl outline-none focus:border-primary focus:ring-4 focus:ring-primary/10 transition-all"
                />
                <button
                  onClick={runTest}
                  className="flex items-center gap-2 px-6 py-3 bg-primary text-white rounded-xl hover:bg-primary-dark transition-colors"
                >
                  <Send className="w-5 h-5" />
                  发送
                </button>
              </div>
            </div>

            {/* 优化建议 */}
            <div className="bg-gradient-to-r from-amber-50 to-orange-50 rounded-xl p-6 border border-amber-200">
              <h4 className="font-semibold text-slate-900 mb-4 flex items-center gap-2">
                <Wand2 className="w-5 h-5 text-amber-600" />
                需要优化吗？
              </h4>
              
              <div className="flex flex-wrap gap-3">
                <button
                  onClick={aiOptimize}
                  className="flex items-center gap-2 px-5 py-2.5 bg-white border-2 border-amber-300 text-amber-700 rounded-lg hover:bg-amber-50 transition-colors font-medium"
                >
                  <Sparkles className="w-4 h-4" />
                  AI 智能优化
                </button>
                
                <button
                  onClick={handlePrev}
                  className="flex items-center gap-2 px-5 py-2.5 bg-white border-2 border-slate-300 text-slate-700 rounded-lg hover:bg-slate-50 transition-colors font-medium"
                >
                  <RotateCcw className="w-4 h-4" />
                  返回调整
                </button>
              </div>

              <div className="mt-4 text-sm text-slate-600">
                <p>💡 提示：你可以随时返回上一步修改配置</p>
              </div>
            </div>

            {/* 生成的提示词预览 */}
            <details className="group">
              <summary className="flex items-center gap-2 cursor-pointer text-sm text-slate-500 hover:text-slate-700">
                <ChevronRight className="w-4 h-4 transition-transform group-open:rotate-90" />
                查看生成的提示词
              </summary>
              <div className="mt-4 bg-slate-900 rounded-xl p-5 overflow-x-auto">
                <pre className="text-sm text-slate-100 whitespace-pre-wrap font-mono">
                  {generatePrompt()}
                </pre>
              </div>
            </details>
          </div>
        )}

        {/* Navigation Buttons */}
        <div className="flex items-center justify-between mt-8 pt-6 border-t border-slate-100">
          <button
            onClick={handlePrev}
            disabled={currentStep === 1}
            className="flex items-center gap-2 px-6 py-3 text-slate-600 hover:bg-slate-100 rounded-xl transition-colors disabled:opacity-50 disabled:cursor-not-allowed"
          >
            <ChevronLeft className="w-5 h-5" />
            上一步
          </button>

          <div className="flex items-center gap-3">
            <button
              onClick={() => localStorage.setItem('roleWizardData', JSON.stringify(data))}
              className="flex items-center gap-2 px-6 py-3 text-slate-600 hover:bg-slate-100 rounded-xl transition-colors"
            >
              <Save className="w-5 h-5" />
              保存进度
            </button>
            
            {currentStep < 3 ? (
              <button
                onClick={handleNext}
                disabled={!canProceed()}
                className="flex items-center gap-2 px-8 py-3 bg-slate-900 text-white rounded-xl hover:bg-slate-800 transition-colors disabled:opacity-50 disabled:cursor-not-allowed shadow-lg shadow-slate-900/20"
              >
                下一步
                <ChevronRight className="w-5 h-5" />
              </button>
            ) : (
              <button
                onClick={handleComplete}
                className="flex items-center gap-2 px-8 py-3 bg-gradient-to-r from-primary to-primary-dark text-white rounded-xl hover:from-primary-dark hover:to-primary transition-all shadow-lg shadow-primary/30"
              >
                <Check className="w-5 h-5" />
                完成创建
              </button>
            )}
          </div>
        </div>
      </div>

      {/* Progress hint */}
      <div className="mt-6 text-center text-sm text-slate-500">
        进度已自动保存，随时可以继续
      </div>
    </div>
  );
};
