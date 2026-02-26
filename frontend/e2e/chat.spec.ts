import { test, expect } from '@playwright/test';

const API_BASE = 'http://localhost:8080/api/v1';

test.describe('对话功能', () => {
  let authToken: string;
  let roleId: string;

  test.beforeEach(async ({ page }) => {
    // 登录获取 token
    const loginResponse = await page.request.post(`${API_BASE}/auth/login`, {
      data: {
        email: 'test@rolecraft.ai',
        password: 'test123',
      },
    });
    authToken = (await loginResponse.json()).data.token;

    // 创建测试角色
    const roleResponse = await page.request.post(`${API_BASE}/roles`, {
      headers: {
        'Authorization': `Bearer ${authToken}`,
        'Content-Type': 'application/json',
      },
      data: {
        name: `Chat Test Role-${Date.now()}`,
        description: '测试角色',
        category: '通用',
        systemPrompt: '你是友好的 AI 助手',
        welcomeMessage: '你好！',
      },
    });
    roleId = (await roleResponse.json()).data.id;
  });

  test('创建会话', async ({ page }) => {
    const response = await page.request.post(`${API_BASE}/chat-sessions`, {
      headers: {
        'Authorization': `Bearer ${authToken}`,
        'Content-Type': 'application/json',
      },
      data: {
        roleId: roleId,
        mode: 'quick',
      },
    });
    
    expect(response.ok()).toBeTruthy();
    const data = await response.json();
    expect(data.data).toHaveProperty('id');
    expect(data.data.roleId).toBe(roleId);
  });

  test('发送消息并获取回复', async ({ page }) => {
    // 创建会话
    const sessionResponse = await page.request.post(`${API_BASE}/chat-sessions`, {
      headers: {
        'Authorization': `Bearer ${authToken}`,
        'Content-Type': 'application/json',
      },
      data: {
        roleId: roleId,
        mode: 'quick',
      },
    });
    const sessionId = (await sessionResponse.json()).data.id;

    // 发送消息
    const chatResponse = await page.request.post(`${API_BASE}/chat/${sessionId}/complete`, {
      headers: {
        'Authorization': `Bearer ${authToken}`,
        'Content-Type': 'application/json',
      },
      data: {
        content: '你好，帮我写个简短的测试',
      },
    });
    
    expect(chatResponse.ok()).toBeTruthy();
    const data = await chatResponse.json();
    expect(data.data).toHaveProperty('userMessage');
    expect(data.data).toHaveProperty('assistantMessage');
    expect(data.data.assistantMessage.content).toBeDefined();
  });

  test('获取会话历史', async ({ page }) => {
    // 创建会话并发送消息
    const sessionResponse = await page.request.post(`${API_BASE}/chat-sessions`, {
      headers: {
        'Authorization': `Bearer ${authToken}`,
        'Content-Type': 'application/json',
      },
      data: {
        roleId: roleId,
        mode: 'quick',
      },
    });
    const sessionId = (await sessionResponse.json()).data.id;

    await page.request.post(`${API_BASE}/chat/${sessionId}/complete`, {
      headers: {
        'Authorization': `Bearer ${authToken}`,
        'Content-Type': 'application/json',
      },
      data: {
        content: '测试消息',
      },
    });

    // 获取会话历史
    const historyResponse = await page.request.get(`${API_BASE}/chat-sessions/${sessionId}`, {
      headers: {
        'Authorization': `Bearer ${authToken}`,
      },
    });
    
    expect(historyResponse.ok()).toBeTruthy();
    const data = await historyResponse.json();
    expect(data.data).toHaveProperty('messages');
    expect(data.data.messages).toBeInstanceOf(Array);
    expect(data.data.messages.length).toBeGreaterThan(0);
  });

  test('Mock AI 回复分类', async ({ page }) => {
    const sessionResponse = await page.request.post(`${API_BASE}/chat-sessions`, {
      headers: {
        'Authorization': `Bearer ${authToken}`,
        'Content-Type': 'application/json',
      },
      data: {
        roleId: roleId,
        mode: 'quick',
      },
    });
    const sessionId = (await sessionResponse.json()).data.id;

    // 测试问候
    const greetingResp = await page.request.post(`${API_BASE}/chat/${sessionId}/complete`, {
      headers: { 'Authorization': `Bearer ${authToken}`, 'Content-Type': 'application/json' },
      data: { content: '你好' },
    });
    const greeting = (await greetingResp.json()).data.assistantMessage.content;
    expect(greeting).toContain('你好');

    // 测试写作
    const writingResp = await page.request.post(`${API_BASE}/chat/${sessionId}/complete`, {
      headers: { 'Authorization': `Bearer ${authToken}`, 'Content-Type': 'application/json' },
      data: { content: '帮我写个文案' },
    });
    const writing = (await writingResp.json()).data.assistantMessage.content;
    expect(writing).toBeDefined();

    // 测试分析
    const analysisResp = await page.request.post(`${API_BASE}/chat/${sessionId}/complete`, {
      headers: { 'Authorization': `Bearer ${authToken}`, 'Content-Type': 'application/json' },
      data: { content: '分析这个数据' },
    });
    const analysis = (await analysisResp.json()).data.assistantMessage.content;
    expect(analysis).toContain('分析');
  });
});
