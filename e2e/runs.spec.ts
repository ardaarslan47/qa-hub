import { expect, test } from './fixtures';
import { createCase, createProject, executionList, executionRow, openRun, runList, startRun } from './steps';

test.beforeEach(async ({ page }) => {
  await createProject(page, 'Mobile');
});

test('Run başlatma: iki case varken run başlatılır, run listesinde görünür, sayfasında iki Untested execution vardır', async ({ page }) => {
  await createCase(page, 'Login works');
  await createCase(page, 'Logout works');

  await startRun(page, 'Mobile', 'Sprint 1');

  await expect(runList(page)).toHaveText(['Sprint 1']);
  await runList(page).getByRole('link', { name: 'Sprint 1' }).click();
  await expect(page.getByRole('heading', { name: 'Sprint 1' })).toBeVisible();
  await expect(executionList(page)).toHaveText(['Login works Untested', 'Logout works Untested']);
});

test('Execution kopyası: Run ekranında her execution ın Description ve Priority değeri görünür', async ({ page }) => {
  await createCase(page, 'Login works', { description: 'Sign in with a valid email and password.', priority: 'High' });
  await createCase(page, 'Logout works', { priority: 'Low' });

  await startRun(page, 'Mobile', 'Sprint 1');
  await openRun(page, 'Mobile', 'Sprint 1');

  const login = executionRow(page, 'Login works');
  await expect(login.getByTestId('execution-priority')).toHaveText('High');
  await expect(login.getByTestId('execution-description')).toHaveText('Sign in with a valid email and password.');
  const logout = executionRow(page, 'Logout works');
  await expect(logout.getByTestId('execution-priority')).toHaveText('Low');
  await expect(logout.getByTestId('execution-description')).toHaveCount(0);
});

test('Boş ad: run başlatılamaz, hata mesajı görünür', async ({ page }) => {
  await createCase(page, 'Login works');

  await startRun(page, 'Mobile', '');

  await expect(page.getByRole('alert')).toHaveText('Name is required.');
  await expect(runList(page)).toHaveCount(0);
});

test('Case yokken run: run başlatılamaz, hata mesajı görünür', async ({ page }) => {
  await startRun(page, 'Mobile', 'Sprint 1');

  await expect(page.getByRole('alert')).toHaveText('Create a case before starting a run.');
  await expect(runList(page)).toHaveCount(0);
});

test('Snapshot: run başladıktan sonra oluşturulan case o run da görünmez, yeni run da görünür', async ({ page }) => {
  await createCase(page, 'Login works');
  await createCase(page, 'Logout works');
  await startRun(page, 'Mobile', 'Sprint 1');
  await createCase(page, 'Reset password works');
  await startRun(page, 'Mobile', 'Sprint 2');

  await openRun(page, 'Mobile', 'Sprint 1');
  await expect(executionList(page)).toHaveText(['Login works Untested', 'Logout works Untested']);

  await openRun(page, 'Mobile', 'Sprint 2');
  await expect(executionList(page)).toHaveText([
    'Login works Untested',
    'Logout works Untested',
    'Reset password works Untested',
  ]);
});
