# SSH Key 配置说明

**时间**: 2026-02-28 22:45  
**目标服务器**: admin@youmind.host

---

## 🔑 SSH Public Key

```
ssh-ed25519 AAAAC3NzaC1lZDI1NTE5AAAAILWHgkj5dSImg9sQ8DHn8bd+sXWtxvkfbWxyEjMONY5l admin@youmind.host
```

---

## 📝 手动配置步骤

### 方法 1: 使用 SSH 命令
```bash
# 1. 测试连接
ssh admin@youmind.host

# 2. 手动添加 Key
ssh admin@youmind.host "mkdir -p ~/.ssh && echo 'ssh-ed25519 AAAAC3NzaC1lZDI1NTE5AAAAILWHgkj5dSImg9sQ8DHn8bd+sXWtxvkfbWxyEjMONY5l admin@youmind.host' >> ~/.ssh/authorized_keys && chmod 700 ~/.ssh && chmod 600 ~/.ssh/authorized_keys"

# 3. 验证
ssh -i ~/.ssh/youmind_key admin@youmind.host
```

### 方法 2: 使用 ssh-copy-id
```bash
# 如果有密码
ssh-copy-id admin@youmind.host

# 或指定 Key
ssh-copy-id -i ~/.ssh/youmind_key.pub admin@youmind.host
```

### 方法 3: 手动复制
```bash
# 1. 登录服务器
ssh admin@youmind.host

# 2. 创建 .ssh 目录
mkdir -p ~/.ssh

# 3. 添加公钥
echo 'ssh-ed25519 AAAAC3NzaC1lZDI1NTE5AAAAILWHgkj5dSImg9sQ8DHn8bd+sXWtxvkfbWxyEjMONY5l admin@youmind.host' >> ~/.ssh/authorized_keys

# 4. 设置权限
chmod 700 ~/.ssh
chmod 600 ~/.ssh/authorized_keys

# 5. 退出
exit

# 6. 测试连接
ssh admin@youmind.host
```

---

## 🔧 SSH 配置文件

在 `~/.ssh/config` 添加：

```
Host youmind
    HostName youmind.host
    User admin
    IdentityFile ~/.ssh/youmind_key
    IdentitiesOnly yes
    AddKeysToAgent yes
    UseKeychain yes
```

然后使用：
```bash
ssh youmind
```

---

## ✅ 验证连接

```bash
# 测试连接
ssh -T admin@youmind.host

# 详细输出（调试用）
ssh -vT admin@youmind.host
```

---

## ⚠️ 常见问题

### 1. 权限问题
```bash
# 确保权限正确
chmod 700 ~/.ssh
chmod 600 ~/.ssh/authorized_keys
```

### 2. 连接被拒绝
```bash
# 检查 SSH 服务
sudo systemctl status sshd

# 检查防火墙
sudo ufw status
```

### 3. Key 格式错误
```bash
# 确保是单行
cat ~/.ssh/authorized_keys | tr -d '\n'
```

---

**创建时间**: 2026-02-28 22:45  
**状态**: 待手动配置 ⏳
