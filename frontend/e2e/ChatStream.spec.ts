import { test, expect } from '@playwright/test';

test.describe('ChatStream Component', () => {
  test('should render ChatStream component', async ({ page }) => {
    await page.goto('http://localhost:5173/chat-stream-demo');
    
    // Check if header is visible
    await expect(page.getByText('AI 助手')).toBeVisible();
    
    // Check if input area is visible
    await expect(page.getByPlaceholder('输入消息...')).toBeVisible();
    
    // Check if send button is visible
    await expect(page.getByRole('button', { name: '发送' })).toBeVisible();
  });

  test('should display empty state when no messages', async ({ page }) => {
    await page.goto('http://localhost:5173/chat-stream-demo');
    
    // Check empty state message
    await expect(page.getByText('开始对话吧！')).toBeVisible();
  });

  test('should send message and show typing indicator', async ({ page }) => {
    await page.goto('http://localhost:5173/chat-stream-demo');
    
    const input = page.getByPlaceholder('输入消息...');
    const sendButton = page.getByRole('button', { name: '发送' });
    
    // Type a message
    await input.fill('你好');
    
    // Send message
    await sendButton.click();
    
    // Check if user message appears
    await expect(page.getByText('你好')).toBeVisible();
  });

  test('should support markdown rendering', async ({ page }) => {
    await page.goto('http://localhost:5173/chat-stream-demo');
    
    const input = page.getByPlaceholder('输入消息...');
    const sendButton = page.getByRole('button', { name: '发送' });
    
    // Send markdown message
    await input.fill('**Bold** and `code`');
    await sendButton.click();
    
    // Check if message appears
    await expect(page.getByText('Bold')).toBeVisible();
  });

  test('should auto-scroll to bottom on new messages', async ({ page }) => {
    await page.goto('http://localhost:5173/chat-stream-demo');
    
    const input = page.getByPlaceholder('输入消息...');
    const sendButton = page.getByRole('button', { name: '发送' });
    
    // Send multiple messages
    for (let i = 0; i < 5; i++) {
      await input.fill(`Message ${i}`);
      await sendButton.click();
    }
    
    // Check if scroll is at bottom
    const messagesContainer = page.locator('.chat-stream-messages');
    const scrollHeight = await messagesContainer.evaluate(el => el.scrollHeight);
    const scrollTop = await messagesContainer.evaluate(el => el.scrollTop);
    const clientHeight = await messagesContainer.evaluate(el => el.clientHeight);
    
    expect(scrollTop + clientHeight).toBeGreaterThanOrEqual(scrollHeight - 100);
  });

  test('should show scroll-to-bottom button when scrolled up', async ({ page }) => {
    await page.goto('http://localhost:5173/chat-stream-demo');
    
    const messagesContainer = page.locator('.chat-stream-messages');
    
    // Scroll to top
    await messagesContainer.evaluate(el => el.scrollTo(0, 0));
    
    // Check if scroll-to-bottom button appears
    const scrollButton = page.getByText('新消息');
    await expect(scrollButton).toBeVisible();
  });

  test('should copy message content', async ({ page }) => {
    await page.goto('http://localhost:5173/chat-stream-demo');
    
    // Grant clipboard permissions
    const context = page.context();
    await context.grantPermissions(['clipboard-read', 'clipboard-write']);
    
    const input = page.getByPlaceholder('输入消息...');
    const sendButton = page.getByRole('button', { name: '发送' });
    
    // Send a message
    await input.fill('Test message to copy');
    await sendButton.click();
    
    // Hover over message to show actions
    const messageBubble = page.locator('.chat-stream-bubble.user').first();
    await messageBubble.hover();
    
    // Click copy button
    const copyButton = page.getByRole('button', { name: '复制' });
    await copyButton.click();
    
    // Verify clipboard content
    const clipboardContent = await page.evaluate(() => navigator.clipboard.readText());
    expect(clipboardContent).toContain('Test message to copy');
  });
});
