// PROPOSED ONLY: not executed. Source-extracted JavaScript, no product module import.
import { readFileSync } from 'node:fs';
import { createHash } from 'node:crypto';
import assert from 'node:assert/strict';
const source = readFileSync(process.argv[2], 'utf8');
assert.equal(createHash('sha256').update(source).digest('hex'),
  '01e3ec2bf5f5af19dcd497fce5a37c8a9394723bca5760ceca53f8b8128626ab');
const add = source.match(/  function addFailure\(error: unknown\) \{\n    if \(!failures.includes\(error\)\) failures.push\(error\);\n  \}/)?.[0];
const pair = source.match(/    if \(pairErrors.length\) \{\n      for \(const error of pairErrors\) addFailure\(error\);\n      throw pairErrors\[0\];\n    \}/)?.[0];
const drain = source.match(/for \(const result of results\) if \(result.status === "rejected"\) addFailure\(result.reason\);/)?.[0];
assert.ok(add && pair && drain, 'exact source blocks must be present');
// Remove only this known parameter annotation; the extracted control flow stays unchanged.
const evaluate = new Function('pairErrors', 'results', `
  const failures = [];
  ${add.replace('error: unknown', 'error')}
  try { ${pair} } catch (error) { addFailure(error); }
  const primary = failures.length;
  ${drain}
  return { failures, primary, cleanup: failures.length - primary };
`);
const a = new Error('a'), b = new Error('b');
for (const [name, errors, expected] of [
  ['one rejection', [a], [a]],
  ['two distinct rejections', [a, b], [a, b]],
  ['same error on both requests', [a, a], [a]],
]) {
  const actual = evaluate(errors, errors.map(reason => ({ status: 'rejected', reason })));
  assert.deepEqual(actual.failures, expected, name);
  assert.equal(actual.primary, expected.length, name);
  assert.equal(actual.cleanup, 0, name);
}
console.log(JSON.stringify({ cases: 3, productImports: 0, http: 0,
  primaryCounts: [1, 2, 1], cleanupCounts: [0, 0, 0] }));
