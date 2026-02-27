import { test, expect } from './fixtures';
import { fillToastEditor } from './editor-helpers';

test.describe('Pagination Controls', () => {
  const uniqueId = Date.now();
  const MIN_RECIPES_FOR_PAGINATION = 65;

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

  async function clickAndWaitForHtmx(page, locator): Promise<void> {
    const responsePromise = page.waitForResponse(response => 
      response.url().includes('/recipes/filter') && response.status() === 200
    );
    await locator.click();
    await responsePromise;
    await page.waitForFunction(() => {
      return document.querySelectorAll('.htmx-request').length === 0 &&
             document.querySelectorAll('.htmx-settling').length === 0;
    }, { timeout: 5000 });
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

      await clickAndWaitForHtmx(page, page.locator('#pagination-control .pagination-btn.pagination-page').filter({ hasText: '2' }));

      const headerActivePage = await getActivePage('#pagination-control', page);
      const footerActivePage = await getActivePage('#pagination-footer', page);

      expect(headerActivePage).toBe(2);
      expect(footerActivePage).toBe(2);
    });

    test('clicking page number in footer updates both controls', async ({ authenticatedPage: page }) => {
      await page.goto('/recipes');

      await clickAndWaitForHtmx(page, page.locator('#pagination-footer .pagination-btn.pagination-page').filter({ hasText: '2' }));

      const headerActivePage = await getActivePage('#pagination-control', page);
      const footerActivePage = await getActivePage('#pagination-footer', page);

      expect(headerActivePage).toBe(2);
      expect(footerActivePage).toBe(2);
    });
  });

  test.describe('Navigation Buttons', () => {
    test('first page button navigates to page 1', async ({ authenticatedPage: page }) => {
      await page.goto('/recipes');

      // Go to page 3
      await clickAndWaitForHtmx(page, page.locator('#pagination-control .pagination-btn.pagination-page').filter({ hasText: '3' }));

      // Click first page button
      await clickAndWaitForHtmx(page, page.locator('#pagination-control .pagination-first'));

      const headerPage = await getActivePage('#pagination-control', page);
      expect(headerPage).toBe(1);
    });

    test('last page button navigates to final page', async ({ authenticatedPage: page }) => {
      await page.goto('/recipes');

      const totalCount = parseInt(await page.locator('#total-count').textContent() || '0');
      const totalPages = Math.ceil(totalCount / 20);

      // Click last page button
      await clickAndWaitForHtmx(page, page.locator('#pagination-control .pagination-last'));

      const headerPage = await getActivePage('#pagination-control', page);
      expect(headerPage).toBe(totalPages);
    });

    test('previous button decrements page', async ({ authenticatedPage: page }) => {
      await page.goto('/recipes');

      // Go to page 3
      await clickAndWaitForHtmx(page, page.locator('#pagination-control .pagination-btn.pagination-page').filter({ hasText: '3' }));

      // Click previous
      await clickAndWaitForHtmx(page, page.locator('#pagination-control .pagination-prev'));

      const headerPage = await getActivePage('#pagination-control', page);
      expect(headerPage).toBe(2);
    });

    test('next button increments page', async ({ authenticatedPage: page }) => {
      await page.goto('/recipes');

      // Click next
      await clickAndWaitForHtmx(page, page.locator('#pagination-control .pagination-next'));

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
      await clickAndWaitForHtmx(page, page.locator('#pagination-control .pagination-last'));

      await expect(page.locator('#pagination-control .pagination-next')).not.toBeVisible();
      await expect(page.locator('#pagination-control .pagination-last')).not.toBeVisible();
    });
  });

  test.describe('Page Size Change', () => {
    test('changing page size preserves approximate position in list', async ({ authenticatedPage: page }) => {
      await page.goto('/recipes');

      // Navigate to page 3 (items 41-60 with page size 20)
      await clickAndWaitForHtmx(page, page.locator('#pagination-control .pagination-btn.pagination-page').filter({ hasText: '3' }));

      const headerPageBefore = await getActivePage('#pagination-control', page);
      expect(headerPageBefore).toBe(3);

      // Change page size to 50
      // This should calculate: offset = (3-1) * 20 = 40, newPage = floor(40/50) + 1 = 1
      await page.locator('#page-size-select').selectOption('50');
      await page.waitForResponse(response => 
        response.url().includes('/recipes/filter') && response.status() === 200
      );
      await page.waitForFunction(() => {
        return document.querySelectorAll('.htmx-request').length === 0 &&
               document.querySelectorAll('.htmx-settling').length === 0;
      }, { timeout: 5000 });

      // With page size 50, item 41 is on page 1 (items 1-50)
      const headerPageAfter = await getActivePage('#pagination-control', page);
      expect(headerPageAfter).toBe(1);

      // Verify the page size control was updated
      const selectedValue = await page.locator('#page-size-select').inputValue();
      expect(selectedValue).toBe('50');
    });

    test('changing page size on page 1 stays on page 1', async ({ authenticatedPage: page }) => {
      await page.goto('/recipes');

      const headerPageBefore = await getActivePage('#pagination-control', page);
      expect(headerPageBefore).toBe(1);

      // Change page size to 50
      await page.locator('#page-size-select').selectOption('50');
      await page.waitForResponse(response => 
        response.url().includes('/recipes/filter') && response.status() === 200
      );
      await page.waitForFunction(() => {
        return document.querySelectorAll('.htmx-request').length === 0 &&
               document.querySelectorAll('.htmx-settling').length === 0;
      }, { timeout: 5000 });

      // Should stay on page 1
      const headerPageAfter = await getActivePage('#pagination-control', page);
      expect(headerPageAfter).toBe(1);
    });
  });
});
