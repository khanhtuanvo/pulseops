import { expect, test, type APIRequestContext, type Browser } from '@playwright/test';

// These tests require the full stack running locally with ENV=test (the e2e
// docker-compose stack). The backend seeds a deterministic team whose id
// matches the e2e auth bypass, with the API key below.
const API_URL = process.env.VITE_API_URL ?? 'http://localhost:8080';
const API_KEY = 'e2e-test-api-key';

// dashboardPage opens an authenticated dashboard for the given role and waits
// until the live subscription websocket has connected. The Hub does not replay
// events, so an alert posted before the socket is open would be missed.
async function dashboardPage(browser: Browser, role = 'OWNER') {
  const context = await browser.newContext({
    extraHTTPHeaders: { 'X-E2E-Test-User': role },
  });
  const page = await context.newPage();

  const websocketOpened = page.waitForEvent('websocket', { timeout: 10_000 }).catch(() => null);
  await page.goto('/dashboard');
  await expect(page.getByRole('heading', { name: /incident dashboard/i })).toBeVisible();
  await websocketOpened;
  // Small buffer for the graphql-ws connection_init/subscribe handshake to
  // register the subscriber on the server-side Hub.
  await page.waitForTimeout(500);

  return { context, page };
}

async function postAlert(request: APIRequestContext, severity = 'CRITICAL') {
  const alertName = `E2E-${Date.now()}-${Math.random().toString(36).slice(2, 8)}`;
  const response = await request.post(`${API_URL}/webhooks/alerts`, {
    headers: { 'X-API-Key': API_KEY, 'Content-Type': 'application/json' },
    data: { source: 'e2e', alertName, severity, environment: 'prod' },
  });
  expect(response.ok()).toBeTruthy();

  return alertName;
}

test('incident appears in the dashboard feed in real time', async ({ browser, request }) => {
  const { context, page } = await dashboardPage(browser);

  const alertName = await postAlert(request, 'CRITICAL');
  const card = page.locator('article').filter({ hasText: alertName });

  await expect(card).toBeVisible({ timeout: 5_000 });
  await expect(card.getByText('CRITICAL')).toBeVisible();
  await expect(card.getByText('TRIGGERED')).toBeVisible();

  await context.close();
});

test('acknowledging an incident updates the card', async ({ browser, request }) => {
  const { context, page } = await dashboardPage(browser, 'RESPONDER');

  const alertName = await postAlert(request);
  const card = page.locator('article').filter({ hasText: alertName });
  await expect(card).toBeVisible({ timeout: 5_000 });

  await card.getByRole('button', { name: 'Ack' }).click();

  await expect(card.getByText('ACKNOWLEDGED')).toBeVisible({ timeout: 3_000 });
  await expect(card.getByRole('button', { name: 'Ack' })).toHaveCount(0);

  await context.close();
});

test('viewer cannot see the acknowledge button', async ({ browser, request }) => {
  const { context, page } = await dashboardPage(browser, 'VIEWER');

  const alertName = await postAlert(request);
  const card = page.locator('article').filter({ hasText: alertName });
  await expect(card).toBeVisible({ timeout: 5_000 });

  await expect(card.getByRole('button', { name: 'Ack' })).toHaveCount(0);

  await context.close();
});

test('two dashboards receive the same incident event', async ({ browser, request }) => {
  const tabA = await dashboardPage(browser);
  const tabB = await dashboardPage(browser);

  const alertName = await postAlert(request);
  const cardA = tabA.page.locator('article').filter({ hasText: alertName });
  const cardB = tabB.page.locator('article').filter({ hasText: alertName });

  await expect(cardA).toBeVisible({ timeout: 5_000 });
  await expect(cardB).toBeVisible({ timeout: 5_000 });

  await tabA.context.close();
  await tabB.context.close();
});
