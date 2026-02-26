import { test, expect } from '@playwright/test';

const API_BASE = 'http://localhost:8080/api/v1';

test.describe('角色管理', () => {
  let authToken: string;

  test.beforeEach(async ({ page }) => {
    // 登录获取 token
    const loginResponse = await page.request.post(`${API_BASE}/auth/login`, {
      data: {
        email: 'test@rolecraft.ai',
        password: 'test123',
      },
    });
    authToken = (await loginResponse.json()).data.token;
  });

  test('获取角色列表', async ({ page }) => {
    const response = await page.request.get(`${API_BASE}/roles`, {
      headers: {
        'Authorization': `Bearer ${authToken}`,
      },
    });
    
    expect(response.ok()).toBeTruthy();
    const data = await response.json();
    expect(data.data).toBeInstanceOf(Array);
  });

  test('创建新角色', async ({ page }) => {
    const newRole = {
      name: `E2E 测试角色-${Date.now()}`,
      description: 'E2E 测试创建的角色',
      category: '通用',
      systemPrompt: '你是一个友好的 AI 助手',
      welcomeMessage: '你好！我是测试助手',
      modelConfig: { temperature: 0.7 },
    };

    const response = await page.request.post(`${API_BASE}/roles`, {
      headers: {
        'Authorization': `Bearer ${authToken}`,
        'Content-Type': 'application/json',
      },
      data: newRole,
    });
    
    expect(response.ok()).toBeTruthy();
    const data = await response.json();
    expect(data.data).toHaveProperty('id');
    expect(data.data.name).toBe(newRole.name);
  });

  test('获取角色模板', async ({ page }) => {
    const response = await page.request.get(`${API_BASE}/roles/templates`);
    
    expect(response.ok()).toBeTruthy();
    const data = await response.json();
    expect(data.data).toBeInstanceOf(Array);
    expect(data.data.length).toBeGreaterThan(0);
  });
});
