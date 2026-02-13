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

test.describe.serial('User Registration and Approval', () => {
  test('registers three test users', async ({ page }) => {
    await registerUser(page, TEST_USERS.approved1);
    await registerUser(page, TEST_USERS.approved2);
    await registerUser(page, TEST_USERS.rejected);
  });

  test('admin approves two users and rejects one', async ({ page }) => {
    // Login as admin
    await loginUser(page, ADMIN_USER.email, ADMIN_USER.password);
    await expect(page.getByRole('link', { name: 'Logout' })).toBeVisible();

    // Navigate to admin registrations page
    await page.goto('/admin/registrations');
    await expect(page.getByRole('heading', { name: /Pending Registration Requests/i })).toBeVisible();

    // Approve first user
    const user1Card = page.locator('.registration-card', { hasText: TEST_USERS.approved1.username });
    await user1Card.getByRole('button', { name: 'Approve' }).click();
    // await expect(page.locator('.success')).toBeVisible();

    // Approve second user
    const user2Card = page.locator('.registration-card', { hasText: TEST_USERS.approved2.username });
    await user2Card.getByRole('button', { name: 'Approve' }).click();
    // await expect(page.locator('.success')).toBeVisible();

    // Reject third user
    const rejectedCard = page.locator('.registration-card', { hasText: TEST_USERS.rejected.username });
    await rejectedCard.getByRole('button', { name: 'Deny' }).click();
    // await expect(page.locator('.success')).toBeVisible();

    // Logout
    await page.getByRole('link', { name: 'Logout' }).click();
  });

  test('approved users can login successfully', async ({ page }) => {
    // Test first approved user
    await loginUser(page, TEST_USERS.approved1.email, TEST_USERS.approved1.password);
    await expect(page).toHaveURL('/');
    await expect(page.locator('.user-greeting')).toContainText(TEST_USERS.approved1.username);
    await page.getByRole('link', { name: 'Logout' }).click();

    // Test second approved user
    await loginUser(page, TEST_USERS.approved2.email, TEST_USERS.approved2.password);
    await expect(page).toHaveURL('/');
    await expect(page.locator('.user-greeting')).toContainText(TEST_USERS.approved2.username);
    await page.getByRole('link', { name: 'Logout' }).click();
  });

  test('rejected user cannot login', async ({ page }) => {
    await loginUser(page, TEST_USERS.rejected.email, TEST_USERS.rejected.password);
    await expect(page.locator('.error')).toBeVisible();
    await expect(page).toHaveURL('/login');
  });
});
