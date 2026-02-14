import { chromium, type Page } from '@playwright/test';
import { ADMIN_USER, TEST_USERS } from './test-users';

const BASE_URL = 'http://localhost:8080';

async function tryLogin(page: Page, email: string, password: string): Promise<boolean> {
  await page.goto(`${BASE_URL}/login`);
  await page.locator('input[name="email"]').fill(email);
  await page.locator('input[name="password"]').fill(password);
  await page.getByRole('button', { name: 'Sign in' }).click();
  await page.waitForURL(url => url.pathname === '/' || url.pathname === '/login');
  return page.url() === `${BASE_URL}/`;
}

async function registerUser(page: Page, user: { username: string; email: string; password: string }) {
  await page.goto(`${BASE_URL}/register`);
  await page.locator('#username').fill(user.username);
  await page.locator('#email').fill(user.email);
  await page.locator('#password').fill(user.password);
  await page.locator('#confirm_password').fill(user.password);
  await page.getByRole('button', { name: 'Send Registration Request' }).click();
  await page.waitForSelector('.success');
}

async function approveUser(page: Page, username: string) {
  await page.goto(`${BASE_URL}/admin/registrations`);
  const userCard = page.locator('.registration-card', { hasText: username });
  if (await userCard.count() > 0) {
    await userCard.getByRole('button', { name: 'Approve' }).click();
  }
}

async function globalSetup() {
  const browser = await chromium.launch();
  const page = await browser.newPage();

  const usersToSetup = [TEST_USERS.approved1, TEST_USERS.approved2];
  const usersNeedingApproval: typeof usersToSetup = [];

  for (const user of usersToSetup) {
    const loginSuccess = await tryLogin(page, user.email, user.password);
    if (loginSuccess) {
      await page.getByRole('link', { name: 'Logout' }).click();
    } else {
      await registerUser(page, user);
      usersNeedingApproval.push(user);
    }
  }

  if (usersNeedingApproval.length > 0) {
    await tryLogin(page, ADMIN_USER.email, ADMIN_USER.password);
    for (const user of usersNeedingApproval) {
      await approveUser(page, user.username);
    }
    await page.getByRole('link', { name: 'Logout' }).click();
  }

  await browser.close();
}

export default globalSetup;
