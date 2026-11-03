import { test as base, expect } from '@playwright/test';
import { spawn } from 'node:child_process';
import { once } from 'node:events';
import { mkdtemp, rm } from 'node:fs/promises';
import { tmpdir } from 'node:os';
import path from 'node:path';

const appBinary = path.resolve(__dirname, '../bin/qa-hub');

async function startApp() {
  const dataDir = await mkdtemp(path.join(tmpdir(), 'qa-hub-e2e-'));
  const app = spawn(appBinary, ['-addr', '127.0.0.1:0', '-db', path.join(dataDir, 'qa-hub.db')], {
    stdio: ['ignore', 'pipe', 'inherit'],
  });

  const baseURL = await new Promise<string>((resolve, reject) => {
    let output = '';
    app.stdout.on('data', (chunk: Buffer) => {
      output += chunk.toString();
      const listening = output.match(/listening on (http:\/\/\S+)/);
      if (listening?.[1]) resolve(listening[1]);
    });
    app.once('error', reject);
    app.once('exit', (code) => reject(new Error(`qa-hub exited with code ${code} before it started listening`)));
  });

  async function stop() {
    if (app.exitCode === null && app.signalCode === null) {
      app.kill('SIGTERM');
      await once(app, 'exit');
    }
    await rm(dataDir, { recursive: true, force: true });
  }

  return { baseURL, stop };
}

export const test = base.extend({
  baseURL: async ({}, use) => {
    const app = await startApp();
    await use(app.baseURL);
    await app.stop();
  },
});

export { expect };
