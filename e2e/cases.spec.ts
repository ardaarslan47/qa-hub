import { expect, test } from './fixtures';
import { caseList, casePriority, createCase, createProject, menuLink, openLibrary, openProject } from './steps';

test('Case oluşturma: üst menüden Library ye gidilir, başlık girilir, case listede görünür', async ({ page }) => {
  await openLibrary(page);
  await expect(page.getByRole('heading', { name: 'Library' })).toBeVisible();

  await page.getByLabel('Title').fill('Login works');
  await page.getByRole('button', { name: 'Create case' }).click();

  await expect(caseList(page)).toHaveText(['Login works']);
});

test('Description ve Priority: case Description ve High Priority ile oluşturulur, Priority listede görünür', async ({ page }) => {
  await openLibrary(page);

  await page.getByLabel('Title').fill('Login works');
  await page.getByLabel('Description').fill('Sign in with a valid email and password.');
  await page.getByLabel('Priority').selectOption('High');
  await page.getByRole('button', { name: 'Create case' }).click();

  await expect(caseList(page)).toHaveText(['Login works']);
  await expect(casePriority(page, 'Login works')).toHaveText('High');
});

test('Varsayılan Priority: formda Medium seçili gelir, Description ve Priority girilmeden oluşan case Medium olur', async ({ page }) => {
  await openLibrary(page);
  await expect(page.getByLabel('Priority')).toHaveValue('Medium');

  await page.getByLabel('Title').fill('Login works');
  await page.getByRole('button', { name: 'Create case' }).click();

  await expect(caseList(page)).toHaveText(['Login works']);
  await expect(casePriority(page, 'Login works')).toHaveText('Medium');
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
