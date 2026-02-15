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

const uniqueId = Date.now();

async function createRecipe(page: Page, title: string): Promise<string> {
  await page.goto('/recipes/create');
  await page.getByRole('textbox', { name: 'Title' }).fill(title);
  await page.locator('#preptime').fill('10');
  await page.locator('#cooktime').fill('20');
  await page.locator('#calories').fill('300');
  await page.locator('#ingredients').fill('- 1 cup rice');
  await page.locator('#instructions').fill('1. Cook rice');
  await page.getByRole('button', { name: /Create Recipe|Submit/i }).click();
  await page.waitForURL(/\/recipes\/\d+/);
  return page.url().match(/\/recipes\/(\d+)/)?.[1] || '';
}

test.describe('Comments', () => {
  test.describe('Adding Comments', () => {
    test('any logged-in user can add a comment to any recipe', async ({ user1Page, user2Page }) => {
      const recipeId = await createRecipe(user1Page, `Comment Test Recipe ${uniqueId}`);

      await user2Page.goto(`/recipes/${recipeId}`);
      await expect(user2Page.locator('#comment-count')).toHaveText('0');
      await expect(user2Page.locator('#no-comments')).toBeVisible();

      await user2Page.locator('#new-comment').fill('This looks delicious!');
      await user2Page.getByRole('button', { name: 'Post Comment' }).click();

      await expect(user2Page.locator('#comment-count')).toHaveText('1');
      await expect(user2Page.locator('#no-comments')).not.toBeVisible();
      await expect(user2Page.locator('.comment-content').first()).toContainText('This looks delicious!');
    });

    test('comment form clears after submission', async ({ user1Page }) => {
      const recipeId = await createRecipe(user1Page, `Comment Clear Test ${uniqueId}`);

      await user1Page.goto(`/recipes/${recipeId}`);
      const textarea = user1Page.locator('#new-comment');
      await textarea.fill('Test comment');
      await user1Page.getByRole('button', { name: 'Post Comment' }).click();

      await expect(user1Page.locator('.comment-content').first()).toContainText('Test comment');
      await expect(textarea).toHaveValue('');
    });

    test('multiple users can comment on the same recipe', async ({ user1Page, user2Page }) => {
      const recipeId = await createRecipe(user1Page, `Multi Comment Test ${uniqueId}`);

      await user1Page.goto(`/recipes/${recipeId}`);
      await user1Page.locator('#new-comment').fill('Author comment');
      await user1Page.getByRole('button', { name: 'Post Comment' }).click();
      await expect(user1Page.locator('#comment-count')).toHaveText('1');

      await user2Page.goto(`/recipes/${recipeId}`);
      await user2Page.locator('#new-comment').fill('User 2 comment');
      await user2Page.getByRole('button', { name: 'Post Comment' }).click();
      await expect(user2Page.locator('#comment-count')).toHaveText('2');

      await user1Page.reload();
      await expect(user1Page.locator('#comment-count')).toHaveText('2');
      await expect(user1Page.locator('.comment')).toHaveCount(2);
    });
  });

  test.describe('Editing Comments', () => {
    test('comment author can edit their own comment', async ({ user1Page }) => {
      const recipeId = await createRecipe(user1Page, `Edit Comment Test ${uniqueId}`);

      await user1Page.goto(`/recipes/${recipeId}`);
      await user1Page.locator('#new-comment').fill('Original comment');
      await user1Page.getByRole('button', { name: 'Post Comment' }).click();

      await expect(user1Page.locator('.comment-content').first()).toContainText('Original comment');
      await expect(user1Page.locator('.comment-actions button').filter({ hasText: 'Edit' }).first()).toBeVisible();

      await user1Page.locator('.comment-actions button').filter({ hasText: 'Edit' }).first().click();
      const editForm = user1Page.locator('.comment-edit-form').first();
      await expect(editForm).toBeVisible();
      await editForm.locator('textarea[name="comment"]').fill('Updated comment');
      await editForm.getByRole('button', { name: 'Save' }).click();

      await expect(user1Page.locator('.comment-content').first()).toContainText('Updated comment');
      await expect(user1Page.locator('.comment-content').first()).not.toContainText('Original comment');
    });

    test('edit cancel button restores original view', async ({ user1Page }) => {
      const recipeId = await createRecipe(user1Page, `Cancel Edit Test ${uniqueId}`);

      await user1Page.goto(`/recipes/${recipeId}`);
      await user1Page.locator('#new-comment').fill('Test comment');
      await user1Page.getByRole('button', { name: 'Post Comment' }).click();

      await user1Page.locator('.comment-actions button').filter({ hasText: 'Edit' }).first().click();
      const editForm = user1Page.locator('.comment-edit-form').first();
      await expect(editForm).toBeVisible();

      await editForm.getByRole('button', { name: 'Cancel' }).click();
      await expect(editForm).not.toBeVisible();
      await expect(user1Page.locator('.comment-content').first()).toBeVisible();
    });

    test('user cannot edit another user\'s comment', async ({ user1Page, user2Page }) => {
      const recipeId = await createRecipe(user1Page, `No Edit Other Test ${uniqueId}`);

      await user1Page.goto(`/recipes/${recipeId}`);
      await user1Page.locator('#new-comment').fill('User 1 comment');
      await user1Page.getByRole('button', { name: 'Post Comment' }).click();
      await expect(user1Page.locator('.comment-content').first()).toContainText('User 1 comment');

      await user2Page.goto(`/recipes/${recipeId}`);
      await expect(user2Page.locator('.comment-content').first()).toContainText('User 1 comment');
      await expect(user2Page.locator('.comment-actions button').filter({ hasText: 'Edit' })).not.toBeVisible();
    });
  });

  test.describe('Deleting Comments', () => {
    test('comment author can delete their own comment', async ({ user1Page }) => {
      const recipeId = await createRecipe(user1Page, `Delete Comment Test ${uniqueId}`);

      await user1Page.goto(`/recipes/${recipeId}`);
      await user1Page.locator('#new-comment').fill('Comment to delete');
      await user1Page.getByRole('button', { name: 'Post Comment' }).click();

      await expect(user1Page.locator('#comment-count')).toHaveText('1');
      await expect(user1Page.locator('.comment')).toHaveCount(1);

      user1Page.on('dialog', dialog => dialog.accept());
      await user1Page.locator('.comment-actions button').filter({ hasText: 'Delete' }).first().click();

      await expect(user1Page.locator('#comment-count')).toHaveText('0');
      await expect(user1Page.locator('.comment')).toHaveCount(0);
    });

    test('user cannot delete another user\'s comment', async ({ user1Page, user2Page }) => {
      const recipeId = await createRecipe(user1Page, `No Delete Other Test ${uniqueId}`);

      await user1Page.goto(`/recipes/${recipeId}`);
      await user1Page.locator('#new-comment').fill('User 1 comment');
      await user1Page.getByRole('button', { name: 'Post Comment' }).click();
      await expect(user1Page.locator('.comment-content').first()).toContainText('User 1 comment');

      await user2Page.goto(`/recipes/${recipeId}`);
      await expect(user2Page.locator('.comment-content').first()).toContainText('User 1 comment');
      await expect(user2Page.locator('.comment-actions button').filter({ hasText: 'Delete' })).not.toBeVisible();
    });

    test('delete confirmation can be cancelled', async ({ user1Page }) => {
      const recipeId = await createRecipe(user1Page, `Cancel Delete Test ${uniqueId}`);

      await user1Page.goto(`/recipes/${recipeId}`);
      await user1Page.locator('#new-comment').fill('Comment to keep');
      await user1Page.getByRole('button', { name: 'Post Comment' }).click();

      await expect(user1Page.locator('#comment-count')).toHaveText('1');

      user1Page.on('dialog', dialog => dialog.dismiss());
      await user1Page.locator('.comment-actions button').filter({ hasText: 'Delete' }).first().click();

      await expect(user1Page.locator('#comment-count')).toHaveText('1');
      await expect(user1Page.locator('.comment')).toHaveCount(1);
    });
  });

  test.describe('Comment Author Permissions', () => {
    test('user sees edit/delete buttons only on their own comments', async ({ user1Page, user2Page }) => {
      const recipeId = await createRecipe(user1Page, `Permissions Test ${uniqueId}`);

      await user1Page.goto(`/recipes/${recipeId}`);
      await user1Page.locator('#new-comment').fill('User 1 comment');
      await user1Page.getByRole('button', { name: 'Post Comment' }).click();
      await expect(user1Page.locator('.comment-content').first()).toContainText('User 1 comment');

      await user2Page.goto(`/recipes/${recipeId}`);
      await user2Page.locator('#new-comment').fill('User 2 comment');
      await user2Page.getByRole('button', { name: 'Post Comment' }).click();
      await expect(user2Page.locator('.comment-content').first()).toContainText('User 2 comment');

      await user1Page.reload();

      const comments = user1Page.locator('.comment');
      await expect(comments).toHaveCount(2);

      const user1Comment = comments.filter({ hasText: 'User 1 comment' });
      const user2Comment = comments.filter({ hasText: 'User 2 comment' });

      await expect(user1Comment.locator('.comment-actions button').filter({ hasText: 'Edit' })).toBeVisible();
      await expect(user1Comment.locator('.comment-actions button').filter({ hasText: 'Delete' })).toBeVisible();
      await expect(user2Comment.locator('.comment-actions')).not.toBeVisible();
    });

    test('recipe author can comment but cannot edit/delete others\' comments', async ({ user1Page, user2Page }) => {
      const recipeId = await createRecipe(user1Page, `Author Perms Test ${uniqueId}`);

      await user2Page.goto(`/recipes/${recipeId}`);
      await user2Page.locator('#new-comment').fill('Visitor comment');
      await user2Page.getByRole('button', { name: 'Post Comment' }).click();

      await user1Page.goto(`/recipes/${recipeId}`);
      const visitorComment = user1Page.locator('.comment').filter({ hasText: 'Visitor comment' });
      await expect(visitorComment.locator('.comment-actions')).not.toBeVisible();

      await user1Page.locator('#new-comment').fill('Author comment');
      await user1Page.getByRole('button', { name: 'Post Comment' }).click();

      const authorComment = user1Page.locator('.comment').filter({ hasText: 'Author comment' });
      await expect(authorComment.locator('.comment-actions button').filter({ hasText: 'Edit' })).toBeVisible();
      await expect(authorComment.locator('.comment-actions button').filter({ hasText: 'Delete' })).toBeVisible();
    });
  });
});
