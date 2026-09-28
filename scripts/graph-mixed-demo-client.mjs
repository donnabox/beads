// Presentation companion: unchanged public BDP client, real HTTP, no mock.
import assert from 'node:assert/strict';
import { createHash } from 'node:crypto';
import { readFile, writeFile } from 'node:fs/promises';
import { join } from 'node:path';
import { pathToFileURL } from 'node:url';
const { BDP_SCOPE: scope, BDP_CHECKOUT: checkout, BDP_EVIDENCE: output } = process.env;
const { BdpClient, createFetchTransport, isBdpClientProblem } = await import(
  pathToFileURL(join(checkout, 'packages/client/dist/index.js')).href);
const expected = JSON.parse(await readFile(join(output, 'expected.json'), 'utf8'));
const requests = [];
const client = new BdpClient({ scope, transport: createFetchTransport(async (input, options) => {
  const response = await fetch(input, { ...options, signal: options?.signal ?? AbortSignal.timeout(15000) });
  const bytes = new Uint8Array(await response.clone().arrayBuffer());
  assert.ok(bytes.length < 2 * 1024 * 1024, 'HTTP body budget exceeded');
  assert.ok(requests.length < 50, 'HTTP request budget exceeded');
  requests.push({ url: String(input), status: response.status, bytes: bytes.length,
    sha256: createHash('sha256').update(bytes).digest('hex') });
  return response;
}, { maximumResponseBodyBytes: 2 * 1024 * 1024, responseTimeoutMs: 15000 }) });
const good = value => { assert.equal(isBdpClientProblem(value), false, JSON.stringify(value)); return value; };
const projectLink = x => Object.fromEntries(['id', 'type', 'source', 'target', 'revision', 'properties'].map(k => [k, x[k]]));
let passed = false;
let failure;
try {
  const discovery = good(await client.discover());
  assert.equal(discovery.profile, 'read');
  assert.equal(discovery.scope, scope);
  for (const name of ['plan', 'release', 'rationale']) {
    const want = expected[name];
    const actual = good(await client.perform({ kind: 'resource', resource: 'bead', id: want.id }));
    for (const key of ['id', 'type', 'revision', 'properties']) assert.deepEqual(actual[key], want[key]);
    const owned = Object.values(actual.ownedLinks).flat().map(projectLink).sort((a, b) => a.id.localeCompare(b.id));
    assert.deepEqual(owned, (want.owned ?? []).map(projectLink).sort((a, b) => a.id.localeCompare(b.id)));
    console.log(`BDP read ${name}: complete properties and owned Links match the CLI`);
  }
  const incident = good(await client.perform({ kind: 'bead-links', bead: expected.plan.id, direction: 'both', limit: 100 }));
  assert.equal(incident.next, null);
  assert.deepEqual(incident.items.map(x => x.id).sort(), [scope + 'links/context', scope + 'links/rationale'].sort());
  const continuationScope = client.createContinuationScope();
  const ids = [];
  try {
    let page = good(await client.perform({ kind: 'collection', collection: 'beads', limit: 1 }, { continuationScope }));
    let pages = 0;
    for (;;) {
      assert.ok(++pages <= 5, 'Bead page budget exceeded');
      ids.push(...page.items.map(x => x.id));
      assert.ok(ids.length <= 4);
      if (page.next === null) break;
      page = good(await client.perform({ kind: 'collection', collection: 'beads', continuation: page.next }, { continuationScope }));
    }
  } finally { client.forgetContinuations(continuationScope); }
  assert.deepEqual(ids.sort(), ['plan', 'rationale', 'release', 'verification'].map(x => scope + 'beads/' + x));
  passed = true;
} catch (error) {
  failure = { name: error.name, message: error.message, stack: error.stack };
  throw error;
} finally {
  try { await client.close(); }
  catch (error) {
    failure ??= { name: error.name, message: error.message, stack: error.stack };
    passed = false;
    process.exitCode = 1;
  }
  await writeFile(join(output, 'client-results.json'), JSON.stringify({ passed, failure, requests, httpRequests: requests.length }, null, 2));
}
