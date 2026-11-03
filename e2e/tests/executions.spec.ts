import { test, expect } from '../fixtures';
import { createCase, executionRow, openRun, recordStatus, startRun, statusCell } from '../actions';

test('6. Status entry: Passed and Failed are recorded and still shown after a reload', async ({ page }) => {
  await createCase(page, 'Login with valid credentials');
  await createCase(page, 'Logout');
  await startRun(page, 'Sprint 1');
  await openRun(page, 'Sprint 1');

  await recordStatus(executionRow(page, 'Login with valid credentials'), 'Passed');
  await recordStatus(executionRow(page, 'Logout'), 'Failed');
  await page.reload();

  await expect(statusCell(executionRow(page, 'Login with valid credentials'), 'Passed')).toBeVisible();
  await expect(statusCell(executionRow(page, 'Logout'), 'Failed')).toBeVisible();
  await expect(page.getByRole('cell', { name: 'Untested', exact: true })).toHaveCount(0);
});

test('7. Overwrite: Passed entered over Failed replaces it', async ({ page }) => {
  await createCase(page, 'Login with valid credentials');
  await startRun(page, 'Sprint 1');
  await openRun(page, 'Sprint 1');
  const row = executionRow(page, 'Login with valid credentials');
  await recordStatus(row, 'Failed');
  await expect(statusCell(row, 'Failed')).toBeVisible();

  await recordStatus(row, 'Passed');

  await expect(statusCell(row, 'Passed')).toBeVisible();
  await expect(statusCell(row, 'Failed')).toHaveCount(0);
});
