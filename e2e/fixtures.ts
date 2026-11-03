import { spawn, type ChildProcess } from 'node:child_process';
import { once } from 'node:events';
import { mkdtemp, rm } from 'node:fs/promises';
import { tmpdir } from 'node:os';
import path from 'node:path';
import { test as base } from '@playwright/test';
import { serverBinary } from './global-setup';

export { expect } from '@playwright/test';

export const test = base.extend({
  baseURL: async ({}, use) => {
    const dir = await mkdtemp(path.join(tmpdir(), 'qa-hub-e2e-'));
    const server = spawn(serverBinary, ['-addr', '127.0.0.1:0', '-db', path.join(dir, 'qa-hub.db')], {
      stdio: ['ignore', 'ignore', 'pipe'],
    });
    try {
      await use(await listeningURL(server));
    } finally {
      if (server.exitCode === null && server.signalCode === null) {
        server.kill('SIGTERM');
        await once(server, 'exit');
      }
      await rm(dir, { recursive: true, force: true });
    }
  },
});

function listeningURL(server: ChildProcess): Promise<string> {
  const stderr = server.stderr!;
  stderr.setEncoding('utf8');
  return new Promise((resolve, reject) => {
    let startup = '';
    const onStartup = (chunk: string) => {
      startup += chunk;
      const match = /listening on (\S+)/.exec(startup);
      if (!match) return;
      stderr.off('data', onStartup);
      stderr.on('data', (chunk: string) => process.stderr.write(chunk));
      resolve(match[1]);
    };
    stderr.on('data', onStartup);
    server.once('error', reject);
    server.once('exit', (code) => reject(new Error(`qa-hub exited with code ${code} before listening:\n${startup}`)));
  });
}
