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

test.describe('Recipe Filtering', () => {
  const uniqueId = Date.now();

  const testRecipes = {
    vegetarianQuick: {
      title: `Filter Test Veggie Quick ${uniqueId}`,
      prepTime: '5',
      cookTime: '10',
      calories: '200',
      ingredients: '- 1 cup vegetables',
      instructions: '1. Cook vegetables',
      tags: ['vegetarian', 'quick'],
    },
    meatSlow: {
      title: `Filter Test Meat Slow ${uniqueId}`,
      prepTime: '30',
      cookTime: '120',
      calories: '800',
      ingredients: '- 1 lb beef',
      instructions: '1. Slow cook beef',
      tags: ['meat', 'slow-cook'],
    },
    dessertLowCal: {
      title: `Filter Test Dessert LowCal ${uniqueId}`,
      prepTime: '15',
      cookTime: '0',
      calories: '150',
      ingredients: '- 1 cup fruit',
      instructions: '1. Prepare fruit dessert',
      tags: ['dessert', 'healthy'],
    },
  };

  async function createRecipe(page: Page, recipe: typeof testRecipes.vegetarianQuick): Promise<void> {
    await page.goto('/recipes/create');
    await page.getByRole('textbox', { name: 'Title' }).fill(recipe.title);
    await page.locator('#preptime').fill(recipe.prepTime);
    await page.locator('#cooktime').fill(recipe.cookTime);
    await page.locator('#calories').fill(recipe.calories);
    await page.locator('#ingredients').fill(recipe.ingredients);
    await page.locator('#instructions').fill(recipe.instructions);

    for (const tag of recipe.tags) {
      await page.locator('#tags-input').fill(tag);
      await page.locator('#tags-input').press('Enter');
    }

    await page.getByRole('button', { name: /Create Recipe|Submit/i }).click();
    await page.waitForURL(/\/recipes\/\d+/);
  }

  async function getVisibleRecipeTitles(page: Page): Promise<string[]> {
    const titles = await page.locator('.recipe-card .recipe-title').allTextContents();
    return titles;
  }

  async function waitForFilterResults(page: Page): Promise<void> {
    await page.waitForResponse(response => 
      response.url().includes('/recipes/filter') && response.status() === 200
    );
    await page.waitForTimeout(100);
  }

  test.beforeAll(async ({ browser }) => {
    const context = await browser.newContext();
    const page = await context.newPage();
    const user = TEST_USERS.approved1;
    
    await page.goto('/login');
    await page.locator('input[name="email"]').fill(user.email);
    await page.locator('input[name="password"]').fill(user.password);
    await page.getByRole('button', { name: 'Sign In' }).click();
    await page.waitForURL('/');

    for (const recipe of Object.values(testRecipes)) {
      await createRecipe(page, recipe);
    }

    await context.close();
  });

  test.describe('Text Search Filter', () => {
    test('filters recipes by title search', async ({ user1Page: page }) => {
      await page.goto('/recipes');

      await page.locator('#search').fill(`Veggie Quick ${uniqueId}`);
      await waitForFilterResults(page);

      const titles = await getVisibleRecipeTitles(page);
      expect(titles.some(t => t.includes('Veggie Quick'))).toBe(true);
      expect(titles.some(t => t.includes('Meat Slow'))).toBe(false);
    });

    test('filters recipes by ingredient search', async ({ user1Page: page }) => {
      await page.goto('/recipes');

      await page.locator('#search').fill('beef');
      await waitForFilterResults(page);

      const titles = await getVisibleRecipeTitles(page);
      expect(titles.some(t => t.includes(`Meat Slow ${uniqueId}`))).toBe(true);
    });
  });

  test.describe('Author Tag Filter', () => {
    test('filters recipes by single author tag', async ({ user1Page: page }) => {
      await page.goto('/recipes');

      await page.locator('#filter-tags-input').fill('vegetarian');
      await page.locator('#filter-tags-input').press('Enter');
      await waitForFilterResults(page);

      const titles = await getVisibleRecipeTitles(page);
      expect(titles.some(t => t.includes(`Veggie Quick ${uniqueId}`))).toBe(true);
      expect(titles.some(t => t.includes(`Meat Slow ${uniqueId}`))).toBe(false);
      expect(titles.some(t => t.includes(`Dessert LowCal ${uniqueId}`))).toBe(false);
    });

    test('filters recipes by multiple author tags (AND logic)', async ({ user1Page: page }) => {
      await page.goto('/recipes');

      await page.locator('#filter-tags-input').fill('vegetarian');
      await page.locator('#filter-tags-input').press('Enter');
      await waitForFilterResults(page);

      await page.locator('#filter-tags-input').fill('quick');
      await page.locator('#filter-tags-input').press('Enter');
      await waitForFilterResults(page);

      const titles = await getVisibleRecipeTitles(page);
      expect(titles.some(t => t.includes(`Veggie Quick ${uniqueId}`))).toBe(true);
    });

    test('removing a tag filter updates results', async ({ user1Page: page }) => {
      await page.goto('/recipes');

      await page.locator('#filter-tags-input').fill('dessert');
      await page.locator('#filter-tags-input').press('Enter');
      await waitForFilterResults(page);

      let titles = await getVisibleRecipeTitles(page);
      expect(titles.some(t => t.includes(`Dessert LowCal ${uniqueId}`))).toBe(true);
      expect(titles.some(t => t.includes(`Veggie Quick ${uniqueId}`))).toBe(false);

      const dessertTag = page.locator('#filter-tags-container .tag').filter({ hasText: 'dessert' });
      await dessertTag.locator('.tag-remove').click();
      await waitForFilterResults(page);

      titles = await getVisibleRecipeTitles(page);
      expect(titles.some(t => t.includes(`Veggie Quick ${uniqueId}`))).toBe(true);
    });
  });

  test.describe('Numeric Filters', () => {
    test('filters by calories less than value', async ({ user1Page: page }) => {
      await page.goto('/recipes');

      // Fill value first (triggers input with delay), then select operator (triggers change immediately)
      // This ensures the filter request includes both values
      await page.locator('#calories_value').fill('300');
      await page.locator('#calories_op').selectOption('lt');
      await waitForFilterResults(page);

      const titles = await getVisibleRecipeTitles(page);
      expect(titles.some(t => t.includes(`Veggie Quick ${uniqueId}`))).toBe(true);
      expect(titles.some(t => t.includes(`Dessert LowCal ${uniqueId}`))).toBe(true);
      expect(titles.some(t => t.includes(`Meat Slow ${uniqueId}`))).toBe(false);
    });

    test('filters by prep time greater than value', async ({ user1Page: page }) => {
      await page.goto('/recipes');

      await page.locator('#prep_time_value').fill('15');
      await page.locator('#prep_time_op').selectOption('gte');
      await waitForFilterResults(page);

      const titles = await getVisibleRecipeTitles(page);
      expect(titles.some(t => t.includes(`Meat Slow ${uniqueId}`))).toBe(true);
      expect(titles.some(t => t.includes(`Dessert LowCal ${uniqueId}`))).toBe(true);
      expect(titles.some(t => t.includes(`Veggie Quick ${uniqueId}`))).toBe(false);
    });

    test('filters by cook time equals value', async ({ user1Page: page }) => {
      await page.goto('/recipes');

      await page.locator('#cook_time_value').fill('120');
      await page.locator('#cook_time_op').selectOption('eq');
      await waitForFilterResults(page);

      const titles = await getVisibleRecipeTitles(page);
      expect(titles.some(t => t.includes(`Meat Slow ${uniqueId}`))).toBe(true);
      expect(titles.some(t => t.includes(`Veggie Quick ${uniqueId}`))).toBe(false);
    });
  });

  test.describe('User Tag Filter', () => {
    test('user can filter by personal tags', async ({ user1Page: page }) => {
      await page.goto('/recipes');
      
      const veggieCard = page.locator('.recipe-card').filter({ hasText: `Veggie Quick ${uniqueId}` }).first();
      await veggieCard.click();
      await page.waitForURL(/\/recipes\/\d+/);

      await page.locator('#user-tags-input').fill('my-favorite');
      await page.locator('#user-tags-input').press('Enter');
      // Wait for the tag to be added before navigating away
      await page.locator('#user-tags-container .tag').filter({ hasText: 'my-favorite' }).waitFor();

      await page.goto('/recipes');
      await page.waitForLoadState('networkidle');

      await page.locator('#filter-user-tags-input').fill('my-favorite');
      await page.locator('#filter-user-tags-input').press('Enter');
      await waitForFilterResults(page);

      const titles = await getVisibleRecipeTitles(page);
      expect(titles.some(t => t.includes(`Veggie Quick ${uniqueId}`))).toBe(true);
      expect(titles.some(t => t.includes(`Meat Slow ${uniqueId}`))).toBe(false);
    });

    test('user tag filter only shows for logged-in users', async ({ browser }) => {
      const context = await browser.newContext();
      const page = await context.newPage();

      await page.goto('/recipes');

      await expect(page.locator('#filter-user-tags-input')).not.toBeVisible();
      await expect(page.locator('#filter-tags-input')).toBeVisible();

      await context.close();
    });

    test('user tags are personal and do not affect other users filter', async ({ user1Page, user2Page }) => {
      await user1Page.goto('/recipes');
      const meatCard = user1Page.locator('.recipe-card').filter({ hasText: `Meat Slow ${uniqueId}` }).first();
      await meatCard.click();
      await user1Page.waitForURL(/\/recipes\/\d+/);

      await user1Page.locator('#user-tags-input').fill('user1-filter-test');
      await user1Page.locator('#user-tags-input').press('Enter');
      // Wait for the tag to be added before continuing
      await user1Page.locator('#user-tags-container .tag').filter({ hasText: 'user1-filter-test' }).waitFor();

      await user2Page.goto('/recipes');
      await user2Page.waitForLoadState('networkidle');

      await user2Page.locator('#filter-user-tags-input').fill('user1-filter-test');
      await user2Page.locator('#filter-user-tags-input').press('Enter');
      await waitForFilterResults(user2Page);

      const titles = await getVisibleRecipeTitles(user2Page);
      expect(titles.some(t => t.includes(`Meat Slow ${uniqueId}`))).toBe(false);
    });
  });

  test.describe('Combined Filters', () => {
    test('combines text search with tag filter', async ({ user1Page: page }) => {
      await page.goto('/recipes');

      await page.locator('#search').fill(`${uniqueId}`);
      await waitForFilterResults(page);

      let titles = await getVisibleRecipeTitles(page);
      expect(titles.length).toBeGreaterThanOrEqual(3);

      await page.locator('#filter-tags-input').fill('healthy');
      await page.locator('#filter-tags-input').press('Enter');
      await waitForFilterResults(page);

      titles = await getVisibleRecipeTitles(page);
      expect(titles.some(t => t.includes(`Dessert LowCal ${uniqueId}`))).toBe(true);
      expect(titles.some(t => t.includes(`Veggie Quick ${uniqueId}`))).toBe(false);
    });

    test('combines tag filter with calorie filter', async ({ user1Page: page }) => {
      await page.goto('/recipes');

      await page.locator('#filter-tags-input').fill('vegetarian');
      await page.locator('#filter-tags-input').press('Enter');
      await waitForFilterResults(page);

      await page.locator('#calories_op').selectOption('lte');
      await page.locator('#calories_value').fill('200');
      await waitForFilterResults(page);

      const titles = await getVisibleRecipeTitles(page);
      expect(titles.some(t => t.includes(`Veggie Quick ${uniqueId}`))).toBe(true);
    });
  });

  test.describe('Clear Filter', () => {
    test('clear button resets all filters', async ({ user1Page: page }) => {
      await page.goto('/recipes');

      await page.locator('#search').fill('test');
      await page.locator('#filter-tags-input').fill('vegetarian');
      await page.locator('#filter-tags-input').press('Enter');
      await page.locator('#calories_op').selectOption('lt');
      await page.locator('#calories_value').fill('500');
      await waitForFilterResults(page);

      await page.getByRole('button', { name: 'Clear' }).click();
      await waitForFilterResults(page);

      await expect(page.locator('#search')).toHaveValue('');
      await expect(page.locator('#calories_op')).toHaveValue('');
      await expect(page.locator('#calories_value')).toHaveValue('');
      await expect(page.locator('#filter-tags-container .tag')).toHaveCount(0);
    });

    test('clearing filter resets to paginated results (max 20 recipes)', async ({ user1Page: page }) => {
      await page.goto('/recipes');

      const initialCount = await page.locator('.recipe-card').count();

      await page.locator('#search').fill(uniqueId.toString());
      await waitForFilterResults(page);

      const filteredCount = await page.locator('.recipe-card').count();
      expect(filteredCount).toBeLessThan(initialCount);

      await page.getByRole('button', { name: 'Clear' }).click();
      await waitForFilterResults(page);

      const clearedCount = await page.locator('.recipe-card').count();
      expect(clearedCount).toBeLessThanOrEqual(20);
    });
  });
});
