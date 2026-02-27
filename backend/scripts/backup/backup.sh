#!/bin/bash

# RoleCraft AI 数据库备份脚本
# 支持 SQLite 和 PostgreSQL 数据库备份

set -e

# 配置
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
BACKUP_DIR="${BACKUP_DIR:-$SCRIPT_DIR/backups}"
DB_PATH="${DB_PATH:-$SCRIPT_DIR/../../rolecraft.db}"
DB_TYPE="${DB_TYPE:-sqlite}"  # sqlite 或 postgres
PG_CONNECTION="${PG_CONNECTION:-}"  # PostgreSQL 连接字符串
RETENTION_DAYS="${RETENTION_DAYS:-30}"  # 保留天数
COMPRESSION="${COMPRESSION:-true}"  # 是否压缩

# 颜色输出
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# 日志函数
log_info() {
    echo -e "${GREEN}[INFO]${NC} $(date '+%Y-%m-%d %H:%M:%S') - $1"
}

log_warn() {
    echo -e "${YELLOW}[WARN]${NC} $(date '+%Y-%m-%d %H:%M:%S') - $1"
}

log_error() {
    echo -e "${RED}[ERROR]${NC} $(date '+%Y-%m-%d %H:%M:%S') - $1"
}

# 创建备份目录
create_backup_dir() {
    if [ ! -d "$BACKUP_DIR" ]; then
        mkdir -p "$BACKUP_DIR"
        log_info "创建备份目录：$BACKUP_DIR"
    fi
}

# 备份 SQLite 数据库
backup_sqlite() {
    local timestamp=$(date '+%Y%m%d_%H%M%S')
    local backup_file="$BACKUP_DIR/rolecraft_${timestamp}.db"
    
    if [ ! -f "$DB_PATH" ]; then
        log_error "数据库文件不存在：$DB_PATH"
        exit 1
    fi
    
    log_info "开始备份 SQLite 数据库..."
    log_info "源文件：$DB_PATH"
    log_info "目标文件：$backup_file"
    
    # 复制数据库文件
    cp "$DB_PATH" "$backup_file"
    
    # 验证备份
    if [ -f "$backup_file" ]; then
        local size=$(du -h "$backup_file" | cut -f1)
        log_info "备份完成，文件大小：$size"
        
        # 压缩备份
        if [ "$COMPRESSION" = "true" ]; then
            log_info "压缩备份文件..."
            gzip "$backup_file"
            backup_file="${backup_file}.gz"
            local compressed_size=$(du -h "$backup_file" | cut -f1)
            log_info "压缩完成，压缩后大小：$compressed_size"
        fi
        
        echo "$backup_file"
        return 0
    else
        log_error "备份文件创建失败"
        return 1
    fi
}

# 备份 PostgreSQL 数据库
backup_postgres() {
    local timestamp=$(date '+%Y%m%d_%H%M%S')
    local backup_file="$BACKUP_DIR/rolecraft_${timestamp}.sql"
    
    if [ -z "$PG_CONNECTION" ]; then
        log_error "PostgreSQL 连接字符串未设置"
        exit 1
    fi
    
    log_info "开始备份 PostgreSQL 数据库..."
    
    # 使用 pg_dump 备份
    PGPASSWORD=$(echo "$PG_CONNECTION" | grep -oP 'password=\K[^ ]+') \
    pg_dump "$PG_CONNECTION" > "$backup_file"
    
    # 验证备份
    if [ -f "$backup_file" ] && [ -s "$backup_file" ]; then
        local size=$(du -h "$backup_file" | cut -f1)
        log_info "备份完成，文件大小：$size"
        
        # 压缩备份
        if [ "$COMPRESSION" = "true" ]; then
            log_info "压缩备份文件..."
            gzip "$backup_file"
            backup_file="${backup_file}.gz"
            local compressed_size=$(du -h "$backup_file" | cut -f1)
            log_info "压缩完成，压缩后大小：$compressed_size"
        fi
        
        echo "$backup_file"
        return 0
    else
        log_error "备份文件创建失败或为空"
        return 1
    fi
}

# 验证备份文件
verify_backup() {
    local backup_file="$1"
    
    log_info "验证备份文件：$backup_file"
    
    if [ ! -f "$backup_file" ]; then
        log_error "备份文件不存在"
        return 1
    fi
    
    # 如果是压缩文件，先解压验证
    if [[ "$backup_file" == *.gz ]]; then
        if ! gzip -t "$backup_file" 2>/dev/null; then
            log_error "压缩文件验证失败"
            return 1
        fi
        log_info "压缩文件完整性验证通过"
    fi
    
    # 对于 SQLite，验证数据库完整性
    if [[ "$backup_file" == *.db ]] || [[ "$backup_file" == *.db.gz ]]; then
        local temp_db=$(mktemp)
        
        if [[ "$backup_file" == *.gz ]]; then
            gunzip -c "$backup_file" > "$temp_db"
        else
            cp "$backup_file" "$temp_db"
        fi
        
        # SQLite 完整性检查
        if command -v sqlite3 &> /dev/null; then
            local integrity=$(sqlite3 "$temp_db" "PRAGMA integrity_check;" 2>&1)
            if [ "$integrity" = "ok" ]; then
                log_info "数据库完整性检查通过"
            else
                log_error "数据库完整性检查失败：$integrity"
                rm -f "$temp_db"
                return 1
            fi
        else
            log_warn "sqlite3 未安装，跳过完整性检查"
        fi
        
        rm -f "$temp_db"
    fi
    
    log_info "备份验证通过"
    return 0
}

# 清理旧备份
cleanup_old_backups() {
    log_info "清理 ${RETENTION_DAYS} 天前的备份..."
    
    local count=0
    while IFS= read -r -d '' file; do
        log_info "删除旧备份：$file"
        rm -f "$file"
        ((count++))
    done < <(find "$BACKUP_DIR" -name "rolecraft_*.db*" -type f -mtime +${RETENTION_DAYS} -print0 2>/dev/null)
    
    while IFS= read -r -d '' file; do
        log_info "删除旧备份：$file"
        rm -f "$file"
        ((count++))
    done < <(find "$BACKUP_DIR" -name "rolecraft_*.sql*" -type f -mtime +${RETENTION_DAYS} -print0 2>/dev/null)
    
    log_info "清理完成，共删除 $count 个旧备份文件"
}

# 列出所有备份
list_backups() {
    log_info "备份文件列表："
    echo ""
    
    if [ -d "$BACKUP_DIR" ]; then
        ls -lh "$BACKUP_DIR"/rolecraft_* 2>/dev/null | awk '{print $9, $5}' || echo "没有找到备份文件"
    else
        echo "备份目录不存在：$BACKUP_DIR"
    fi
}

# 恢复备份
restore_backup() {
    local backup_file="$1"
    
    if [ -z "$backup_file" ]; then
        log_error "请指定备份文件"
        echo "用法：$0 restore <备份文件>"
        exit 1
    fi
    
    if [ ! -f "$backup_file" ]; then
        log_error "备份文件不存在：$backup_file"
        exit 1
    fi
    
    log_warn "即将恢复备份：$backup_file"
    log_warn "当前数据库将被覆盖！"
    read -p "确定要继续吗？(y/N): " confirm
    
    if [ "$confirm" != "y" ] && [ "$confirm" != "Y" ]; then
        log_info "恢复已取消"
        exit 0
    fi
    
    # 创建当前数据库的备份
    log_info "创建当前数据库的备份..."
    backup_sqlite
    
    if [[ "$backup_file" == *.gz ]]; then
        log_info "解压并恢复备份..."
        gunzip -c "$backup_file" > "$DB_PATH"
    else
        log_info "恢复备份..."
        cp "$backup_file" "$DB_PATH"
    fi
    
    log_info "恢复完成"
    
    # 验证恢复的数据库
    verify_backup "$DB_PATH"
}

# 显示帮助
show_help() {
    echo "RoleCraft AI 数据库备份脚本"
    echo ""
    echo "用法：$0 <命令> [选项]"
    echo ""
    echo "命令:"
    echo "  backup      创建数据库备份"
    echo "  restore     恢复数据库备份"
    echo "  verify      验证备份文件"
    echo "  list        列出所有备份"
    echo "  cleanup     清理旧备份"
    echo "  help        显示帮助信息"
    echo ""
    echo "环境变量:"
    echo "  BACKUP_DIR      备份目录 (默认：./backups)"
    echo "  DB_PATH         数据库文件路径 (默认：./rolecraft.db)"
    echo "  DB_TYPE         数据库类型：sqlite 或 postgres (默认：sqlite)"
    echo "  PG_CONNECTION   PostgreSQL 连接字符串"
    echo "  RETENTION_DAYS  备份保留天数 (默认：30)"
    echo "  COMPRESSION     是否压缩备份：true/false (默认：true)"
    echo ""
    echo "示例:"
    echo "  $0 backup                    # 创建备份"
    echo "  $0 restore backups/xxx.db    # 恢复备份"
    echo "  $0 list                      # 列出备份"
    echo "  BACKUP_DIR=/data/backups $0 backup  # 指定备份目录"
}

# 主函数
main() {
    local command="${1:-help}"
    
    create_backup_dir
    
    case "$command" in
        backup)
            if [ "$DB_TYPE" = "postgres" ]; then
                backup_postgres
            else
                backup_sqlite
            fi
            
            if [ $? -eq 0 ]; then
                log_info "备份成功"
                cleanup_old_backups
            else
                log_error "备份失败"
                exit 1
            fi
            ;;
        
        restore)
            restore_backup "$2"
            ;;
        
        verify)
            if [ -z "$2" ]; then
                log_error "请指定备份文件"
                exit 1
            fi
            verify_backup "$2"
            ;;
        
        list)
            list_backups
            ;;
        
        cleanup)
            cleanup_old_backups
            ;;
        
        help|--help|-h)
            show_help
            ;;
        
        *)
            log_error "未知命令：$command"
            show_help
            exit 1
            ;;
    esac
}

# 执行主函数
main "$@"
