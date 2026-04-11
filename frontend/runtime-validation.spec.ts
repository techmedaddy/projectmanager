import { test, expect } from '@playwright/test';

test('runtime flow: register -> login -> projects -> task create/edit/filter and no CORS console errors', async ({ page }) => {
  const baseUrl = 'http://localhost:3001';
  const apiUrl = 'http://localhost:8080';
  const email = `runtime_${Date.now()}@example.com`;
  const password = 'password123';
  const projectName = `Runtime Project ${Date.now()}`;
  const taskTitle = `Runtime Task ${Date.now()}`;

  const consoleIssues: string[] = [];
  const corsIssues: string[] = [];
  const tasksRequestUrls: string[] = [];

  const corsPatterns = [/cors/i, /access-control-allow-origin/i, /blocked by cors policy/i];

  page.on('console', (msg) => {
    const text = msg.text();
    if (msg.type() === 'error') {
      consoleIssues.push(text);
    }
    if (corsPatterns.some((pattern) => pattern.test(text))) {
      corsIssues.push(text);
    }
  });

  page.on('pageerror', (err) => {
    consoleIssues.push(`pageerror: ${err.message}`);
  });

  page.on('requestfinished', (request) => {
    const url = request.url();
    if (url.includes('/projects/') && url.includes('/tasks')) {
      tasksRequestUrls.push(url);
    }
  });

  // Register
  await page.goto(`${baseUrl}/register`);
  await page.getByLabel('Full Name').fill('Runtime Validate');
  await page.getByLabel('Email').fill(email);
  await page.getByLabel('Password').fill(password);
  await page.getByRole('button', { name: 'Create account' }).click();
  await expect(page).toHaveURL(/\/login$/);

  // Login
  await page.getByLabel('Email').fill(email);
  await page.getByLabel('Password').fill(password);
  await page.getByRole('button', { name: 'Sign in' }).click();
  await expect(page).toHaveURL(/\/projects$/);

  // Create owned project through API with current token (stable setup for task flow)
  const projectId = await page.evaluate(async ({ api, name }) => {
    const token = localStorage.getItem('taskflow_token');
    const response = await fetch(`${api}/projects`, {
      method: 'POST',
      headers: {
        Authorization: `Bearer ${token}`,
        'Content-Type': 'application/json',
      },
      body: JSON.stringify({ name, description: 'Runtime validation project' }),
    });
    const body = await response.json();
    return body.id as string;
  }, { api: apiUrl, name: projectName });

  expect(projectId).toBeTruthy();

  await page.goto(`${baseUrl}/projects/${projectId}`);
  await expect(page.getByRole('button', { name: 'Add Task' })).toBeVisible();

  // Create task in UI
  await page.getByRole('button', { name: 'Add Task' }).click();
  await page.getByLabel('Title').fill(taskTitle);
  await page.getByLabel('Description').fill('Created in runtime validation');
  await page.getByLabel('Due Date').fill('2026-05-01');
  await page.getByRole('button', { name: 'Create Task' }).click();
  await expect(page.getByText(taskTitle).first()).toBeVisible();

  // Edit task: set assignee and done status
  await page.getByText(taskTitle).first().click();
  const userId = await page.evaluate(async (api) => {
    const token = localStorage.getItem('taskflow_token');
    const response = await fetch(`${api}/auth/me`, {
      headers: {
        Authorization: `Bearer ${token}`,
      },
    });
    const body = await response.json();
    return body.user.id as string;
  }, apiUrl);

  expect(userId).toBeTruthy();

  await page.getByLabel('Assignee ID').fill(userId);
  await page.locator('form').last().locator('select').first().selectOption('done');
  await page.getByRole('button', { name: 'Save Changes' }).click();

  // Filter flow: status + assignee
  await page.locator('select').nth(0).selectOption('done');
  await page.locator('select').nth(1).selectOption(userId);

  await expect(page.getByText(taskTitle).first()).toBeVisible();

  // Reset filter flow
  await page.getByRole('button', { name: 'Clear filters' }).click();

  // Request validation for filters
  const statusRequestSeen = tasksRequestUrls.some((url) => /\/tasks\?/.test(url) && /status=done/.test(url));
  const assigneeRequestSeen = tasksRequestUrls.some((url) => /\/tasks\?/.test(url) && /assignee=/.test(url));

  expect(statusRequestSeen).toBeTruthy();
  expect(assigneeRequestSeen).toBeTruthy();

  // Browser console CORS validation
  expect(corsIssues).toEqual([]);

  // Keep non-CORS console issues visible in output for debugging, but do not fail automatically.
  if (consoleIssues.length > 0) {
    console.log('Non-CORS console issues observed:', consoleIssues);
  }
});
