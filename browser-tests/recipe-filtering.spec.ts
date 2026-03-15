import { test as base, expect, Page } from '@playwright/test';
import { TEST_USERS } from './test-users';
import { fillToastEditor } from './editor-helpers';

type AuthFixtures = {
  user1Page: Page;
  user2Page: Page;
};

async function clickTagAddButton(page: Page, tagInputId: string): Promise<void> {
  const addBtn = page.locator(`#${tagInputId}-add-btn`);
  await addBtn.click();
  await page.locator(`#${tagInputId}-input`).waitFor({ state: 'visible' });
}

async function expandFiltersOnMobile(page: Page): Promise<void> {
  const filterToggle = page.locator('.filter-toggle');
  if (await filterToggle.isVisible()) {
    const filtersContainer = page.locator('.filters-collapsible');
    const isExpanded = await filtersContainer.evaluate(el => el.classList.contains('expanded'));
    if (!isExpanded) {
      await filterToggle.click();
      await page.locator('#filter-form').waitFor({ state: 'visible' });
    }
  }
}

const test = base.extend<AuthFixtures>({
  user1Page: async ({ browser }, use) => {
    const context = await browser.newContext();
    const page = await context.newPage();
    const user = TEST_USERS.approved1;
    await page.goto('/login');
    await page.locator('input[name="email"]').fill(user.email);
    await page.locator('input[name="password"]').fill(user.password);
    await Promise.all([
      page.waitForURL('/'),
      page.getByRole('button', { name: 'Sign In' }).click(),
    ]);
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
    await Promise.all([
      page.waitForURL('/'),
      page.getByRole('button', { name: 'Sign In' }).click(),
    ]);
    await use(page);
    await context.close();
  },
});

test.describe.serial('Recipe Filtering', () => {
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

    await Promise.all([
      page.waitForURL(/\/recipes\/\d+/),
      page.getByRole('button', { name: /Create Recipe|Submit/i }).click(),
    ]);
  }

  async function getVisibleRecipeTitles(page: Page): Promise<string[]> {
    const titles = await page.locator('.recipe-card .recipe-title').allTextContents();
    return titles;
  }

  async function triggerFilterAndWait(page: Page, action: () => Promise<void>): Promise<void> {
    const responsePromise = page.waitForResponse(response => 
      response.url().includes('/recipes/filter') && response.status() === 200
    );
    await action();
    await responsePromise;
    await page.waitForFunction(() => {
      return document.querySelectorAll('.htmx-request').length === 0 &&
             document.querySelectorAll('.htmx-settling').length === 0;
    }, { timeout: 5000 });
  }

  async function createSimpleRecipe(page: Page, title: string): Promise<void> {
    await page.goto('/recipes/create');
    await page.getByRole('textbox', { name: 'Title' }).fill(title);
    await page.locator('#preptime').fill('10');
    await page.locator('#cooktime').fill('20');
    await page.locator('#calories').fill('300');
    await fillToastEditor(page, 'ingredients-editor', '- 1 ingredient');
    await fillToastEditor(page, 'instructions-editor', '1. Do something');
    await Promise.all([
      page.waitForURL(/\/recipes\/\d+/),
      page.getByRole('button', { name: /Create Recipe|Submit/i }).click(),
    ]);
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
    await Promise.all([
      page.waitForURL('/'),
      page.getByRole('button', { name: 'Sign In' }).click(),
    ]);

    for (const recipe of Object.values(testRecipes)) {
      await createRecipe(page, recipe);
    }

    await context.close();
  });

  test.describe('Text Search Filter', () => {
    test('filters recipes by title search', async ({ user1Page: page }) => {
      await page.goto('/recipes');
      await expandFiltersOnMobile(page);

      await triggerFilterAndWait(page, () => page.locator('#search').fill(`Veggie Quick ${uniqueId}`));

      const titles = await getVisibleRecipeTitles(page);
      expect(titles.some(t => t.includes('Veggie Quick'))).toBe(true);
      expect(titles.some(t => t.includes('Meat Slow'))).toBe(false);
    });

    test('filters recipes by ingredient search', async ({ user1Page: page }) => {
      await page.goto('/recipes');
      await expandFiltersOnMobile(page);

      await triggerFilterAndWait(page, () => page.locator('#search').fill('beef'));

      const titles = await getVisibleRecipeTitles(page);
      expect(titles.some(t => t.includes(`Meat Slow ${uniqueId}`))).toBe(true);
    });
  });

  test.describe('Author Tag Filter', () => {
    test('filters recipes by single author tag', async ({ user1Page: page }) => {
      await page.goto('/recipes');
      await expandFiltersOnMobile(page);

      await page.locator('#filter-tags-input').fill('vegetarian');
      await triggerFilterAndWait(page, () => page.locator('#filter-tags-input').press('Enter'));

      const titles = await getVisibleRecipeTitles(page);
      expect(titles.some(t => t.includes(`Veggie Quick ${uniqueId}`))).toBe(true);
      expect(titles.some(t => t.includes(`Meat Slow ${uniqueId}`))).toBe(false);
      expect(titles.some(t => t.includes(`Dessert LowCal ${uniqueId}`))).toBe(false);
    });

    test('filters recipes by multiple author tags (AND logic)', async ({ user1Page: page }) => {
      await page.goto('/recipes');
      await expandFiltersOnMobile(page);

      await page.locator('#filter-tags-input').fill('vegetarian');
      await triggerFilterAndWait(page, () => page.locator('#filter-tags-input').press('Enter'));

      await page.locator('#filter-tags-input').fill('quick');
      await triggerFilterAndWait(page, () => page.locator('#filter-tags-input').press('Enter'));

      const titles = await getVisibleRecipeTitles(page);
      expect(titles.some(t => t.includes(`Veggie Quick ${uniqueId}`))).toBe(true);
    });

    test('removing a tag filter updates results', async ({ user1Page: page }) => {
      await page.goto('/recipes');
      await expandFiltersOnMobile(page);

      await page.locator('#filter-tags-input').fill('dessert');
      await triggerFilterAndWait(page, () => page.locator('#filter-tags-input').press('Enter'));

      let titles = await getVisibleRecipeTitles(page);
      expect(titles.some(t => t.includes(`Dessert LowCal ${uniqueId}`))).toBe(true);
      expect(titles.some(t => t.includes(`Veggie Quick ${uniqueId}`))).toBe(false);

      const dessertTag = page.locator('#filter-tags-container .tag').filter({ hasText: 'dessert' });
      await triggerFilterAndWait(page, () => dessertTag.locator('.tag-remove').click());

      await triggerFilterAndWait(page, () => page.locator('#search').fill(`Veggie Quick ${uniqueId}`));

      titles = await getVisibleRecipeTitles(page);
      expect(titles.some(t => t.includes(`Veggie Quick ${uniqueId}`))).toBe(true);
    });
  });

  test.describe('Numeric Filters', () => {
    test('filters by calories less than value', async ({ user1Page: page }) => {
      await page.goto('/recipes');
      await expandFiltersOnMobile(page);

      await page.locator('#calories_value').fill('300');
      await triggerFilterAndWait(page, () => page.locator('#calories_op').selectOption('lt'));

      const titles = await getVisibleRecipeTitles(page);
      expect(titles.some(t => t.includes(`Veggie Quick ${uniqueId}`))).toBe(true);
      expect(titles.some(t => t.includes(`Dessert LowCal ${uniqueId}`))).toBe(true);
      expect(titles.some(t => t.includes(`Meat Slow ${uniqueId}`))).toBe(false);
    });

    test('filters by prep time greater than value', async ({ user1Page: page }) => {
      await page.goto('/recipes');
      await expandFiltersOnMobile(page);

      await page.locator('#prep_time_value').fill('15');
      await triggerFilterAndWait(page, () => page.locator('#prep_time_op').selectOption('gte'));

      const titles = await getVisibleRecipeTitles(page);
      expect(titles.some(t => t.includes(`Meat Slow ${uniqueId}`))).toBe(true);
      expect(titles.some(t => t.includes(`Dessert LowCal ${uniqueId}`))).toBe(true);
      expect(titles.some(t => t.includes(`Veggie Quick ${uniqueId}`))).toBe(false);
    });

    test('filters by cook time equals value', async ({ user1Page: page }) => {
      await page.goto('/recipes');
      await expandFiltersOnMobile(page);

      await page.locator('#cook_time_value').fill('120');
      await triggerFilterAndWait(page, () => page.locator('#cook_time_op').selectOption('eq'));

      const titles = await getVisibleRecipeTitles(page);
      expect(titles.some(t => t.includes(`Meat Slow ${uniqueId}`))).toBe(true);
      expect(titles.some(t => t.includes(`Veggie Quick ${uniqueId}`))).toBe(false);
    });
  });

  test.describe('User Tag Filter', () => {
    test('user can filter by personal tags', async ({ user1Page: page }) => {
      await page.goto('/recipes');
      await expandFiltersOnMobile(page);
      
      await triggerFilterAndWait(page, () => page.locator('#search').fill(`Veggie Quick ${uniqueId}`));
      
      const veggieCard = page.locator('.recipe-card').filter({ hasText: `Veggie Quick ${uniqueId}` }).first();
      await Promise.all([
        page.waitForURL(/\/recipes\/\d+/),
        veggieCard.click(),
      ]);

      await clickTagAddButton(page, 'user-tags');
      await page.locator('#user-tags-input').fill('my-favorite');
      await page.locator('#user-tags-input').press('Enter');
      await page.locator('#user-tags-container .tag').filter({ hasText: 'my-favorite' }).waitFor();

      await page.goto('/recipes');
      await page.waitForLoadState('networkidle');
      await expandFiltersOnMobile(page);

      await page.locator('#filter-user-tags-input').fill('my-favorite');
      await triggerFilterAndWait(page, () => page.locator('#filter-user-tags-input').press('Enter'));

      const titles = await getVisibleRecipeTitles(page);
      expect(titles.some(t => t.includes(`Veggie Quick ${uniqueId}`))).toBe(true);
      expect(titles.some(t => t.includes(`Meat Slow ${uniqueId}`))).toBe(false);
    });

    test('user tag filter only shows for logged-in users', async ({ browser }) => {
      const context = await browser.newContext();
      const page = await context.newPage();

      await page.goto('/recipes');
      await expandFiltersOnMobile(page);

      await expect(page.locator('#filter-user-tags-input')).not.toBeVisible();
      await expect(page.locator('#filter-tags-input')).toBeVisible();

      await context.close();
    });

    test('user tags are personal and do not affect other users filter', async ({ user1Page, user2Page }) => {
      await user1Page.goto('/recipes');
      await expandFiltersOnMobile(user1Page);
      
      await triggerFilterAndWait(user1Page, () => user1Page.locator('#search').fill(`Meat Slow ${uniqueId}`));
      
      const meatCard = user1Page.locator('.recipe-card').filter({ hasText: `Meat Slow ${uniqueId}` }).first();
      await Promise.all([
        user1Page.waitForURL(/\/recipes\/\d+/),
        meatCard.click(),
      ]);

      await clickTagAddButton(user1Page, 'user-tags');
      await user1Page.locator('#user-tags-input').fill('user1-filter-test');
      await user1Page.locator('#user-tags-input').press('Enter');
      await user1Page.locator('#user-tags-container .tag').filter({ hasText: 'user1-filter-test' }).waitFor();

      await user2Page.goto('/recipes');
      await user2Page.waitForLoadState('networkidle');
      await expandFiltersOnMobile(user2Page);

      await user2Page.locator('#filter-user-tags-input').fill('user1-filter-test');
      await triggerFilterAndWait(user2Page, () => user2Page.locator('#filter-user-tags-input').press('Enter'));

      const titles = await getVisibleRecipeTitles(user2Page);
      expect(titles.some(t => t.includes(`Meat Slow ${uniqueId}`))).toBe(false);
    });
  });

  test.describe('Combined Filters', () => {
    test('combines text search with tag filter', async ({ user1Page: page }) => {
      await page.goto('/recipes');
      await expandFiltersOnMobile(page);

      await triggerFilterAndWait(page, () => page.locator('#search').fill(`${uniqueId}`));

      let titles = await getVisibleRecipeTitles(page);
      expect(titles.length).toBeGreaterThanOrEqual(3);

      await page.locator('#filter-tags-input').fill('healthy');
      await triggerFilterAndWait(page, () => page.locator('#filter-tags-input').press('Enter'));

      titles = await getVisibleRecipeTitles(page);
      expect(titles.some(t => t.includes(`Dessert LowCal ${uniqueId}`))).toBe(true);
      expect(titles.some(t => t.includes(`Veggie Quick ${uniqueId}`))).toBe(false);
    });

    test('combines tag filter with calorie filter', async ({ user1Page: page }) => {
      await page.goto('/recipes');
      await expandFiltersOnMobile(page);

      await page.locator('#filter-tags-input').fill('vegetarian');
      await triggerFilterAndWait(page, () => page.locator('#filter-tags-input').press('Enter'));

      await page.locator('#calories_value').fill('200');
      await triggerFilterAndWait(page, () => page.locator('#calories_op').selectOption('lte'));

      const titles = await getVisibleRecipeTitles(page);
      expect(titles.some(t => t.includes(`Veggie Quick ${uniqueId}`))).toBe(true);
    });
  });

  test.describe('Recipe Count Indicator', () => {
    test('count updates when filtering narrows results', async ({ user1Page: page }) => {
      await page.goto('/recipes');
      await expandFiltersOnMobile(page);

      const initialTotal = parseInt(await page.locator('#total-count').textContent() || '0');

      await page.locator('#filter-tags-input').fill('vegetarian');
      await triggerFilterAndWait(page, () => page.locator('#filter-tags-input').press('Enter'));

      const filteredTotal = parseInt(await page.locator('#total-count').textContent() || '0');

      expect(filteredTotal).toBeLessThan(initialTotal);
    });

    test('count shows 0-0 when no recipes match filter', async ({ user1Page: page }) => {
      await page.goto('/recipes');
      await expandFiltersOnMobile(page);

      await triggerFilterAndWait(page, () => page.locator('#search').fill('xyznonexistent12345'));

      const countElement = page.locator('#total-count');
      await expect(countElement).toContainText('0');
    });

    test('count updates when clearing filters', async ({ user1Page: page }) => {
      await ensureMinimumRecipes(page, 10);
      
      await page.goto('/recipes');
      await expandFiltersOnMobile(page);

      await triggerFilterAndWait(page, () => page.locator('#search').fill(uniqueId.toString()));

      const filteredTotal = parseInt(await page.locator('#total-count').textContent() || '0');
      expect(filteredTotal).toBeGreaterThan(0);
      expect(filteredTotal).toBeLessThanOrEqual(20);

      await triggerFilterAndWait(page, () => page.getByRole('button', { name: 'Clear' }).click());

      const clearedTotal = parseInt(await page.locator('#total-count').textContent() || '0');
      expect(clearedTotal).toBeGreaterThan(filteredTotal);
    });
  });



  test.describe('Authored By Me Filter', () => {
    const authorFilterId = Date.now();

    test.beforeAll(async ({ browser }) => {
      const context1 = await browser.newContext();
      const page1 = await context1.newPage();
      await page1.goto('/login');
      await page1.locator('input[name="email"]').fill(TEST_USERS.approved1.email);
      await page1.locator('input[name="password"]').fill(TEST_USERS.approved1.password);
      await Promise.all([
        page1.waitForURL('/'),
        page1.getByRole('button', { name: 'Sign In' }).click(),
      ]);

      await page1.goto('/recipes/create');
      await page1.getByRole('textbox', { name: 'Title' }).fill(`User1 Recipe ${authorFilterId}`);
      await page1.locator('#preptime').fill('10');
      await page1.locator('#cooktime').fill('20');
      await page1.locator('#calories').fill('300');
      await fillToastEditor(page1, 'ingredients-editor', '- test ingredient');
      await fillToastEditor(page1, 'instructions-editor', '1. test instruction');
      await Promise.all([
        page1.waitForURL(/\/recipes\/\d+/),
        page1.getByRole('button', { name: /Create Recipe|Submit/i }).click(),
      ]);
      await context1.close();

      const context2 = await browser.newContext();
      const page2 = await context2.newPage();
      await page2.goto('/login');
      await page2.locator('input[name="email"]').fill(TEST_USERS.approved2.email);
      await page2.locator('input[name="password"]').fill(TEST_USERS.approved2.password);
      await Promise.all([
        page2.waitForURL('/'),
        page2.getByRole('button', { name: 'Sign In' }).click(),
      ]);

      await page2.goto('/recipes/create');
      await page2.getByRole('textbox', { name: 'Title' }).fill(`User2 Recipe ${authorFilterId}`);
      await page2.locator('#preptime').fill('15');
      await page2.locator('#cooktime').fill('25');
      await page2.locator('#calories').fill('400');
      await fillToastEditor(page2, 'ingredients-editor', '- another ingredient');
      await fillToastEditor(page2, 'instructions-editor', '1. another instruction');
      await Promise.all([
        page2.waitForURL(/\/recipes\/\d+/),
        page2.getByRole('button', { name: /Create Recipe|Submit/i }).click(),
      ]);
      await context2.close();
    });

    test('filters to show only recipes authored by current user', async ({ user1Page: page }) => {
      await page.goto('/recipes');
      await expandFiltersOnMobile(page);

      await triggerFilterAndWait(page, () => page.locator('#search').fill(authorFilterId.toString()));

      let titles = await getVisibleRecipeTitles(page);
      expect(titles.some(t => t.includes(`User1 Recipe ${authorFilterId}`))).toBe(true);
      expect(titles.some(t => t.includes(`User2 Recipe ${authorFilterId}`))).toBe(true);

      await triggerFilterAndWait(page, () => page.locator('#authored_by_me').check());

      titles = await getVisibleRecipeTitles(page);
      expect(titles.some(t => t.includes(`User1 Recipe ${authorFilterId}`))).toBe(true);
      expect(titles.some(t => t.includes(`User2 Recipe ${authorFilterId}`))).toBe(false);
    });

    test('unchecking filter shows all recipes again', async ({ user1Page: page }) => {
      await page.goto('/recipes');
      await expandFiltersOnMobile(page);

      await triggerFilterAndWait(page, () => page.locator('#search').fill(authorFilterId.toString()));

      await triggerFilterAndWait(page, () => page.locator('#authored_by_me').check());

      let titles = await getVisibleRecipeTitles(page);
      expect(titles.some(t => t.includes(`User2 Recipe ${authorFilterId}`))).toBe(false);

      await triggerFilterAndWait(page, () => page.locator('#authored_by_me').uncheck());

      titles = await getVisibleRecipeTitles(page);
      expect(titles.some(t => t.includes(`User1 Recipe ${authorFilterId}`))).toBe(true);
      expect(titles.some(t => t.includes(`User2 Recipe ${authorFilterId}`))).toBe(true);
    });

    test('clear button resets authored by me checkbox', async ({ user1Page: page }) => {
      await page.goto('/recipes');
      await expandFiltersOnMobile(page);

      await triggerFilterAndWait(page, () => page.locator('#authored_by_me').check());

      await expect(page.locator('#authored_by_me')).toBeChecked();

      await triggerFilterAndWait(page, () => page.getByRole('button', { name: 'Clear' }).click());

      await expect(page.locator('#authored_by_me')).not.toBeChecked();
    });

    test('authored by me filter only shows for logged-in users', async ({ browser }) => {
      const context = await browser.newContext();
      const page = await context.newPage();

      await page.goto('/recipes');
      await expandFiltersOnMobile(page);

      await expect(page.locator('#authored_by_me')).not.toBeVisible();

      await context.close();
    });

    test('user2 sees only their recipes when filtering', async ({ user2Page: page }) => {
      await page.goto('/recipes');
      await expandFiltersOnMobile(page);

      await triggerFilterAndWait(page, () => page.locator('#search').fill(authorFilterId.toString()));

      await triggerFilterAndWait(page, () => page.locator('#authored_by_me').check());

      const titles = await getVisibleRecipeTitles(page);
      expect(titles.some(t => t.includes(`User2 Recipe ${authorFilterId}`))).toBe(true);
      expect(titles.some(t => t.includes(`User1 Recipe ${authorFilterId}`))).toBe(false);
    });
  });

  test.describe('Clear Filter', () => {
    test('clear button resets all filters', async ({ user1Page: page }) => {
      await page.goto('/recipes');
      await expandFiltersOnMobile(page);

      await page.locator('#search').fill('test');
      await page.locator('#filter-tags-input').fill('vegetarian');
      await page.locator('#filter-tags-input').press('Enter');
      await page.locator('#calories_op').selectOption('lt');
      await page.locator('#calories_value').fill('500');
      // Wait for the last filter to complete
      await page.waitForTimeout(500);

      await triggerFilterAndWait(page, () => page.getByRole('button', { name: 'Clear' }).click());

      await expect(page.locator('#search')).toHaveValue('');
      await expect(page.locator('#calories_op')).toHaveValue('');
      await expect(page.locator('#calories_value')).toHaveValue('');
      await expect(page.locator('#filter-tags-container .tag')).toHaveCount(0);
    });

    test('clearing filter resets to paginated results (max 20 recipes)', async ({ user1Page: page }) => {
      await page.goto('/recipes');
      await expandFiltersOnMobile(page);

      const initialCount = await page.locator('.recipe-card').count();

      await triggerFilterAndWait(page, () => page.locator('#search').fill(uniqueId.toString()));

      const filteredCount = await page.locator('.recipe-card').count();
      expect(filteredCount).toBeLessThan(initialCount);

      await triggerFilterAndWait(page, () => page.getByRole('button', { name: 'Clear' }).click());

      const clearedCount = await page.locator('.recipe-card').count();
      expect(clearedCount).toBeLessThanOrEqual(20);
    });
  });
});
