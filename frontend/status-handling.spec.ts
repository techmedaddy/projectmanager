import { test, expect } from '@playwright/test';

const baseUrl = 'http://localhost:3001';
const apiUrl = 'http://localhost:8080';

async function registerAndLogin(page: any, suffix: string) {
  const email = `status_${suffix}_${Date.now()}@example.com`;
  const password = 'password123';

  await page.goto(`${baseUrl}/register`);
  await page.getByLabel('Full Name').fill(`Status ${suffix}`);
  await page.getByLabel('Email').fill(email);
  await page.getByLabel('Password').fill(password);
  await page.getByRole('button', { name: 'Create account' }).click();
  await expect(page).toHaveURL(/\/login$/);

  await page.getByLabel('Email').fill(email);
  await page.getByLabel('Password').fill(password);
  await page.getByRole('button', { name: 'Sign in' }).click();
  await expect(page).toHaveURL(/\/projects$/);

  const token = await page.evaluate(() => localStorage.getItem('taskflow_token'));
  return { email, password, token };
}

test('403 and 404 are visible in project detail UI', async ({ browser }) => {
  const ownerPage = await browser.newPage();
  const viewerPage = await browser.newPage();

  // owner setup
  const owner = await registerAndLogin(ownerPage, 'owner');
  const ownerProjectId = await ownerPage.evaluate(async ({ api }) => {
    const token = localStorage.getItem('taskflow_token');
    const response = await fetch(`${api}/projects`, {
      method: 'POST',
      headers: {
        Authorization: `Bearer ${token}`,
        'Content-Type': 'application/json',
      },
      body: JSON.stringify({ name: `Private ${Date.now()}`, description: 'forbidden test' }),
    });
    const body = await response.json();
    return body.id as string;
  }, { api: apiUrl });
  expect(ownerProjectId).toBeTruthy();

  // viewer cannot access owner's project => 403 visible
  await registerAndLogin(viewerPage, 'viewer');
  await viewerPage.goto(`${baseUrl}/projects/${ownerProjectId}`);
  await expect(viewerPage.getByText('You do not have access to this project.')).toBeVisible();
  await expect(viewerPage.getByText('Error code: 403')).toBeVisible();

  // non-existent project => 404 visible
  await viewerPage.goto(`${baseUrl}/projects/aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa`);
  await expect(viewerPage.getByText('Project not found.')).toBeVisible();
  await expect(viewerPage.getByText('Error code: 404')).toBeVisible();

  await ownerPage.close();
  await viewerPage.close();
});

test('401 handling is visible via protected-route redirect', async ({ page }) => {
  await page.goto(`${baseUrl}/projects`);
  // Ensure explicit unauthenticated route behavior is visible to user.
  await expect(page).toHaveURL(/\/login$/);
  await expect(page.getByText('Welcome back')).toBeVisible();
});
