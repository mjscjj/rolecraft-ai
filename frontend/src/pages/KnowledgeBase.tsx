import type { FC } from 'react'; import { useState } from 'react';
import { 
  Upload, 
  FileText, 
  MoreVertical, 
  Search, 
  Filter,
  Trash2,
  Download,
  CheckCircle,
  Clock,
  AlertCircle
} from 'lucide-react';
import type { Document } from '../types';

const mockDocuments: Document[] = [
  { id: '1', name: '产品手册.pdf', fileType: 'PDF', fileSize: 2621440, status: 'completed', createdAt: '2024-01-15' },
  { id: '2', name: '公司制度.docx', fileType: 'DOCX', fileSize: 1258291, status: 'processing', createdAt: '2024-01-18' },
  { id: '3', name: '竞品分析.pdf', fileType: 'PDF', fileSize: 5347737, status: 'completed', createdAt: '2024-01-19' },
  { id: '4', name: '营销策略2024.pptx', fileType: 'PPTX', fileSize: 4194304, status: 'completed', createdAt: '2024-01-20' },
  { id: '5', name: '用户调研报告.pdf', fileType: 'PDF', fileSize: 3145728, status: 'failed', createdAt: '2024-01-20' },
];

const formatFileSize = (bytes: number): string => {
  if (bytes === 0) return '0 Bytes';
  const k = 1024;
  const sizes = ['Bytes', 'KB', 'MB', 'GB'];
  const i = Math.floor(Math.log(bytes) / Math.log(k));
  return parseFloat((bytes / Math.pow(k, i)).toFixed(2)) + ' ' + sizes[i];
};

const getStatusIcon = (status: string) => {
  switch (status) {
    case 'completed':
      return <CheckCircle className="w-5 h-5 text-green-500" />;
    case 'processing':
      return <Clock className="w-5 h-5 text-blue-500" />;
    case 'failed':
      return <AlertCircle className="w-5 h-5 text-red-500" />;
    default:
      return <Clock className="w-5 h-5 text-slate-400" />;
  }
};

const getStatusText = (status: string) => {
  switch (status) {
    case 'completed':
      return '已就绪';
    case 'processing':
      return '处理中';
    case 'failed':
      return '处理失败';
    default:
      return '等待中';
  }
};

const getFileIcon = (fileType: string) => {
  const colorClass = fileType === 'PDF' ? 'text-red-500 bg-red-50' :
                    fileType === 'DOCX' ? 'text-blue-500 bg-blue-50' :
                    fileType === 'PPTX' ? 'text-orange-500 bg-orange-50' :
                    'text-slate-500 bg-slate-50';
  
  return (
    <div className={`w-10 h-10 rounded-lg flex items-center justify-center ${colorClass}`}>
      <FileText className="w-5 h-5" />
    </div>
  );
};

export const KnowledgeBase: FC = () => {
  const [documents] = useState<Document[]>(mockDocuments);
  const [selectedDocs, setSelectedDocs] = useState<string[]>([]);
  const [searchQuery, setSearchQuery] = useState('');

  const toggleSelect = (docId: string) => {
    setSelectedDocs(prev => 
      prev.includes(docId)
        ? prev.filter(id => id !== docId)
        : [...prev, docId]
    );
  };

  const toggleSelectAll = () => {
    if (selectedDocs.length === documents.length) {
      setSelectedDocs([]);
    } else {
      setSelectedDocs(documents.map(d => d.id));
    }
  };

  const filteredDocs = documents.filter(doc => 
    doc.name.toLowerCase().includes(searchQuery.toLowerCase())
  );

  return (
    <div className="space-y-6">
      {/* Header */}
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-2xl font-bold text-slate-900">知识库</h1>
          <p className="text-slate-500 mt-1">管理你的文档，为 AI 角色提供知识支持</p>
        </div>
        <button className="flex items-center gap-2 px-4 py-2.5 bg-slate-900 text-white rounded-lg hover:bg-slate-800 transition-colors">
          <Upload className="w-5 h-5" />
          上传文档
        </button>
      </div>

      {/* Stats */}
      <div className="grid grid-cols-4 gap-4">
        <div className="bg-white p-4 rounded-xl border border-slate-100">
          <p className="text-sm text-slate-500">文档总数</p>
          <p className="text-2xl font-bold text-slate-900 mt-1">{documents.length}</p>
        </div>
        <div className="bg-white p-4 rounded-xl border border-slate-100">
          <p className="text-sm text-slate-500">已就绪</p>
          <p className="text-2xl font-bold text-green-600 mt-1">
            {documents.filter(d => d.status === 'completed').length}
          </p>
        </div>
        <div className="bg-white p-4 rounded-xl border border-slate-100">
          <p className="text-sm text-slate-500">处理中</p>
          <p className="text-2xl font-bold text-blue-600 mt-1">
            {documents.filter(d => d.status === 'processing').length}
          </p>
        </div>
        <div className="bg-white p-4 rounded-xl border border-slate-100">
          <p className="text-sm text-slate-500">存储空间</p>
          <p className="text-2xl font-bold text-slate-900 mt-1">
            {formatFileSize(documents.reduce((acc, d) => acc + d.fileSize, 0))}
          </p>
        </div>
      </div>

      {/* Toolbar */}
      <div className="flex items-center justify-between">
        <div className="flex items-center gap-4">
          <div className="relative">
            <Search className="absolute left-3 top-1/2 -translate-y-1/2 w-5 h-5 text-slate-400" />
            <input
              type="text"
              placeholder="搜索文档..."
              value={searchQuery}
              onChange={(e) => setSearchQuery(e.target.value)}
              className="pl-10 pr-4 py-2 bg-white border border-slate-200 rounded-lg outline-none focus:border-primary focus:ring-2 focus:ring-primary/20 transition-all w-64"
            />
          </div>
          <button className="flex items-center gap-2 px-3 py-2 text-slate-600 hover:bg-white hover:border-slate-200 border border-transparent rounded-lg transition-all">
            <Filter className="w-4 h-4" />
            筛选
          </button>
        </div>

        {selectedDocs.length > 0 && (
          <div className="flex items-center gap-2">
            <span className="text-sm text-slate-500">已选择 {selectedDocs.length} 项</span>
            <button className="p-2 text-red-500 hover:bg-red-50 rounded-lg transition-colors">
              <Trash2 className="w-5 h-5" />
            </button>
          </div>
        )}
      </div>

      {/* Document Table */}
      <div className="bg-white rounded-xl border border-slate-200 overflow-hidden">
        <table className="w-full">
          <thead className="bg-slate-50 border-b border-slate-200">
            <tr>
              <th className="px-4 py-3 text-left">
                <input
                  type="checkbox"
                  checked={selectedDocs.length === documents.length && documents.length > 0}
                  onChange={toggleSelectAll}
                  className="w-4 h-4 rounded border-slate-300 text-primary focus:ring-primary"
                />
              </th>
              <th className="px-4 py-3 text-left text-sm font-medium text-slate-700">文档名称</th>
              <th className="px-4 py-3 text-left text-sm font-medium text-slate-700">类型</th>
              <th className="px-4 py-3 text-left text-sm font-medium text-slate-700">大小</th>
              <th className="px-4 py-3 text-left text-sm font-medium text-slate-700">状态</th>
              <th className="px-4 py-3 text-left text-sm font-medium text-slate-700">上传时间</th>
              <th className="px-4 py-3 text-left text-sm font-medium text-slate-700">操作</th>
            </tr>
          </thead>
          <tbody className="divide-y divide-slate-100">
            {filteredDocs.map(doc => (
              <tr key={doc.id} className="hover:bg-slate-50 transition-colors">
                <td className="px-4 py-4">
                  <input
                    type="checkbox"
                    checked={selectedDocs.includes(doc.id)}
                    onChange={() => toggleSelect(doc.id)}
                    className="w-4 h-4 rounded border-slate-300 text-primary focus:ring-primary"
                  />
                </td>
                <td className="px-4 py-4">
                  <div className="flex items-center gap-3">
                    {getFileIcon(doc.fileType)}
                    <span className="font-medium text-slate-900">{doc.name}</span>
                  </div>
                </td>
                <td className="px-4 py-4 text-sm text-slate-600">{doc.fileType}</td>
                <td className="px-4 py-4 text-sm text-slate-600">{formatFileSize(doc.fileSize)}</td>
                <td className="px-4 py-4">
                  <div className="flex items-center gap-2">
                    {getStatusIcon(doc.status)}
                    <span className="text-sm text-slate-600">{getStatusText(doc.status)}</span>
                  </div>
                </td>
                <td className="px-4 py-4 text-sm text-slate-500">{doc.createdAt}</td>
                <td className="px-4 py-4">
                  <div className="flex items-center gap-1">
                    <button className="p-2 text-slate-400 hover:text-slate-600 hover:bg-slate-100 rounded-lg transition-colors">
                      <Download className="w-4 h-4" />
                    </button>
                    <button className="p-2 text-slate-400 hover:text-slate-600 hover:bg-slate-100 rounded-lg transition-colors">
                      <MoreVertical className="w-4 h-4" />
                    </button>
                  </div>
                </td>
              </tr>
            ))}
          </tbody>
        </table>

        {filteredDocs.length === 0 && (
          <div className="text-center py-16">
            <div className="w-16 h-16 bg-slate-100 rounded-full flex items-center justify-center mx-auto mb-4">
              <FileText className="w-8 h-8 text-slate-400" />
            </div>
            <p className="text-slate-500">暂无文档</p>
            <button className="mt-4 text-primary hover:underline">上传第一个文档</button>
          </div>
        )}
      </div>

      {/* Upload Area (Drop Zone) */}
      <div className="border-2 border-dashed border-slate-300 rounded-xl p-8 text-center hover:border-primary hover:bg-primary/5 transition-colors cursor-pointer">
        <div className="w-16 h-16 bg-slate-100 rounded-full flex items-center justify-center mx-auto mb-4">
          <Upload className="w-8 h-8 text-slate-400" />
        </div>
        <p className="text-slate-700 font-medium">拖拽文件到此处上传</p>
        <p className="text-sm text-slate-500 mt-1">支持 PDF、Word、TXT、Markdown 等格式</p>
        <p className="text-xs text-slate-400 mt-2">单个文件最大 50MB</p>
      </div>
    </div>
  );
};