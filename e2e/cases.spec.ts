import { expect, test } from './fixtures';
import {
  caseList,
  casePriority,
  casesUsingPrecondition,
  createCase,
  createProject,
  deleteCase,
  editCase,
  enterStatus,
  executionList,
  executionRow,
  menuLink,
  openCase,
  openCaseDeletion,
  openLibrary,
  openProject,
  openRun,
  startRun,
} from './steps';

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

test('Case düzenleme: Library den case sayfası açılır, alanlar düzenlenir, yeni değerler Library de görünür', async ({ page }) => {
  await createCase(page, 'Login works', { description: 'Sign in with a valid email and password.', priority: 'Low' });

  await openCase(page, 'Login works');
  await expect(page.getByLabel('Title')).toHaveValue('Login works');
  await expect(page.getByLabel('Description')).toHaveValue('Sign in with a valid email and password.');
  await expect(page.getByLabel('Priority')).toHaveValue('Low');

  await page.getByLabel('Title').fill('Sign in works');
  await page.getByLabel('Description').fill('Sign in with single sign-on.');
  await page.getByLabel('Priority').selectOption('High');
  await page.getByRole('button', { name: 'Save case' }).click();

  await expect(page.getByRole('heading', { name: 'Library' })).toBeVisible();
  await expect(caseList(page)).toHaveText(['Sign in works']);
  await expect(casePriority(page, 'Sign in works')).toHaveText('High');
  await openCase(page, 'Sign in works');
  await expect(page.getByLabel('Description')).toHaveValue('Sign in with single sign-on.');
});

test('Düzenleme başlamış run ı etkilemez: eski run eski değerleri, yeni run yeni değerleri gösterir', async ({ page }) => {
  await createProject(page, 'Mobile');
  await createCase(page, 'Login works', { description: 'Sign in with a valid email and password.', priority: 'Low' });
  await startRun(page, 'Mobile', 'Sprint 1');

  await editCase(page, 'Login works', {
    title: 'Sign in works',
    description: 'Sign in with single sign-on.',
    priority: 'High',
  });
  await startRun(page, 'Mobile', 'Sprint 2');

  await openRun(page, 'Mobile', 'Sprint 1');
  await expect(executionList(page)).toHaveText(['Login works Untested']);
  const old = executionRow(page, 'Login works');
  await expect(old.getByTestId('execution-priority')).toHaveText('Low');
  await expect(old.getByTestId('execution-description')).toHaveText('Sign in with a valid email and password.');

  await openRun(page, 'Mobile', 'Sprint 2');
  await expect(executionList(page)).toHaveText(['Sign in works Untested']);
  const fresh = executionRow(page, 'Sign in works');
  await expect(fresh.getByTestId('execution-priority')).toHaveText('High');
  await expect(fresh.getByTestId('execution-description')).toHaveText('Sign in with single sign-on.');
});

test('Case silme: case sayfasından onay adımına gidilir, onaydan sonra case Library de görünmez', async ({ page }) => {
  await createCase(page, 'Login works');
  await createCase(page, 'Logout works');
  await openCaseDeletion(page, 'Login works');

  await page.getByRole('button', { name: 'Delete case' }).click();

  await expect(page.getByRole('heading', { name: 'Library' })).toBeVisible();
  await expect(caseList(page)).toHaveText(['Logout works']);
});

test('Silme onayı: case i Precondition olarak kullanan case ler listelenir, Cancel case i silmez', async ({ page }) => {
  await createCase(page, 'Login works');
  await createCase(page, 'Checkout works', { preconditions: ['Login works'] });
  await openCaseDeletion(page, 'Login works');
  await expect(casesUsingPrecondition(page)).toHaveText(['Checkout works']);

  await page.getByRole('link', { name: 'Cancel' }).click();

  await expect(page.getByLabel('Title')).toHaveValue('Login works');
  await openLibrary(page);
  await expect(caseList(page)).toHaveText(['Login works', 'Checkout works']);
});

test('Silinen case in execution ı: Run ekranında kopyalarıyla görünür, status girilir', async ({ page }) => {
  await createProject(page, 'Mobile');
  await createCase(page, 'Login works');
  await createCase(page, 'Checkout works', {
    description: 'Pay with a card.',
    priority: 'High',
    preconditions: ['Login works'],
  });
  await startRun(page, 'Mobile', 'Sprint 1');

  await deleteCase(page, 'Checkout works');
  await deleteCase(page, 'Login works');

  await openRun(page, 'Mobile', 'Sprint 1');
  await expect(executionList(page)).toHaveText(['Login works Untested', 'Checkout works Untested']);
  const deleted = executionRow(page, 'Checkout works');
  await expect(deleted.getByTestId('execution-priority')).toHaveText('High');
  await expect(deleted.getByTestId('execution-description')).toHaveText('Pay with a card.');
  await expect(deleted.getByTestId('execution-precondition')).toHaveText(['Login works']);

  await enterStatus(page, 'Checkout works', 'Passed');

  await expect(executionList(page)).toHaveText(['Login works Untested', 'Checkout works Passed']);
  await page.reload();
  await expect(executionList(page)).toHaveText(['Login works Untested', 'Checkout works Passed']);
});
