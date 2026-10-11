import assert from 'node:assert/strict';
import { execFileSync } from 'node:child_process';
import { mkdtemp, readFile, rm, writeFile } from 'node:fs/promises';
import { tmpdir } from 'node:os';
import { join } from 'node:path';
import { fileURLToPath } from 'node:url';
import { test } from 'node:test';
import { build } from 'esbuild';

const buildScript = fileURLToPath(new URL('./build.mjs', import.meta.url));

test('generated HTML preserves bundled JavaScript and dollar replacement sequences', async () => {
  const dir = await mkdtemp(join(tmpdir(), 'counter-build-test-'));
  try {
    const values = ['$&', '$`', "$'", '$$'];
    await writeFile(join(dir, 'app.js'), `export const special = ${JSON.stringify(values)};`);
    await writeFile(join(dir, 'counter.template.html'), '<html><head></head><body><!--APP_SCRIPT--></body></html>');
    const result = await build({ absWorkingDir: dir, entryPoints: ['app.js'], bundle: true, minify: true, format: 'esm', write: false });
    const bundle = result.outputFiles[0].text;
    execFileSync(process.execPath, [buildScript], { cwd: dir });
    const html = await readFile(join(dir, 'counter.html'), 'utf8');
    const script = html.match(/<script type="module">([\s\S]*)<\/script>/)?.[1];
    assert.equal(script, bundle);
    for (const value of values) assert.ok(script.includes(value), `missing literal ${value}`);
    await writeFile(join(dir, 'generated.mjs'), script);
    execFileSync(process.execPath, ['--check', join(dir, 'generated.mjs')]);
  } finally {
    await rm(dir, { recursive: true, force: true });
  }
});
