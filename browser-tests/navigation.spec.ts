import { test, expect } from '@playwright/test';

test.describe('Navigation', () => {
  test('navigates from landing page to recipe overview', async ({ page }) => {
    await page.goto('/');

    await expect(page).toHaveTitle(/Recipe Book|Taste/);

    await page.getByRole('link', { name: 'Browse Recipes' }).click();

    await page.waitForURL('/recipes');
    await expect(page.getByRole('heading', { name: 'Recipe Collection', level: 1 })).toBeVisible();
  });
});
