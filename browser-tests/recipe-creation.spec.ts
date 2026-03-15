import { test, expect } from './fixtures';
import { fillToastEditor } from './editor-helpers';

test.describe.serial('Recipe Creation', () => {
  const testRecipe = {
    title: `Test Recipe ${Date.now()}`,
    description: 'A delicious **test recipe** with rich flavors.',
    source: 'https://example.com/recipes/test',
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
    await Promise.all([
      page.waitForURL('/recipes/create'),
      page.getByRole('link', { name: 'Submit A Recipe' }).click(),
    ]);
    await expect(page.getByRole('heading', { name: /Submit.*Recipe/i, level: 1 })).toBeVisible();

    // Fill in the recipe form
    await page.getByRole('textbox', { name: 'Title' }).fill(testRecipe.title);
    await page.locator('#description').fill(testRecipe.description);
    await page.locator('#source').fill(testRecipe.source);
    await page.locator('#preptime').fill(testRecipe.prepTime);
    await page.locator('#cooktime').fill(testRecipe.cookTime);
    await page.locator('#calories').fill(testRecipe.calories);
    await fillToastEditor(page, 'ingredients-editor', testRecipe.ingredients);
    await fillToastEditor(page, 'instructions-editor', testRecipe.instructions);

    // Submit the form
    await page.getByRole('button', { name: /Create Recipe|Submit/i }).click();

    // Should redirect to the new recipe's page
    await page.waitForURL(/\/recipes\/\d+/);
    
    // Verify we're on the recipe detail page with correct content
    await expect(page.getByRole('heading', { name: testRecipe.title, level: 1 })).toBeVisible();
    // Description is rendered as markdown in its own section
    await expect(page.getByRole('heading', { name: 'Description', level: 2 })).toBeVisible();
    await expect(page.locator('.recipe-section').filter({ hasText: 'Description' }).getByText('test recipe')).toBeVisible();
    // Source is a URL, so it should be rendered as a link
    await expect(page.locator('.recipe-source a[href="https://example.com/recipes/test"]')).toBeVisible();
    await expect(page.locator('.recipe-meta').getByText(`${testRecipe.prepTime} min`)).toBeVisible();
    await expect(page.locator('.recipe-meta').getByText(`${testRecipe.cookTime} min`)).toBeVisible();
    await expect(page.locator('.recipe-meta .meta-item .meta-value').getByText(testRecipe.calories)).toBeVisible();
    await expect(page.getByText('2 cups flour')).toBeVisible();
    await expect(page.getByText('Mix dry ingredients')).toBeVisible();

    // Navigate back to recipe overview
    await page.goto('/recipes');

    // Verify the new recipe appears in the list (use exact match to avoid matching "Cancel Test Recipe...")
    await expect(page.getByText(testRecipe.title, { exact: true })).toBeVisible();

    // Click on the recipe to navigate to its detail page
    await Promise.all([
      page.waitForURL(/\/recipes\/\d+/),
      page.getByText(testRecipe.title, { exact: true }).click(),
    ]);

    // Verify the recipe details again
    await expect(page.getByRole('heading', { name: testRecipe.title, level: 1 })).toBeVisible();
  });
});
