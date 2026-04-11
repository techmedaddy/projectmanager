import { test, expect, Page } from '@playwright/test';

const baseUrl = 'http://localhost:3001';
const apiUrl = 'http://localhost:8080';

async function registerAndLogin(page: Page, suffix: string) {
  const email = `responsive_${suffix}_${Date.now()}@example.com`;
  const password = 'password123';

  await page.goto(`${baseUrl}/register`);
  await page.getByLabel('Full Name').fill(`Responsive ${suffix}`);
  await page.getByLabel('Email').fill(email);
  await page.getByLabel('Password').fill(password);
  await page.getByRole('button', { name: 'Create account' }).click();
  await expect(page).toHaveURL(/\/login$/);

  await page.getByLabel('Email').fill(email);
  await page.getByLabel('Password').fill(password);
  await page.getByRole('button', { name: 'Sign in' }).click();
  await expect(page).toHaveURL(/\/projects$/);
}

async function mockAuthenticatedUser(page: Page) {
  await page.addInitScript(() => {
    localStorage.setItem('taskflow_token', 'qa-token');
  });

  await page.route('**/auth/me', async (route) => {
    await route.fulfill({
      status: 200,
      contentType: 'application/json',
      body: JSON.stringify({
        user: {
          id: '11111111-1111-1111-1111-111111111111',
          name: 'QA User',
          email: 'qa@example.com',
        },
      }),
    });
  });
}

test('responsive layout check at 375px and 1280px', async ({ page }) => {
  // 375px
  await page.setViewportSize({ width: 375, height: 812 });
  await page.goto(`${baseUrl}/login`);
  await expect(page.getByText('Welcome back')).toBeVisible();

  await page.goto(`${baseUrl}/register`);
  await expect(page.getByText('Create an account')).toBeVisible();

  await registerAndLogin(page, 'mobile');

  const projectId = await page.evaluate(async (api) => {
    const token = localStorage.getItem('taskflow_token');
    const response = await fetch(`${api}/projects`, {
      method: 'POST',
      headers: {
        Authorization: `Bearer ${token}`,
        'Content-Type': 'application/json',
      },
      body: JSON.stringify({ name: `Responsive Project ${Date.now()}`, description: 'layout check' }),
    });
    const body = await response.json();
    return body.id as string;
  }, apiUrl);

  await page.goto(`${baseUrl}/projects`);
  await expect(page.getByText('Projects')).toBeVisible();

  await page.goto(`${baseUrl}/projects/${projectId}`);
  await expect(page.getByRole('button', { name: 'Add Task' })).toBeVisible();

  const mobileOverflow = await page.evaluate(() => document.documentElement.scrollWidth <= window.innerWidth + 1);
  expect(mobileOverflow).toBeTruthy();

  // 1280px
  await page.setViewportSize({ width: 1280, height: 800 });
  await page.goto(`${baseUrl}/projects`);
  await expect(page.getByText('Projects')).toBeVisible();

  await page.goto(`${baseUrl}/projects/${projectId}`);
  await expect(page.getByRole('button', { name: 'Add Task' })).toBeVisible();

  const desktopOverflow = await page.evaluate(() => document.documentElement.scrollWidth <= window.innerWidth + 1);
  expect(desktopOverflow).toBeTruthy();
});

test('projects page shows loading, error, and empty states', async ({ page }) => {
  await mockAuthenticatedUser(page);

  // loading state
  await page.route(`${apiUrl}/projects`, async (route) => {
    if (route.request().method() !== 'GET') return route.continue();
    await new Promise((resolve) => setTimeout(resolve, 1500));
    await route.fulfill({
      status: 200,
      contentType: 'application/json',
      body: JSON.stringify({ projects: [] }),
    });
  });

  const loadingNav = page.goto(`${baseUrl}/projects`);
  await expect(page.getByText('Projects')).not.toBeVisible();
  await loadingNav;
  await expect(page.getByText('No projects yet')).toBeVisible();

  await page.unroute(`${apiUrl}/projects`);

  // error state
  await page.route(`${apiUrl}/projects`, async (route) => {
    if (route.request().method() !== 'GET') return route.continue();
    await route.fulfill({
      status: 500,
      contentType: 'application/json',
      body: JSON.stringify({ error: 'internal server error' }),
    });
  });

  await page.goto(`${baseUrl}/projects`);
  await expect(page.getByText('Failed to load projects. Please try again.')).toBeVisible();

  await page.unroute(`${apiUrl}/projects`);

  // empty state
  await page.route(`${apiUrl}/projects`, async (route) => {
    if (route.request().method() !== 'GET') return route.continue();
    await route.fulfill({
      status: 200,
      contentType: 'application/json',
      body: JSON.stringify({ projects: [] }),
    });
  });

  await page.goto(`${baseUrl}/projects`);
  await expect(page.getByText('No projects yet')).toBeVisible();
});

test('project detail shows loading, error, and filtered empty states', async ({ page }) => {
  await mockAuthenticatedUser(page);

  // project loading
  await page.route(`${apiUrl}/projects/22222222-2222-2222-2222-222222222222`, async (route) => {
    await new Promise((resolve) => setTimeout(resolve, 1500));
    await route.fulfill({
      status: 200,
      contentType: 'application/json',
      body: JSON.stringify({
        project: {
          id: '22222222-2222-2222-2222-222222222222',
          name: 'Detail QA',
          description: 'detail states',
          owner_id: '11111111-1111-1111-1111-111111111111',
          created_at: new Date().toISOString(),
        },
        tasks: [
          {
            id: '33333333-3333-3333-3333-333333333333',
            title: 'Todo Task',
            description: 'Task desc',
            status: 'todo',
            priority: 'medium',
            project_id: '22222222-2222-2222-2222-222222222222',
            assignee_id: '11111111-1111-1111-1111-111111111111',
            creator_id: '11111111-1111-1111-1111-111111111111',
            due_date: null,
            created_at: new Date().toISOString(),
            updated_at: new Date().toISOString(),
          },
        ],
      }),
    });
  });

  // tasks loading then success
  await page.route(`${apiUrl}/projects/22222222-2222-2222-2222-222222222222/tasks**`, async (route) => {
    await new Promise((resolve) => setTimeout(resolve, 1200));
    const url = new URL(route.request().url());
    const status = url.searchParams.get('status');
    const tasks = status === 'done'
      ? []
      : [
          {
            id: '33333333-3333-3333-3333-333333333333',
            title: 'Todo Task',
            description: 'Task desc',
            status: 'todo',
            priority: 'medium',
            project_id: '22222222-2222-2222-2222-222222222222',
            assignee_id: '11111111-1111-1111-1111-111111111111',
            creator_id: '11111111-1111-1111-1111-111111111111',
            due_date: null,
            created_at: new Date().toISOString(),
            updated_at: new Date().toISOString(),
          },
        ];

    await route.fulfill({
      status: 200,
      contentType: 'application/json',
      body: JSON.stringify({ tasks }),
    });
  });

  const detailNav = page.goto(`${baseUrl}/projects/22222222-2222-2222-2222-222222222222`);
  await expect(page.getByRole('button', { name: 'Add Task' })).not.toBeVisible();
  await detailNav;
  await expect(page.getByRole('button', { name: 'Add Task' })).toBeVisible();

  await page.locator('select').nth(0).selectOption('done');
  await expect(page.getByText('Updating…')).toBeVisible();
  await expect(page.getByText('No tasks match current filters')).toBeVisible();

  // tasks error state
  await page.unroute(`${apiUrl}/projects/22222222-2222-2222-2222-222222222222/tasks**`);
  await page.route(`${apiUrl}/projects/22222222-2222-2222-2222-222222222222/tasks**`, async (route) => {
    await route.fulfill({
      status: 500,
      contentType: 'application/json',
      body: JSON.stringify({ error: 'internal server error' }),
    });
  });

  await page.goto(`${baseUrl}/projects/22222222-2222-2222-2222-222222222222`);
  await expect(page.getByText('Failed to load filtered tasks.')).toBeVisible();

  // project error state
  await page.unroute(`${apiUrl}/projects/22222222-2222-2222-2222-222222222222`);
  await page.route(`${apiUrl}/projects/22222222-2222-2222-2222-222222222222`, async (route) => {
    await route.fulfill({
      status: 404,
      contentType: 'application/json',
      body: JSON.stringify({ error: 'not found' }),
    });
  });

  await page.goto(`${baseUrl}/projects/22222222-2222-2222-2222-222222222222`);
  await expect(page.getByText('Project not found.')).toBeVisible();
});
