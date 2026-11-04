import type { Page } from '@playwright/test';

export async function createProject(page: Page, name: string) {
  await page.goto('/projects');
  await page.getByLabel('Name').fill(name);
  await page.getByRole('button', { name: 'Create project' }).click();
}

export function projectList(page: Page) {
  return page.getByRole('list', { name: 'Projects' }).getByRole('listitem');
}

export async function openProject(page: Page, name: string) {
  await page.goto('/projects');
  await projectList(page).getByRole('link', { name, exact: true }).click();
}

export function projectSection(page: Page, name: string) {
  return page.getByRole('navigation', { name: 'Project' }).getByRole('link', { name, exact: true });
}

export async function openRuns(page: Page, project: string) {
  await openProject(page, project);
  await projectSection(page, 'Runs').click();
}

export function menuLink(page: Page, name: string) {
  return page.getByRole('navigation', { name: 'Main' }).getByRole('link', { name, exact: true });
}

export async function openLibrary(page: Page) {
  await page.goto('/projects');
  await menuLink(page, 'Library').click();
}

export async function createCase(
  page: Page,
  title: string,
  fields: { description?: string; priority?: 'Low' | 'Medium' | 'High' } = {},
) {
  await openLibrary(page);
  await page.getByLabel('Title').fill(title);
  if (fields.description !== undefined) await page.getByLabel('Description').fill(fields.description);
  if (fields.priority !== undefined) await page.getByLabel('Priority').selectOption(fields.priority);
  await page.getByRole('button', { name: 'Create case' }).click();
}

function caseRows(page: Page) {
  return page.getByRole('list', { name: 'Cases' }).getByRole('listitem');
}

export async function openCase(page: Page, title: string) {
  await openLibrary(page);
  await caseRows(page).getByRole('link', { name: title, exact: true }).click();
}

export async function editCase(
  page: Page,
  title: string,
  fields: { title?: string; description?: string; priority?: 'Low' | 'Medium' | 'High' },
) {
  await openCase(page, title);
  if (fields.title !== undefined) await page.getByLabel('Title').fill(fields.title);
  if (fields.description !== undefined) await page.getByLabel('Description').fill(fields.description);
  if (fields.priority !== undefined) await page.getByLabel('Priority').selectOption(fields.priority);
  await page.getByRole('button', { name: 'Save case' }).click();
}

export function caseList(page: Page) {
  return caseRows(page).getByTestId('case-title');
}

export function casePriority(page: Page, title: string) {
  return caseRows(page).filter({ hasText: title }).getByTestId('case-priority');
}

export async function startRun(page: Page, project: string, name: string) {
  await openRuns(page, project);
  await page.getByLabel('Name').fill(name);
  await page.getByRole('button', { name: 'Start run' }).click();
}

export function runList(page: Page) {
  return page.getByRole('list', { name: 'Runs' }).getByRole('listitem');
}

export async function openRun(page: Page, project: string, name: string) {
  await openRuns(page, project);
  await runList(page).getByRole('link', { name, exact: true }).click();
}

function executionRows(page: Page) {
  return page.getByRole('list', { name: 'Executions' }).getByRole('listitem');
}

export function executionList(page: Page) {
  return executionRows(page).getByTestId('execution');
}

export function executionRow(page: Page, title: string) {
  return executionRows(page).filter({ hasText: title });
}

export async function enterStatus(page: Page, title: string, status: 'Passed' | 'Failed') {
  await executionRow(page, title).getByRole('button', { name: status }).click();
}
