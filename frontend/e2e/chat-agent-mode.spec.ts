import { test, expect } from '@playwright/test';

const API_BASE = 'http://127.0.0.1:8080/api/v1';

test.describe('Chat + Agent 模式', () => {
  test('标准模式与深度模式均可对话，且深度模式自动添加 @agent', async ({ page }) => {
    test.setTimeout(180000);
    const email = `mode.${Date.now()}@rolecraft.ai`;
    const password = 'test123456';

    const registerResponse = await page.request.post(`${API_BASE}/auth/register`, {
      data: {
        email,
        password,
        name: 'Mode Test User',
      },
    });
    expect(registerResponse.ok()).toBeTruthy();
    const registerData = await registerResponse.json();
    const token = registerData.data.token as string;
    const user = registerData.data.user;

    const roleResponse = await page.request.post(`${API_BASE}/roles`, {
      headers: {
        Authorization: `Bearer ${token}`,
        'Content-Type': 'application/json',
      },
      data: {
        name: `Mode Role ${Date.now()}`,
        description: '用于验证 chat/agent 模式',
        category: '测试',
        systemPrompt: '你是一个可靠的助手。',
        welcomeMessage: '你好',
      },
    });
    expect(roleResponse.ok()).toBeTruthy();
    const roleId = (await roleResponse.json()).data.id as string;

    await page.addInitScript(([authToken, userObj]) => {
      localStorage.setItem('token', authToken);
      localStorage.setItem('user', JSON.stringify(userObj));
    }, [token, user]);

    await page.goto(`/chat/${roleId}`);
    await expect(page.getByText('欢迎使用 RoleCraft')).toBeVisible({ timeout: 15000 });
    await page.getByRole('button', { name: /新建对话/ }).click();

    const input = page.locator('textarea.chat-input');
    const sendButton = page.locator('button.send-btn');
    const modeSelect = page.locator('select[title="对话模式"]');
    await expect(input).toBeVisible({ timeout: 15000 });

    await modeSelect.selectOption('normal');
    await input.fill('请用一句话介绍你自己');
    await sendButton.click();

    const assistantBubbles = page.locator('.message.assistant .message-bubble');
    await expect(assistantBubbles.last()).toContainText(/.+/, { timeout: 45000 });
    await expect(assistantBubbles.last()).not.toContainText('生成失败，请重试。');
    await expect(assistantBubbles.last()).not.toContainText('Mock AI 助手');

    let deepRequestBody = '';
    page.on('request', (request) => {
      if (request.url().includes('/stream-with-thinking') && request.method() === 'POST') {
        deepRequestBody = request.postData() || '';
      }
    });

    await modeSelect.selectOption('deep');
    await input.fill('给我 3 条关于前端性能优化的建议');
    await sendButton.click();

    await expect(page.locator('.agent-mode-badge').last()).toHaveText('Agent', { timeout: 45000 });
    await expect(page.locator('.message.assistant .message-bubble').last()).toContainText(/.+/, {
      timeout: 45000,
    });
    expect(deepRequestBody).toContain('@agent');
  });
});
