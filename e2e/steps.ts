import { expect, type Page } from '@playwright/test';

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

export async function openRun(page: Page, name: string) {
  await page.goto('/runs');
  await page.getByRole('link', { name }).click();
}

function executionRows(page: Page) {
  return page.getByRole('list', { name: 'Executions' }).getByRole('listitem');
}

function executionRow(page: Page, title: string) {
  return executionRows(page).filter({ has: page.getByText(title, { exact: true }) });
}

export async function enterStatus(page: Page, title: string, status: 'Passed' | 'Failed') {
  await executionRow(page, title).getByRole('button', { name: status }).click();
}

export async function expectExecutions(page: Page, expected: [title: string, status: string][]) {
  await expect(executionRows(page)).toHaveCount(expected.length);
  for (const [title, status] of expected) {
    await expect(executionRow(page, title).getByRole('status')).toHaveText(status);
  }
}
