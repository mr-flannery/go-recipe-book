import { test as base, expect, Page } from '@playwright/test';
import { TEST_USERS } from './test-users';
import { fillToastEditor } from './editor-helpers';

type AuthFixtures = {
  user1Page: Page;
  user2Page: Page;
};

const test = base.extend<AuthFixtures>({
  user1Page: async ({ browser }, use) => {
    const context = await browser.newContext();
    const page = await context.newPage();
    const user = TEST_USERS.approved1;
    await page.goto('/login');
    await page.locator('input[name="email"]').fill(user.email);
    await page.locator('input[name="password"]').fill(user.password);
    await page.getByRole('button', { name: 'Sign In' }).click();
    await page.waitForURL('/');
    await use(page);
    await context.close();
  },
  user2Page: async ({ browser }, use) => {
    const context = await browser.newContext();
    const page = await context.newPage();
    const user = TEST_USERS.approved2;
    await page.goto('/login');
    await page.locator('input[name="email"]').fill(user.email);
    await page.locator('input[name="password"]').fill(user.password);
    await page.getByRole('button', { name: 'Sign In' }).click();
    await page.waitForURL('/');
    await use(page);
    await context.close();
  },
});

test.describe('Recipe Deletion', () => {
  const uniqueId = Date.now();

  test('author can delete recipe and non-author cannot', async ({ user1Page, user2Page }) => {
    const testRecipe = {
      title: `Delete Test Recipe ${uniqueId}`,
      prepTime: '15',
      cookTime: '30',
      calories: '400',
      ingredients: '- 2 cups flour\n- 1 cup sugar',
      instructions: '1. Mix ingredients\n2. Bake at 350F',
    };

    // User 1 creates a recipe
    await user1Page.goto('/recipes/create');
    await user1Page.getByRole('textbox', { name: 'Title' }).fill(testRecipe.title);
    await user1Page.locator('#preptime').fill(testRecipe.prepTime);
    await user1Page.locator('#cooktime').fill(testRecipe.cookTime);
    await user1Page.locator('#calories').fill(testRecipe.calories);
    await fillToastEditor(user1Page, 'ingredients-editor', testRecipe.ingredients);
    await fillToastEditor(user1Page, 'instructions-editor', testRecipe.instructions);

    await user1Page.getByRole('button', { name: /Create Recipe|Submit/i }).click();
    await user1Page.waitForURL(/\/recipes\/\d+/);

    const url = user1Page.url();
    const recipeId = url.match(/\/recipes\/(\d+)/)?.[1] || '';
    expect(recipeId).toBeTruthy();

    // Verify author sees Delete button
    await expect(user1Page.getByRole('button', { name: 'Delete' })).toBeVisible();

    // Non-author (user2) views the same recipe - should NOT see Delete button
    await user2Page.goto(`/recipes/${recipeId}`);
    await expect(user2Page.getByRole('heading', { name: testRecipe.title, level: 1 })).toBeVisible();
    await expect(user2Page.getByRole('button', { name: 'Delete' })).not.toBeVisible();

    // Non-author tries to delete via API directly - should be forbidden
    const deleteResponse = await user2Page.request.delete(`/recipes/${recipeId}/delete`);
    expect(deleteResponse.status()).toBe(403);
    const responseText = await deleteResponse.text();
    expect(responseText).toContain('Forbidden');

    // Verify recipe still exists after failed delete attempt
    await user2Page.reload();
    await expect(user2Page.getByRole('heading', { name: testRecipe.title, level: 1 })).toBeVisible();

    // Author deletes the recipe
    user1Page.on('dialog', dialog => dialog.accept());
    await user1Page.getByRole('button', { name: 'Delete' }).click();

    // Should redirect to recipes list
    await user1Page.waitForURL('/recipes');

    // Verify recipe is no longer on the overview page
    await expect(user1Page.getByText(testRecipe.title)).not.toBeVisible();

    // Verify recipe is not accessible anymore
    const viewResponse = await user1Page.goto(`/recipes/${recipeId}`);
    expect(viewResponse?.status()).toBe(404);
  });
});
