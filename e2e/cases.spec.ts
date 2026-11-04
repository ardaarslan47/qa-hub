import { expect, test } from './fixtures';
import { caseList, createCase, createProject, menuLink, openLibrary, openProject } from './steps';

test('Case oluşturma: üst menüden Library ye gidilir, başlık girilir, case listede görünür', async ({ page }) => {
  await openLibrary(page);
  await expect(page.getByRole('heading', { name: 'Library' })).toBeVisible();

  await page.getByLabel('Title').fill('Login works');
  await page.getByRole('button', { name: 'Create case' }).click();

  await expect(caseList(page)).toHaveText(['Login works']);
});

test('Boş başlık: case oluşturulamaz, hata mesajı görünür, listeye bir şey eklenmez', async ({ page }) => {
  await createCase(page, 'Login works');
  await expect(caseList(page)).toHaveText(['Login works']);

  await createCase(page, '');

  await expect(page.getByRole('alert')).toHaveText('Title is required.');
  await expect(caseList(page)).toHaveText(['Login works']);
});

test('Library project e bağlı değildir: project in içinden de menüden açılır ve aynı case leri gösterir', async ({ page }) => {
  await createCase(page, 'Login works');
  await createProject(page, 'Mobile');
  await openProject(page, 'Mobile');

  await menuLink(page, 'Library').click();

  await expect(page.getByRole('heading', { name: 'Library' })).toBeVisible();
  await expect(caseList(page)).toHaveText(['Login works']);
});
