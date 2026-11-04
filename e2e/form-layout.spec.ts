import type { Locator, Page } from '@playwright/test';
import { expect, test } from './fixtures';
import { createCase, createProject, openCase, openLibrary, openRuns } from './steps';

const viewports = [
  { name: 'masaüstü', size: { width: 1280, height: 720 } },
  { name: 'mobil', size: { width: 375, height: 812 } },
];

async function boxOf(locator: Locator) {
  const box = await locator.boundingBox();
  expect(box, 'element is rendered').not.toBeNull();
  return box!;
}

async function expectCaseFormLayout(page: Page, submit: string) {
  const form = await boxOf(page.locator('main > form'));
  for (const label of ['Title', 'Description', 'Priority']) {
    const field = await boxOf(page.getByLabel(label));
    expect(field.width, `${label} fills the form`).toBeGreaterThanOrEqual(form.width - 1);
  }
  const priority = await boxOf(page.getByLabel('Priority'));
  const button = await boxOf(page.getByRole('button', { name: submit }));
  expect(button.y, `${submit} sits below Priority`).toBeGreaterThanOrEqual(priority.y + priority.height);
}

async function expectInlineSubmit(page: Page, label: string, submit: string) {
  const field = await boxOf(page.getByLabel(label));
  const button = await boxOf(page.getByRole('button', { name: submit }));
  expect(button.x, `${submit} sits right of ${label}`).toBeGreaterThanOrEqual(field.x + field.width);
  expect(button.y, `${submit} shares the row of ${label}`).toBeLessThan(field.y + field.height);
}

for (const { name, size } of viewports) {
  test.describe(`Form yerleşimi, ${name}`, () => {
    test.use({ viewport: size });

    test('Library formu: alanlar form genişliğini doldurur, Create case Priority satırının altında durur', async ({ page }) => {
      await openLibrary(page);

      await expectCaseFormLayout(page, 'Create case');
    });

    test('Case sayfası formu: alanlar form genişliğini doldurur, Save case Priority satırının altında durur', async ({ page }) => {
      await createCase(page, 'Login works');
      await openCase(page, 'Login works');

      await expectCaseFormLayout(page, 'Save case');
    });

    test('Tek alanlı formlar: düğme alanın yanında durur', async ({ page }) => {
      await createProject(page, 'Mobile');
      await expectInlineSubmit(page, 'Name', 'Create project');

      await openRuns(page, 'Mobile');
      await expectInlineSubmit(page, 'Name', 'Start run');
    });
  });
}
