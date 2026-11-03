import { test, expect } from './fixtures';

test('creating a case shows it in the list', async ({ page }) => {
  await page.goto('/cases');

  await page.getByLabel('Title').fill('Login works');
  await page.getByRole('button', { name: 'Create case' }).click();

  await expect(page.getByRole('list', { name: 'Cases' }).getByRole('listitem')).toHaveText(['Login works']);
});

test('a blank title is rejected with an error message and no case is added', async ({ page }) => {
  await page.goto('/cases');

  await page.getByRole('button', { name: 'Create case' }).click();

  await expect(page.getByRole('alert')).toHaveText('Title is required');
  await expect(page.getByRole('listitem')).toHaveCount(0);
});
