import { expect, test } from './fixtures';
import {
  casePreconditions,
  createCase,
  createProject,
  editCase,
  enterStatus,
  executionList,
  executionRow,
  openCase,
  openRun,
  startRun,
} from './steps';

test('Precondition ekleme: case e iki Precondition eklenir, ikisi de case sayfasında görünür', async ({ page }) => {
  await createCase(page, 'Login works');
  await createCase(page, 'Profile loads');
  await createCase(page, 'Checkout works');

  await editCase(page, 'Checkout works', { preconditions: ['Login works', 'Profile loads'] });

  await openCase(page, 'Checkout works');
  await expect(casePreconditions(page)).toHaveText(['Login works', 'Profile loads']);
});

test('Case oluştururken Precondition: seçilen case ler yeni case in sayfasında görünür', async ({ page }) => {
  await createCase(page, 'Login works');

  await createCase(page, 'Checkout works', { preconditions: ['Login works'] });

  await openCase(page, 'Checkout works');
  await expect(casePreconditions(page)).toHaveText(['Login works']);
});

test('Zincir: A→B→C de A nın sayfasında B görünür, C görünmez', async ({ page }) => {
  await createCase(page, 'Case C');
  await createCase(page, 'Case B', { preconditions: ['Case C'] });
  await createCase(page, 'Case A', { preconditions: ['Case B'] });

  await openCase(page, 'Case A');

  await expect(casePreconditions(page)).toHaveText(['Case B']);
  await expect(page.getByText('Case C')).toBeHidden();
});

test('Döngü: A→B varken B ye A seçilirse hata görünür, seçim korunur, bağ oluşmaz', async ({ page }) => {
  await createCase(page, 'Case B');
  await createCase(page, 'Case A', { preconditions: ['Case B'] });

  await editCase(page, 'Case B', { preconditions: ['Case A'] });

  await expect(page.getByRole('alert')).toHaveText(
    'A case cannot be its own Precondition, directly or through other cases.',
  );
  await expect(page.getByLabel('Case A', { exact: true })).toBeVisible();
  await expect(page.getByLabel('Case A', { exact: true })).toBeChecked();
  await openCase(page, 'Case B');
  await expect(casePreconditions(page)).toHaveCount(0);
});

test('Run ekranı: execution ın Precondition başlıkları görünür, Precondition Failed olsa bile Passed girilir', async ({ page }) => {
  await createProject(page, 'Mobile');
  await createCase(page, 'Login works');
  await createCase(page, 'Profile loads');
  await createCase(page, 'Checkout works', { preconditions: ['Login works', 'Profile loads'] });
  await startRun(page, 'Mobile', 'Sprint 1');
  await openRun(page, 'Mobile', 'Sprint 1');

  await expect(executionRow(page, 'Checkout works').getByTestId('execution-precondition')).toHaveText([
    'Login works',
    'Profile loads',
  ]);

  await enterStatus(page, 'Login works', 'Failed');
  await expect(executionList(page)).toHaveText([
    'Login works Failed',
    'Profile loads Untested',
    'Checkout works Untested',
  ]);
  await enterStatus(page, 'Checkout works', 'Passed');

  await expect(executionList(page)).toHaveText([
    'Login works Failed',
    'Profile loads Untested',
    'Checkout works Passed',
  ]);
});

test('Precondition düzenleme başlamış run ı etkilemez: eski run eski başlığı gösterir', async ({ page }) => {
  await createProject(page, 'Mobile');
  await createCase(page, 'Login works');
  await createCase(page, 'Checkout works', { preconditions: ['Login works'] });
  await startRun(page, 'Mobile', 'Sprint 1');

  await editCase(page, 'Login works', { title: 'Sign in works' });

  await openRun(page, 'Mobile', 'Sprint 1');
  await expect(executionRow(page, 'Checkout works').getByTestId('execution-precondition')).toHaveText(['Login works']);
});
