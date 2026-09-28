// A small independent public-client Read chapter. No writer or mock transport.
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
async function observedFetch(input, options = {}) {
  assert.ok(requests.length < 16, 'HTTP request budget exceeded');
  const url = input instanceof Request ? input.url : String(input);
  const method = options.method ?? (input instanceof Request ? input.method : 'GET');
  const requestHeaders = options.headers ?? (input instanceof Request ? input.headers : undefined);
  const response = await fetch(input, { ...options, signal: options.signal ?? AbortSignal.timeout(15000) });
  const bytes = new Uint8Array(await response.clone().arrayBuffer());
  assert.ok(bytes.length <= 2 * 1024 * 1024, 'HTTP body budget exceeded');
  const bodyFile = `http-${String(requests.length + 1).padStart(2, '0')}.body`;
  await writeFile(join(output, bodyFile), bytes);
  requests.push({ url, method, status: response.status,
    requestHeaders: Object.fromEntries(new Headers(requestHeaders)),
    headers: Object.fromEntries(response.headers), bodyFile, bytes: bytes.length,
    sha256: createHash('sha256').update(bytes).digest('hex') });
  return response;
}
const client = new BdpClient({ scope, transport: createFetchTransport(observedFetch,
  { maximumResponseBodyBytes: 2 * 1024 * 1024, responseTimeoutMs: 15000 }) });
const good = value => { assert.equal(isBdpClientProblem(value), false, JSON.stringify(value)); return value; };
const byID = rows => [...rows].sort((a, b) => a.id < b.id ? -1 : a.id > b.id ? 1 : 0);
function common(want) {
  const result = Object.fromEntries(['id', 'type', 'revision', 'properties'].map(key => [key, want[key]]));
  if (want.attribution.status === 'claimed') {
    result.attribution = { principal: want.attribution.actor, status: 'claimed' };
  } else {
    assert.equal(want.attribution.status, 'unknown');
    assert.equal(want.attribution.actor, '');
  }
  return result;
}
function wireLink(want) {
  return { ...common(want), source: want.source, target: want.target };
}
function wireBead(want) {
  // Existing immutable descriptors: Issues own Dependency group (even empty);
  // Memory wildcard ownership has only its populated installed-Type groups.
  const ownedLinks = want.type === scope + 'types/preview-issue-v2'
    ? { [scope + 'types/preview-blocks-v1']: [] } : {};
  for (const link of want.owned) (ownedLinks[link.type] ??= []).push(wireLink(link));
  for (const key of Object.keys(ownedLinks)) ownedLinks[key] = byID(ownedLinks[key]);
  return { ...common(want), ownedLinks };
}
let passed = false;
let failure;
try {
  const discovery = good(await client.discover());
  assert.equal(discovery.profile, 'read');
  assert.equal(discovery.scope, scope);
  for (const want of expected.beads) {
    const actual = good(await client.perform({ kind: 'resource', resource: 'bead', id: want.id }));
    assert.deepEqual(actual, wireBead(want));
    console.log(`BDP ${want.id}: complete current properties, revision, attribution and owned Links match CLI`);
  }
  const linkPage = good(await client.perform({ kind: 'collection', collection: 'links', limit: 100 }));
  assert.equal(linkPage.next, null);
  assert.deepEqual(byID(linkPage.items), byID(expected.links.map(wireLink)));
  const beadPage = good(await client.perform({ kind: 'collection', collection: 'beads', limit: 100 }));
  assert.equal(beadPage.next, null);
  assert.deepEqual(byID(beadPage.items), byID(expected.beads.map(wireBead)));
  const incident = good(await client.perform({ kind: 'bead-links', bead: expected.plan.id, direction: 'both', limit: 100 }));
  assert.equal(incident.next, null);
  const expectedIncident = expected.links.filter(link => link.source === expected.plan.id || link.target === expected.plan.id);
  assert.deepEqual(byID(incident.items), byID(expectedIncident.map(wireLink)));
  const baseline = requests.find(row => row.url === expected.plan.id && row.status === 200);
  assert.ok(baseline, `fresh plan resource response absent from observed URLs: ${requests.map(row => row.url).join(', ')}`);
  assert.ok(baseline.headers.etag, 'fresh plan resource response lacks ETag');
  // The pinned client exposes no conditional-result convenience here; use the
  // same real observed transport for this explicit HTTP conditional request.
  const unchanged = await observedFetch(expected.plan.id, { headers: { Accept: 'application/json', 'If-None-Match': baseline.headers.etag } });
  assert.equal(unchanged.status, 304);
  assert.equal((await unchanged.arrayBuffer()).byteLength, 0);
  assert.equal(unchanged.headers.get('etag'), baseline.headers.etag);
  assert.equal(requests.length, 10);
  console.log('Complete Bead/Link inventories and incident Links match; unchanged plan returns304. HTTP remains read-only.');
  passed = true;
} catch (error) {
  failure = { name: error.name, message: error.message, stack: error.stack };
  process.exitCode = 1;
} finally {
  try { await client.close(); }
  catch (error) {
    failure ??= { name: error.name, message: error.message, stack: error.stack };
    passed = false;
    process.exitCode = 1;
  }
  await writeFile(join(output, 'client-results.json'), JSON.stringify({ passed, failure: failure ?? null, qualification: false,
    httpRequests: requests.length, requests, limitation: 'Nine public-client reads plus one real conditional GET; no HTTP writes or History.' }, null, 2));
}
