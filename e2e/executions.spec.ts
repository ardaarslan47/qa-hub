import { expect, test } from './fixtures';
import { createCase, enterStatus, executionList, openRun, startRun } from './steps';

test('Status girme: bir execution a Passed, diğerine Failed girilir, sayfa yenilenince de görünür', async ({ page }) => {
  await createCase(page, 'Login works');
  await createCase(page, 'Logout works');
  await startRun(page, 'Sprint 1');
  await openRun(page, 'Sprint 1');

  await enterStatus(page, 'Login works', 'Passed');
  await expect(executionList(page)).toHaveText(['Login works Passed', 'Logout works Untested']);
  await enterStatus(page, 'Logout works', 'Failed');
  await expect(executionList(page)).toHaveText(['Login works Passed', 'Logout works Failed']);

  await page.reload();

  await expect(executionList(page)).toHaveText(['Login works Passed', 'Logout works Failed']);
});

test('Üzerine yazma: Failed girilmiş execution a Passed girilir, status u Passed olur', async ({ page }) => {
  await createCase(page, 'Login works');
  await createCase(page, 'Logout works');
  await startRun(page, 'Sprint 1');
  await openRun(page, 'Sprint 1');
  await enterStatus(page, 'Login works', 'Failed');
  await expect(executionList(page)).toHaveText(['Login works Failed', 'Logout works Untested']);

  await enterStatus(page, 'Login works', 'Passed');

  await expect(executionList(page)).toHaveText(['Login works Passed', 'Logout works Untested']);
});
