import type { Page } from '@playwright/test';

export async function createCase(page: Page, title: string): Promise<void> {
  await page.goto('/cases');
  await page.getByLabel('Title').fill(title);
  await page.getByRole('button', { name: 'Create case' }).click();
}
