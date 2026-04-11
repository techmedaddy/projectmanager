import { chromium } from 'playwright';

const baseUrl = 'http://localhost:3001';
const apiUrl = 'http://localhost:8080';

const corsPatterns = [/cors/i, /access-control-allow-origin/i, /blocked by cors policy/i];
const consoleErrors = [];
const corsConsoleIssues = [];
const taskRequestUrls = [];

function assert(condition, message) {
  if (!condition) throw new Error(message);
}

const email = `runtime_${Date.now()}@example.com`;
const password = 'password123';
const name = 'Runtime Validate';
const projectName = `Runtime Project ${Date.now()}`;
const taskTitle = `Runtime Task ${Date.now()}`;

const browser = await chromium.launch({ headless: true });
const page = await browser.newPage({ viewport: { width: 1280, height: 800 } });

page.on('console', (msg) => {
  const text = msg.text();
  if (msg.type() === 'error') {
    consoleErrors.push(text);
  }
  if (corsPatterns.some((p) => p.test(text))) {
    corsConsoleIssues.push(text);
  }
});

page.on('pageerror', (err) => {
  consoleErrors.push(`pageerror: ${err.message}`);
});

page.on('requestfinished', (request) => {
  const url = request.url();
  if (url.includes('/projects/') && url.includes('/tasks')) {
    taskRequestUrls.push(url);
  }
});

try {
  await page.goto(`${baseUrl}/register`, { waitUntil: 'networkidle' });
  await page.getByLabel('Full Name').fill(name);
  await page.getByLabel('Email').fill(email);
  await page.getByLabel('Password').fill(password);
  await page.getByRole('button', { name: 'Create account' }).click();
  await page.waitForURL('**/login', { timeout: 15000 });

  await page.getByLabel('Email').fill(email);
  await page.getByLabel('Password').fill(password);
  await page.getByRole('button', { name: 'Sign in' }).click();
  await page.waitForURL('**/projects', { timeout: 15000 });

  await page.getByRole('button', { name: 'New Project' }).click();
  await page.getByLabel('Name').fill(projectName);
  await page.getByLabel('Description (Optional)').fill('Runtime validation project');
  await page.getByRole('button', { name: /^Create$/ }).click();
  await page.getByRole('link', { name: projectName }).click();
  await page.waitForURL('**/projects/*', { timeout: 15000 });

  await page.getByRole('button', { name: 'Add Task' }).click();
  await page.getByLabel('Title').fill(taskTitle);
  await page.getByLabel('Description').fill('Created in runtime validation');
  await page.getByLabel('Due Date').fill('2026-05-01');
  await page.getByRole('button', { name: 'Create Task' }).click();
  await page.getByText(taskTitle, { exact: false }).first().waitFor({ timeout: 15000 });

  await page.getByText(taskTitle, { exact: false }).first().click();
  const userId = await page.evaluate(async (api) => {
    const token = localStorage.getItem('taskflow_token');
    const res = await fetch(`${api}/auth/me`, {
      headers: token ? { Authorization: `Bearer ${token}` } : {},
    });
    const body = await res.json();
    return body?.user?.id ?? '';
  }, apiUrl);

  assert(Boolean(userId), 'Could not resolve logged-in user id for assignee flow');

  await page.getByLabel('Assignee ID').fill(userId);
  await page.locator('label:has-text("Status")').locator('xpath=..').locator('select').selectOption('done');
  await page.getByRole('button', { name: 'Save Changes' }).click();

  await page.locator('label:has-text("Status")').locator('xpath=..').locator('select').first().selectOption('done');
  await page.locator('label:has-text("Assignee")').locator('xpath=..').locator('select').selectOption(userId);

  await page.getByText(taskTitle, { exact: false }).first().waitFor({ timeout: 10000 });

  await page.getByRole('button', { name: 'Clear filters' }).click();

  const hasStatusFilterRequest = taskRequestUrls.some((u) => /\/tasks\?/.test(u) && /status=done/.test(u));
  const hasAssigneeFilterRequest = taskRequestUrls.some((u) => /\/tasks\?/.test(u) && /assignee=/.test(u));

  assert(hasStatusFilterRequest, 'No tasks API request detected with status filter');
  assert(hasAssigneeFilterRequest, 'No tasks API request detected with assignee filter');

  assert(corsConsoleIssues.length === 0, `CORS-related browser console issue detected: ${corsConsoleIssues.join(' | ')}`);

  console.log('✅ Runtime flow validation passed');
  console.log(`Registered user: ${email}`);
  console.log(`Project: ${projectName}`);
  console.log(`Task: ${taskTitle}`);
  console.log('Observed filtered task requests:');
  [...new Set(taskRequestUrls)].forEach((u) => console.log(`- ${u}`));
  if (consoleErrors.length) {
    console.log('Console errors observed (non-CORS):');
    consoleErrors.forEach((e) => console.log(`- ${e}`));
  }
} finally {
  await browser.close();
}
