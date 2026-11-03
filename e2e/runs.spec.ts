import { test, expect } from './fixtures';
import { createCase, startRun } from './steps';

test('starting a run shows it in the list and its page has an Untested execution per case', async ({ page }) => {
  await createCase(page, 'Login works');
  await createCase(page, 'Logout works');

  await page.getByRole('link', { name: 'Runs' }).click();
  await page.getByLabel('Name').fill('Sprint 1');
  await page.getByRole('button', { name: 'Start run' }).click();

  const runs = page.getByRole('list', { name: 'Runs' }).getByRole('listitem');
  await expect(runs).toHaveText(['Sprint 1']);

  await runs.getByRole('link', { name: 'Sprint 1' }).click();

  await expect(page.getByRole('heading', { name: 'Sprint 1' })).toBeVisible();
  await expect(page.getByRole('list', { name: 'Executions' }).getByRole('listitem')).toHaveText([
    'Login works Untested',
    'Logout works Untested',
  ]);
});

test('a blank run name is rejected with an error message and no run is added', async ({ page }) => {
  await createCase(page, 'Login works');
  await startRun(page, '');

  await expect(page.getByRole('alert')).toHaveText('Name is required');
  await expect(page.getByRole('listitem')).toHaveCount(0);
});

test('a run cannot be started while there are no cases', async ({ page }) => {
  await startRun(page, 'Sprint 1');

  await expect(page.getByRole('alert')).toHaveText('Create at least one case before starting a run');
  await expect(page.getByRole('listitem')).toHaveCount(0);
});

test('a case created after a run started is not in that run but is in a new run', async ({ page }) => {
  await createCase(page, 'Login works');
  await startRun(page, 'Sprint 1');
  await createCase(page, 'Logout works');
  await startRun(page, 'Sprint 2');

  const executions = page.getByRole('list', { name: 'Executions' }).getByRole('listitem');

  await page.getByRole('link', { name: 'Sprint 1' }).click();
  await expect(executions).toHaveText(['Login works Untested']);

  await page.goto('/runs');
  await page.getByRole('link', { name: 'Sprint 2' }).click();
  await expect(executions).toHaveText(['Login works Untested', 'Logout works Untested']);
});
