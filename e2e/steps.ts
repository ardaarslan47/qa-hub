import type { Page } from '@playwright/test';

export async function createCase(page: Page, title: string) {
  await page.goto('/cases');
  await page.getByLabel('Title').fill(title);
  await page.getByRole('button', { name: 'Create case' }).click();
}

export async function startRun(page: Page, name: string) {
  await page.goto('/runs');
  await page.getByLabel('Name').fill(name);
  await page.getByRole('button', { name: 'Start run' }).click();
}
