import { test, expect } from './fixtures';

test.describe('Load More Recipes', () => {
  const RECIPES_PER_PAGE = 20;
  const MIN_RECIPES_FOR_TEST = 21;

  async function getRecipeCount(page): Promise<number> {
    return await page.locator('.recipe-card').count();
  }

  async function createRecipe(page, index: number): Promise<void> {
    await page.goto('/recipes/create');
    await page.getByRole('textbox', { name: 'Title' }).fill(`Load More Test Recipe ${Date.now()}-${index}`);
    await page.locator('#preptime').fill('10');
    await page.locator('#cooktime').fill('20');
    await page.locator('#calories').fill('300');
    await page.locator('#ingredients').fill('- Test ingredient');
    await page.locator('#instructions').fill('Test instructions');
    await page.getByRole('button', { name: /Create Recipe|Submit/i }).click();
    await page.waitForURL(/\/recipes\/\d+/);
  }

  test('displays 20 recipes initially and loads more when clicking Load More button', async ({ authenticatedPage: page }) => {
    await page.goto('/recipes');

    let currentCount = await getRecipeCount(page);
    const loadMoreButton = page.getByRole('button', { name: 'Load More' });
    const hasLoadMore = await loadMoreButton.isVisible().catch(() => false);

    if (currentCount < RECIPES_PER_PAGE || !hasLoadMore) {
      const recipesToCreate = MIN_RECIPES_FOR_TEST - currentCount;
      for (let i = 0; i < recipesToCreate; i++) {
        await createRecipe(page, i);
      }
      await page.goto('/recipes');
    }

    const initialCount = await getRecipeCount(page);
    expect(initialCount).toBe(RECIPES_PER_PAGE);

    await expect(page.getByRole('button', { name: 'Load More' })).toBeVisible();

    await page.getByRole('button', { name: 'Load More' }).click();

    await page.waitForFunction(
      (expectedMin) => document.querySelectorAll('.recipe-card').length > expectedMin,
      RECIPES_PER_PAGE,
      { timeout: 5000 }
    );

    const newCount = await getRecipeCount(page);
    expect(newCount).toBeGreaterThan(RECIPES_PER_PAGE);
  });
});
