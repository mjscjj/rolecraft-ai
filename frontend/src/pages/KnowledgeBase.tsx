import type { FC } from 'react';
import { useState, useEffect, useRef, useCallback } from 'react';
import { 
  Upload, 
  FileText, 
  Search, 
  Trash2,
  Edit2,
  CheckCircle,
  Clock,
  AlertCircle,
  Folder,
  FolderPlus,
  Tag,
  Layers,
  ChevronRight,
  ChevronDown,
  Move,
  File,
  X,
  Download,
  Eye,
  History,
  Filter,
  SortAsc,
  SortDesc,
  Grid,
  List as ListIcon,
  Copy,
  AlertTriangle,
  Check
} from 'lucide-react';
import documentApi from '../api/document';

// ==================== 类型定义 ====================

interface Document {
  id: string;
  name: string;
  fileType: string;
  fileSize: number;
  status: string;
  filePath?: string;
  createdAt: string;
  updatedAt?: string;
  folderId?: string;
  tags?: string[];
  description?: string;
  chunkCount?: number;
  similarity?: number;
}

interface Folder {
  id: string;
  name: string;
  parentId?: string;
  children?: Folder[];
  documentCount?: number;
  createdAt: string;
}

interface SearchHistory {
  id: string;
  query: string;
  timestamp: string;
  resultCount: number;
}

interface UploadProgress {
  fileName: string;
  progress: number;
  status: 'pending' | 'uploading' | 'processing' | 'completed' | 'failed';
  message?: string;
}

// ==================== 工具函数 ====================

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
      return <Clock className="w-5 h-5 text-blue-500 animate-spin" />;
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

const highlightText = (text: string, query: string) => {
  if (!query) return text;
  const parts = text.split(new RegExp(`(${query})`, 'gi'));
  return parts.map((part, i) => 
    part.toLowerCase() === query.toLowerCase() ? 
      <mark key={i} className="bg-yellow-200 px-0.5 rounded">{part}</mark> : 
      part
  );
};

// ==================== 主组件 ====================

export const KnowledgeBase: FC = () => {
  // 文档状态
  const [documents, setDocuments] = useState<Document[]>([]);
  const [loading, setLoading] = useState(true);
  const [searchQuery, setSearchQuery] = useState('');
  const [pollingDocs, setPollingDocs] = useState<Set<string>>(new Set());
  
  // 批量操作状态
  const [selectedDocs, setSelectedDocs] = useState<Set<string>>(new Set());
  const [selectAll, setSelectAll] = useState(false);
  const [uploading, setUploading] = useState(false);
  const [uploadProgress, setUploadProgress] = useState<UploadProgress[]>([]);
  
  // 文件夹管理
  const [folders, setFolders] = useState<Folder[]>([]);
  const [currentFolderId, setCurrentFolderId] = useState<string | undefined>();
  const [showFolderTree, setShowFolderTree] = useState(false);
  const [newFolderName, setNewFolderName] = useState('');
  const [showCreateFolder, setShowCreateFolder] = useState(false);
  const [showMoveModal, setShowMoveModal] = useState(false);
  
  // 智能分类
  const [autoTags, setAutoTags] = useState<Record<string, string[]>>({});
  const [duplicateDocs, setDuplicateDocs] = useState<{id: string, similarIds: string[]}[]>([]);
  
  // 文档预览
  const [previewDoc, setPreviewDoc] = useState<Document | null>(null);
  const [previewContent, setPreviewContent] = useState('');
  const [previewLoading, setPreviewLoading] = useState(false);
  const [showVersionHistory, setShowVersionHistory] = useState(false);
  
  // 搜索优化
  const [searchFilters, setSearchFilters] = useState({
    type: '',
    status: '',
    folder: '',
    dateFrom: '',
    dateTo: ''
  });
  const [searchHistory, setSearchHistory] = useState<SearchHistory[]>([]);
  const [sortBy, setSortBy] = useState<'name' | 'date' | 'size' | 'relevance'>('date');
  const [sortOrder, setSortOrder] = useState<'asc' | 'desc'>('desc');
  const [viewMode, setViewMode] = useState<'list' | 'grid'>('list');
  
  // 编辑状态
  const [editingDoc, setEditingDoc] = useState<Document | null>(null);
  const [editName, setEditName] = useState('');
  const [editTags, setEditTags] = useState<string[]>([]);
  
  // 标签批量更新
  const [showTagModal, setShowTagModal] = useState(false);
  const [bulkTags, setBulkTags] = useState<string[]>([]);
  const [operationError, setOperationError] = useState('');
  
  const fileInputRef = useRef<HTMLInputElement>(null);
  const token = typeof window !== 'undefined' ? localStorage.getItem('token') : null;

  // ==================== 生命周期 ====================

  useEffect(() => {
    loadDocuments();
    loadFolders();
    loadSearchHistory();
  }, []);

  useEffect(() => {
    const processingDocs = documents.filter(d => d.status === 'processing');
    if (processingDocs.length === 0) return;

    const interval = setInterval(() => {
      processingDocs.forEach(doc => {
        pollDocumentStatus(doc.id);
      });
    }, 2000);

    return () => clearInterval(interval);
  }, [documents]);

  // 检测重复文档
  useEffect(() => {
    detectDuplicates();
  }, [documents]);

  // ==================== 数据加载 ====================

  const loadDocuments = async () => {
    if (!token) {
      setLoading(false);
      setOperationError('请先登录后使用知识库');
      return;
    }
    try {
      const data = await documentApi.list({
        type: searchFilters.type || undefined,
        status: searchFilters.status || undefined,
        folder: currentFolderId,
      });
      setDocuments(data);
      generateAutoTags(data);
      setOperationError('');
    } catch (err: any) {
      console.error('Failed to load documents:', err);
      setOperationError(err?.message || '加载文档失败');
    } finally {
      setLoading(false);
    }
  };

  const loadFolders = async () => {
    if (!token) return;
    try {
      const data = await documentApi.listFolders();
      setFolders(data);
      setOperationError('');
    } catch (err: any) {
      console.error('Failed to load folders:', err);
      setOperationError(err?.message || '加载文件夹失败');
    }
  };

  const loadSearchHistory = () => {
    const saved = localStorage.getItem('searchHistory');
    if (saved) {
      setSearchHistory(JSON.parse(saved));
    }
  };

  // ==================== 智能分类 ====================

  const generateAutoTags = async (docs: Document[]) => {
    const tagsMap: Record<string, string[]> = {};
    
    docs.forEach(doc => {
      const tags: string[] = [];
      
      // 基于文件名生成标签
      const nameLower = doc.name.toLowerCase();
      if (nameLower.includes('合同') || nameLower.includes('agreement')) tags.push('合同');
      if (nameLower.includes('报告') || nameLower.includes('report')) tags.push('报告');
      if (nameLower.includes('发票') || nameLower.includes('invoice')) tags.push('财务');
      if (nameLower.includes('技术') || nameLower.includes('technical')) tags.push('技术');
      if (nameLower.includes('产品') || nameLower.includes('product')) tags.push('产品');
      
      // 基于文件类型
      if (doc.fileType === 'pdf') tags.push('PDF');
      if (doc.fileType === 'doc' || doc.fileType === 'docx') tags.push('Word');
      
      tagsMap[doc.id] = tags;
    });
    
    setAutoTags(tagsMap);
  };

  const detectDuplicates = () => {
    const duplicates: {id: string, similarIds: string[]}[] = [];
    const docMap = new Map(documents.map(d => [d.name.toLowerCase(), d]));
    
    documents.forEach(doc => {
      const similarIds: string[] = [];
      documents.forEach(otherDoc => {
        if (doc.id !== otherDoc.id) {
          // 简单相似度检测：文件名相似度
          const similarity = calculateStringSimilarity(
            doc.name.toLowerCase(),
            otherDoc.name.toLowerCase()
          );
          if (similarity > 0.85) {
            similarIds.push(otherDoc.id);
          }
        }
      });
      
      if (similarIds.length > 0) {
        duplicates.push({ id: doc.id, similarIds });
      }
    });
    
    setDuplicateDocs(duplicates);
  };

  const calculateStringSimilarity = (s1: string, s2: string): number => {
    const longer = s1.length > s2.length ? s1 : s2;
    const shorter = s1.length > s2.length ? s2 : s1;
    if (longer.length === 0) return 1.0;
    const editDistance = levenshteinDistance(longer, shorter);
    return (longer.length - editDistance) / longer.length;
  };

  const levenshteinDistance = (s: string, t: string): number => {
    const m = s.length, n = t.length;
    const dp = Array(m + 1).fill(null).map(() => Array(n + 1).fill(0));
    
    for (let i = 0; i <= m; i++) dp[i][0] = i;
    for (let j = 0; j <= n; j++) dp[0][j] = j;
    
    for (let i = 1; i <= m; i++) {
      for (let j = 1; j <= n; j++) {
        if (s[i-1] === t[j-1]) dp[i][j] = dp[i-1][j-1];
        else dp[i][j] = 1 + Math.min(dp[i-1][j], dp[i][j-1], dp[i-1][j-1]);
      }
    }
    
    return dp[m][n];
  };

  // ==================== 批量操作 ====================

  const handleFileSelect = async (e: React.ChangeEvent<HTMLInputElement>) => {
    const files = Array.from(e.target.files || []);
    if (files.length === 0 || !token) return;

    setUploading(true);
    const uploads: UploadProgress[] = files.map(f => ({
      fileName: f.name,
      progress: 0,
      status: 'pending'
    }));
    setUploadProgress(uploads);

    for (let i = 0; i < files.length; i++) {
      const file = files[i];
      updateUploadProgress(i, { status: 'uploading', progress: 10 });
      
      try {
        updateUploadProgress(i, { progress: 50 });
        const data = await documentApi.uploadWithFolder(file, currentFolderId);
        const uploadedDoc = Array.isArray(data) ? data[0] : data;
        if (uploadedDoc) {
          updateUploadProgress(i, { 
            status: 'processing', 
            progress: 75,
            message: '正在处理...'
          });
          setDocuments(prev => [uploadedDoc as Document, ...prev]);
        }
      } catch (err: any) {
        updateUploadProgress(i, { 
          status: 'failed', 
          message: err?.message || '上传失败'
        });
        setOperationError(err?.message || '上传失败');
      }
    }

    setUploading(false);
    setTimeout(() => setUploadProgress([]), 3000);
    if (fileInputRef.current) fileInputRef.current.value = '';
  };

  const updateUploadProgress = (index: number, updates: Partial<UploadProgress>) => {
    setUploadProgress(prev => prev.map((item, i) => 
      i === index ? { ...item, ...updates } : item
    ));
  };

  const handleBulkDelete = async () => {
    if (!token || selectedDocs.size === 0) return;
    if (!confirm(`确定要删除选中的 ${selectedDocs.size} 个文档吗？`)) return;

    try {
      await documentApi.batchDelete(Array.from(selectedDocs));
      setDocuments(prev => prev.filter(d => !selectedDocs.has(d.id)));
      setSelectedDocs(new Set());
      setSelectAll(false);
      setOperationError('');
    } catch (err: any) {
      console.error('Failed to bulk delete:', err);
      setOperationError(err?.message || '批量删除失败');
    }
  };

  const handleBulkMove = () => {
    if (selectedDocs.size === 0) return;
    setShowMoveModal(true);
  };

  const handleBulkTagUpdate = () => {
    if (selectedDocs.size === 0) return;
    setShowTagModal(true);
  };

  const applyBulkTags = async () => {
    if (!token || selectedDocs.size === 0) return;

    try {
      await documentApi.batchUpdateTags(Array.from(selectedDocs), bulkTags);
      setDocuments(prev => prev.map(d => 
        selectedDocs.has(d.id) ? { ...d, tags: bulkTags } : d
      ));
      setShowTagModal(false);
      setBulkTags([]);
      setOperationError('');
    } catch (err: any) {
      console.error('Failed to update tags:', err);
      setOperationError(err?.message || '批量更新标签失败');
    }
  };

  const toggleSelectDoc = (docId: string) => {
    const newSelected = new Set(selectedDocs);
    if (newSelected.has(docId)) {
      newSelected.delete(docId);
    } else {
      newSelected.add(docId);
    }
    setSelectedDocs(newSelected);
    setSelectAll(newSelected.size === documents.length);
  };

  const toggleSelectAll = () => {
    if (selectAll) {
      setSelectedDocs(new Set());
    } else {
      setSelectedDocs(new Set(documents.map(d => d.id)));
    }
    setSelectAll(!selectAll);
  };

  // ==================== 文件夹管理 ====================

  const createFolder = async () => {
    if (!token || !newFolderName.trim()) return;

    try {
      await documentApi.createFolder(newFolderName, currentFolderId);
      await loadFolders();
      setShowCreateFolder(false);
      setNewFolderName('');
      setOperationError('');
    } catch (err: any) {
      console.error('Failed to create folder:', err);
      setOperationError(err?.message || '创建文件夹失败');
    }
  };

  const deleteFolder = async (folderId: string) => {
    if (!token || !confirm('确定要删除这个文件夹吗？')) return;

    try {
      await documentApi.deleteFolder(folderId);
      await loadFolders();
      setOperationError('');
    } catch (err: any) {
      console.error('Failed to delete folder:', err);
      setOperationError(err?.message || '删除文件夹失败');
    }
  };

  const moveDocumentsToFolder = async (folderId: string) => {
    if (!token) return;

    try {
      await documentApi.batchMove(Array.from(selectedDocs), folderId);
      setDocuments(prev => prev.map(d => 
        selectedDocs.has(d.id) ? { ...d, folderId } : d
      ));
      setShowMoveModal(false);
      setSelectedDocs(new Set());
      setOperationError('');
    } catch (err: any) {
      console.error('Failed to move documents:', err);
      setOperationError(err?.message || '批量移动失败');
    }
  };

  // ==================== 文档预览 ====================

  const previewDocument = async (doc: Document) => {
    setPreviewDoc(doc);
    setPreviewLoading(true);
    
    try {
      const data = await documentApi.preview(doc.id, doc.fileType);

      if (doc.fileType === 'pdf' && data.blob) {
        // PDF 预览
        const blob = data.blob;
        const url = URL.createObjectURL(blob);
        setPreviewContent(url);
      } else {
        // 文本类预览
        setPreviewContent(data.content || '');
      }
      setOperationError('');
    } catch (err: any) {
      console.error('Failed to preview:', err);
      setPreviewContent('无法加载预览');
      setOperationError(err?.message || '文档预览失败');
    } finally {
      setPreviewLoading(false);
    }
  };

  const downloadDocument = async (doc: Document) => {
    if (!token) return;
    
    try {
      const blob = await documentApi.download(doc.id);
      const url = window.URL.createObjectURL(blob);
      const a = document.createElement('a');
      a.href = url;
      a.download = doc.name;
      a.click();
      window.URL.revokeObjectURL(url);
      setOperationError('');
    } catch (err: any) {
      console.error('Failed to download:', err);
      setOperationError(err?.message || '下载失败');
    }
  };

  // ==================== 搜索优化 ====================

  const handleSearch = useCallback(async () => {
    if (!token || !searchQuery.trim()) {
      loadDocuments();
      return;
    }

    try {
      const data = await documentApi.search({
        query: searchQuery,
        filters: searchFilters,
        sortBy,
        sortOrder,
      });
      setDocuments(data.documents);

      const historyItem: SearchHistory = {
        id: Date.now().toString(),
        query: searchQuery,
        timestamp: new Date().toISOString(),
        resultCount: data.documents.length
      };

      const newHistory = [historyItem, ...searchHistory].slice(0, 10);
      setSearchHistory(newHistory);
      localStorage.setItem('searchHistory', JSON.stringify(newHistory));
      setOperationError('');
    } catch (err: any) {
      console.error('Failed to search:', err);
      setOperationError(err?.message || '搜索失败');
    }
  }, [searchQuery, searchFilters, sortBy, sortOrder, token]);

  useEffect(() => {
    const timer = setTimeout(handleSearch, 300);
    return () => clearTimeout(timer);
  }, [handleSearch]);

  // ==================== 文档状态轮询 ====================

  const pollDocumentStatus = async (docId: string) => {
    if (!token || pollingDocs.has(docId)) return;
    
    setPollingDocs(prev => new Set(prev).add(docId));
    
    try {
      const data = await documentApi.getStatus(docId);
      setDocuments(prev => prev.map(doc => 
        doc.id === docId 
          ? { ...doc, status: data.status, updatedAt: data.updatedAt }
          : doc
      ));
    } catch (err) {
      console.error('Failed to poll status:', err);
    } finally {
      setPollingDocs(prev => {
        const next = new Set(prev);
        next.delete(docId);
        return next;
      });
    }
  };

  // ==================== 删除操作 ====================

  const handleDelete = async (docId: string) => {
    if (!token || !confirm('确定要删除这个文档吗？')) return;

    try {
      await documentApi.delete(docId);
      setDocuments(documents.filter(d => d.id !== docId));
      setOperationError('');
    } catch (err: any) {
      console.error('Failed to delete:', err);
      setOperationError(err?.message || '删除文档失败');
    }
  };

  // ==================== 编辑操作 ====================

  const handleEdit = (doc: Document) => {
    setEditingDoc(doc);
    setEditName(doc.name);
    setEditTags(doc.tags || autoTags[doc.id] || []);
  };

  const handleUpdate = async () => {
    if (!token || !editingDoc) return;

    try {
      const data = await documentApi.update(editingDoc.id, {
        name: editName,
        tags: editTags,
      });
      setDocuments(documents.map(d => 
        d.id === editingDoc.id ? { ...d, ...data } : d
      ));
      setEditingDoc(null);
      setOperationError('');
    } catch (err: any) {
      console.error('Failed to update:', err);
      setOperationError(err?.message || '更新文档失败');
    }
  };

  // ==================== 过滤和排序 ====================

  const filteredDocs = documents.filter(doc => {
    const matchesSearch = doc.name.toLowerCase().includes(searchQuery.toLowerCase());
    const matchesType = !searchFilters.type || doc.fileType === searchFilters.type;
    const matchesStatus = !searchFilters.status || doc.status === searchFilters.status;
    const matchesFolder = !currentFolderId || doc.folderId === currentFolderId;
    return matchesSearch && matchesType && matchesStatus && matchesFolder;
  });

  const sortedDocs = [...filteredDocs].sort((a, b) => {
    let comparison = 0;
    switch (sortBy) {
      case 'name':
        comparison = a.name.localeCompare(b.name);
        break;
      case 'size':
        comparison = a.fileSize - b.fileSize;
        break;
      case 'date':
        comparison = new Date(a.createdAt).getTime() - new Date(b.createdAt).getTime();
        break;
      case 'relevance':
        comparison = (b.similarity || 0) - (a.similarity || 0);
        break;
    }
    return sortOrder === 'asc' ? comparison : -comparison;
  });

  const stats = {
    total: documents.length,
    completed: documents.filter(d => d.status === 'completed').length,
    processing: documents.filter(d => d.status === 'processing').length,
    failed: documents.filter(d => d.status === 'failed').length,
    selected: selectedDocs.size,
  };

  // ==================== 渲染 ====================

  return (
    <div className="space-y-6">
      {/* Header */}
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-2xl font-bold text-slate-900">知识库</h1>
          <p className="text-slate-500 mt-1">管理你的文档和知识资源</p>
        </div>
        <div className="flex items-center gap-2">
          <label className="flex items-center gap-2 px-4 py-2 bg-primary text-white rounded-lg hover:bg-primary-dark transition-colors cursor-pointer">
            <Upload className="w-5 h-5" />
            <span>{uploading ? '上传中...' : '上传文档'}</span>
            <input
              ref={fileInputRef}
              type="file"
              accept=".pdf,.doc,.docx,.txt,.md"
              multiple
              onChange={handleFileSelect}
              disabled={uploading}
              className="hidden"
            />
          </label>
        </div>
      </div>

      {operationError && (
        <div className="flex items-center justify-between rounded-xl border border-red-200 bg-red-50 px-4 py-3 text-sm text-red-700">
          <span>{operationError}</span>
          <button
            onClick={() => {
              setOperationError('');
              loadDocuments();
              loadFolders();
            }}
            className="rounded-md bg-white px-3 py-1 text-red-600 hover:bg-red-100"
          >
            重试
          </button>
        </div>
      )}

      {/* Stats */}
      <div className="grid grid-cols-5 gap-4">
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
        {stats.selected > 0 && (
          <div className="bg-blue-50 p-4 rounded-xl border border-blue-200">
            <p className="text-sm text-blue-600">已选择</p>
            <p className="text-2xl font-bold text-blue-700">{stats.selected}</p>
          </div>
        )}
      </div>

      {/* 批量操作工具栏 */}
      {selectedDocs.size > 0 && (
        <div className="bg-blue-50 border border-blue-200 rounded-xl p-4 flex items-center justify-between">
          <span className="text-blue-700 font-medium">已选择 {selectedDocs.size} 个文档</span>
          <div className="flex items-center gap-2">
            <button
              onClick={handleBulkTagUpdate}
              className="px-3 py-1.5 bg-white border border-blue-200 rounded-lg text-blue-600 hover:bg-blue-100 transition-colors flex items-center gap-2"
            >
              <Tag className="w-4 h-4" />
              批量标签
            </button>
            <button
              onClick={handleBulkMove}
              className="px-3 py-1.5 bg-white border border-blue-200 rounded-lg text-blue-600 hover:bg-blue-100 transition-colors flex items-center gap-2"
            >
              <Move className="w-4 h-4" />
              批量移动
            </button>
            <button
              onClick={handleBulkDelete}
              className="px-3 py-1.5 bg-red-500 text-white rounded-lg hover:bg-red-600 transition-colors flex items-center gap-2"
            >
              <Trash2 className="w-4 h-4" />
              批量删除
            </button>
          </div>
        </div>
      )}

      {/* 搜索和过滤 */}
      <div className="bg-white p-4 rounded-xl border border-slate-200 space-y-4">
        <div className="flex items-center gap-4">
          <div className="relative flex-1">
            <Search className="absolute left-3 top-1/2 -translate-y-1/2 w-5 h-5 text-slate-400" />
            <input
              type="text"
              placeholder="搜索文档..."
              value={searchQuery}
              onChange={(e) => setSearchQuery(e.target.value)}
              className="w-full pl-10 pr-4 py-2 border border-slate-200 rounded-lg focus:outline-none focus:ring-2 focus:ring-primary"
            />
          </div>
          
          <button
            onClick={() => setShowFolderTree(!showFolderTree)}
            className="px-3 py-2 border border-slate-200 rounded-lg hover:bg-slate-50 flex items-center gap-2"
          >
            <Folder className="w-5 h-5" />
            文件夹
          </button>
          
          <select
            value={searchFilters.type}
            onChange={(e) => setSearchFilters({...searchFilters, type: e.target.value})}
            className="px-3 py-2 border border-slate-200 rounded-lg focus:outline-none"
          >
            <option value="">所有类型</option>
            <option value="pdf">PDF</option>
            <option value="doc">Word</option>
            <option value="txt">文本</option>
          </select>
          
          <select
            value={sortBy}
            onChange={(e) => setSortBy(e.target.value as any)}
            className="px-3 py-2 border border-slate-200 rounded-lg focus:outline-none"
          >
            <option value="date">按日期</option>
            <option value="name">按名称</option>
            <option value="size">按大小</option>
          </select>
          
          <button
            onClick={() => setSortOrder(sortOrder === 'asc' ? 'desc' : 'asc')}
            className="p-2 border border-slate-200 rounded-lg hover:bg-slate-50"
          >
            {sortOrder === 'asc' ? <SortAsc className="w-5 h-5" /> : <SortDesc className="w-5 h-5" />}
          </button>
          
          <div className="flex items-center gap-1 border border-slate-200 rounded-lg p-1">
            <button
              onClick={() => setViewMode('list')}
              className={`p-1.5 rounded ${viewMode === 'list' ? 'bg-slate-100' : ''}`}
            >
              <ListIcon className="w-4 h-4" />
            </button>
            <button
              onClick={() => setViewMode('grid')}
              className={`p-1.5 rounded ${viewMode === 'grid' ? 'bg-slate-100' : ''}`}
            >
              <Grid className="w-4 h-4" />
            </button>
          </div>
        </div>
        
        {/* 文件夹树 */}
        {showFolderTree && (
          <div className="border-t pt-4">
            <div className="flex items-center justify-between mb-2">
              <h3 className="font-medium text-slate-700">文件夹</h3>
              <button
                onClick={() => setShowCreateFolder(true)}
                className="text-blue-600 hover:text-blue-700 flex items-center gap-1 text-sm"
              >
                <FolderPlus className="w-4 h-4" />
                新建文件夹
              </button>
            </div>
            <div className="space-y-1">
              <button
                onClick={() => setCurrentFolderId(undefined)}
                className={`w-full text-left px-3 py-2 rounded-lg flex items-center gap-2 ${!currentFolderId ? 'bg-blue-50 text-blue-600' : 'hover:bg-slate-50'}`}
              >
                <Folder className="w-4 h-4" />
                全部文档
              </button>
              {folders.map(folder => (
                <button
                  key={folder.id}
                  onClick={() => setCurrentFolderId(folder.id)}
                  className={`w-full text-left px-3 py-2 rounded-lg flex items-center justify-between ${currentFolderId === folder.id ? 'bg-blue-50 text-blue-600' : 'hover:bg-slate-50'}`}
                >
                  <div className="flex items-center gap-2">
                    <ChevronRight className="w-4 h-4" />
                    <Folder className="w-4 h-4" />
                    {folder.name}
                  </div>
                  <span className="text-xs text-slate-400">{folder.documentCount} 个文档</span>
                </button>
              ))}
            </div>
          </div>
        )}
      </div>

      {/* 上传进度 */}
      {uploadProgress.length > 0 && (
        <div className="bg-white rounded-xl border border-slate-200 p-4 space-y-2">
          <h3 className="font-medium text-slate-700">上传进度</h3>
          {uploadProgress.map((upload, i) => (
            <div key={i} className="flex items-center gap-3">
              <File className="w-4 h-4 text-slate-400" />
              <span className="flex-1 text-sm text-slate-600">{upload.fileName}</span>
              <div className="w-32 bg-slate-200 rounded-full h-2 overflow-hidden">
                <div 
                  className={`h-2 rounded-full transition-all ${
                    upload.status === 'failed' ? 'bg-red-500' :
                    upload.status === 'completed' ? 'bg-green-500' : 'bg-blue-500'
                  }`}
                  style={{ width: `${upload.progress}%` }}
                />
              </div>
              <span className="text-xs text-slate-500 w-20">
                {upload.message || `${upload.progress}%`}
              </span>
            </div>
          ))}
        </div>
      )}

      {/* 重复文档提示 */}
      {duplicateDocs.length > 0 && (
        <div className="bg-yellow-50 border border-yellow-200 rounded-xl p-4">
          <div className="flex items-start gap-3">
            <AlertTriangle className="w-5 h-5 text-yellow-600 mt-0.5" />
            <div>
              <h3 className="font-medium text-yellow-800">发现重复文档</h3>
              <p className="text-sm text-yellow-700 mt-1">
                检测到 {duplicateDocs.length} 个文档可能存在重复，建议检查并清理。
              </p>
            </div>
          </div>
        </div>
      )}

      {/* 文档列表 */}
      {loading ? (
        <div className="text-center py-20 text-slate-500">加载中...</div>
      ) : sortedDocs.length === 0 ? (
        <div className="text-center py-20 text-slate-500">
          <FileText className="w-16 h-16 mx-auto mb-4 opacity-20" />
          <p>暂无文档</p>
          <p className="text-sm mt-2">上传 PDF、Word、TXT 等格式文档</p>
        </div>
      ) : viewMode === 'list' ? (
        <div className="bg-white rounded-xl border border-slate-200 overflow-hidden">
          <table className="w-full">
            <thead className="bg-slate-50 border-b border-slate-200">
              <tr>
                <th className="text-left py-3 px-4">
                  <input
                    type="checkbox"
                    checked={selectAll}
                    onChange={toggleSelectAll}
                    className="rounded border-slate-300"
                  />
                </th>
                <th className="text-left py-3 px-4 text-sm font-medium text-slate-600">文档名称</th>
                <th className="text-left py-3 px-4 text-sm font-medium text-slate-600">类型</th>
                <th className="text-left py-3 px-4 text-sm font-medium text-slate-600">大小</th>
                <th className="text-left py-3 px-4 text-sm font-medium text-slate-600">标签</th>
                <th className="text-left py-3 px-4 text-sm font-medium text-slate-600">状态</th>
                <th className="text-left py-3 px-4 text-sm font-medium text-slate-600">上传时间</th>
                <th className="text-right py-3 px-4 text-sm font-medium text-slate-600">操作</th>
              </tr>
            </thead>
            <tbody>
              {sortedDocs.map((doc) => (
                <tr key={doc.id} className="border-b border-slate-100 hover:bg-slate-50">
                  <td className="py-3 px-4">
                    <input
                      type="checkbox"
                      checked={selectedDocs.has(doc.id)}
                      onChange={() => toggleSelectDoc(doc.id)}
                      className="rounded border-slate-300"
                    />
                  </td>
                  <td className="py-3 px-4">
                    <div className="flex items-center gap-3">
                      {getFileIcon(doc.fileType)}
                      <div>
                        <span className="font-medium text-slate-900">
                          {highlightText(doc.name, searchQuery)}
                        </span>
                        {duplicateDocs.find(d => d.id === doc.id) && (
                          <div className="flex items-center gap-1 mt-1">
                            <AlertTriangle className="w-3 h-3 text-yellow-600" />
                            <span className="text-xs text-yellow-600">可能重复</span>
                          </div>
                        )}
                      </div>
                    </div>
                  </td>
                  <td className="py-3 px-4">
                    <span className="text-sm text-slate-600 uppercase">{doc.fileType}</span>
                  </td>
                  <td className="py-3 px-4">
                    <span className="text-sm text-slate-600">{formatFileSize(doc.fileSize)}</span>
                  </td>
                  <td className="py-3 px-4">
                    <div className="flex items-center gap-1 flex-wrap max-w-xs">
                      {(doc.tags || autoTags[doc.id] || []).slice(0, 3).map((tag, i) => (
                        <span key={i} className="px-2 py-0.5 bg-blue-50 text-blue-600 text-xs rounded-full">
                          {tag}
                        </span>
                      ))}
                    </div>
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
                        onClick={() => previewDocument(doc)}
                        className="p-1.5 text-slate-400 hover:text-blue-500 hover:bg-blue-50 rounded transition-colors"
                        title="预览"
                      >
                        <Eye className="w-4 h-4" />
                      </button>
                      <button
                        onClick={() => handleEdit(doc)}
                        className="p-1.5 text-slate-400 hover:text-blue-500 hover:bg-blue-50 rounded transition-colors"
                        title="重命名"
                      >
                        <Edit2 className="w-4 h-4" />
                      </button>
                      <button
                        onClick={() => downloadDocument(doc)}
                        className="p-1.5 text-slate-400 hover:text-green-500 hover:bg-green-50 rounded transition-colors"
                        title="下载"
                      >
                        <Download className="w-4 h-4" />
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
      ) : (
        <div className="grid grid-cols-3 gap-4">
          {sortedDocs.map((doc) => (
            <div key={doc.id} className="bg-white rounded-xl border border-slate-200 p-4 hover:shadow-lg transition-shadow">
              <div className="flex items-start justify-between mb-3">
                <input
                  type="checkbox"
                  checked={selectedDocs.has(doc.id)}
                  onChange={() => toggleSelectDoc(doc.id)}
                  className="rounded border-slate-300"
                />
                <div className="flex items-center gap-1">
                  <button
                    onClick={() => previewDocument(doc)}
                    className="p-1.5 text-slate-400 hover:text-blue-500 rounded transition-colors"
                  >
                    <Eye className="w-4 h-4" />
                  </button>
                  <button
                    onClick={() => handleDelete(doc.id)}
                    className="p-1.5 text-slate-400 hover:text-red-500 rounded transition-colors"
                  >
                    <Trash2 className="w-4 h-4" />
                  </button>
                </div>
              </div>
              <div className="flex items-center gap-3 mb-3">
                {getFileIcon(doc.fileType)}
                <div className="flex-1 min-w-0">
                  <h3 className="font-medium text-slate-900 truncate" title={doc.name}>
                    {highlightText(doc.name, searchQuery)}
                  </h3>
                  <p className="text-xs text-slate-500">{formatFileSize(doc.fileSize)}</p>
                </div>
              </div>
              <div className="flex items-center gap-2 mb-3">
                {getStatusIcon(doc.status)}
                <span className="text-xs text-slate-600">{getStatusText(doc.status)}</span>
              </div>
              <div className="flex items-center gap-1 flex-wrap">
                {(doc.tags || autoTags[doc.id] || []).slice(0, 3).map((tag, i) => (
                  <span key={i} className="px-2 py-0.5 bg-blue-50 text-blue-600 text-xs rounded-full">
                    {tag}
                  </span>
                ))}
              </div>
            </div>
          ))}
        </div>
      )}

      {/* 编辑弹窗 */}
      {editingDoc && (
        <div className="fixed inset-0 bg-black/50 flex items-center justify-center z-50">
          <div className="bg-white rounded-xl p-6 w-full max-w-md">
            <h3 className="text-lg font-semibold text-slate-900 mb-4">编辑文档</h3>
            <div className="space-y-4">
              <div>
                <label className="block text-sm font-medium text-slate-700 mb-1">名称</label>
                <input
                  type="text"
                  value={editName}
                  onChange={(e) => setEditName(e.target.value)}
                  className="w-full px-3 py-2 border border-slate-200 rounded-lg focus:outline-none focus:ring-2 focus:ring-primary"
                  autoFocus
                />
              </div>
              <div>
                <label className="block text-sm font-medium text-slate-700 mb-1">标签</label>
                <div className="flex items-center gap-2 flex-wrap">
                  {editTags.map((tag, i) => (
                    <span key={i} className="px-2 py-1 bg-blue-50 text-blue-600 text-sm rounded-lg flex items-center gap-1">
                      {tag}
                      <button onClick={() => setEditTags(editTags.filter((_, idx) => idx !== i))}>
                        <X className="w-3 h-3" />
                      </button>
                    </span>
                  ))}
                  <input
                    type="text"
                    placeholder="添加标签..."
                    className="flex-1 px-3 py-1.5 border border-slate-200 rounded-lg focus:outline-none text-sm"
                    onKeyDown={(e) => {
                      if (e.key === 'Enter' && e.currentTarget.value.trim()) {
                        setEditTags([...editTags, e.currentTarget.value.trim()]);
                        e.currentTarget.value = '';
                      }
                    }}
                  />
                </div>
              </div>
            </div>
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

      {/* 预览弹窗 */}
      {previewDoc && (
        <div className="fixed inset-0 bg-black/50 flex items-center justify-center z-50">
          <div className="bg-white rounded-xl w-full max-w-4xl max-h-[80vh] flex flex-col">
            <div className="flex items-center justify-between p-4 border-b">
              <h3 className="text-lg font-semibold text-slate-900">{previewDoc.name}</h3>
              <button
                onClick={() => setPreviewDoc(null)}
                className="p-2 hover:bg-slate-100 rounded-lg"
              >
                <X className="w-5 h-5" />
              </button>
            </div>
            <div className="flex-1 overflow-auto p-4">
              {previewLoading ? (
                <div className="text-center py-20 text-slate-500">加载中...</div>
              ) : previewDoc.fileType === 'pdf' ? (
                <iframe src={previewContent} className="w-full h-[60vh]" />
              ) : (
                <pre className="bg-slate-50 p-4 rounded-lg overflow-auto text-sm font-mono">
                  {previewContent}
                </pre>
              )}
            </div>
          </div>
        </div>
      )}

      {/* 批量移动弹窗 */}
      {showMoveModal && (
        <div className="fixed inset-0 bg-black/50 flex items-center justify-center z-50">
          <div className="bg-white rounded-xl p-6 w-full max-w-md">
            <h3 className="text-lg font-semibold text-slate-900 mb-4">移动到文件夹</h3>
            <div className="space-y-2">
              {folders.map(folder => (
                <button
                  key={folder.id}
                  onClick={() => moveDocumentsToFolder(folder.id)}
                  className="w-full text-left px-4 py-3 hover:bg-slate-50 rounded-lg flex items-center gap-3"
                >
                  <Folder className="w-5 h-5 text-slate-400" />
                  {folder.name}
                </button>
              ))}
            </div>
            <button
              onClick={() => setShowMoveModal(false)}
              className="w-full mt-4 px-4 py-2 border border-slate-200 rounded-lg text-slate-600 hover:bg-slate-50 transition-colors"
            >
              取消
            </button>
          </div>
        </div>
      )}

      {/* 批量标签弹窗 */}
      {showTagModal && (
        <div className="fixed inset-0 bg-black/50 flex items-center justify-center z-50">
          <div className="bg-white rounded-xl p-6 w-full max-w-md">
            <h3 className="text-lg font-semibold text-slate-900 mb-4">批量更新标签</h3>
            <div className="flex items-center gap-2 flex-wrap mb-4">
              {bulkTags.map((tag, i) => (
                <span key={i} className="px-2 py-1 bg-blue-50 text-blue-600 text-sm rounded-lg flex items-center gap-1">
                  {tag}
                  <button onClick={() => setBulkTags(bulkTags.filter((_, idx) => idx !== i))}>
                    <X className="w-3 h-3" />
                  </button>
                </span>
              ))}
              <input
                type="text"
                placeholder="添加标签..."
                className="flex-1 px-3 py-1.5 border border-slate-200 rounded-lg focus:outline-none text-sm"
                onKeyDown={(e) => {
                  if (e.key === 'Enter' && e.currentTarget.value.trim()) {
                    setBulkTags([...bulkTags, e.currentTarget.value.trim()]);
                    e.currentTarget.value = '';
                  }
                }}
              />
            </div>
            <div className="flex gap-3">
              <button
                onClick={() => setShowTagModal(false)}
                className="flex-1 px-4 py-2 border border-slate-200 rounded-lg text-slate-600 hover:bg-slate-50 transition-colors"
              >
                取消
              </button>
              <button
                onClick={applyBulkTags}
                className="flex-1 px-4 py-2 bg-primary text-white rounded-lg hover:bg-primary-dark transition-colors"
              >
                应用
              </button>
            </div>
          </div>
        </div>
      )}

      {/* 创建文件夹弹窗 */}
      {showCreateFolder && (
        <div className="fixed inset-0 bg-black/50 flex items-center justify-center z-50">
          <div className="bg-white rounded-xl p-6 w-full max-w-md">
            <h3 className="text-lg font-semibold text-slate-900 mb-4">新建文件夹</h3>
            <input
              type="text"
              value={newFolderName}
              onChange={(e) => setNewFolderName(e.target.value)}
              placeholder="文件夹名称"
              className="w-full px-3 py-2 border border-slate-200 rounded-lg focus:outline-none focus:ring-2 focus:ring-primary"
              autoFocus
              onKeyDown={(e) => e.key === 'Enter' && createFolder()}
            />
            <div className="flex gap-3 mt-6">
              <button
                onClick={() => {
                  setShowCreateFolder(false);
                  setNewFolderName('');
                }}
                className="flex-1 px-4 py-2 border border-slate-200 rounded-lg text-slate-600 hover:bg-slate-50 transition-colors"
              >
                取消
              </button>
              <button
                onClick={createFolder}
                className="flex-1 px-4 py-2 bg-primary text-white rounded-lg hover:bg-primary-dark transition-colors"
              >
                创建
              </button>
            </div>
          </div>
        </div>
      )}
    </div>
  );
};
