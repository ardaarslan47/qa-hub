import { test as base, expect } from '@playwright/test';
import { spawn } from 'node:child_process';
import { mkdtemp, rm } from 'node:fs/promises';
import { tmpdir } from 'node:os';
import path from 'node:path';

const binary = process.env.QAHUB_BIN ?? path.resolve(__dirname, '../bin/qahub');

type RunningApp = {
  url: string;
  log: () => string;
  stop: () => Promise<void>;
};

async function startApp(): Promise<RunningApp> {
  const dir = await mkdtemp(path.join(tmpdir(), 'qahub-e2e-'));
  const server = spawn(binary, ['-addr', '127.0.0.1:0', '-db', path.join(dir, 'qahub.db')], {
    stdio: ['ignore', 'ignore', 'pipe'],
  });
  let log = '';
  const exited = new Promise<void>((resolve) => server.once('exit', () => resolve()));

  const stop = async () => {
    if (server.exitCode === null && server.signalCode === null) {
      server.kill('SIGTERM');
      await exited;
    }
    await rm(dir, { recursive: true, force: true });
  };

  try {
    const url = await new Promise<string>((resolve, reject) => {
      server.stderr.on('data', (chunk) => {
        log += chunk;
        const match = /listening on (http:\/\/\S+)/.exec(log);
        if (match) resolve(match[1]);
      });
      server.once('error', reject);
      server.once('exit', (code) => reject(new Error(`qahub exited with code ${code} before listening:\n${log}`)));
    });
    await waitUntilAnswers(url);
    return { url, log: () => log, stop };
  } catch (error) {
    await stop();
    throw error;
  }
}

async function waitUntilAnswers(url: string): Promise<void> {
  const deadline = Date.now() + 10_000;
  for (;;) {
    try {
      await fetch(url);
      return;
    } catch (error) {
      if (Date.now() > deadline) throw error;
      await new Promise((resolve) => setTimeout(resolve, 50));
    }
  }
}

export const test = base.extend({
  baseURL: async ({}, use, testInfo) => {
    const app = await startApp();
    try {
      await use(app.url);
    } finally {
      await app.stop();
      if (testInfo.status !== testInfo.expectedStatus) {
        await testInfo.attach('qahub.log', { body: app.log(), contentType: 'text/plain' });
      }
    }
  },
});

export { expect };
