import { expect, type Page } from '@playwright/test';

export type Theme = 'light' | 'dark';

export const themes: Theme[] = ['light', 'dark'];

export function opposite(theme: Theme): Theme {
  return theme === 'dark' ? 'light' : 'dark';
}

export function themeToggle(page: Page) {
  return page.getByRole('button', { name: 'Dark theme' });
}

export function themeOfBackground(color: string): Theme {
  const channels = color.match(/\d+(\.\d+)?/g)?.map(Number);
  if (!channels || channels.length < 3) throw new Error(`${color} is not an rgb color`);
  const [r, g, b] = channels;
  return (r * 299 + g * 587 + b * 114) / 1000 < 128 ? 'dark' : 'light';
}

export async function appliedTheme(page: Page): Promise<Theme> {
  const background = await page.evaluate(() => getComputedStyle(document.documentElement).backgroundColor);
  return themeOfBackground(background);
}

export async function expectAppliedTheme(page: Page, theme: Theme) {
  await expect.poll(() => appliedTheme(page)).toBe(theme);
  await expect(themeToggle(page)).toHaveAttribute('aria-pressed', String(theme === 'dark'));
}
