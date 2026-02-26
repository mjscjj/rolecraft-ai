import { test, expect } from '@playwright/test';

test.describe('E2E 测试截图', () => {
  test('前端首页截图', async ({ page }) => {
    await page.goto('/');
    await page.waitForTimeout(2000);
    await page.screenshot({ 
      path: 'e2e/screenshots/01-dashboard.png',
      fullPage: true 
    });
  });

  test('登录流程截图', async ({ page }) => {
    await page.goto('/');
    await page.waitForTimeout(1000);
    
    // 截图登录前
    await page.screenshot({ 
      path: 'e2e/screenshots/02-before-login.png',
      fullPage: true 
    });
  });

  test('角色列表截图', async ({ page }) => {
    await page.goto('/roles');
    await page.waitForTimeout(2000);
    await page.screenshot({ 
      path: 'e2e/screenshots/03-roles.png',
      fullPage: true 
    });
  });

  test('对话界面截图', async ({ page }) => {
    // 先登录获取 token
    const API_BASE = 'http://localhost:8080/api/v1';
    const loginResp = await page.request.post(`${API_BASE}/auth/login`, {
      data: { email: 'test@rolecraft.ai', password: 'test123' }
    });
    const token = (await loginResp.json()).data.token;

    // 创建角色
    const roleResp = await page.request.post(`${API_BASE}/roles`, {
      headers: { 'Authorization': `Bearer ${token}` },
      data: {
        name: '截图测试角色',
        description: '测试',
        category: '通用',
        systemPrompt: 'test',
        welcomeMessage: '你好'
      }
    });
    const roleId = (await roleResp.json()).data.id;

    // 访问对话页面
    await page.goto(`/chat/${roleId}`);
    await page.waitForTimeout(2000);
    await page.screenshot({ 
      path: 'e2e/screenshots/04-chat.png',
      fullPage: true 
    });
  });
});
