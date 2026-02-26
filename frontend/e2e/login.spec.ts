import { test, expect } from '@playwright/test';

const API_BASE = 'http://localhost:8080/api/v1';

test.describe('认证流程', () => {
  test('用户登录成功', async ({ page }) => {
    await page.goto('/');
    
    // 检查页面加载
    await expect(page).toHaveTitle(/RoleCraft/);
    
    // 登录 API 测试
    const loginResponse = await page.request.post(`${API_BASE}/auth/login`, {
      data: {
        email: 'test@rolecraft.ai',
        password: 'test123',
      },
    });
    
    expect(loginResponse.ok()).toBeTruthy();
    const data = await loginResponse.json();
    expect(data.data).toHaveProperty('token');
    expect(data.data.user).toHaveProperty('email', 'test@rolecraft.ai');
  });

  test('用户注册', async ({ page }) => {
    const email = `test_${Date.now()}@rolecraft.ai`;
    
    const registerResponse = await page.request.post(`${API_BASE}/auth/register`, {
      data: {
        email: email,
        password: 'test123',
        name: 'E2E Test User',
      },
    });
    
    expect(registerResponse.ok()).toBeTruthy();
    const data = await registerResponse.json();
    expect(data.data).toHaveProperty('token');
  });

  test('错误密码登录失败', async ({ page }) => {
    const loginResponse = await page.request.post(`${API_BASE}/auth/login`, {
      data: {
        email: 'test@rolecraft.ai',
        password: 'wrongpassword',
      },
    });
    
    expect(loginResponse.ok()).toBeFalsy();
    const data = await loginResponse.json();
    expect(data).toHaveProperty('error');
  });

  test('Token 认证', async ({ page }) => {
    // 先登录获取 token
    const loginResponse = await page.request.post(`${API_BASE}/auth/login`, {
      data: {
        email: 'test@rolecraft.ai',
        password: 'test123',
      },
    });
    
    const { token } = (await loginResponse.json()).data;
    expect(token).toBeDefined();
    
    // 使用 token 访问受保护接口
    const userResponse = await page.request.get(`${API_BASE}/users/me`, {
      headers: {
        'Authorization': `Bearer ${token}`,
      },
    });
    
    expect(userResponse.ok()).toBeTruthy();
    const userData = await userResponse.json();
    expect(userData.data).toHaveProperty('email', 'test@rolecraft.ai');
  });
});
