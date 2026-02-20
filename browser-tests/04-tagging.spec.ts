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

test.describe('Tagging', () => {
  const uniqueId = Date.now();

  test.describe('Author Tags', () => {
    test('author can add tags on recipe creation and they persist', async ({ user1Page }) => {
      const testRecipe = {
        title: `Tag Test Recipe ${uniqueId}`,
        prepTime: '10',
        cookTime: '20',
        calories: '300',
        ingredients: '- 1 cup rice',
        instructions: '1. Cook rice',
        tags: ['vegetarian', 'easy'],
      };

      await user1Page.goto('/recipes/create');

      await user1Page.getByRole('textbox', { name: 'Title' }).fill(testRecipe.title);
      await user1Page.locator('#preptime').fill(testRecipe.prepTime);
      await user1Page.locator('#cooktime').fill(testRecipe.cookTime);
      await user1Page.locator('#calories').fill(testRecipe.calories);
      await user1Page.locator('#ingredients').fill(testRecipe.ingredients);
      await user1Page.locator('#instructions').fill(testRecipe.instructions);

      for (const tag of testRecipe.tags) {
        await user1Page.locator('#tags-input').fill(tag);
        await user1Page.locator('#tags-input').press('Enter');
      }

      await expect(user1Page.locator('#tags-container .tag')).toHaveCount(2);
      await expect(user1Page.locator('#tags-container .tag').getByText('vegetarian')).toBeVisible();
      await expect(user1Page.locator('#tags-container .tag').getByText('easy')).toBeVisible();

      await user1Page.getByRole('button', { name: /Create Recipe|Submit/i }).click();
      await user1Page.waitForURL(/\/recipes\/\d+/);

      await expect(user1Page.locator('#author-tags-container .tag').getByText('vegetarian')).toBeVisible();
      await expect(user1Page.locator('#author-tags-container .tag').getByText('easy')).toBeVisible();
    });

    test('author can add and remove tags via live editing on view page', async ({ user1Page }) => {
      const testRecipe = {
        title: `Live Tag Edit Recipe ${uniqueId}`,
        prepTime: '5',
        cookTime: '10',
        calories: '200',
        ingredients: '- 1 banana',
        instructions: '1. Eat banana',
        initialTag: 'healthy',
      };

      await user1Page.goto('/recipes/create');
      await user1Page.getByRole('textbox', { name: 'Title' }).fill(testRecipe.title);
      await user1Page.locator('#preptime').fill(testRecipe.prepTime);
      await user1Page.locator('#cooktime').fill(testRecipe.cookTime);
      await user1Page.locator('#calories').fill(testRecipe.calories);
      await user1Page.locator('#ingredients').fill(testRecipe.ingredients);
      await user1Page.locator('#instructions').fill(testRecipe.instructions);

      await user1Page.locator('#tags-input').fill(testRecipe.initialTag);
      await user1Page.locator('#tags-input').press('Enter');

      await user1Page.getByRole('button', { name: /Create Recipe|Submit/i }).click();
      await user1Page.waitForURL(/\/recipes\/\d+/);

      const recipeUrl = user1Page.url();

      await expect(user1Page.locator('#author-tags-container .tag').getByText(testRecipe.initialTag)).toBeVisible();
      await expect(user1Page.locator('#author-tags-input')).toBeVisible();

      await user1Page.locator('#author-tags-input').fill('snack');
      await user1Page.locator('#author-tags-input').press('Enter');

      await user1Page.waitForURL(recipeUrl);
      await expect(user1Page.locator('#author-tags-container .tag').getByText('snack')).toBeVisible();

      await expect(user1Page.locator('#author-tags-container .tag').getByText('healthy')).toBeVisible();
      const healthyTag = user1Page.locator('#author-tags-container .tag').filter({ hasText: 'healthy' });
      await healthyTag.locator('.tag-remove').click();

      await user1Page.waitForURL(recipeUrl);
      await expect(user1Page.locator('#author-tags-container .tag').getByText('healthy')).not.toBeVisible();
      await expect(user1Page.locator('#author-tags-container .tag').getByText('snack')).toBeVisible();
    });

    test('non-author cannot edit author tags on view page', async ({ user1Page, user2Page }) => {
      const testRecipe = {
        title: `Non-Author Tag Test ${uniqueId}`,
        prepTime: '5',
        cookTime: '10',
        calories: '150',
        ingredients: '- 1 apple',
        instructions: '1. Eat apple',
        tag: 'fruit',
      };

      await user1Page.goto('/recipes/create');
      await user1Page.getByRole('textbox', { name: 'Title' }).fill(testRecipe.title);
      await user1Page.locator('#preptime').fill(testRecipe.prepTime);
      await user1Page.locator('#cooktime').fill(testRecipe.cookTime);
      await user1Page.locator('#calories').fill(testRecipe.calories);
      await user1Page.locator('#ingredients').fill(testRecipe.ingredients);
      await user1Page.locator('#instructions').fill(testRecipe.instructions);
      await user1Page.locator('#tags-input').fill(testRecipe.tag);
      await user1Page.locator('#tags-input').press('Enter');
      await user1Page.getByRole('button', { name: /Create Recipe|Submit/i }).click();
      await user1Page.waitForURL(/\/recipes\/\d+/);

      const url = user1Page.url();
      const recipeId = url.match(/\/recipes\/(\d+)/)?.[1] || '';

      await expect(user1Page.locator('#author-tags-input')).toBeVisible();
      await expect(user1Page.locator('#author-tags-container .tag-remove')).toBeVisible();

      await user2Page.goto(`/recipes/${recipeId}`);

      await expect(user2Page.locator('#author-tags-container .tag').getByText('fruit')).toBeVisible();
      await expect(user2Page.locator('#author-tags-input')).not.toBeVisible();
      await expect(user2Page.locator('#author-tags-container .tag-remove')).not.toBeVisible();
    });
  });

  test.describe('User Tags', () => {
    test('logged-in user can add and remove personal tags on view page', async ({ user1Page }) => {
      const testRecipe = {
        title: `User Tag Test Recipe ${uniqueId}`,
        prepTime: '5',
        cookTime: '10',
        calories: '100',
        ingredients: '- 1 orange',
        instructions: '1. Peel and eat',
      };

      await user1Page.goto('/recipes/create');
      await user1Page.getByRole('textbox', { name: 'Title' }).fill(testRecipe.title);
      await user1Page.locator('#preptime').fill(testRecipe.prepTime);
      await user1Page.locator('#cooktime').fill(testRecipe.cookTime);
      await user1Page.locator('#calories').fill(testRecipe.calories);
      await user1Page.locator('#ingredients').fill(testRecipe.ingredients);
      await user1Page.locator('#instructions').fill(testRecipe.instructions);
      await user1Page.getByRole('button', { name: /Create Recipe|Submit/i }).click();
      await user1Page.waitForURL(/\/recipes\/\d+/);

      const recipeUrl = user1Page.url();

      await expect(user1Page.locator('#user-tags-input')).toBeVisible();

      await user1Page.locator('#user-tags-input').fill('favorite');
      await user1Page.locator('#user-tags-input').press('Enter');

      await user1Page.waitForURL(recipeUrl);
      await expect(user1Page.locator('#user-tags-container .tag').getByText('favorite')).toBeVisible();

      await user1Page.locator('#user-tags-input').fill('try-later');
      await user1Page.locator('#user-tags-input').press('Enter');

      await user1Page.waitForURL(recipeUrl);
      await expect(user1Page.locator('#user-tags-container .tag').getByText('try-later')).toBeVisible();

      const favoriteTag = user1Page.locator('#user-tags-container .tag').filter({ hasText: 'favorite' });
      await favoriteTag.locator('.tag-remove').click();

      await user1Page.waitForURL(recipeUrl);
      await expect(user1Page.locator('#user-tags-container .tag').getByText('favorite')).not.toBeVisible();
      await expect(user1Page.locator('#user-tags-container .tag').getByText('try-later')).toBeVisible();
    });

    test('user tags are individual per user', async ({ user1Page, user2Page }) => {
      const testRecipe = {
        title: `Individual User Tags Test ${uniqueId}`,
        prepTime: '5',
        cookTime: '10',
        calories: '120',
        ingredients: '- 1 pear',
        instructions: '1. Slice and eat',
      };

      await user1Page.goto('/recipes/create');
      await user1Page.getByRole('textbox', { name: 'Title' }).fill(testRecipe.title);
      await user1Page.locator('#preptime').fill(testRecipe.prepTime);
      await user1Page.locator('#cooktime').fill(testRecipe.cookTime);
      await user1Page.locator('#calories').fill(testRecipe.calories);
      await user1Page.locator('#ingredients').fill(testRecipe.ingredients);
      await user1Page.locator('#instructions').fill(testRecipe.instructions);
      await user1Page.getByRole('button', { name: /Create Recipe|Submit/i }).click();
      await user1Page.waitForURL(/\/recipes\/\d+/);

      const url = user1Page.url();
      const recipeId = url.match(/\/recipes\/(\d+)/)?.[1] || '';

      await user1Page.locator('#user-tags-input').fill('user1-personal');
      await user1Page.locator('#user-tags-input').press('Enter');

      await user1Page.waitForURL(url);
      await expect(user1Page.locator('#user-tags-container .tag').getByText('user1-personal')).toBeVisible();

      await user2Page.goto(`/recipes/${recipeId}`);
      await expect(user2Page.locator('#user-tags-container .tag').getByText('user1-personal')).not.toBeVisible();

      await user2Page.locator('#user-tags-input').fill('user2-personal');
      await user2Page.locator('#user-tags-input').press('Enter');

      await user2Page.waitForURL(`/recipes/${recipeId}`);
      await expect(user2Page.locator('#user-tags-container .tag').getByText('user2-personal')).toBeVisible();
      await expect(user2Page.locator('#user-tags-container .tag').getByText('user1-personal')).not.toBeVisible();

      await user1Page.reload();
      await expect(user1Page.locator('#user-tags-container .tag').getByText('user1-personal')).toBeVisible();
      await expect(user1Page.locator('#user-tags-container .tag').getByText('user2-personal')).not.toBeVisible();
    });
  });

  test.describe('Tag Editing on Create/Update Pages', () => {
    test('author can modify tags on update page and changes persist', async ({ user1Page }) => {
      const testRecipe = {
        title: `Update Tags Test ${uniqueId}`,
        prepTime: '5',
        cookTime: '10',
        calories: '100',
        ingredients: '- 1 mango',
        instructions: '1. Cut and serve',
        initialTags: ['tropical', 'sweet'],
      };

      await user1Page.goto('/recipes/create');
      await user1Page.getByRole('textbox', { name: 'Title' }).fill(testRecipe.title);
      await user1Page.locator('#preptime').fill(testRecipe.prepTime);
      await user1Page.locator('#cooktime').fill(testRecipe.cookTime);
      await user1Page.locator('#calories').fill(testRecipe.calories);
      await user1Page.locator('#ingredients').fill(testRecipe.ingredients);
      await user1Page.locator('#instructions').fill(testRecipe.instructions);

      for (const tag of testRecipe.initialTags) {
        await user1Page.locator('#tags-input').fill(tag);
        await user1Page.locator('#tags-input').press('Enter');
      }

      await user1Page.getByRole('button', { name: /Create Recipe|Submit/i }).click();
      await user1Page.waitForURL(/\/recipes\/\d+/);

      const url = user1Page.url();
      const recipeId = url.match(/\/recipes\/(\d+)/)?.[1] || '';

      await user1Page.getByRole('link', { name: 'Edit Recipe' }).click();
      await user1Page.waitForURL(`/recipes/${recipeId}/update`);

      await expect(user1Page.locator('#tags-container .tag').getByText('tropical')).toBeVisible();
      await expect(user1Page.locator('#tags-container .tag').getByText('sweet')).toBeVisible();

      const tropicalTag = user1Page.locator('#tags-container .tag').filter({ hasText: 'tropical' });
      await tropicalTag.locator('.tag-remove').click();

      await user1Page.locator('#tags-input').fill('dessert');
      await user1Page.locator('#tags-input').press('Enter');

      await expect(user1Page.locator('#tags-container .tag').getByText('tropical')).not.toBeVisible();
      await expect(user1Page.locator('#tags-container .tag').getByText('sweet')).toBeVisible();
      await expect(user1Page.locator('#tags-container .tag').getByText('dessert')).toBeVisible();

      await user1Page.getByRole('button', { name: 'Update Recipe' }).click();
      await user1Page.waitForURL(`/recipes/${recipeId}`);

      await expect(user1Page.locator('#author-tags-container .tag').getByText('tropical')).not.toBeVisible();
      await expect(user1Page.locator('#author-tags-container .tag').getByText('sweet')).toBeVisible();
      await expect(user1Page.locator('#author-tags-container .tag').getByText('dessert')).toBeVisible();
    });
  });
});
