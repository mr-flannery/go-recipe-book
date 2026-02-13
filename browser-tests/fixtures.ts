import { test as base, Page } from '@playwright/test';
import { TEST_USERS } from './test-users';

type AuthFixtures = {
  authenticatedPage: Page;
};

export const test = base.extend<AuthFixtures>({
  authenticatedPage: async ({ page }, use) => {
    const user = TEST_USERS.approved1;
    await page.goto('/login');
    await page.locator('input[name="email"]').fill(user.email);
    await page.locator('input[name="password"]').fill(user.password);
    await page.getByRole('button', { name: 'Sign In' }).click();
    
    await page.waitForURL('/');
    
    await use(page);
  },
});

export { expect } from '@playwright/test';
