import { test, expect } from './fixtures';

test.describe('Recipe Creation', () => {
  const testRecipe = {
    title: `Test Recipe ${Date.now()}`,
    prepTime: '15',
    cookTime: '30',
    calories: '450',
    ingredients: '- 2 cups flour\n- 1 cup sugar\n- 3 eggs',
    instructions: '1. Mix dry ingredients\n2. Add eggs\n3. Bake at 350°F for 30 minutes',
  };

  test('creates a new recipe and verifies it appears on overview and detail page', async ({ authenticatedPage: page }) => {
    // Navigate to recipe overview page
    await page.goto('/recipes');
    await expect(page.getByRole('heading', { name: 'Recipe Collection', level: 1 })).toBeVisible();

    // Navigate to create recipe page
    await page.getByRole('link', { name: 'Submit A Recipe' }).click();
    await page.waitForURL('/recipes/create');
    await expect(page.getByRole('heading', { name: /Submit.*Recipe/i, level: 1 })).toBeVisible();

    // Fill in the recipe form
    await page.getByRole('textbox', { name: 'Title' }).fill(testRecipe.title);
    await page.locator('#preptime').fill(testRecipe.prepTime);
    await page.locator('#cooktime').fill(testRecipe.cookTime);
    await page.locator('#calories').fill(testRecipe.calories);
    await page.locator('#ingredients').fill(testRecipe.ingredients);
    await page.locator('#instructions').fill(testRecipe.instructions);

    // Submit the form
    await page.getByRole('button', { name: /Create Recipe|Submit/i }).click();

    // Should redirect to the new recipe's page
    await page.waitForURL(/\/recipes\/\d+/);
    
    // Verify we're on the recipe detail page with correct content
    await expect(page.getByRole('heading', { name: testRecipe.title, level: 1 })).toBeVisible();
    await expect(page.locator('.recipe-meta').getByText(`${testRecipe.prepTime} min`)).toBeVisible();
    await expect(page.locator('.recipe-meta').getByText(`${testRecipe.cookTime} min`)).toBeVisible();
    await expect(page.locator('.recipe-meta .meta-item .meta-value').getByText(testRecipe.calories)).toBeVisible();
    await expect(page.getByText('2 cups flour')).toBeVisible();
    await expect(page.getByText('Mix dry ingredients')).toBeVisible();

    // Navigate back to recipe overview
    await page.goto('/recipes');

    // Verify the new recipe appears in the list
    await expect(page.getByText(testRecipe.title)).toBeVisible();

    // Click on the recipe to navigate to its detail page
    await page.getByText(testRecipe.title).click();
    await page.waitForURL(/\/recipes\/\d+/);

    // Verify the recipe details again
    await expect(page.getByRole('heading', { name: testRecipe.title, level: 1 })).toBeVisible();
  });
});
