import type { Page } from '@playwright/test';
import { expect, test } from './fixtures';
import { appliedTheme, expectAppliedTheme, opposite, themeOfBackground, themeToggle, themes } from './theme';
import { createCase, createProject, openCase, openLibrary, openProject, openRun, startRun } from './steps';

async function expectToggleAtTopRight(page: Page) {
  await expect(themeToggle(page)).toBeVisible();
  const box = (await themeToggle(page).boundingBox())!;
  const viewport = page.viewportSize()!;
  expect(box.x, 'toggle is in the right half').toBeGreaterThan(viewport.width / 2);
  expect(box.y + box.height, 'toggle is in the top quarter').toBeLessThan(viewport.height / 4);
}

test('Düğme: her sayfada sağ üstte görünür', async ({ page }) => {
  await createProject(page, 'Mobile');
  await createCase(page, 'Login works');
  await startRun(page, 'Mobile', 'Sprint 1');

  await page.goto('/projects');
  await expectToggleAtTopRight(page);
  await openLibrary(page);
  await expectToggleAtTopRight(page);
  await openCase(page, 'Login works');
  await expectToggleAtTopRight(page);
  await openProject(page, 'Mobile');
  await expectToggleAtTopRight(page);
  await openRun(page, 'Mobile', 'Sprint 1');
  await expectToggleAtTopRight(page);
});

for (const system of themes) {
  test.describe(`Sistem teması ${system}`, () => {
    test.use({ colorScheme: system });

    test('Tercih yokken sistem teması uygulanır', async ({ page }) => {
      await page.goto('/projects');

      await expectAppliedTheme(page, system);
    });

    test('Düğme temayı tersine çevirir, ikinci basış geri alır', async ({ page }) => {
      await page.goto('/projects');

      await themeToggle(page).click();
      await expectAppliedTheme(page, opposite(system));

      await themeToggle(page).click();
      await expectAppliedTheme(page, system);
    });

    test('Seçim yenilemeden ve sayfa geçişinden sonra korunur', async ({ page }) => {
      await page.goto('/projects');
      await themeToggle(page).click();
      await expectAppliedTheme(page, opposite(system));

      await page.reload();
      await expectAppliedTheme(page, opposite(system));

      await openLibrary(page);
      await expectAppliedTheme(page, opposite(system));

      await page.goBack();
      await expectAppliedTheme(page, opposite(system));
    });

    test('Sayfa yüklenirken yanlış tema hiç görünmez: seçim, gövde oluşmadan önce uygulanır', async ({ page }) => {
      await page.goto('/projects');
      await themeToggle(page).click();
      await expectAppliedTheme(page, opposite(system));

      await page.addInitScript(() => {
        const recordWhenBodyExists = new MutationObserver(() => {
          if (!document.body) return;
          recordWhenBodyExists.disconnect();
          Object.assign(window, {
            backgroundWhenBodyAppeared: getComputedStyle(document.documentElement).backgroundColor,
          });
        });
        recordWhenBodyExists.observe(document, { childList: true, subtree: true });
      });
      await page.reload();

      const background = await page.evaluate(
        () => (window as unknown as { backgroundWhenBodyAppeared: string }).backgroundWhenBodyAppeared,
      );
      expect(themeOfBackground(background)).toBe(opposite(system));
    });
  });
}

test('Tercih yokken sistem teması değişince sayfa ve düğme onu izler', async ({ page }) => {
  await page.emulateMedia({ colorScheme: 'light' });
  await page.goto('/projects');
  await expectAppliedTheme(page, 'light');

  await page.emulateMedia({ colorScheme: 'dark' });

  await expectAppliedTheme(page, 'dark');
});

for (const system of themes) {
  test.describe(`JavaScript kapalı, sistem teması ${system}`, () => {
    test.use({ javaScriptEnabled: false, colorScheme: system });

    test('Düğme görünmez, sayfalar sistem temasıyla çalışır', async ({ page }) => {
      await page.goto('/projects');
      await page.getByLabel('Name').fill('Mobile');
      await page.getByRole('button', { name: 'Create project' }).click();

      await expect(page.getByRole('list', { name: 'Projects' }).getByRole('link', { name: 'Mobile' })).toBeVisible();
      await expect(themeToggle(page)).toHaveCount(0);
      expect(await appliedTheme(page)).toBe(system);
    });
  });
}
