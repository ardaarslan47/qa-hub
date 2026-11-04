import { expect, test } from './fixtures';
import {
  caseList,
  createCase,
  createProject,
  executionList,
  menuLink,
  openCases,
  openProject,
  openRun,
  openRuns,
  projectList,
  runList,
  startRun,
} from './steps';

test('Project oluşturma: ad girilir, project listede görünür, tıklanınca ana sayfasında adı görünür', async ({ page }) => {
  await createProject(page, 'Mobile');

  await expect(projectList(page)).toHaveText(['Mobile']);
  await projectList(page).getByRole('link', { name: 'Mobile' }).click();
  await expect(page.getByRole('heading', { name: 'Mobile' })).toBeVisible();
});

test('Boş ad: project oluşturulamaz, hata mesajı görünür, listeye bir şey eklenmez', async ({ page }) => {
  await createProject(page, 'Mobile');
  await expect(projectList(page)).toHaveText(['Mobile']);

  await createProject(page, '');

  await expect(page.getByRole('alert')).toHaveText('Name is required.');
  await expect(projectList(page)).toHaveText(['Mobile']);
});

test('Aynı ad: Mobile varken mobile oluşturulamaz, hata mesajı görünür, listeye bir şey eklenmez', async ({ page }) => {
  await createProject(page, 'Mobile');
  await expect(projectList(page)).toHaveText(['Mobile']);

  await createProject(page, 'mobile');

  await expect(page.getByRole('alert')).toHaveText('A project with this name already exists.');
  await expect(projectList(page)).toHaveText(['Mobile']);
});

test('Project ana sayfası: Cases ve Runs linkleri ilgili sayfaları açar, menüdeki project adı ana sayfaya döner', async ({ page }) => {
  await createProject(page, 'Mobile');
  await openProject(page, 'Mobile');

  await page.getByRole('link', { name: 'Cases', exact: true }).click();
  await expect(page.getByRole('heading', { name: 'Cases' })).toBeVisible();

  await menuLink(page, 'Mobile').click();
  await expect(page.getByRole('heading', { name: 'Mobile' })).toBeVisible();

  await page.getByRole('link', { name: 'Runs', exact: true }).click();
  await expect(page.getByRole('heading', { name: 'Runs' })).toBeVisible();
});

test('İzolasyon: iki project te açılan case ler ve run lar birbirinin listesinde görünmez', async ({ page }) => {
  await createProject(page, 'Mobile');
  await createProject(page, 'Web');
  await createCase(page, 'Mobile', 'Login works');
  await createCase(page, 'Web', 'Checkout works');
  await startRun(page, 'Mobile', 'Sprint 1');
  await startRun(page, 'Web', 'Sprint 2');

  await openCases(page, 'Mobile');
  await expect(caseList(page)).toHaveText(['Login works']);
  await openRuns(page, 'Mobile');
  await expect(runList(page)).toHaveText(['Sprint 1']);

  await openCases(page, 'Web');
  await expect(caseList(page)).toHaveText(['Checkout works']);
  await openRuns(page, 'Web');
  await expect(runList(page)).toHaveText(['Sprint 2']);
});

test('Run sadece kendi project indeki case lerden execution üretir', async ({ page }) => {
  await createProject(page, 'Mobile');
  await createProject(page, 'Web');
  await createCase(page, 'Mobile', 'Login works');
  await createCase(page, 'Web', 'Checkout works');

  await startRun(page, 'Mobile', 'Sprint 1');

  await openRun(page, 'Mobile', 'Sprint 1');
  await expect(executionList(page)).toHaveText(['Login works Untested']);
});

test('Başka project in URL si altından açılan run 404 döner', async ({ page }) => {
  await createProject(page, 'Mobile');
  await createProject(page, 'Web');
  await createCase(page, 'Mobile', 'Login works');
  await startRun(page, 'Mobile', 'Sprint 1');

  await openRun(page, 'Mobile', 'Sprint 1');
  await expect(page.getByRole('heading', { name: 'Sprint 1' })).toBeVisible();
  const runPath = new URL(page.url()).pathname;
  await openProject(page, 'Web');
  await expect(page.getByRole('heading', { name: 'Web' })).toBeVisible();
  const webPath = new URL(page.url()).pathname;

  const response = await page.goto(runPath.replace(/^\/projects\/\d+/, webPath));

  expect(response?.status()).toBe(404);
});
