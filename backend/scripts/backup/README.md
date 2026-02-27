# RoleCraft AI 备份系统

## 概述

本备份系统提供数据库的自动备份、验证、恢复和清理功能。

## 快速开始

### 1. 手动备份

```bash
# 进入备份目录
cd backend/scripts/backup

# 创建备份
./backup.sh backup

# 或使用 Go 版本
go run backup.go
```

### 2. 恢复备份

```bash
# 列出所有备份
./backup.sh list

# 恢复指定备份
./backup.sh restore backups/rolecraft_20260227_120000.db.gz
```

### 3. 验证备份

```bash
./backup.sh verify backups/rolecraft_20260227_120000.db.gz
```

## 自动备份配置

### 使用 Cron（Linux/macOS）

编辑 crontab：
```bash
crontab -e
```

添加以下行（每天凌晨 2 点备份）：
```cron
0 2 * * * /path/to/rolecraft-ai/backend/scripts/backup/backup.sh backup >> /var/log/rolecraft-backup.log 2>&1
```

### 使用 systemd Timer（Linux）

1. 创建服务文件 `/etc/systemd/system/rolecraft-backup.service`:

```ini
[Unit]
Description=RoleCraft AI Database Backup
After=network.target

[Service]
Type=oneshot
User=rolecraft
WorkingDirectory=/path/to/rolecraft-ai/backend/scripts/backup
Environment=DB_PATH=/path/to/rolecraft.db
Environment=BACKUP_DIR=/path/to/backups
Environment=RETENTION_DAYS=30
ExecStart=/path/to/rolecraft-ai/backend/scripts/backup/backup.sh backup
```

2. 创建定时器文件 `/etc/systemd/system/rolecraft-backup.timer`:

```ini
[Unit]
Description=Run RoleCraft AI Backup Daily
Requires=rolecraft-backup.service

[Timer]
OnCalendar=*-*-* 02:00:00
Persistent=true

[Install]
WantedBy=timers.target
```

3. 启用并启动定时器：

```bash
sudo systemctl daemon-reload
sudo systemctl enable rolecraft-backup.timer
sudo systemctl start rolecraft-backup.timer

# 查看状态
systemctl list-timers
```

### 使用 Windows 任务计划程序

1. 打开"任务计划程序"
2. 创建基本任务
3. 设置触发器（每天 2:00）
4. 操作：启动程序
   - 程序：`bash.exe`（需要安装 Git Bash 或 WSL）
   - 参数：`/path/to/backup.sh backup`

## 环境变量

| 变量名 | 说明 | 默认值 |
|--------|------|--------|
| `BACKUP_DIR` | 备份文件存储目录 | `./backups` |
| `DB_PATH` | 数据库文件路径 | `./rolecraft.db` |
| `DB_TYPE` | 数据库类型：sqlite/postgres | `sqlite` |
| `PG_CONNECTION` | PostgreSQL 连接字符串 | - |
| `RETENTION_DAYS` | 备份保留天数 | `30` |
| `COMPRESSION` | 是否压缩备份 | `true` |

## 备份策略建议

### 小型项目（个人使用）
- 频率：每天 1 次
- 保留：7 天
- 存储：本地磁盘

### 中型项目（团队使用）
- 频率：每天 2 次（凌晨 2 点和下午 2 点）
- 保留：30 天
- 存储：本地 + 云存储（S3/OSS）

### 大型项目（生产环境）
- 频率：每小时 1 次（增量）+ 每天 1 次（完整）
- 保留：90 天
- 存储：多地备份（本地 + 异地 + 云存储）

## 备份文件结构

```
backups/
├── rolecraft_20260227_020000.db.gz    # 压缩的 SQLite 备份
├── rolecraft_20260227_140000.db.gz
├── rolecraft_20260226_020000.db.gz
└── rolecraft_backup_20260227_020000.zip  # 完整备份（含配置）
```

## 恢复测试

**重要：** 定期测试备份恢复流程，确保备份可用。

```bash
# 1. 创建测试数据库
cp rolecraft.db rolecraft.db.backup

# 2. 恢复备份
./backup.sh restore backups/rolecraft_20260227_020000.db.gz

# 3. 验证数据
sqlite3 rolecraft.db "SELECT COUNT(*) FROM users;"

# 4. 恢复原数据库
mv rolecraft.db.backup rolecraft.db
```

## 监控和告警

### 检查备份状态

```bash
# 检查最新备份
ls -lt backups/ | head -5

# 检查备份大小（异常大小可能表示问题）
du -h backups/*.gz
```

### 备份监控脚本

创建 `monitor_backup.sh`:

```bash
#!/bin/bash
BACKUP_DIR="./backups"
MAX_AGE_HOURS=26  # 超过 26 小时没有新备份则告警

LATEST=$(ls -t $BACKUP_DIR/rolecraft_*.db.gz 2>/dev/null | head -1)
if [ -z "$LATEST" ]; then
    echo "ERROR: No backups found!"
    exit 1
fi

AGE=$(( ($(date +%s) - $(stat -c %Y "$LATEST")) / 3600 ))
if [ $AGE -gt $MAX_AGE_HOURS ]; then
    echo "WARNING: Latest backup is $AGE hours old!"
    exit 1
fi

echo "OK: Latest backup is $AGE hours old"
exit 0
```

## 故障排除

### 问题：备份失败，提示"database is locked"

**解决：** 确保没有其他进程正在写入数据库。可以在备份前暂停应用服务。

### 问题：备份文件过大

**解决：** 
1. 启用压缩（默认已启用）
2. 定期清理旧数据
3. 考虑使用增量备份

### 问题：恢复后数据不一致

**解决：**
1. 确保备份时数据库处于一致状态
2. 使用事务进行备份
3. 考虑使用 WAL 模式

## 最佳实践

1. ✅ **3-2-1 规则**：至少 3 份备份，2 种不同介质，1 份异地存储
2. ✅ **定期测试恢复**：每月至少测试一次恢复流程
3. ✅ **监控备份状态**：设置告警，确保备份按时执行
4. ✅ **加密敏感数据**：生产环境备份应加密存储
5. ✅ **文档化流程**：确保团队成员都知道如何恢复数据

## 安全注意事项

- 备份文件可能包含敏感数据，应设置适当的文件权限
- 考虑对备份文件进行加密
- 不要将备份文件提交到版本控制系统
- 云存储备份时，确保使用加密传输和存储

## 联系支持

如有问题，请联系：support@rolecraft.ai
