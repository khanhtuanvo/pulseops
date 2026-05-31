import { expect, test, type Browser } from '@playwright/test';

async function authenticatedPage(browser: Browser, role = 'OWNER') {
  const context = await browser.newContext({
    extraHTTPHeaders: {
      'X-E2E-Test-User': role,
    },
  });
  const page = await context.newPage();
  return { context, page };
}

test('unauthenticated user is redirected to login', async ({ page }) => {
  await page.goto('/dashboard');

  await expect(page).toHaveURL(/\/login/);
  await expect(page.getByRole('button', { name: /sign in with google/i })).toBeVisible();
});

test('login page renders correctly', async ({ page }) => {
  await page.goto('/login');

  await expect(page.getByText('PulseOps')).toBeVisible();
  await expect(page.getByRole('button', { name: /sign in with google/i })).toBeEnabled();
});

test('authenticated user sees dashboard', async ({ browser }) => {
  const { context, page } = await authenticatedPage(browser);
  await page.goto('/dashboard');

  await expect(page).toHaveURL(/\/dashboard/);
  await expect(page.getByRole('heading', { name: /incident dashboard/i })).toBeVisible();
  await expect(page.getByText(/no incidents match|triggered/i).first()).toBeVisible();

  await context.close();
});

test('logout redirects to login', async ({ browser }) => {
  const { context, page } = await authenticatedPage(browser);
  await page.goto('/dashboard');

  await page.getByRole('button', { name: /logout/i }).click();

  await expect(page).toHaveURL(/\/login/);
  const cookies = await context.cookies();
  expect(cookies.find((cookie) => cookie.name === 'session')).toBeUndefined();

  await context.close();
});
