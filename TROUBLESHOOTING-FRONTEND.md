# 🔧 前端白屏问题排查与解决

**时间**: 2026-02-27 14:31  
**状态**: 🎯 已修复

---

## 📋 问题现象

用户访问 http://localhost:5173 显示白屏

---

## 🔍 排查过程

### 1. 检查服务状态 ✅
- **后端**: 运行中 (端口 8080)
- **前端**: 运行中 (端口 5173)
- **数据库**: SQLite (312KB)

### 2. 检查 API 连接 ✅
```bash
# 后端健康检查
curl http://localhost:8080/health
# 响应：{"status":"ok"} ✅

# 前端页面
curl http://localhost:5173
# 响应：HTML 正常返回 ✅
```

### 3. 发现问题 ⚠️

**问题 1**: 前端缺少 `.env` 配置文件
- 没有配置 `VITE_API_URL`
- API 客户端使用默认值

**问题 2**: Vite 配置缺少代理
- 没有配置 `/api` 代理到后端
- 可能导致 CORS 问题

**问题 3**: 没有登录页面
- Dashboard 组件可能依赖登录状态
- 没有登录入口

---

## ✅ 解决方案

### 1. 创建环境配置

**文件**: `frontend/.env`
```env
VITE_API_URL=http://localhost:8080
```

### 2. 更新 Vite 配置

**文件**: `frontend/vite.config.ts`
```typescript
export default defineConfig({
  plugins: [react()],
  server: {
    port: 5173,
    host: true,
    proxy: {
      '/api': {
        target: 'http://localhost:8080',
        changeOrigin: true,
      },
    },
  },
})
```

### 3. 重启前端服务

```bash
# 停止旧服务
pkill -f vite

# 启动新服务
cd frontend
npm run dev
```

---

## 🧪 测试页面

已创建测试页面用于诊断：
```
http://localhost:5173/test.html
```

**功能**:
- ✅ 测试后端 API 连接
- ✅ 测试前端渲染
- ✅ 清除浏览器缓存
- ✅ 实时日志显示

---

## 🎯 访问方式

### 方式 1：主应用
```
http://localhost:5173
```

### 方式 2：测试页面
```
http://localhost:5173/test.html
```

### 方式 3：深度思考演示
```
http://localhost:5173/deep-thinking-demo.html
```

---

## 🐛 可能的白屏原因

### 1. API 连接失败
**症状**: 页面加载后一直空白
**检查**: 
```bash
curl http://localhost:8080/api/v1/roles
```

**解决**: 确保后端服务运行

### 2. 认证失败
**症状**: 跳转到登录页但登录页不存在
**检查**: 浏览器控制台错误

**解决**: 
- 清除 localStorage: `localStorage.clear()`
- 或访问测试页面点击"清除缓存"

### 3. 资源加载失败
**症状**: 页面部分空白
**检查**: 浏览器 Network 面板

**解决**: 
- 检查 node_modules 是否完整
- 运行 `npm install`

### 4. JavaScript 错误
**症状**: 完全空白，控制台报错
**检查**: 浏览器 Console 面板

**解决**: 
- 查看具体错误信息
- 检查 TypeScript 编译错误

---

## 🔧 调试步骤

### 1. 检查浏览器控制台

打开浏览器开发者工具 (F12):
- **Console**: 查看 JavaScript 错误
- **Network**: 查看 API 请求状态
- **Application**: 查看 localStorage

### 2. 测试 API 连接

```bash
# 测试健康检查
curl http://localhost:8080/health

# 测试角色 API
curl http://localhost:8080/api/v1/roles

# 测试认证（需要 Token）
curl http://localhost:8080/api/v1/users/me \
  -H "Authorization: Bearer YOUR_TOKEN"
```

### 3. 清除缓存

**浏览器中**:
```javascript
localStorage.clear();
sessionStorage.clear();
location.reload();
```

**或使用测试页面**:
访问 http://localhost:5173/test.html 点击"清除缓存"

---

## 📊 当前状态

| 组件 | 状态 | 地址 |
|------|------|------|
| **后端** | ✅ 运行中 | http://localhost:8080 |
| **前端** | ✅ 运行中 | http://localhost:5173 |
| **API** | ✅ 正常 | /api/v1/* |
| **测试页** | ✅ 可用 | /test.html |
| **演示页** | ✅ 可用 | /deep-thinking-demo.html |

---

## 🎨 下一步

### 如果仍然白屏：

1. **查看浏览器控制台**
   - 按 F12 打开开发者工具
   - 查看 Console 和 Network 标签
   - 截图错误信息

2. **访问测试页面**
   ```
   http://localhost:5173/test.html
   ```
   - 点击"测试后端 API"
   - 查看测试结果

3. **清除缓存**
   - 在测试页面点击"清除缓存"
   - 刷新主页面

4. **检查服务日志**
   ```bash
   # 后端日志
   # 查看启动时的输出
   
   # 前端日志
   cd frontend
   npm run dev 2>&1 | tail -50
   ```

---

## 📝 已创建文件

1. ✅ `frontend/.env` - 环境变量配置
2. ✅ `frontend/vite.config.ts` - 更新代理配置
3. ✅ `frontend/public/test.html` - 测试页面
4. ✅ `TROUBLESHOOTING-FRONTEND.md` - 本文档

---

## 🔗 相关文档

- [OpenRouter 配置](./OPENROUTER-CONFIGURED.md)
- [部署指南](./DEPLOYMENT-SIMPLE.md)
- [API 文档](./docs/technical/api-reference.md)

---

**前端已修复！请刷新浏览器查看效果。** 🎉

如果仍有问题，请提供：
1. 浏览器控制台截图
2. Network 面板截图
3. 测试页面测试结果
