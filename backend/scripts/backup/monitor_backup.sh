#!/bin/bash

# RoleCraft AI 备份监控脚本
# 检查备份状态，发送告警

set -e

# 配置
BACKUP_DIR="${BACKUP_DIR:-./backups}"
MAX_AGE_HOURS="${MAX_AGE_HOURS:-26}"  # 超过 26 小时没有新备份则告警
MIN_SIZE_BYTES="${MIN_SIZE_BYTES:-1024}"  # 最小备份大小（1KB）
ALERT_EMAIL="${ALERT_EMAIL:-}"  # 告警邮箱
ALERT_WEBHOOK="${ALERT_WEBHOOK:-}"  # 告警 Webhook

# 颜色输出
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m'

log_info() {
    echo -e "${GREEN}[INFO]${NC} $(date '+%Y-%m-%d %H:%M:%S') - $1"
}

log_warn() {
    echo -e "${YELLOW}[WARN]${NC} $(date '+%Y-%m-%d %H:%M:%S') - $1"
}

log_error() {
    echo -e "${RED}[ERROR]${NC} $(date '+%Y-%m-%d %H:%M:%S') - $1"
}

# 发送告警
send_alert() {
    local subject="$1"
    local message="$2"
    
    # 邮件告警
    if [ -n "$ALERT_EMAIL" ] && command -v mail &> /dev/null; then
        echo "$message" | mail -s "[RoleCraft Backup Alert] $subject" "$ALERT_EMAIL"
    fi
    
    # Webhook 告警（钉钉/企业微信/Slack）
    if [ -n "$ALERT_WEBHOOK" ]; then
        curl -s -X POST "$ALERT_WEBHOOK" \
            -H "Content-Type: application/json" \
            -d "{\"msgtype\":\"text\",\"text\":{\"content\":\"[RoleCraft Backup Alert]\\n$subject\\n$message\"}}"
    fi
    
    # 输出到标准错误
    log_error "$subject: $message"
}

# 检查最新备份
check_latest_backup() {
    log_info "检查最新备份..."
    
    # 查找最新备份文件
    LATEST=$(ls -t "$BACKUP_DIR"/rolecraft_*.db.gz 2>/dev/null | head -1)
    
    if [ -z "$LATEST" ]; then
        # 尝试不带压缩的
        LATEST=$(ls -t "$BACKUP_DIR"/rolecraft_*.db 2>/dev/null | head -1)
    fi
    
    if [ -z "$LATEST" ]; then
        send_alert "No Backups Found" "没有找到任何备份文件！请检查备份系统是否正常运行。"
        return 1
    fi
    
    log_info "最新备份：$LATEST"
    
    # 检查备份时间
    BACKUP_TIME=$(stat -c %Y "$LATEST" 2>/dev/null || stat -f %m "$LATEST" 2>/dev/null)
    CURRENT_TIME=$(date +%s)
    AGE_HOURS=$(( (CURRENT_TIME - BACKUP_TIME) / 3600 ))
    
    log_info "备份时间：$AGE_HOURS 小时前"
    
    if [ $AGE_HOURS -gt $MAX_AGE_HOURS ]; then
        send_alert "Backup Too Old" "最新备份已是 $AGE_HOURS 小时前，超过阈值 ($MAX_AGE_HOURS 小时)！"
        return 1
    fi
    
    # 检查备份大小
    BACKUP_SIZE=$(stat -c %s "$LATEST" 2>/dev/null || stat -f %z "$LATEST" 2>/dev/null)
    
    if [ $BACKUP_SIZE -lt $MIN_SIZE_BYTES ]; then
        send_alert "Backup Too Small" "备份文件过小 ($BACKUP_SIZE bytes)，可能备份失败！"
        return 1
    fi
    
    log_info "备份大小：$BACKUP_SIZE bytes"
    
    # 检查备份数量
    BACKUP_COUNT=$(ls "$BACKUP_DIR"/rolecraft_*.db* 2>/dev/null | wc -l)
    log_info "备份总数：$BACKUP_COUNT"
    
    if [ $BACKUP_COUNT -lt 1 ]; then
        send_alert "No Backups" "备份目录为空！"
        return 1
    fi
    
    # 检查磁盘空间
    DISK_USAGE=$(df "$BACKUP_DIR" | tail -1 | awk '{print $5}' | sed 's/%//')
    log_info "磁盘使用率：$DISK_USAGE%"
    
    if [ $DISK_USAGE -gt 90 ]; then
        send_alert "Low Disk Space" "备份目录磁盘使用率超过 90%！"
        return 1
    fi
    
    log_info "✓ 备份检查通过"
    return 0
}

# 检查备份完整性
check_backup_integrity() {
    log_info "检查备份完整性..."
    
    LATEST=$(ls -t "$BACKUP_DIR"/rolecraft_*.db.gz 2>/dev/null | head -1)
    
    if [ -z "$LATEST" ]; then
        LATEST=$(ls -t "$BACKUP_DIR"/rolecraft_*.db 2>/dev/null | head -1)
    fi
    
    if [ -z "$LATEST" ]; then
        log_warn "没有备份文件可检查"
        return 0
    fi
    
    # 检查 gzip 完整性
    if [[ "$LATEST" == *.gz ]]; then
        if ! gzip -t "$LATEST" 2>/dev/null; then
            send_alert "Backup Corrupted" "备份文件损坏：$LATEST"
            return 1
        fi
        log_info "✓ Gzip 完整性检查通过"
    fi
    
    # SQLite 完整性检查（如果有 sqlite3）
    if command -v sqlite3 &> /dev/null; then
        TEMP_DB=$(mktemp)
        
        if [[ "$LATEST" == *.gz ]]; then
            gunzip -c "$LATEST" > "$TEMP_DB"
        else
            cp "$LATEST" "$TEMP_DB"
        fi
        
        INTEGRITY=$(sqlite3 "$TEMP_DB" "PRAGMA integrity_check;" 2>&1)
        rm -f "$TEMP_DB"
        
        if [ "$INTEGRITY" != "ok" ]; then
            send_alert "Database Corrupted" "数据库完整性检查失败：$INTEGRITY"
            return 1
        fi
        log_info "✓ SQLite 完整性检查通过"
    else
        log_warn "sqlite3 未安装，跳过数据库完整性检查"
    fi
    
    return 0
}

# 生成备份报告
generate_report() {
    log_info "生成备份报告..."
    
    echo "======================================"
    echo "RoleCraft AI 备份状态报告"
    echo "生成时间：$(date '+%Y-%m-%d %H:%M:%S')"
    echo "======================================"
    echo ""
    
    # 备份统计
    echo "备份统计:"
    TOTAL_COUNT=$(ls "$BACKUP_DIR"/rolecraft_*.db* 2>/dev/null | wc -l)
    TOTAL_SIZE=$(du -sh "$BACKUP_DIR" 2>/dev/null | cut -f1)
    echo "  - 备份总数：$TOTAL_COUNT"
    echo "  - 总大小：$TOTAL_SIZE"
    echo ""
    
    # 最新备份
    echo "最新备份:"
    ls -lht "$BACKUP_DIR"/rolecraft_*.db* 2>/dev/null | head -5 | while read line; do
        echo "  $line"
    done
    echo ""
    
    # 磁盘使用
    echo "磁盘使用:"
    df -h "$BACKUP_DIR" | tail -1 | awk '{printf "  - 使用率：%s (已用 %s, 可用 %s)\n", $5, $3, $4}'
    echo ""
    
    echo "======================================"
}

# 主函数
main() {
    local command="${1:-check}"
    
    case "$command" in
        check)
            check_latest_backup
            EXIT_CODE=$?
            check_backup_integrity
            EXIT_CODE=$((EXIT_CODE || $?))
            exit $EXIT_CODE
            ;;
        
        report)
            generate_report
            ;;
        
        full)
            check_latest_backup
            check_backup_integrity
            generate_report
            ;;
        
        *)
            echo "用法：$0 {check|report|full}"
            echo ""
            echo "命令:"
            echo "  check   检查备份状态"
            echo "  report  生成备份报告"
            echo "  full    完整检查（检查 + 报告）"
            exit 1
            ;;
    esac
}

main "$@"
