import { expect, type Locator, type Page } from '@playwright/test';

export async function createCase(page: Page, title: string): Promise<void> {
  await page.goto('/cases');
  await page.getByLabel('Title').fill(title);
  await page.getByRole('button', { name: 'Create case' }).click();
}

export async function startRun(page: Page, name: string): Promise<void> {
  await page.goto('/runs');
  await page.getByLabel('Name').fill(name);
  await page.getByRole('button', { name: 'Start run' }).click();
}

export async function openRun(page: Page, name: string): Promise<void> {
  await page.goto('/runs');
  await page.getByRole('list', { name: 'Runs' }).getByRole('link', { name, exact: true }).click();
  await expect(page.getByRole('heading', { level: 1, name, exact: true })).toBeVisible();
}

export function caseList(page: Page): Locator {
  return page.getByRole('list', { name: 'Cases' }).getByRole('listitem');
}

export function runList(page: Page): Locator {
  return page.getByRole('list', { name: 'Runs' }).getByRole('listitem');
}

export function executionRows(page: Page): Locator {
  return page
    .getByRole('table', { name: 'Executions' })
    .getByRole('row')
    .filter({ has: page.getByRole('cell') });
}

export function executionRow(page: Page, title: string): Locator {
  return executionRows(page).filter({ has: page.getByRole('cell', { name: title, exact: true }) });
}

export function statusCell(row: Locator, status: string): Locator {
  return row.getByRole('cell', { name: status, exact: true });
}
