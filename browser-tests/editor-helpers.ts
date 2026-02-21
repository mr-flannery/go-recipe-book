import { Page } from '@playwright/test';

/**
 * Fill a TOAST UI Editor with the specified content.
 * Uses the editor's API directly via page.evaluate() since contenteditable
 * divs don't support Playwright's .fill() method.
 * 
 * @param page - Playwright page object
 * @param editorId - The ID of the editor container div (e.g., 'ingredients-editor')
 * @param content - The markdown content to fill
 */
export async function fillToastEditor(page: Page, editorId: string, content: string): Promise<void> {
  // Wait for the editor to be initialized
  const editorContainer = page.locator(`#${editorId}`);
  await editorContainer.locator('.toastui-editor-md-container').waitFor({ state: 'visible' });
  
  // Use the TOAST UI Editor API to set content directly
  await page.evaluate(({ editorId, content }) => {
    const editor = (window as any).RecipeEditor?.get(editorId);
    if (editor) {
      editor.setMarkdown(content);
    } else {
      throw new Error(`Editor not found: ${editorId}`);
    }
  }, { editorId, content });
}

/**
 * Clear a TOAST UI Editor's content.
 * 
 * @param page - Playwright page object
 * @param editorId - The ID of the editor container div (e.g., 'ingredients-editor')
 */
export async function clearToastEditor(page: Page, editorId: string): Promise<void> {
  await page.evaluate((editorId) => {
    const editor = (window as any).RecipeEditor?.get(editorId);
    if (editor) {
      editor.setMarkdown('');
    } else {
      throw new Error(`Editor not found: ${editorId}`);
    }
  }, editorId);
}
