import { expect, test } from './fixtures';
import { expectAppliedTheme, themeToggle } from './theme';
import { menuLink } from './steps';

test.use({
  channel: 'chromium',
  colorScheme: 'light',
  launchOptions: {
    ignoreDefaultArgs: ['--disable-back-forward-cache'],
    args: ['--enable-features=BackForwardCache'],
  },
});

test('Geri/ileri önbelleği: başka sayfada yapılan seçim, geri dönülen sayfada da görünür', async ({ page }) => {
  await page.goto('/projects');
  await page.evaluate(() => Object.assign(window, { servedFromBackForwardCache: true }));
  await menuLink(page, 'Library').click();
  await themeToggle(page).click();
  await expectAppliedTheme(page, 'dark');

  await page.goBack({ waitUntil: 'commit' });

  expect(await page.evaluate(() => 'servedFromBackForwardCache' in window), 'page was restored from the cache').toBe(true);
  await expectAppliedTheme(page, 'dark');
});
