-- 知识库管理增强 - 数据库迁移脚本
-- 版本：v2.0
-- 日期：2026-02-27

-- 创建 folders 表
CREATE TABLE IF NOT EXISTS folders (
  id TEXT PRIMARY KEY,
  user_id TEXT NOT NULL,
  name TEXT NOT NULL,
  parent_id TEXT,
  created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
  updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
  FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
);

-- 为 folders 表添加索引
CREATE INDEX IF NOT EXISTS idx_folders_user_id ON folders(user_id);
CREATE INDEX IF NOT EXISTS idx_folders_parent_id ON folders(parent_id);

-- 为 documents 表添加 folder_id 字段
ALTER TABLE documents ADD COLUMN folder_id TEXT;

-- 为 folder_id 添加索引
CREATE INDEX IF NOT EXISTS idx_documents_folder_id ON documents(folder_id);

-- 为 documents 表添加 tags 支持 (通过 metadata 字段，无需额外字段)
-- metadata 字段已存在，存储 JSON 格式的标签等信息

-- 添加外键约束 (SQLite 3.35.0+ 支持)
-- ALTER TABLE documents ADD CONSTRAINT fk_documents_folder 
--   FOREIGN KEY (folder_id) REFERENCES folders(id) ON DELETE SET NULL;

-- 插入示例数据 (可选)
-- INSERT INTO folders (id, user_id, name, parent_id) 
-- VALUES ('folder-001', 'user-001', '项目文档', NULL);

-- 验证迁移
-- SELECT name FROM sqlite_master WHERE type='table' AND name='folders';
-- PRAGMA table_info(documents);
