import { expect, test } from './fixtures';
import {
  createCase,
  createProject,
  executionList,
  menuLink,
  openProject,
  openRun,
  openRuns,
  projectList,
  projectSection,
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

test('Project ana sayfası: Project gezintisinde Runs görünür Cases görünmez, Runs sayfayı açar, menüdeki project adı ana sayfaya döner', async ({ page }) => {
  await createProject(page, 'Mobile');
  await openProject(page, 'Mobile');

  await expect(projectSection(page, 'Runs')).toBeVisible();
  await expect(page.getByRole('link', { name: 'Cases', exact: true })).toHaveCount(0);

  await projectSection(page, 'Runs').click();
  await expect(page.getByRole('heading', { name: 'Runs' })).toBeVisible();

  await menuLink(page, 'Mobile').click();
  await expect(page.getByRole('heading', { name: 'Mobile' })).toBeVisible();
});

test('İzolasyon: iki project te başlatılan run lar birbirinin listesinde görünmez', async ({ page }) => {
  await createProject(page, 'Mobile');
  await createProject(page, 'Web');
  await createCase(page, 'Login works');
  await startRun(page, 'Mobile', 'Sprint 1');
  await startRun(page, 'Web', 'Sprint 2');

  await openRuns(page, 'Mobile');
  await expect(runList(page)).toHaveText(['Sprint 1']);

  await openRuns(page, 'Web');
  await expect(runList(page)).toHaveText(['Sprint 2']);
});

test('Library nin her case i iki project in run ında da execution olur', async ({ page }) => {
  await createProject(page, 'Mobile');
  await createProject(page, 'Web');
  await createCase(page, 'Login works');
  await createCase(page, 'Checkout works');

  await startRun(page, 'Mobile', 'Sprint 1');
  await startRun(page, 'Web', 'Sprint 2');

  await openRun(page, 'Mobile', 'Sprint 1');
  await expect(executionList(page)).toHaveText(['Login works Untested', 'Checkout works Untested']);
  await openRun(page, 'Web', 'Sprint 2');
  await expect(executionList(page)).toHaveText(['Login works Untested', 'Checkout works Untested']);
});

test('Başka project in URL si altından açılan run 404 döner', async ({ page }) => {
  await createProject(page, 'Mobile');
  await createProject(page, 'Web');
  await createCase(page, 'Login works');
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
