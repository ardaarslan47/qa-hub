import type { Page } from '@playwright/test';
import { expect, test } from './fixtures';
import { createCase, createProject, openRun, startRun } from './steps';

type Theme = 'light' | 'dark';
type Rgb = [number, number, number];

const themes: Theme[] = ['light', 'dark'];

function parseOpaqueRgb(color: string): Rgb {
  const channels = color.match(/[\d.]+/g)?.map(Number) ?? [];
  const [r, g, b, alpha = 1] = channels;
  expect(channels.length, `${color} is an rgb color`).toBeGreaterThanOrEqual(3);
  expect(alpha, `${color} is opaque`).toBe(1);
  return [r, g, b];
}

function luminance([r, g, b]: Rgb): number {
  const [lr, lg, lb] = [r, g, b].map((channel) => {
    const c = channel / 255;
    return c <= 0.03928 ? c / 12.92 : ((c + 0.055) / 1.055) ** 2.4;
  });
  return 0.2126 * lr + 0.7152 * lg + 0.0722 * lb;
}

function contrastRatio(a: Rgb, b: Rgb): number {
  const [lighter, darker] = [luminance(a), luminance(b)].sort((x, y) => y - x);
  return (lighter + 0.05) / (darker + 0.05);
}

async function appliedTheme(page: Page): Promise<Theme> {
  const background = await page.evaluate(() => getComputedStyle(document.documentElement).backgroundColor);
  return luminance(parseOpaqueRgb(background)) < 0.5 ? 'dark' : 'light';
}

function statusButton(page: Page, status: 'Passed' | 'Failed') {
  return page.getByRole('list', { name: 'Executions' }).getByRole('button', { name: status });
}

async function buttonColors(page: Page, status: 'Passed' | 'Failed') {
  const { text, background } = await statusButton(page, status).evaluate((button) => {
    const style = getComputedStyle(button);
    return { text: style.color, background: style.backgroundColor };
  });
  return { text: parseOpaqueRgb(text), background: parseOpaqueRgb(background) };
}

test('Stil dosyaları: sayfa yalnızca kendi sunucusundan stil yükler, Pico lisansı CSS in yanında sunulur', async ({ page, request, baseURL }) => {
  const origin = new URL(baseURL!).origin;
  const requested: URL[] = [];
  const stylesheets: URL[] = [];
  page.on('request', (r) => requested.push(new URL(r.url())));
  page.on('response', (r) => {
    if (r.request().resourceType() === 'stylesheet' && r.ok()) stylesheets.push(new URL(r.url()));
  });

  await page.goto('/projects');
  await page.waitForLoadState('load');

  expect(requested.map((url) => url.origin)).toEqual(requested.map(() => origin));
  expect(stylesheets).toHaveLength(2);
  const pico = stylesheets.find((url) => url.pathname.includes('pico'));
  expect(pico, 'Pico stylesheet was loaded').toBeDefined();
  const rulesPerSheet = await page.evaluate(() => [...document.styleSheets].map((sheet) => sheet.cssRules.length));
  expect(rulesPerSheet.every((count) => count > 0)).toBe(true);

  const license = await request.get(new URL('pico.LICENSE.txt', pico).href);
  expect(license.ok()).toBe(true);
  const text = await license.text();
  expect(text).toContain('Copyright 2019-2025');
  expect(text).toContain('Permission is hereby granted, free of charge');
});

for (const system of themes) {
  test.describe(`Sistem teması ${system}`, () => {
    test.use({ colorScheme: system });

    test(`Sayfa sistemin ${system} temasıyla açılır`, async ({ page }) => {
      await page.goto('/projects');

      expect(await appliedTheme(page)).toBe(system);
    });

    test(`Status düğmeleri: Passed ve Failed farklı renkte ve okunaklı`, async ({ page }) => {
      await createProject(page, 'Mobile');
      await createCase(page, 'Login works');
      await startRun(page, 'Mobile', 'Sprint 1');
      await openRun(page, 'Mobile', 'Sprint 1');
      await expect(statusButton(page, 'Passed')).toBeVisible();

      const passed = await buttonColors(page, 'Passed');
      const failed = await buttonColors(page, 'Failed');

      expect(passed.background).not.toEqual(failed.background);
      expect(contrastRatio(passed.text, passed.background)).toBeGreaterThanOrEqual(4.5);
      expect(contrastRatio(failed.text, failed.background)).toBeGreaterThanOrEqual(4.5);
    });
  });
}
