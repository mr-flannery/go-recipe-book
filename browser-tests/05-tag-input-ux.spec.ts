import { test as base, expect, Page } from '@playwright/test';
import { TEST_USERS } from './test-users';
import { fillToastEditor } from './editor-helpers';

type AuthFixtures = {
  userPage: Page;
};

const test = base.extend<AuthFixtures>({
  userPage: async ({ browser }, use) => {
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
});

test.describe('Tag Input UX', () => {
  const uniqueId = Date.now();

  async function createRecipeAndNavigate(page: Page, title: string) {
    await page.goto('/recipes/create');
    await page.getByRole('textbox', { name: 'Title' }).fill(title);
    await page.locator('#preptime').fill('5');
    await page.locator('#cooktime').fill('10');
    await page.locator('#calories').fill('100');
    await fillToastEditor(page, 'ingredients-editor', '- 1 item');
    await fillToastEditor(page, 'instructions-editor', '1. Do something');
    await page.getByRole('button', { name: /Create Recipe|Submit/i }).click();
    await page.waitForURL(/\/recipes\/\d+/);
  }

  test.describe('Suggestions Visibility', () => {
    test('no suggestions shown on empty focus', async ({ userPage }) => {
      await userPage.goto('/recipes/create');
      
      const input = userPage.locator('#tags-input');
      const suggestions = userPage.locator('#tags-suggestions');
      
      await input.focus();
      await userPage.waitForTimeout(100);
      
      await expect(suggestions).not.toBeVisible();
    });

    test('suggestions hidden when input cleared', async ({ userPage }) => {
      await userPage.goto('/recipes/create');
      
      const input = userPage.locator('#tags-input');
      const suggestions = userPage.locator('#tags-suggestions');
      
      await input.fill('test');
      await expect(suggestions).toBeVisible();
      
      await input.fill('');
      await expect(suggestions).not.toBeVisible();
    });

    test('suggestions appear when typing', async ({ userPage }) => {
      await userPage.goto('/recipes/create');
      
      const input = userPage.locator('#tags-input');
      const suggestions = userPage.locator('#tags-suggestions');
      
      await input.fill('a');
      await expect(suggestions).toBeVisible();
    });
  });

  test.describe('Keyboard Navigation', () => {
    test('arrow down highlights first suggestion', async ({ userPage }) => {
      await userPage.goto('/recipes/create');
      
      const input = userPage.locator('#tags-input');
      const suggestions = userPage.locator('#tags-suggestions');
      
      await input.fill('test');
      await expect(suggestions).toBeVisible();
      
      await input.press('ArrowDown');
      
      const firstSuggestion = suggestions.locator('.tag-suggestion').first();
      await expect(firstSuggestion).toHaveClass(/selected/);
    });

    test('arrow up and down cycle through suggestions', async ({ userPage }) => {
      await userPage.goto('/recipes/create');
      
      const input = userPage.locator('#tags-input');
      const suggestions = userPage.locator('#tags-suggestions');
      
      await input.fill('test');
      await expect(suggestions).toBeVisible();
      
      await input.press('ArrowDown');
      const firstSuggestion = suggestions.locator('.tag-suggestion').first();
      await expect(firstSuggestion).toHaveClass(/selected/);
      
      await input.press('ArrowDown');
      const suggestionItems = suggestions.locator('.tag-suggestion');
      const count = await suggestionItems.count();
      if (count > 1) {
        await expect(suggestionItems.nth(1)).toHaveClass(/selected/);
        await expect(firstSuggestion).not.toHaveClass(/selected/);
      }
      
      await input.press('ArrowUp');
      await expect(firstSuggestion).toHaveClass(/selected/);
    });

    test('selected suggestion scrolls into view', async ({ userPage }) => {
      await createRecipeAndNavigate(userPage, `Scroll Test Recipe ${uniqueId}`);
      
      const input = userPage.locator('#author-tags-input');
      const suggestions = userPage.locator('#author-tags-suggestions');
      
      await input.fill('a');
      await expect(suggestions).toBeVisible();
      
      const suggestionItems = suggestions.locator('.tag-suggestion');
      const count = await suggestionItems.count();
      
      if (count > 4) {
        for (let i = 0; i < count; i++) {
          await input.press('ArrowDown');
        }
        
        const lastItem = suggestionItems.last();
        await expect(lastItem).toHaveClass(/selected/);
        
        const isVisible = await lastItem.isVisible();
        expect(isVisible).toBe(true);
      }
    });

    test('escape clears input and hides suggestions', async ({ userPage }) => {
      await userPage.goto('/recipes/create');
      
      const input = userPage.locator('#tags-input');
      const suggestions = userPage.locator('#tags-suggestions');
      
      await input.fill('test');
      await expect(suggestions).toBeVisible();
      await expect(input).toHaveValue('test');
      
      await input.press('Escape');
      
      await expect(suggestions).not.toBeVisible();
      await expect(input).toHaveValue('');
    });

    test('enter selects highlighted suggestion', async ({ userPage }) => {
      await userPage.goto('/recipes/create');
      
      const input = userPage.locator('#tags-input');
      const suggestions = userPage.locator('#tags-suggestions');
      const tagsContainer = userPage.locator('#tags-container');
      
      await input.fill('newtag');
      await expect(suggestions).toBeVisible();
      
      await input.press('ArrowDown');
      const selectedText = await suggestions.locator('.tag-suggestion.selected').textContent();
      
      await input.press('Enter');
      
      await expect(suggestions).not.toBeVisible();
      await expect(input).toHaveValue('');
      
      if (selectedText) {
        const cleanText = selectedText.replace(/^Create "(.+)"$/, '$1');
        await expect(tagsContainer.locator('.tag')).toContainText(cleanText);
      }
    });
  });

  test.describe('Overlay Behavior', () => {
    test('suggestions overlay uses absolute positioning', async ({ userPage }) => {
      await userPage.goto('/recipes/create');
      
      const input = userPage.locator('#tags-input');
      const suggestions = userPage.locator('#tags-suggestions');
      
      await input.fill('test');
      await expect(suggestions).toBeVisible();
      
      const position = await suggestions.evaluate((el) => {
        return window.getComputedStyle(el).position;
      });
      
      expect(position).toBe('absolute');
    });
  });

  test.describe('Remove Button', () => {
    test('clicking remove button deletes tag', async ({ userPage }) => {
      await userPage.goto('/recipes/create');
      
      const input = userPage.locator('#tags-input');
      const tagsContainer = userPage.locator('#tags-container');
      
      await input.fill('removeme');
      await input.press('Enter');
      
      await expect(tagsContainer.locator('.tag')).toHaveCount(1);
      await expect(tagsContainer.locator('.tag')).toContainText('removeme');
      
      await tagsContainer.locator('.tag-remove').click();
      
      await expect(tagsContainer.locator('.tag')).toHaveCount(0);
    });

    test('remove button has pointer cursor', async ({ userPage }) => {
      await userPage.goto('/recipes/create');
      
      const input = userPage.locator('#tags-input');
      const tagsContainer = userPage.locator('#tags-container');
      
      await input.fill('cursortest');
      await input.press('Enter');
      
      const removeButton = tagsContainer.locator('.tag-remove');
      await expect(removeButton).toBeVisible();
      
      const cursor = await removeButton.evaluate((el) => {
        return window.getComputedStyle(el).cursor;
      });
      
      expect(cursor).toBe('pointer');
    });
  });
});
