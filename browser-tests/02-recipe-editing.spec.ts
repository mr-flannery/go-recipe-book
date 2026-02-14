import { test as base, expect, Page } from '@playwright/test';
import { TEST_USERS } from './test-users';

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

test.describe('Recipe Editing', () => {
  const uniqueId = Date.now();
  const testRecipe = {
    title: `Edit Test Recipe ${uniqueId}`,
    prepTime: '10',
    cookTime: '20',
    calories: '300',
    ingredients: '- 1 cup rice\n- 2 cups water',
    instructions: '1. Boil water\n2. Add rice\n3. Simmer for 20 minutes',
    tags: ['dinner', 'easy'],
  };

  let recipeId: string;

  test('author can create, edit, save changes, and non-author cannot edit', async ({ user1Page, user2Page }) => {
    // User 1 creates a recipe
    await user1Page.goto('/recipes/create');
    await user1Page.getByRole('textbox', { name: 'Title' }).fill(testRecipe.title);
    await user1Page.locator('#preptime').fill(testRecipe.prepTime);
    await user1Page.locator('#cooktime').fill(testRecipe.cookTime);
    await user1Page.locator('#calories').fill(testRecipe.calories);
    await user1Page.locator('#ingredients').fill(testRecipe.ingredients);
    await user1Page.locator('#instructions').fill(testRecipe.instructions);

    // Add tags
    for (const tag of testRecipe.tags) {
      await user1Page.locator('#tags-input').fill(tag);
      await user1Page.locator('#tags-input').press('Enter');
    }

    await user1Page.getByRole('button', { name: /Create Recipe|Submit/i }).click();
    await user1Page.waitForURL(/\/recipes\/\d+/);

    // Extract recipe ID from URL
    const url = user1Page.url();
    recipeId = url.match(/\/recipes\/(\d+)/)?.[1] || '';
    expect(recipeId).toBeTruthy();

    // Verify author sees Edit button
    await expect(user1Page.getByRole('link', { name: 'Edit Recipe' })).toBeVisible();

    // Non-author (user2) views the same recipe - should NOT see Edit button
    await user2Page.goto(`/recipes/${recipeId}`);
    await expect(user2Page.getByRole('heading', { name: testRecipe.title, level: 1 })).toBeVisible();
    await expect(user2Page.getByRole('link', { name: 'Edit Recipe' })).not.toBeVisible();

    // Non-author tries to access edit page directly - should be forbidden or redirected
    const response = await user2Page.goto(`/recipes/update?id=${recipeId}`);
    // The page loads but submitting changes should be forbidden
    // Let's verify non-author cannot submit changes
    await user2Page.locator('#title').fill('Hacked Title');
    await user2Page.getByRole('button', { name: 'Update Recipe' }).click();
    // Should get forbidden response - check we're not redirected to the recipe page with new title
    await expect(user2Page.getByText('Forbidden')).toBeVisible();

    // Author edits the recipe - modify all fields
    await user1Page.goto(`/recipes/update?id=${recipeId}`);
    await expect(user1Page.getByRole('heading', { name: 'Edit Recipe', level: 1 })).toBeVisible();

    const updatedRecipe = {
      title: `Updated Recipe ${uniqueId}`,
      prepTime: '25',
      cookTime: '45',
      calories: '550',
      ingredients: '- 2 cups flour\n- 1 cup milk\n- 2 eggs',
      instructions: '1. Mix ingredients\n2. Bake at 375F\n3. Cool and serve',
      newTag: 'breakfast',
    };

    // Clear and fill new values
    await user1Page.locator('#title').clear();
    await user1Page.locator('#title').fill(updatedRecipe.title);
    await user1Page.locator('#preptime').clear();
    await user1Page.locator('#preptime').fill(updatedRecipe.prepTime);
    await user1Page.locator('#cooktime').clear();
    await user1Page.locator('#cooktime').fill(updatedRecipe.cookTime);
    await user1Page.locator('#calories').clear();
    await user1Page.locator('#calories').fill(updatedRecipe.calories);
    await user1Page.locator('#ingredients').clear();
    await user1Page.locator('#ingredients').fill(updatedRecipe.ingredients);
    await user1Page.locator('#instructions').clear();
    await user1Page.locator('#instructions').fill(updatedRecipe.instructions);

    // Remove existing tags and add a new one
    const existingTags = user1Page.locator('#tags-container .tag');
    const tagCount = await existingTags.count();
    for (let i = tagCount - 1; i >= 0; i--) {
      await existingTags.nth(i).locator('.tag-remove').click();
    }
    await user1Page.locator('#tags-input').fill(updatedRecipe.newTag);
    await user1Page.locator('#tags-input').press('Enter');

    // Save changes
    await user1Page.getByRole('button', { name: 'Update Recipe' }).click();
    await user1Page.waitForURL(`/recipes/${recipeId}`);

    // Verify all changes are persisted
    await expect(user1Page.getByRole('heading', { name: updatedRecipe.title, level: 1 })).toBeVisible();
    await expect(user1Page.locator('.recipe-meta').getByText(`${updatedRecipe.prepTime} min`)).toBeVisible();
    await expect(user1Page.locator('.recipe-meta').getByText(`${updatedRecipe.cookTime} min`)).toBeVisible();
    await expect(user1Page.locator('.recipe-meta .meta-item .meta-value').getByText(updatedRecipe.calories)).toBeVisible();
    await expect(user1Page.getByText('2 cups flour')).toBeVisible();
    await expect(user1Page.getByText('Mix ingredients')).toBeVisible();
    await expect(user1Page.locator('.tag').getByText(updatedRecipe.newTag)).toBeVisible();
    // Old tags should not be present
    await expect(user1Page.locator('.tag').getByText('dinner')).not.toBeVisible();
    await expect(user1Page.locator('.tag').getByText('easy')).not.toBeVisible();
  });

  test('canceling edit discards all changes including tags', async ({ user1Page }) => {
    // Create a new recipe first
    const cancelTestRecipe = {
      title: `Cancel Test Recipe ${uniqueId}`,
      prepTime: '5',
      cookTime: '10',
      calories: '200',
      ingredients: '- 1 banana\n- 1 cup yogurt',
      instructions: '1. Blend together\n2. Serve cold',
      tags: ['healthy', 'quick'],
    };

    await user1Page.goto('/recipes/create');
    await user1Page.getByRole('textbox', { name: 'Title' }).fill(cancelTestRecipe.title);
    await user1Page.locator('#preptime').fill(cancelTestRecipe.prepTime);
    await user1Page.locator('#cooktime').fill(cancelTestRecipe.cookTime);
    await user1Page.locator('#calories').fill(cancelTestRecipe.calories);
    await user1Page.locator('#ingredients').fill(cancelTestRecipe.ingredients);
    await user1Page.locator('#instructions').fill(cancelTestRecipe.instructions);

    for (const tag of cancelTestRecipe.tags) {
      await user1Page.locator('#tags-input').fill(tag);
      await user1Page.locator('#tags-input').press('Enter');
    }

    await user1Page.getByRole('button', { name: /Create Recipe|Submit/i }).click();
    await user1Page.waitForURL(/\/recipes\/\d+/);

    const cancelRecipeUrl = user1Page.url();
    const cancelRecipeId = cancelRecipeUrl.match(/\/recipes\/(\d+)/)?.[1] || '';

    // Go to edit page
    await user1Page.getByRole('link', { name: 'Edit Recipe' }).click();
    await user1Page.waitForURL(`/recipes/update?id=${cancelRecipeId}`);

    // Make changes to all fields
    await user1Page.locator('#title').clear();
    await user1Page.locator('#title').fill('This Should Not Be Saved');
    await user1Page.locator('#preptime').clear();
    await user1Page.locator('#preptime').fill('99');
    await user1Page.locator('#cooktime').clear();
    await user1Page.locator('#cooktime').fill('88');
    await user1Page.locator('#calories').clear();
    await user1Page.locator('#calories').fill('9999');
    await user1Page.locator('#ingredients').clear();
    await user1Page.locator('#ingredients').fill('- Should not be saved');
    await user1Page.locator('#instructions').clear();
    await user1Page.locator('#instructions').fill('1. This should not appear');

    // Modify tags - remove existing and add new ones
    const cancelExistingTags = user1Page.locator('#tags-container .tag');
    const cancelTagCount = await cancelExistingTags.count();
    for (let i = cancelTagCount - 1; i >= 0; i--) {
      await cancelExistingTags.nth(i).locator('.tag-remove').click();
    }
    await user1Page.locator('#tags-input').fill('discarded-tag');
    await user1Page.locator('#tags-input').press('Enter');

    // Cancel instead of saving
    await user1Page.getByRole('link', { name: 'Cancel' }).click();
    await user1Page.waitForURL(`/recipes/${cancelRecipeId}`);

    // Verify original data is still intact
    await expect(user1Page.getByRole('heading', { name: cancelTestRecipe.title, level: 1 })).toBeVisible();
    await expect(user1Page.locator('.recipe-meta').getByText(`${cancelTestRecipe.prepTime} min`)).toBeVisible();
    await expect(user1Page.locator('.recipe-meta').getByText(`${cancelTestRecipe.cookTime} min`)).toBeVisible();
    await expect(user1Page.locator('.recipe-meta .meta-item .meta-value').getByText(cancelTestRecipe.calories)).toBeVisible();
    await expect(user1Page.getByText('1 banana')).toBeVisible();
    await expect(user1Page.getByText('Blend together')).toBeVisible();
    
    // Original tags should still be present
    await expect(user1Page.locator('.tag').getByText('healthy')).toBeVisible();
    await expect(user1Page.locator('.tag').getByText('quick')).toBeVisible();
    
    // Discarded changes should NOT be present
    await expect(user1Page.getByText('This Should Not Be Saved')).not.toBeVisible();
    await expect(user1Page.getByText('99 min')).not.toBeVisible();
    await expect(user1Page.getByText('Should not be saved')).not.toBeVisible();
    await expect(user1Page.locator('.tag').getByText('discarded-tag')).not.toBeVisible();
  });
});
