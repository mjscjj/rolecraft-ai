#!/bin/bash
# RoleCraft AI 服务监控脚本

LOG_FILE="rolecraft-ai/logs/monitor.log"
mkdir -p rolecraft-ai/logs

echo "=== RoleCraft AI 服务监控 ===" | tee -a $LOG_FILE
echo "时间: $(date)" | tee -a $LOG_FILE

# 检查后端
BACKEND_STATUS=$(curl -s -o /dev/null -w "%{http_code}" http://localhost:8080/health)
if [ "$BACKEND_STATUS" = "200" ]; then
    echo "[OK] 后端服务正常" | tee -a $LOG_FILE
else
    echo "[ERROR] 后端服务异常 (HTTP $BACKEND_STATUS)" | tee -a $LOG_FILE
    # 尝试重启
    cd rolecraft-ai/backend && unset DATABASE_URL && go run cmd/server/main.go > /dev/null 2>&1 &
    echo "[ACTION] 尝试重启后端服务" | tee -a $LOG_FILE
fi

# 检查前端
FRONTEND_STATUS=$(curl -s -o /dev/null -w "%{http_code}" http://localhost:5173)
if [ "$FRONTEND_STATUS" = "200" ]; then
    echo "[OK] 前端服务正常" | tee -a $LOG_FILE
else
    echo "[ERROR] 前端服务异常 (HTTP $FRONTEND_STATUS)" | tee -a $LOG_FILE
fi

# 检查数据库
DB_EXISTS=$(ls rolecraft-ai/backend/rolecraft.db 2>/dev/null)
if [ -n "$DB_EXISTS" ]; then
    DB_SIZE=$(ls -lh rolecraft-ai/backend/rolecraft.db | awk '{print $5}')
    echo "[OK] 数据库正常 (大小: $DB_SIZE)" | tee -a $LOG_FILE
else
    echo "[ERROR] 数据库文件不存在" | tee -a $LOG_FILE
fi

# 检查进程
BACKEND_PID=$(pgrep -f "go run cmd/server/main.go" | head -1)
FRONTEND_PID=$(pgrep -f "vite" | head -1)

echo "进程ID: 后端=$BACKEND_PID 前端=$FRONTEND_PID" | tee -a $LOG_FILE
echo "---" | tee -a $LOG_FILE