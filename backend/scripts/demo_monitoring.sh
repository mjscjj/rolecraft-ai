#!/bin/bash

# RoleCraft AI 监控功能演示脚本

set -e

# 颜色
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

echo -e "${BLUE}======================================${NC}"
echo -e "${BLUE}RoleCraft AI 监控功能演示${NC}"
echo -e "${BLUE}======================================${NC}"
echo ""

# 检查是否在正确的目录
if [ ! -f "cmd/server/main.go" ]; then
    echo -e "${YELLOW}请在 backend 目录下运行此脚本${NC}"
    exit 1
fi

# 1. 创建日志目录
echo -e "${GREEN}[1/5] 创建日志目录...${NC}"
mkdir -p logs
echo "✅ 日志目录已创建：logs/"

# 2. 创建备份目录
echo -e "${GREEN}[2/5] 创建备份目录...${NC}"
mkdir -p scripts/backup/backups
echo "✅ 备份目录已创建：scripts/backup/backups/"

# 3. 运行测试
echo -e "${GREEN}[3/5] 运行监控模块测试...${NC}"
go test ./internal/middleware/... -v | grep -E "(PASS|FAIL|RUN)" | head -20

# 4. 运行健康检查测试
echo -e "${GREEN}[4/5] 运行健康检查测试...${NC}"
go test ./internal/api/handler/health_test.go ./internal/api/handler/health.go -v | grep -E "(PASS|FAIL)"

# 5. 创建示例备份
echo -e "${GREEN}[5/5] 创建示例备份...${NC}"
cd scripts/backup
if [ -f "../../rolecraft.db" ]; then
    ./backup.sh backup 2>&1 | grep -E "(INFO|成功)" || true
else
    echo "⚠️  数据库文件不存在，跳过备份演示"
fi
cd ../..

echo ""
echo -e "${BLUE}======================================${NC}"
echo -e "${BLUE}演示完成！${NC}"
echo -e "${BLUE}======================================${NC}"
echo ""
echo -e "${GREEN}✅ 已完成的配置:${NC}"
echo "   - 日志中间件 (internal/middleware/logger.go)"
echo "   - 性能监控 (internal/middleware/performance.go)"
echo "   - 健康检查 (internal/api/handler/health.go)"
echo "   - 备份系统 (scripts/backup/)"
echo ""
echo -e "${YELLOW}📝 下一步:${NC}"
echo "   1. 启动服务：go run cmd/server/main.go"
echo "   2. 访问健康检查：curl http://localhost:8080/api/v1/health"
echo "   3. 查看性能指标：curl http://localhost:8080/api/v1/metrics"
echo "   4. 查看日志：tail -f logs/rolecraft.log"
echo ""
echo -e "${BLUE}📚 详细文档:${NC}"
echo "   - docs/STABILITY_MONITORING.md"
echo "   - scripts/backup/README.md"
echo ""
