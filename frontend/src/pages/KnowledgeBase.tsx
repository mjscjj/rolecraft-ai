import type { FC } from 'react';
import { useState, useEffect } from 'react';
import { 
  Upload, 
  FileText, 
  MoreVertical, 
  Search, 
  Trash2,
  Edit2,
  CheckCircle,
  Clock,
  AlertCircle,
  X
} from 'lucide-react';

const API_BASE = 'http://localhost:8080/api/v1';

interface Document {
  id: string;
  name: string;
  fileType: string;
  fileSize: number;
  status: string;
  filePath?: string;
  createdAt: string;
  updatedAt?: string;
}

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
  const colorClass = fileType === 'pdf' ? 'text-red-500 bg-red-50' :
                    fileType === 'doc' || fileType === 'docx' ? 'text-blue-500 bg-blue-50' :
                    fileType === 'txt' || fileType === 'md' ? 'text-slate-500 bg-slate-50' :
                    'text-slate-500 bg-slate-50';
  
  return (
    <div className={`w-10 h-10 rounded-lg flex items-center justify-center ${colorClass}`}>
      <FileText className="w-5 h-5" />
    </div>
  );
};

export const KnowledgeBase: FC = () => {
  const [documents, setDocuments] = useState<Document[]>([]);
  const [loading, setLoading] = useState(true);
  const [uploading, setUploading] = useState(false);
  const [searchQuery, setSearchQuery] = useState('');
  const [editingDoc, setEditingDoc] = useState<Document | null>(null);
  const [editName, setEditName] = useState('');

  const token = typeof window !== 'undefined' ? localStorage.getItem('token') : null;

  // 加载文档列表
  useEffect(() => {
    loadDocuments();
  }, []);

  const loadDocuments = async () => {
    if (!token) return;
    try {
      const res = await fetch(`${API_BASE}/documents`, {
        headers: { 'Authorization': `Bearer ${token}` },
      });
      const data = await res.json();
      if (data.data) {
        setDocuments(data.data);
      }
    } catch (err) {
      console.error('Failed to load documents:', err);
    } finally {
      setLoading(false);
    }
  };

  const handleUpload = async (e: React.ChangeEvent<HTMLInputElement>) => {
    const file = e.target.files?.[0];
    if (!file || !token) return;

    setUploading(true);
    const formData = new FormData();
    formData.append('file', file);

    try {
      const res = await fetch(`${API_BASE}/documents`, {
        method: 'POST',
        headers: { 'Authorization': `Bearer ${token}` },
        body: formData,
      });
      const data = await res.json();
      if (data.data) {
        setDocuments([data.data, ...documents]);
      }
    } catch (err) {
      console.error('Failed to upload:', err);
    } finally {
      setUploading(false);
      e.target.value = '';
    }
  };

  const handleDelete = async (docId: string) => {
    if (!token || !confirm('确定要删除这个文档吗？')) return;

    try {
      const res = await fetch(`${API_BASE}/documents/${docId}`, {
        method: 'DELETE',
        headers: { 'Authorization': `Bearer ${token}` },
      });
      const data = await res.json();
      if (data.code === 200) {
        setDocuments(documents.filter(d => d.id !== docId));
      }
    } catch (err) {
      console.error('Failed to delete:', err);
    }
  };

  const handleEdit = (doc: Document) => {
    setEditingDoc(doc);
    setEditName(doc.name);
  };

  const handleUpdate = async () => {
    if (!token || !editingDoc) return;

    try {
      const res = await fetch(`${API_BASE}/documents/${editingDoc.id}`, {
        method: 'PUT',
        headers: {
          'Authorization': `Bearer ${token}`,
          'Content-Type': 'application/json',
        },
        body: JSON.stringify({ name: editName }),
      });
      const data = await res.json();
      if (data.data) {
        setDocuments(documents.map(d => 
          d.id === editingDoc.id ? { ...d, ...data.data } : d
        ));
        setEditingDoc(null);
      }
    } catch (err) {
      console.error('Failed to update:', err);
    }
  };

  const filteredDocs = documents.filter(doc => 
    doc.name.toLowerCase().includes(searchQuery.toLowerCase())
  );

  const stats = {
    total: documents.length,
    completed: documents.filter(d => d.status === 'completed').length,
    processing: documents.filter(d => d.status === 'processing').length,
    failed: documents.filter(d => d.status === 'failed').length,
  };

  return (
    <div className="space-y-6">
      {/* Header */}
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-2xl font-bold text-slate-900">知识库</h1>
          <p className="text-slate-500 mt-1">管理你的文档和知识资源</p>
        </div>
        <label className="flex items-center gap-2 px-4 py-2 bg-primary text-white rounded-lg hover:bg-primary-dark transition-colors cursor-pointer">
          <Upload className="w-5 h-5" />
          <span>{uploading ? '上传中...' : '上传文档'}</span>
          <input
            type="file"
            accept=".pdf,.doc,.docx,.txt,.md"
            onChange={handleUpload}
            disabled={uploading}
            className="hidden"
          />
        </label>
      </div>

      {/* Stats */}
      <div className="grid grid-cols-4 gap-4">
        <div className="bg-white p-4 rounded-xl border border-slate-200">
          <p className="text-sm text-slate-500">总文档数</p>
          <p className="text-2xl font-bold text-slate-900">{stats.total}</p>
        </div>
        <div className="bg-white p-4 rounded-xl border border-slate-200">
          <p className="text-sm text-slate-500">已就绪</p>
          <p className="text-2xl font-bold text-green-600">{stats.completed}</p>
        </div>
        <div className="bg-white p-4 rounded-xl border border-slate-200">
          <p className="text-sm text-slate-500">处理中</p>
          <p className="text-2xl font-bold text-blue-600">{stats.processing}</p>
        </div>
        <div className="bg-white p-4 rounded-xl border border-slate-200">
          <p className="text-sm text-slate-500">失败</p>
          <p className="text-2xl font-bold text-red-600">{stats.failed}</p>
        </div>
      </div>

      {/* Search */}
      <div className="relative">
        <Search className="absolute left-3 top-1/2 -translate-y-1/2 w-5 h-5 text-slate-400" />
        <input
          type="text"
          placeholder="搜索文档..."
          value={searchQuery}
          onChange={(e) => setSearchQuery(e.target.value)}
          className="w-full pl-10 pr-4 py-2 border border-slate-200 rounded-lg focus:outline-none focus:ring-2 focus:ring-primary"
        />
      </div>

      {/* Document List */}
      {loading ? (
        <div className="text-center py-20 text-slate-500">加载中...</div>
      ) : filteredDocs.length === 0 ? (
        <div className="text-center py-20 text-slate-500">
          <FileText className="w-16 h-16 mx-auto mb-4 opacity-20" />
          <p>暂无文档</p>
          <p className="text-sm mt-2">上传 PDF、Word、TXT 等格式文档</p>
        </div>
      ) : (
        <div className="bg-white rounded-xl border border-slate-200 overflow-hidden">
          <table className="w-full">
            <thead className="bg-slate-50 border-b border-slate-200">
              <tr>
                <th className="text-left py-3 px-4 text-sm font-medium text-slate-600">文档名称</th>
                <th className="text-left py-3 px-4 text-sm font-medium text-slate-600">类型</th>
                <th className="text-left py-3 px-4 text-sm font-medium text-slate-600">大小</th>
                <th className="text-left py-3 px-4 text-sm font-medium text-slate-600">状态</th>
                <th className="text-left py-3 px-4 text-sm font-medium text-slate-600">上传时间</th>
                <th className="text-right py-3 px-4 text-sm font-medium text-slate-600">操作</th>
              </tr>
            </thead>
            <tbody>
              {filteredDocs.map((doc) => (
                <tr key={doc.id} className="border-b border-slate-100 hover:bg-slate-50">
                  <td className="py-3 px-4">
                    <div className="flex items-center gap-3">
                      {getFileIcon(doc.fileType)}
                      <span className="font-medium text-slate-900">{doc.name}</span>
                    </div>
                  </td>
                  <td className="py-3 px-4">
                    <span className="text-sm text-slate-600 uppercase">{doc.fileType}</span>
                  </td>
                  <td className="py-3 px-4">
                    <span className="text-sm text-slate-600">{formatFileSize(doc.fileSize)}</span>
                  </td>
                  <td className="py-3 px-4">
                    <div className="flex items-center gap-2">
                      {getStatusIcon(doc.status)}
                      <span className="text-sm text-slate-600">{getStatusText(doc.status)}</span>
                    </div>
                  </td>
                  <td className="py-3 px-4">
                    <span className="text-sm text-slate-600">
                      {new Date(doc.createdAt).toLocaleDateString('zh-CN')}
                    </span>
                  </td>
                  <td className="py-3 px-4 text-right">
                    <div className="flex items-center justify-end gap-2">
                      <button
                        onClick={() => handleEdit(doc)}
                        className="p-1.5 text-slate-400 hover:text-blue-500 hover:bg-blue-50 rounded transition-colors"
                        title="重命名"
                      >
                        <Edit2 className="w-4 h-4" />
                      </button>
                      <button
                        onClick={() => handleDelete(doc.id)}
                        className="p-1.5 text-slate-400 hover:text-red-500 hover:bg-red-50 rounded transition-colors"
                        title="删除"
                      >
                        <Trash2 className="w-4 h-4" />
                      </button>
                    </div>
                  </td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>
      )}

      {/* Edit Modal */}
      {editingDoc && (
        <div className="fixed inset-0 bg-black/50 flex items-center justify-center z-50">
          <div className="bg-white rounded-xl p-6 w-full max-w-md">
            <h3 className="text-lg font-semibold text-slate-900 mb-4">重命名文档</h3>
            <input
              type="text"
              value={editName}
              onChange={(e) => setEditName(e.target.value)}
              className="w-full px-3 py-2 border border-slate-200 rounded-lg focus:outline-none focus:ring-2 focus:ring-primary"
              autoFocus
            />
            <div className="flex gap-3 mt-6">
              <button
                onClick={() => setEditingDoc(null)}
                className="flex-1 px-4 py-2 border border-slate-200 rounded-lg text-slate-600 hover:bg-slate-50 transition-colors"
              >
                取消
              </button>
              <button
                onClick={handleUpdate}
                className="flex-1 px-4 py-2 bg-primary text-white rounded-lg hover:bg-primary-dark transition-colors"
              >
                保存
              </button>
            </div>
          </div>
        </div>
      )}
    </div>
  );
};
