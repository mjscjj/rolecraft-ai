import { test, expect } from '@playwright/test';
import * as fs from 'fs';
import * as path from 'path';

const API_BASE = 'http://localhost:8080/api/v1';

// 生成唯一的测试用户邮箱
const generateUniqueEmail = () => `e2e_${Date.now()}_${Math.random().toString(36).substring(7)}@rolecraft.ai`;

// 辅助函数：注册并获取 token
const registerAndGetToken = async (request: any) => {
  const email = generateUniqueEmail();
  const registerResponse = await request.post(`${API_BASE}/auth/register`, {
    data: { email, password: 'E2ETest123!', name: 'Test User' },
  });
  expect(registerResponse.ok()).toBeTruthy();
  return { email, token: (await registerResponse.json()).data.token };
};

// 辅助函数：创建角色
const createRole = async (request: any, token: string) => {
  const response = await request.post(`${API_BASE}/roles`, {
    headers: { 'Authorization': `Bearer ${token}`, 'Content-Type': 'application/json' },
    data: {
      name: `Test Role ${Date.now()}`,
      description: 'Test',
      category: 'Test',
      systemPrompt: 'You are helpful',
      welcomeMessage: 'Hello',
    },
  });
  expect(response.ok()).toBeTruthy();
  return (await response.json()).data.id;
};

/**
 * E2E 集成测试 - 完整用户流程
 */
test.describe('端到端集成测试', () => {
  
  test.describe('1. 用户注册与登录', () => {
    test('用户注册成功', async ({ request }) => {
      const email = generateUniqueEmail();
      const registerResponse = await request.post(`${API_BASE}/auth/register`, {
        data: { email, password: 'E2ETest123!', name: 'E2E User' },
      });
      
      expect(registerResponse.ok()).toBeTruthy();
      const data = await registerResponse.json();
      expect(data.data).toHaveProperty('token');
      expect(data.data.user).toHaveProperty('email', email);
    });

    test('使用注册的账号登录', async ({ request }) => {
      const email = generateUniqueEmail();
      
      // 先注册
      await request.post(`${API_BASE}/auth/register`, {
        data: { email, password: 'E2ETest123!', name: 'Test User' },
      });

      // 再登录
      const loginResponse = await request.post(`${API_BASE}/auth/login`, {
        data: { email, password: 'E2ETest123!' },
      });
      
      expect(loginResponse.ok()).toBeTruthy();
      const data = await loginResponse.json();
      expect(data.data).toHaveProperty('token');
    });

    test('获取当前用户信息', async ({ request }) => {
      const { email, token } = await registerAndGetToken(request);
      
      const userResponse = await request.get(`${API_BASE}/users/me`, {
        headers: { 'Authorization': `Bearer ${token}` },
      });
      
      expect(userResponse.ok()).toBeTruthy();
      const data = await userResponse.json();
      expect(data.data).toHaveProperty('email', email);
    });
  });

  test.describe('2. 角色创建与管理', () => {
    test('创建测试角色', async ({ request }) => {
      const { token } = await registerAndGetToken(request);
      
      const roleData = {
        name: `E2E 测试助手-${Date.now()}`,
        description: '端到端集成测试',
        category: '通用',
        systemPrompt: '你是 AI 助手',
        welcomeMessage: '你好',
        modelConfig: { temperature: 0.7, maxTokens: 2048 },
      };

      const response = await request.post(`${API_BASE}/roles`, {
        headers: { 'Authorization': `Bearer ${token}`, 'Content-Type': 'application/json' },
        data: roleData,
      });
      
      expect(response.ok()).toBeTruthy();
      const data = await response.json();
      expect(data.data).toHaveProperty('id');
      expect(data.data.name).toBe(roleData.name);
    });

    test('获取角色列表', async ({ request }) => {
      const { token } = await registerAndGetToken(request);
      
      // 先创建一个角色
      await createRole(request, token);

      const response = await request.get(`${API_BASE}/roles`, {
        headers: { 'Authorization': `Bearer ${token}` },
      });
      
      expect(response.ok()).toBeTruthy();
      const data = await response.json();
      expect(data.data).toBeInstanceOf(Array);
      expect(data.data.length).toBeGreaterThan(0);
    });

    test('获取角色详情', async ({ request }) => {
      const { token } = await registerAndGetToken(request);
      const roleId = await createRole(request, token);

      const response = await request.get(`${API_BASE}/roles/${roleId}`, {
        headers: { 'Authorization': `Bearer ${token}` },
      });
      
      expect(response.ok()).toBeTruthy();
      const data = await response.json();
      expect(data.data).toHaveProperty('id', roleId);
    });
  });

  test.describe('3. 文档上传与处理', () => {
    test('上传测试文档 (TXT)', async ({ request }) => {
      const { token } = await registerAndGetToken(request);
      
      const testContent = `RoleCraft AI 测试文档
========================
这是一个用于端到端集成测试的文档。
主要内容：AI 角色创建平台，支持文档上传和向量搜索。
关键词：人工智能、角色扮演、文档处理`;
      
      const testFilePath = path.join('/tmp', `e2e_test_${Date.now()}.txt`);
      fs.writeFileSync(testFilePath, testContent, 'utf-8');

      const response = await request.post(`${API_BASE}/documents`, {
        headers: { 'Authorization': `Bearer ${token}` },
        multipart: {
          file: fs.createReadStream(testFilePath),
          name: 'e2e_test_document.txt',
        },
      });
      
      // 清理临时文件
      fs.unlinkSync(testFilePath);
      
      expect(response.ok()).toBeTruthy();
      const data = await response.json();
      expect(data.data).toHaveProperty('id');
    });

    test('获取文档列表', async ({ request }) => {
      const { token } = await registerAndGetToken(request);
      
      const response = await request.get(`${API_BASE}/documents`, {
        headers: { 'Authorization': `Bearer ${token}` },
      });
      
      expect(response.ok()).toBeTruthy();
      const data = await response.json();
      expect(data.data).toBeInstanceOf(Array);
    });

    test('文档状态查询', async ({ request }) => {
      const { token } = await registerAndGetToken(request);
      
      // 先上传文档
      const testContent = 'Test content';
      const testFilePath = path.join('/tmp', `e2e_status_${Date.now()}.txt`);
      fs.writeFileSync(testFilePath, testContent, 'utf-8');

      const uploadResponse = await request.post(`${API_BASE}/documents`, {
        headers: { 'Authorization': `Bearer ${token}` },
        multipart: {
          file: fs.createReadStream(testFilePath),
          name: 'status_test.txt',
        },
      });
      fs.unlinkSync(testFilePath);
      
      const docId = (await uploadResponse.json()).data.id;

      // 查询状态
      const statusResponse = await request.get(`${API_BASE}/documents/${docId}/status`, {
        headers: { 'Authorization': `Bearer ${token}` },
      });
      
      expect(statusResponse.ok()).toBeTruthy();
      const data = await statusResponse.json();
      expect(data.data).toHaveProperty('status');
    });
  });

  test.describe('4. 对话功能测试', () => {
    test('创建会话', async ({ request }) => {
      const { token } = await registerAndGetToken(request);
      const roleId = await createRole(request, token);

      const response = await request.post(`${API_BASE}/chat-sessions`, {
        headers: { 'Authorization': `Bearer ${token}`, 'Content-Type': 'application/json' },
        data: { roleId },
      });
      
      expect(response.ok()).toBeTruthy();
      const data = await response.json();
      expect(data.data).toHaveProperty('id');
      expect(data.data.roleId).toBe(roleId);
    });

    test('发送消息并获取回复', async ({ request }) => {
      const { token } = await registerAndGetToken(request);
      const roleId = await createRole(request, token);

      // 创建会话
      const sessionResponse = await request.post(`${API_BASE}/chat-sessions`, {
        headers: { 'Authorization': `Bearer ${token}`, 'Content-Type': 'application/json' },
        data: { roleId },
      });
      const sessionId = (await sessionResponse.json()).data.id;

      // 发送消息
      const chatResponse = await request.post(`${API_BASE}/chat/${sessionId}/complete`, {
        headers: { 'Authorization': `Bearer ${token}`, 'Content-Type': 'application/json' },
        data: { content: '你好' },
      });
      
      expect(chatResponse.ok()).toBeTruthy();
      const data = await chatResponse.json();
      expect(data.data).toHaveProperty('assistantMessage');
    });

    test('获取会话历史', async ({ request }) => {
      const { token } = await registerAndGetToken(request);
      const roleId = await createRole(request, token);

      // 创建会话并发送消息
      const sessionResponse = await request.post(`${API_BASE}/chat-sessions`, {
        headers: { 'Authorization': `Bearer ${token}`, 'Content-Type': 'application/json' },
        data: { roleId },
      });
      const sessionId = (await sessionResponse.json()).data.id;

      await request.post(`${API_BASE}/chat/${sessionId}/complete`, {
        headers: { 'Authorization': `Bearer ${token}`, 'Content-Type': 'application/json' },
        data: { content: 'Hello' },
      });

      // 获取历史
      const historyResponse = await request.get(`${API_BASE}/chat-sessions/${sessionId}`, {
        headers: { 'Authorization': `Bearer ${token}` },
      });
      
      expect(historyResponse.ok()).toBeTruthy();
      const data = await historyResponse.json();
      expect(data.data.session).toHaveProperty('id', sessionId);
      expect(data.data).toHaveProperty('messages');
    });
  });

  test.describe('5. 完整流程验证', () => {
    test('验证完整用户流程', async ({ request }) => {
      // 1. 注册
      const { email, token } = await registerAndGetToken(request);
      
      // 2. 创建角色
      const roleId = await createRole(request, token);
      
      // 3. 创建会话
      const sessionResponse = await request.post(`${API_BASE}/chat-sessions`, {
        headers: { 'Authorization': `Bearer ${token}`, 'Content-Type': 'application/json' },
        data: { roleId },
      });
      const sessionId = (await sessionResponse.json()).data.id;
      
      // 4. 发送消息
      const chatResponse = await request.post(`${API_BASE}/chat/${sessionId}/complete`, {
        headers: { 'Authorization': `Bearer ${token}`, 'Content-Type': 'application/json' },
        data: { content: 'Hi' },
      });
      expect(chatResponse.ok()).toBeTruthy();
      
      // 5. 验证用户数据
      const userResponse = await request.get(`${API_BASE}/users/me`, {
        headers: { 'Authorization': `Bearer ${token}` },
      });
      expect(userResponse.ok()).toBeTruthy();
      const userData = await userResponse.json();
      expect(userData.data.email).toBe(email);
      
      // 6. 验证角色
      const roleResponse = await request.get(`${API_BASE}/roles/${roleId}`, {
        headers: { 'Authorization': `Bearer ${token}` },
      });
      expect(roleResponse.ok()).toBeTruthy();
      
      // 7. 验证会话
      const getSessionResponse = await request.get(`${API_BASE}/chat-sessions/${sessionId}`, {
        headers: { 'Authorization': `Bearer ${token}` },
      });
      expect(getSessionResponse.ok()).toBeTruthy();
    });
  });

  test.describe('6. 错误处理测试', () => {
    test('使用错误密码登录失败', async ({ request }) => {
      const email = generateUniqueEmail();
      await request.post(`${API_BASE}/auth/register`, {
        data: { email, password: 'CorrectPassword123!', name: 'Test' },
      });

      const loginResponse = await request.post(`${API_BASE}/auth/login`, {
        data: { email, password: 'WrongPassword' },
      });
      
      expect(loginResponse.ok()).toBeFalsy();
      expect(loginResponse.status()).toBe(401);
    });

    test('未授权访问失败', async ({ request }) => {
      const response = await request.get(`${API_BASE}/roles`, {
        headers: { 'Authorization': 'Bearer invalid_token' },
      });
      
      expect(response.ok()).toBeFalsy();
      expect(response.status()).toBe(401);
    });

    test('访问不存在的角色失败', async ({ request }) => {
      const { token } = await registerAndGetToken(request);
      
      const response = await request.get(`${API_BASE}/roles/non-existent-id`, {
        headers: { 'Authorization': `Bearer ${token}` },
      });
      
      expect(response.ok()).toBeFalsy();
      expect(response.status()).toBe(404);
    });
  });

  test.describe('性能测试', () => {
    test('对话 API 响应时间', async ({ request }) => {
      const { token } = await registerAndGetToken(request);
      const roleId = await createRole(request, token);

      const sessionResponse = await request.post(`${API_BASE}/chat-sessions`, {
        headers: { 'Authorization': `Bearer ${token}`, 'Content-Type': 'application/json' },
        data: { roleId },
      });
      const sessionId = (await sessionResponse.json()).data.id;

      const startTime = Date.now();
      const chatResponse = await request.post(`${API_BASE}/chat/${sessionId}/complete`, {
        headers: { 'Authorization': `Bearer ${token}`, 'Content-Type': 'application/json' },
        data: { content: 'Hello' },
      });
      const responseTime = Date.now() - startTime;
      
      expect(chatResponse.ok()).toBeTruthy();
      expect(responseTime).toBeLessThan(10000); // < 10 秒
    });
  });
});
