import { execFileSync } from 'node:child_process';
import path from 'node:path';

const repoRoot = path.resolve(__dirname, '..');

export const serverBinary = path.join(repoRoot, 'bin', 'qa-hub');

export default function buildServer() {
  execFileSync('make', ['build'], { cwd: repoRoot, stdio: 'inherit' });
}
