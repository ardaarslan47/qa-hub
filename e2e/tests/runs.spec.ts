import { test, expect } from '../fixtures';
import { createCase, executionRow, executionRows, openRun, runList, startRun, statusCell } from '../actions';

test('3. Run start: a run over two cases is listed and has two Untested executions', async ({ page }) => {
  await createCase(page, 'Login with valid credentials');
  await createCase(page, 'Logout');

  await startRun(page, 'Sprint 1');

  await expect(runList(page)).toHaveText(['Sprint 1']);
  await openRun(page, 'Sprint 1');
  await expect(executionRows(page)).toHaveCount(2);
  await expect(statusCell(executionRow(page, 'Login with valid credentials'), 'Untested')).toBeVisible();
  await expect(statusCell(executionRow(page, 'Logout'), 'Untested')).toBeVisible();
});

test('4. Blank name: no run is started and an error message is shown', async ({ page }) => {
  await createCase(page, 'Login with valid credentials');

  await startRun(page, '');

  await expect(page.getByRole('alert')).toHaveText('Name is required.');
  await expect(runList(page)).toHaveCount(0);
});

test('5. No cases: no run is started and an error message is shown', async ({ page }) => {
  await startRun(page, 'Sprint 1');

  await expect(page.getByRole('alert')).toHaveText('Create a case before starting a run.');
  await expect(runList(page)).toHaveCount(0);
});

test('8. Snapshot: a case created after a run started is only in runs started later', async ({ page }) => {
  await createCase(page, 'Login with valid credentials');
  await startRun(page, 'Sprint 1');
  await createCase(page, 'Logout');
  await startRun(page, 'Sprint 2');

  await openRun(page, 'Sprint 1');
  await expect(executionRows(page)).toHaveCount(1);
  await expect(executionRow(page, 'Login with valid credentials')).toBeVisible();

  await openRun(page, 'Sprint 2');
  await expect(executionRows(page)).toHaveCount(2);
  await expect(executionRow(page, 'Login with valid credentials')).toBeVisible();
  await expect(executionRow(page, 'Logout')).toBeVisible();
});
