import { test, expect } from '@playwright/test';

const baseUrl = 'http://localhost:3001';
const apiUrl = 'http://localhost:8080';

test('production frontend has no browser console errors on core pages', async ({ page }) => {
  const email = `console_${Date.now()}@example.com`;
  const password = 'password123';

  const consoleErrors: string[] = [];

  page.on('console', (msg) => {
    if (msg.type() === 'error') {
      consoleErrors.push(msg.text());
    }
  });

  page.on('pageerror', (err) => {
    consoleErrors.push(`pageerror: ${err.message}`);
  });

  // Register and login
  await page.goto(`${baseUrl}/register`);
  await page.getByLabel('Full Name').fill('Console QA');
  await page.getByLabel('Email').fill(email);
  await page.getByLabel('Password').fill(password);
  await page.getByRole('button', { name: 'Create account' }).click();
  await expect(page).toHaveURL(/\/login$/);

  await page.getByLabel('Email').fill(email);
  await page.getByLabel('Password').fill(password);
  await page.getByRole('button', { name: 'Sign in' }).click();
  await expect(page).toHaveURL(/\/projects$/);

  // Create project via API and visit project detail page
  const projectId = await page.evaluate(async (api) => {
    const token = localStorage.getItem('taskflow_token');
    const response = await fetch(`${api}/projects`, {
      method: 'POST',
      headers: {
        Authorization: `Bearer ${token}`,
        'Content-Type': 'application/json',
      },
      body: JSON.stringify({ name: `Console Project ${Date.now()}` }),
    });

    const body = await response.json();
    return body.id as string;
  }, apiUrl);

  expect(projectId).toBeTruthy();

  await page.goto(`${baseUrl}/projects/${projectId}`);
  await expect(page.getByRole('button', { name: 'Add Task' })).toBeVisible();

  expect(consoleErrors).toEqual([]);
});
