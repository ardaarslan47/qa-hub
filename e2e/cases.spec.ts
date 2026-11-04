import { expect, test } from './fixtures';
import { caseList, createCase, createProject } from './steps';

test.beforeEach(async ({ page }) => {
  await createProject(page, 'Mobile');
});

test('Case oluşturma: başlık girilir, case listede görünür', async ({ page }) => {
  await createCase(page, 'Mobile', 'Login works');

  await expect(caseList(page)).toHaveText(['Login works']);
});

test('Boş başlık: case oluşturulamaz, hata mesajı görünür, listeye bir şey eklenmez', async ({ page }) => {
  await createCase(page, 'Mobile', 'Login works');
  await expect(caseList(page)).toHaveText(['Login works']);

  await createCase(page, 'Mobile', '');

  await expect(page.getByRole('alert')).toHaveText('Title is required.');
  await expect(caseList(page)).toHaveText(['Login works']);
});
