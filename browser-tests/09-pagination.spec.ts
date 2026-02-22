import { test, expect } from './fixtures';
import { fillToastEditor } from './editor-helpers';

test.describe('Pagination Controls', () => {
  const uniqueId = Date.now();
  const MIN_RECIPES_FOR_PAGINATION = 45;

  async function createSimpleRecipe(page, title: string): Promise<void> {
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

  async function ensureMinimumRecipes(page, minCount: number): Promise<void> {
    await page.goto('/recipes');
    const totalCount = parseInt(await page.locator('#total-count').textContent() || '0');
    
    if (totalCount < minCount) {
      const recipesToCreate = minCount - totalCount;
      for (let i = 0; i < recipesToCreate; i++) {
        await createSimpleRecipe(page, `Pagination Test Recipe ${uniqueId}-${i}`);
      }
    }
  }

  async function waitForFilterResults(page): Promise<void> {
    await page.waitForResponse(response => 
      response.url().includes('/recipes/filter') && response.status() === 200
    );
    await page.waitForTimeout(200);
  }

  async function getActivePage(paginationSelector: string, page): Promise<number> {
    const activePageElement = page.locator(`${paginationSelector} .pagination-page.current`);
    if (await activePageElement.count() === 0) {
      return 0;
    }
    const text = await activePageElement.textContent();
    return parseInt(text || '0');
  }

  test.beforeAll(async ({ browser }) => {
    const context = await browser.newContext();
    const page = await context.newPage();
    const { TEST_USERS } = await import('./test-users');
    const user = TEST_USERS.approved1;
    
    await page.goto('/login');
    await page.locator('input[name="email"]').fill(user.email);
    await page.locator('input[name="password"]').fill(user.password);
    await page.getByRole('button', { name: 'Sign In' }).click();
    await page.waitForURL('/');

    await ensureMinimumRecipes(page, MIN_RECIPES_FOR_PAGINATION);

    await context.close();
  });

  test.describe('Header and Footer Pagination', () => {
    test('both pagination controls are visible on page load', async ({ authenticatedPage: page }) => {
      await page.goto('/recipes');

      await expect(page.locator('#pagination-control')).toBeVisible();
      await expect(page.locator('#pagination-footer')).toBeVisible();
    });

    test('both pagination controls show same active page on initial load', async ({ authenticatedPage: page }) => {
      await page.goto('/recipes');

      const headerActivePage = await getActivePage('#pagination-control', page);
      const footerActivePage = await getActivePage('#pagination-footer', page);

      expect(headerActivePage).toBe(1);
      expect(footerActivePage).toBe(1);
    });

    test('clicking page number in header updates both controls', async ({ authenticatedPage: page }) => {
      await page.goto('/recipes');

      await page.locator('#pagination-control .pagination-btn.pagination-page').filter({ hasText: '2' }).click();
      await waitForFilterResults(page);

      const headerActivePage = await getActivePage('#pagination-control', page);
      const footerActivePage = await getActivePage('#pagination-footer', page);

      expect(headerActivePage).toBe(2);
      expect(footerActivePage).toBe(2);
    });

    test('clicking page number in footer updates both controls', async ({ authenticatedPage: page }) => {
      await page.goto('/recipes');

      await page.locator('#pagination-footer .pagination-btn.pagination-page').filter({ hasText: '2' }).click();
      await waitForFilterResults(page);

      const headerActivePage = await getActivePage('#pagination-control', page);
      const footerActivePage = await getActivePage('#pagination-footer', page);

      expect(headerActivePage).toBe(2);
      expect(footerActivePage).toBe(2);
    });
  });

  test.describe('Load More and Load Previous Independent Tracking', () => {
    test('Load More updates only footer pagination', async ({ authenticatedPage: page }) => {
      await page.goto('/recipes');

      // Navigate to page 2 first
      await page.locator('#pagination-control .pagination-btn.pagination-page').filter({ hasText: '2' }).click();
      await waitForFilterResults(page);

      const headerBefore = await getActivePage('#pagination-control', page);
      const footerBefore = await getActivePage('#pagination-footer', page);
      expect(headerBefore).toBe(2);
      expect(footerBefore).toBe(2);

      // Click Load More
      await page.getByRole('button', { name: 'Load More' }).click();
      await waitForFilterResults(page);

      const headerAfter = await getActivePage('#pagination-control', page);
      const footerAfter = await getActivePage('#pagination-footer', page);

      // Header should stay at page 2, footer should advance to page 3
      expect(headerAfter).toBe(2);
      expect(footerAfter).toBe(3);
    });

    test('Load Previous updates only header pagination', async ({ authenticatedPage: page }) => {
      await page.goto('/recipes');

      // Navigate to page 3 first
      await page.locator('#pagination-control .pagination-btn.pagination-page').filter({ hasText: '3' }).click();
      await waitForFilterResults(page);

      const headerBefore = await getActivePage('#pagination-control', page);
      const footerBefore = await getActivePage('#pagination-footer', page);
      expect(headerBefore).toBe(3);
      expect(footerBefore).toBe(3);

      // Click Load Previous
      await page.getByRole('button', { name: 'Load Previous' }).click();
      await waitForFilterResults(page);

      const headerAfter = await getActivePage('#pagination-control', page);
      const footerAfter = await getActivePage('#pagination-footer', page);

      // Header should go back to page 2, footer should stay at page 3
      expect(headerAfter).toBe(2);
      expect(footerAfter).toBe(3);
    });

    test('combined Load More and Load Previous track boundaries independently', async ({ authenticatedPage: page }) => {
      await page.goto('/recipes');

      // Navigate to page 3
      await page.locator('#pagination-control .pagination-btn.pagination-page').filter({ hasText: '3' }).click();
      await waitForFilterResults(page);

      // Click Load More - footer should show page 4
      await page.getByRole('button', { name: 'Load More' }).click();
      await waitForFilterResults(page);

      let headerPage = await getActivePage('#pagination-control', page);
      let footerPage = await getActivePage('#pagination-footer', page);
      expect(headerPage).toBe(3);
      expect(footerPage).toBe(4);

      // Click Load Previous - header should show page 2
      await page.getByRole('button', { name: 'Load Previous' }).click();
      await waitForFilterResults(page);

      headerPage = await getActivePage('#pagination-control', page);
      footerPage = await getActivePage('#pagination-footer', page);
      expect(headerPage).toBe(2);
      expect(footerPage).toBe(4);
    });
  });

  test.describe('Navigation Buttons', () => {
    test('first page button navigates to page 1', async ({ authenticatedPage: page }) => {
      await page.goto('/recipes');

      // Go to page 3
      await page.locator('#pagination-control .pagination-btn.pagination-page').filter({ hasText: '3' }).click();
      await waitForFilterResults(page);

      // Click first page button
      await page.locator('#pagination-control .pagination-first').click();
      await waitForFilterResults(page);

      const headerPage = await getActivePage('#pagination-control', page);
      expect(headerPage).toBe(1);
    });

    test('last page button navigates to final page', async ({ authenticatedPage: page }) => {
      await page.goto('/recipes');

      const totalCount = parseInt(await page.locator('#total-count').textContent() || '0');
      const totalPages = Math.ceil(totalCount / 20);

      // Click last page button
      await page.locator('#pagination-control .pagination-last').click();
      await waitForFilterResults(page);

      const headerPage = await getActivePage('#pagination-control', page);
      expect(headerPage).toBe(totalPages);
    });

    test('previous button decrements page', async ({ authenticatedPage: page }) => {
      await page.goto('/recipes');

      // Go to page 3
      await page.locator('#pagination-control .pagination-btn.pagination-page').filter({ hasText: '3' }).click();
      await waitForFilterResults(page);

      // Click previous
      await page.locator('#pagination-control .pagination-prev').click();
      await waitForFilterResults(page);

      const headerPage = await getActivePage('#pagination-control', page);
      expect(headerPage).toBe(2);
    });

    test('next button increments page', async ({ authenticatedPage: page }) => {
      await page.goto('/recipes');

      // Click next
      await page.locator('#pagination-control .pagination-next').click();
      await waitForFilterResults(page);

      const headerPage = await getActivePage('#pagination-control', page);
      expect(headerPage).toBe(2);
    });

    test('navigation buttons not visible on first page for going back', async ({ authenticatedPage: page }) => {
      await page.goto('/recipes');

      await expect(page.locator('#pagination-control .pagination-first')).not.toBeVisible();
      await expect(page.locator('#pagination-control .pagination-prev')).not.toBeVisible();
    });

    test('navigation buttons not visible on last page for going forward', async ({ authenticatedPage: page }) => {
      await page.goto('/recipes');

      // Go to last page
      await page.locator('#pagination-control .pagination-last').click();
      await waitForFilterResults(page);

      await expect(page.locator('#pagination-control .pagination-next')).not.toBeVisible();
      await expect(page.locator('#pagination-control .pagination-last')).not.toBeVisible();
    });
  });

  test.describe('Load More Button Behavior', () => {
    test('load more button disappears when navigating to last page', async ({ authenticatedPage: page }) => {
      await page.goto('/recipes');

      // Navigate to last page using pagination
      await page.locator('#pagination-control .pagination-last').click();
      await waitForFilterResults(page);

      // Load More should not be visible on last page
      await expect(page.getByRole('button', { name: 'Load More' })).not.toBeVisible();

      // Verify we loaded all remaining recipes
      const totalCount = parseInt(await page.locator('#total-count').textContent() || '0');
      const recipeCount = await page.locator('.recipe-card').count();
      const expectedOnLastPage = totalCount % 20 || 20;
      expect(recipeCount).toBe(expectedOnLastPage);
    });

    test('load previous button appears when navigating past first page', async ({ authenticatedPage: page }) => {
      await page.goto('/recipes');

      // Initially Load Previous should not be visible
      await expect(page.getByRole('button', { name: 'Load Previous' })).not.toBeVisible();

      // Navigate to page 2
      await page.locator('#pagination-control .pagination-btn.pagination-page').filter({ hasText: '2' }).click();
      await waitForFilterResults(page);

      // Load Previous should now be visible
      await expect(page.getByRole('button', { name: 'Load Previous' })).toBeVisible();
    });

    test('load previous button disappears when all previous pages are loaded', async ({ authenticatedPage: page }) => {
      await page.goto('/recipes');

      // Navigate to page 2
      await page.locator('#pagination-control .pagination-btn.pagination-page').filter({ hasText: '2' }).click();
      await waitForFilterResults(page);

      // Click Load Previous to load page 1
      await page.getByRole('button', { name: 'Load Previous' }).click();
      await waitForFilterResults(page);

      // Load Previous should disappear since we're now at page 1
      await expect(page.getByRole('button', { name: 'Load Previous' })).not.toBeVisible();
    });
  });
});
