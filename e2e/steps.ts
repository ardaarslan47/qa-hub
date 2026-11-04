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

function projectSection(page: Page, name: 'Cases' | 'Runs') {
  return page.getByRole('navigation', { name: 'Project' }).getByRole('link', { name, exact: true });
}

export async function openCases(page: Page, project: string) {
  await openProject(page, project);
  await projectSection(page, 'Cases').click();
}

export async function openRuns(page: Page, project: string) {
  await openProject(page, project);
  await projectSection(page, 'Runs').click();
}

export function menuLink(page: Page, name: string) {
  return page.getByRole('navigation', { name: 'Main' }).getByRole('link', { name, exact: true });
}

export async function createCase(page: Page, project: string, title: string) {
  await openCases(page, project);
  await page.getByLabel('Title').fill(title);
  await page.getByRole('button', { name: 'Create case' }).click();
}

export function caseList(page: Page) {
  return page.getByRole('list', { name: 'Cases' }).getByRole('listitem');
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

export async function enterStatus(page: Page, title: string, status: 'Passed' | 'Failed') {
  await executionRows(page).filter({ hasText: title }).getByRole('button', { name: status }).click();
}
