#!/bin/bash

# RoleCraft AI 停止脚本

echo "🛑 停止 RoleCraft AI..."

# 停止后端
if [ -f /tmp/rolecraft-backend.pid ]; then
    kill $(cat /tmp/rolecraft-backend.pid) 2>/dev/null
    rm /tmp/rolecraft-backend.pid
fi

# 停止前端
if [ -f /tmp/rolecraft-frontend.pid ]; then
    kill $(cat /tmp/rolecraft-frontend.pid) 2>/dev/null
    rm /tmp/rolecraft-frontend.pid
fi

# 停止 Docker 服务
docker-compose down

echo "✅ 所有服务已停止"
