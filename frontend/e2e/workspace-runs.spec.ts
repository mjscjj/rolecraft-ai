import { expect, test } from '@playwright/test';

const API_BASE = process.env.E2E_API_BASE || 'http://localhost:8080/api/v1';

test.describe('工作区协商记录', () => {
  test('可查看 run 列表和详情', async ({ page, request }) => {
    const now = Date.now();
    const email = `workspace_e2e_${now}@rolecraft.ai`;
    const password = 'E2ETest123!';

    const registerResp = await request.post(`${API_BASE}/auth/register`, {
      data: {
        email,
        password,
        name: 'Workspace E2E User',
      },
    });
    expect(registerResp.ok()).toBeTruthy();
    const registerData = await registerResp.json();
    const token = registerData?.data?.token as string;
    const user = registerData?.data?.user as Record<string, unknown>;
    expect(token).toBeTruthy();

    const authHeaders = {
      Authorization: `Bearer ${token}`,
      'Content-Type': 'application/json',
    };

    const createWorkResp = await request.post(`${API_BASE}/workspaces`, {
      headers: authHeaders,
      data: {
        name: `E2E Workspace ${now}`,
        description: '验证多 Agent 协商记录展示',
        type: 'analyze',
        triggerType: 'manual',
        inputSource: 'e2e:test',
        reportRule: '输出摘要与步骤',
      },
    });
    expect(createWorkResp.ok()).toBeTruthy();
    const createWorkData = await createWorkResp.json();
    const workId = createWorkData?.data?.id as string;
    expect(workId).toBeTruthy();

    const runResp = await request.post(`${API_BASE}/workspaces/${workId}/run`, {
      headers: {
        Authorization: `Bearer ${token}`,
      },
    });
    expect(runResp.ok()).toBeTruthy();
    const runData = await runResp.json();
    const runId = runData?.data?.run?.id as string;
    expect(runId).toBeTruthy();

    await page.addInitScript(
      ({ authToken, authUser }) => {
        localStorage.setItem('token', authToken);
        localStorage.setItem('user', JSON.stringify(authUser));
      },
      { authToken: token, authUser: user }
    );

    await page.goto('/workspaces');
    await expect(page.locator(`[data-testid="workspace-item-${workId}"]`)).toBeVisible();

    await page.click(`[data-testid="workspace-toggle-runs-${workId}"]`);
    await expect(page.locator(`[data-testid="workspace-runs-${workId}"]`)).toBeVisible();
    await expect(page.locator(`[data-testid="workspace-run-item-${runId}"]`)).toBeVisible();

    await page.click(`[data-testid="workspace-run-detail-${runId}"]`);
    await expect(page.getByText('执行详情')).toBeVisible();
    await expect(page.getByText('多 Agent 协商过程')).toBeVisible();
  });
});
