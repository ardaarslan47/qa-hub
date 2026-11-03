import { test, expect } from '../fixtures';
import { caseList, createCase } from '../actions';

test('1. Case creation: the title entered appears in the case list', async ({ page }) => {
  await createCase(page, 'Login with valid credentials');

  await expect(caseList(page)).toHaveText(['Login with valid credentials']);
});

test('2. Blank title: no case is created and an error message is shown', async ({ page }) => {
  await createCase(page, 'Existing case');

  await createCase(page, '');

  await expect(page.getByRole('alert')).toHaveText('Title is required.');
  await expect(caseList(page)).toHaveText(['Existing case']);
});
