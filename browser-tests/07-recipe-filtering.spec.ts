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
    await fillToastEditor(page, 'ingredients-editor', recipe.ingredients);
    await fillToastEditor(page, 'instructions-editor', recipe.instructions);

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

  async function createSimpleRecipe(page: Page, title: string): Promise<void> {
    await page.goto('/recipes/create');
    await page.getByRole('textbox', { name: 'Title' }).fill(title);
    await page.locator('#preptime').fill('10');
    await page.locator('#cooktime').fill('20');
    await page.locator('#calories').fill('300');
    await fillToastEditor(page, 'ingredients-editor', '- 1 ingredient');
    await fillToastEditor(page, 'instructions-editor', '1. Do something');
    await page.getByRole('button', { name: /Create Recipe|Submit/i }).click();
    await page.waitForURL(/\/recipes\/\d+/);
  }

  async function ensureMinimumRecipes(page: Page, minCount: number): Promise<void> {
    await page.goto('/recipes');
    const totalCount = parseInt(await page.locator('#total-count').textContent() || '0');
    
    if (totalCount < minCount) {
      const recipesToCreate = minCount - totalCount;
      for (let i = 0; i < recipesToCreate; i++) {
        await createSimpleRecipe(page, `Pagination Test Recipe ${uniqueId}-${i}`);
      }
    }
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

      // After removing the filter, search for our specific test recipe to verify it can now be found
      await page.locator('#search').fill(`Veggie Quick ${uniqueId}`);
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
      
      // Search for the specific test recipe since it may not be on first page
      await page.locator('#search').fill(`Veggie Quick ${uniqueId}`);
      await waitForFilterResults(page);
      
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
      
      // Search for the specific test recipe since it may not be on first page
      await user1Page.locator('#search').fill(`Meat Slow ${uniqueId}`);
      await waitForFilterResults(user1Page);
      
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

  test.describe('Recipe Count Indicator', () => {
    test('shows recipe range on initial page load', async ({ user1Page: page }) => {
      await page.goto('/recipes');

      const countElement = page.locator('#recipe-count');
      await expect(countElement).toBeVisible();
      await expect(countElement).toContainText('of');

      const countText = await countElement.textContent() || '';
      const match = countText.match(/(\d+)-(\d+) of (\d+)/);
      expect(match).not.toBeNull();
      
      const rangeStart = parseInt(match![1]);
      const rangeEnd = parseInt(match![2]);
      const totalCount = parseInt(match![3]);
      
      expect(rangeStart).toBe(1);
      expect(rangeEnd).toBeGreaterThan(0);
      expect(rangeEnd).toBeLessThanOrEqual(20);
      expect(totalCount).toBeGreaterThanOrEqual(rangeEnd);
    });

    test('count updates when filtering narrows results', async ({ user1Page: page }) => {
      await page.goto('/recipes');

      const initialTotal = parseInt(await page.locator('#total-count').textContent() || '0');

      await page.locator('#filter-tags-input').fill('vegetarian');
      await page.locator('#filter-tags-input').press('Enter');
      await waitForFilterResults(page);

      const countText = await page.locator('#recipe-count').textContent() || '';
      const match = countText.match(/(\d+)-(\d+) of (\d+)/);
      expect(match).not.toBeNull();
      
      const filteredRangeEnd = parseInt(match![2]);
      const filteredTotal = parseInt(match![3]);

      expect(filteredTotal).toBeLessThan(initialTotal);
      expect(filteredRangeEnd).toBeLessThanOrEqual(filteredTotal);
    });

    test('count shows 0-0 when no recipes match filter', async ({ user1Page: page }) => {
      await page.goto('/recipes');

      await page.locator('#search').fill('xyznonexistent12345');
      await waitForFilterResults(page);

      const countElement = page.locator('#recipe-count');
      await expect(countElement).toContainText('0-0 of 0');
    });

    test('count updates when clearing filters', async ({ user1Page: page }) => {
      // Ensure we have more recipes than just our test recipes so clearing shows more
      await ensureMinimumRecipes(page, 10);
      
      await page.goto('/recipes');

      // Apply a filter to narrow results
      await page.locator('#search').fill(uniqueId.toString());
      await waitForFilterResults(page);

      const filteredTotal = parseInt(await page.locator('#total-count').textContent() || '0');
      // Test recipes contain uniqueId, so we should have a small number
      expect(filteredTotal).toBeGreaterThan(0);
      expect(filteredTotal).toBeLessThanOrEqual(20);

      // Clear filters
      await page.getByRole('button', { name: 'Clear' }).click();
      await waitForFilterResults(page);

      // After clearing, total should be higher than filtered (all recipes)
      const clearedTotal = parseInt(await page.locator('#total-count').textContent() || '0');
      expect(clearedTotal).toBeGreaterThan(filteredTotal);
    });
  });

  test.describe('Pagination and Page Markers', () => {
    test('page marker appears when loading more recipes', async ({ user1Page: page }) => {
      // Ensure we have enough recipes for pagination (need > 20)
      await ensureMinimumRecipes(page, 25);
      
      await page.goto('/recipes');

      // Initially there should be no page markers
      await expect(page.locator('.page-marker')).toHaveCount(0);

      // Click Load More
      await page.getByRole('button', { name: 'Load More' }).click();
      await page.waitForResponse(response => 
        response.url().includes('/recipes/filter') && response.status() === 200
      );

      // Page marker should appear with range format (e.g., "21-40 of 1214")
      const pageMarker = page.locator('.page-marker').first();
      await expect(pageMarker).toBeVisible();
      const markerText = await pageMarker.textContent() || '';
      expect(markerText).toMatch(/\d+-\d+ of \d+/);
    });

    test('top count stays static after loading more', async ({ user1Page: page }) => {
      // Ensure we have enough recipes for pagination (need > 20)
      await ensureMinimumRecipes(page, 25);
      
      await page.goto('/recipes');

      const initialCountText = await page.locator('#recipe-count').textContent() || '';
      const initialMatch = initialCountText.match(/(\d+)-(\d+) of (\d+)/);
      expect(initialMatch).not.toBeNull();
      const initialRangeStart = parseInt(initialMatch![1]);
      const initialRangeEnd = parseInt(initialMatch![2]);
      const initialTotal = parseInt(initialMatch![3]);
      expect(initialRangeEnd).toBeLessThanOrEqual(20);

      // Click Load More
      await page.getByRole('button', { name: 'Load More' }).click();
      await page.waitForResponse(response => 
        response.url().includes('/recipes/filter') && response.status() === 200
      );
      await page.waitForTimeout(100);

      // Top count indicator should remain unchanged (static 1-20 of N)
      const afterCountText = await page.locator('#recipe-count').textContent() || '';
      const afterMatch = afterCountText.match(/(\d+)-(\d+) of (\d+)/);
      expect(afterMatch).not.toBeNull();
      expect(parseInt(afterMatch![1])).toBe(initialRangeStart);
      expect(parseInt(afterMatch![2])).toBe(initialRangeEnd);
      expect(parseInt(afterMatch![3])).toBe(initialTotal);
    });

    test('load more button disappears when all recipes are loaded', async ({ user1Page: page }) => {
      await page.goto('/recipes');

      const totalCount = parseInt(await page.locator('#total-count').textContent() || '0');
      
      // Skip test if we need many pages (would take too long)
      if (totalCount > 60 || totalCount <= 20) {
        test.skip();
        return;
      }

      // Keep clicking Load More until it disappears
      while (await page.getByRole('button', { name: 'Load More' }).isVisible()) {
        await page.getByRole('button', { name: 'Load More' }).click();
        await page.waitForResponse(response => 
          response.url().includes('/recipes/filter') && response.status() === 200
        );
        await page.waitForTimeout(100);
      }

      // Count total recipe cards to verify all loaded
      const recipeCards = await page.locator('.recipe-card').count();
      expect(recipeCards).toBe(totalCount);

      // Load more button should be gone
      await expect(page.getByRole('button', { name: 'Load More' })).not.toBeVisible();
    });
  });

  test.describe('Authored By Me Filter', () => {
    const authorFilterId = Date.now();

    test.beforeAll(async ({ browser }) => {
      // User1 creates a recipe
      const context1 = await browser.newContext();
      const page1 = await context1.newPage();
      await page1.goto('/login');
      await page1.locator('input[name="email"]').fill(TEST_USERS.approved1.email);
      await page1.locator('input[name="password"]').fill(TEST_USERS.approved1.password);
      await page1.getByRole('button', { name: 'Sign In' }).click();
      await page1.waitForURL('/');

      await page1.goto('/recipes/create');
      await page1.getByRole('textbox', { name: 'Title' }).fill(`User1 Recipe ${authorFilterId}`);
      await page1.locator('#preptime').fill('10');
      await page1.locator('#cooktime').fill('20');
      await page1.locator('#calories').fill('300');
      await fillToastEditor(page1, 'ingredients-editor', '- test ingredient');
      await fillToastEditor(page1, 'instructions-editor', '1. test instruction');
      await page1.getByRole('button', { name: /Create Recipe|Submit/i }).click();
      await page1.waitForURL(/\/recipes\/\d+/);
      await context1.close();

      // User2 creates a recipe
      const context2 = await browser.newContext();
      const page2 = await context2.newPage();
      await page2.goto('/login');
      await page2.locator('input[name="email"]').fill(TEST_USERS.approved2.email);
      await page2.locator('input[name="password"]').fill(TEST_USERS.approved2.password);
      await page2.getByRole('button', { name: 'Sign In' }).click();
      await page2.waitForURL('/');

      await page2.goto('/recipes/create');
      await page2.getByRole('textbox', { name: 'Title' }).fill(`User2 Recipe ${authorFilterId}`);
      await page2.locator('#preptime').fill('15');
      await page2.locator('#cooktime').fill('25');
      await page2.locator('#calories').fill('400');
      await fillToastEditor(page2, 'ingredients-editor', '- another ingredient');
      await fillToastEditor(page2, 'instructions-editor', '1. another instruction');
      await page2.getByRole('button', { name: /Create Recipe|Submit/i }).click();
      await page2.waitForURL(/\/recipes\/\d+/);
      await context2.close();
    });

    test('filters to show only recipes authored by current user', async ({ user1Page: page }) => {
      await page.goto('/recipes');

      // Both recipes should be visible initially when searching by the unique ID
      await page.locator('#search').fill(authorFilterId.toString());
      await waitForFilterResults(page);

      let titles = await getVisibleRecipeTitles(page);
      expect(titles.some(t => t.includes(`User1 Recipe ${authorFilterId}`))).toBe(true);
      expect(titles.some(t => t.includes(`User2 Recipe ${authorFilterId}`))).toBe(true);

      // Check "My recipes" filter
      await page.locator('#authored_by_me').check();
      await waitForFilterResults(page);

      titles = await getVisibleRecipeTitles(page);
      expect(titles.some(t => t.includes(`User1 Recipe ${authorFilterId}`))).toBe(true);
      expect(titles.some(t => t.includes(`User2 Recipe ${authorFilterId}`))).toBe(false);
    });

    test('unchecking filter shows all recipes again', async ({ user1Page: page }) => {
      await page.goto('/recipes');

      await page.locator('#search').fill(authorFilterId.toString());
      await waitForFilterResults(page);

      await page.locator('#authored_by_me').check();
      await waitForFilterResults(page);

      let titles = await getVisibleRecipeTitles(page);
      expect(titles.some(t => t.includes(`User2 Recipe ${authorFilterId}`))).toBe(false);

      await page.locator('#authored_by_me').uncheck();
      await waitForFilterResults(page);

      titles = await getVisibleRecipeTitles(page);
      expect(titles.some(t => t.includes(`User1 Recipe ${authorFilterId}`))).toBe(true);
      expect(titles.some(t => t.includes(`User2 Recipe ${authorFilterId}`))).toBe(true);
    });

    test('clear button resets authored by me checkbox', async ({ user1Page: page }) => {
      await page.goto('/recipes');

      await page.locator('#authored_by_me').check();
      await waitForFilterResults(page);

      await expect(page.locator('#authored_by_me')).toBeChecked();

      await page.getByRole('button', { name: 'Clear' }).click();
      await waitForFilterResults(page);

      await expect(page.locator('#authored_by_me')).not.toBeChecked();
    });

    test('authored by me filter only shows for logged-in users', async ({ browser }) => {
      const context = await browser.newContext();
      const page = await context.newPage();

      await page.goto('/recipes');

      await expect(page.locator('#authored_by_me')).not.toBeVisible();

      await context.close();
    });

    test('user2 sees only their recipes when filtering', async ({ user2Page: page }) => {
      await page.goto('/recipes');

      await page.locator('#search').fill(authorFilterId.toString());
      await waitForFilterResults(page);

      await page.locator('#authored_by_me').check();
      await waitForFilterResults(page);

      const titles = await getVisibleRecipeTitles(page);
      expect(titles.some(t => t.includes(`User2 Recipe ${authorFilterId}`))).toBe(true);
      expect(titles.some(t => t.includes(`User1 Recipe ${authorFilterId}`))).toBe(false);
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
