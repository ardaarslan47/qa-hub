import type { Page } from '@playwright/test';

export async function createProject(page: Page, name: string) {
  await page.goto('/projects');
  await page.getByLabel('Name').fill(name);
  await page.getByRole('button', { name: 'Create project' }).click();
}

export function projectList(page: Page) {
  return page.getByRole('list', { name: 'Projects' }).getByRole('listitem');
}

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

function executionRows(page: Page) {
  return page.getByRole('list', { name: 'Executions' }).getByRole('listitem');
}

export function executionList(page: Page) {
  return executionRows(page).getByTestId('execution');
}

export async function enterStatus(page: Page, title: string, status: 'Passed' | 'Failed') {
  await executionRows(page).filter({ hasText: title }).getByRole('button', { name: status }).click();
}
