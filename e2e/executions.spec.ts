import { test } from './fixtures';
import { createCase, enterStatus, expectExecutions, openRun, startRun } from './steps';

test('statuses entered on two executions are shown and survive a page reload', async ({ page }) => {
  await createCase(page, 'Login works');
  await createCase(page, 'Logout works');
  await startRun(page, 'Sprint 1');
  await openRun(page, 'Sprint 1');

  await enterStatus(page, 'Login works', 'Passed');
  await enterStatus(page, 'Logout works', 'Failed');

  await expectExecutions(page, [
    ['Login works', 'Passed'],
    ['Logout works', 'Failed'],
  ]);

  await page.reload();

  await expectExecutions(page, [
    ['Login works', 'Passed'],
    ['Logout works', 'Failed'],
  ]);
});

test('entering Passed on a Failed execution overwrites its status', async ({ page }) => {
  await createCase(page, 'Login works');
  await startRun(page, 'Sprint 1');
  await openRun(page, 'Sprint 1');

  await enterStatus(page, 'Login works', 'Failed');
  await expectExecutions(page, [['Login works', 'Failed']]);

  await enterStatus(page, 'Login works', 'Passed');
  await expectExecutions(page, [['Login works', 'Passed']]);
});
