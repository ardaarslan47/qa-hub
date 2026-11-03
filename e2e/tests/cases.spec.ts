import { test, expect } from '../fixtures';
import { createCase } from '../actions';

test('1. Case creation: the title entered appears in the case list', async ({ page }) => {
  await createCase(page, 'Login with valid credentials');

  await expect(page.getByRole('list', { name: 'Cases' }).getByRole('listitem')).toHaveText([
    'Login with valid credentials',
  ]);
});

test('2. Blank title: no case is created and an error message is shown', async ({ page }) => {
  await createCase(page, 'Existing case');

  await createCase(page, '');

  await expect(page.getByRole('alert')).toHaveText('Title is required.');
  await expect(page.getByRole('list', { name: 'Cases' }).getByRole('listitem')).toHaveText(['Existing case']);
});
