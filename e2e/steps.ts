import type { Page } from '@playwright/test';

export async function createCase(page: Page, title: string) {
  await page.goto('/cases');
  await page.getByLabel('Title').fill(title);
  await page.getByRole('button', { name: 'Create case' }).click();
}

export function caseList(page: Page) {
  return page.getByRole('list', { name: 'Cases' }).getByRole('listitem');
}

export async function startRun(page: Page, name: string) {
  await page.goto('/runs');
  await page.getByLabel('Name').fill(name);
  await page.getByRole('button', { name: 'Start run' }).click();
}

export function runList(page: Page) {
  return page.getByRole('list', { name: 'Runs' }).getByRole('listitem');
}

export async function openRun(page: Page, name: string) {
  await page.goto('/runs');
  await page.getByRole('link', { name, exact: true }).click();
}

export function executionList(page: Page) {
  return page.getByRole('list', { name: 'Executions' }).getByRole('listitem');
}
