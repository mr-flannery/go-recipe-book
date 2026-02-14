import { test, expect, Page } from '@playwright/test';
import { ADMIN_USER, TEST_USERS } from './test-users';

async function registerUser(page: Page, user: { username: string; email: string; password: string }) {
  await page.goto('/register');
  await page.locator('#username').fill(user.username);
  await page.locator('#email').fill(user.email);
  await page.locator('#password').fill(user.password);
  await page.locator('#confirm_password').fill(user.password);
  await page.getByRole('button', { name: 'Send Registration Request' }).click();
  await expect(page.locator('.success')).toBeVisible();
}

async function loginUser(page: Page, email: string, password: string) {
  await page.goto('/login');
  await page.locator('input[name="email"]').fill(email);
  await page.locator('input[name="password"]').fill(password);
  await page.getByRole('button', { name: 'Sign in' }).click();
}

test.describe('User Registration', () => {
  test('approved users can login successfully', async ({ page }) => {
    await loginUser(page, TEST_USERS.approved1.email, TEST_USERS.approved1.password);
    await expect(page).toHaveURL('/');
    await expect(page.locator('.user-greeting')).toContainText(TEST_USERS.approved1.username);
    await page.getByRole('link', { name: 'Logout' }).click();

    await loginUser(page, TEST_USERS.approved2.email, TEST_USERS.approved2.password);
    await expect(page).toHaveURL('/');
    await expect(page.locator('.user-greeting')).toContainText(TEST_USERS.approved2.username);
    await page.getByRole('link', { name: 'Logout' }).click();
  });

  test('rejected user cannot login', async ({ page }) => {
    const rejectedUser = {
      username: `rejecteduser_${Date.now()}`,
      email: `rejected_${Date.now()}@example.com`,
      password: 'RejectedPass789!',
    };

    await registerUser(page, rejectedUser);

    await loginUser(page, ADMIN_USER.email, ADMIN_USER.password);
    await page.goto('/admin/registrations');
    const rejectedCard = page.locator('.registration-card', { hasText: rejectedUser.username });
    await rejectedCard.getByRole('button', { name: 'Deny' }).click();
    await page.getByRole('link', { name: 'Logout' }).click();

    await loginUser(page, rejectedUser.email, rejectedUser.password);
    await expect(page.locator('.error')).toBeVisible();
    await expect(page).toHaveURL('/login');
  });
});
