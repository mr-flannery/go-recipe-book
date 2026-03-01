import { Page } from '@playwright/test';

/**
 * Fill an EasyMDE editor with the specified content.
 * Uses the editor's API directly via page.evaluate().
 * 
 * For backwards compatibility, accepts IDs like 'ingredients-editor' 
 * and maps them to textarea IDs like 'ingredients'.
 * 
 * @param page - Playwright page object
 * @param editorId - The ID of the editor (e.g., 'ingredients-editor' or 'ingredients')
 * @param content - The markdown content to fill
 */
export async function fillToastEditor(page: Page, editorId: string, content: string): Promise<void> {
  // Map old editor container IDs to actual textarea IDs
  const textareaId = editorId.replace(/-editor$/, '');
  
  // Wait for the EasyMDE editor to be initialized via the RecipeEditor API
  await page.waitForFunction(
    (id) => !!(window as any).RecipeEditor?.get(id),
    textareaId,
    { timeout: 10000 }
  );
  
  // Use the EasyMDE API to set content directly
  await page.evaluate(({ textareaId, content }) => {
    const editor = (window as any).RecipeEditor?.get(textareaId);
    if (editor) {
      editor.value(content);
    } else {
      throw new Error(`Editor not found: ${textareaId}`);
    }
  }, { textareaId, content });
}

/**
 * Clear an EasyMDE editor's content.
 * 
 * @param page - Playwright page object
 * @param editorId - The ID of the editor (e.g., 'ingredients-editor' or 'ingredients')
 */
export async function clearToastEditor(page: Page, editorId: string): Promise<void> {
  const textareaId = editorId.replace(/-editor$/, '');
  
  await page.evaluate((textareaId) => {
    const editor = (window as any).RecipeEditor?.get(textareaId);
    if (editor) {
      editor.value('');
    } else {
      throw new Error(`Editor not found: ${textareaId}`);
    }
  }, textareaId);
}
