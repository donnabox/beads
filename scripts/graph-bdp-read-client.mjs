// Real network demonstration using the unchanged, independently built public
// BDP client. The Python owner creates the workspace and owns every process.
import assert from 'node:assert/strict';
import { execFile } from 'node:child_process';
import { createHash } from 'node:crypto';
import { readFile, writeFile } from 'node:fs/promises';
import { join } from 'node:path';
import { pathToFileURL } from 'node:url';
import { promisify } from 'node:util';

const scope = process.env.BDP_SCOPE;
const checkout = process.env.BDP_CHECKOUT;
const output = process.env.BDP_EVIDENCE;
assert.ok(scope && checkout && output && process.env.BDP_BD);
const { BdpClient, createFetchTransport, isBdpClientProblem } = await import(
  pathToFileURL(join(checkout, 'packages/client/dist/index.js')).href);
const { parseBeadRecord, parseBeadCollection, parseLinkCollection } = await import(
  pathToFileURL(join(checkout, 'packages/protocol/dist/index.js')).href);
const checks = [];
const network = [];
const artifacts = {};
const digest = bytes => createHash('sha256').update(bytes).digest('hex');
const pass = name => { checks.push(name); console.log(`PASS ${name}`); };
const success = value => { assert.equal(isBdpClientProblem(value), false, JSON.stringify(value)); return value; };
const id = path => scope + path;
const planID = id('beads/plan');
const contextID = id('links/context');
const memoryType = id('types/preview-memory-v2');
const relatedType = id('types/preview-related-v2');

// Observational decoration only: delegates to native fetch without synthesizing
// any status, body, routing, or discovery. It does not log credentials.
async function observedFetch(input, init = {}) {
  const url = typeof input === 'string' ? input : input instanceof URL ? input.href : input.url;
  const response = await fetch(input, { ...init, signal: init.signal ?? AbortSignal.timeout(15_000) });
  assert.ok(network.length < 200, 'network request cap');
  const clone = response.clone();
  const bytes = new Uint8Array(await clone.arrayBuffer());
  assert.ok(bytes.length <= 2 * 1024 * 1024, 'network receipt body cap');
  network.push({ url, method: init.method ?? 'GET', status: response.status,
    headers: Object.fromEntries(response.headers), bodyBytes: bytes.length, bodySha256: digest(bytes) });
  return response;
}

const client = new BdpClient({ scope, transport: createFetchTransport(observedFetch,
  { maximumResponseBodyBytes: 2 * 1024 * 1024, responseTimeoutMs: 15_000 }) });
const perform = async (request, options) => success(await client.perform(request, options));

async function collect(collection, parameters = {}) {
  const continuationScope = client.createContinuationScope();
  const pages = [];
  try {
    let page = await perform({ kind: 'collection', collection, limit: 1, ...parameters }, { continuationScope });
    for (;;) {
      assert.ok(pages.length < 20, 'collection page cap');
      pages.push(page);
      if (page.next === null) break;
      page = await perform({ kind: 'collection', collection, continuation: page.next }, { continuationScope });
    }
    const items = pages.flatMap(page => page.items);
    assert.equal(new Set(items.map(item => item.id)).size, items.length, 'duplicate collection identity');
    return { pages, items };
  } finally {
    client.forgetContinuations(continuationScope);
  }
}

async function http(path, options = {}) {
  const response = await observedFetch(path.startsWith('http') ? path : id(path),
    { redirect: 'manual', ...options });
  const bytes = new Uint8Array(await response.arrayBuffer());
  return { response, bytes, text: new TextDecoder().decode(bytes) };
}

let failure;
try {
  artifacts.discovery = success(await client.discover());
  assert.equal(artifacts.discovery.scope, scope);
  assert.equal(artifacts.discovery.profile, 'read');
  pass('public client service-desc discovery and strict Read discovery');
  const plan = await perform({ kind: 'resource', resource: 'bead', id: planID });
  artifacts.planBefore = plan;
  assert.equal(plan.type, memoryType);
  assert.equal(plan.properties.title, 'Plan — 雪');
  assert.equal(plan.ownedLinks[relatedType][0].properties.note, 'before page');
  for (const [resource, resourceID] of [['type', plan.type], ['bead', id('beads/work')], ['link', contextID]]) {
    const value = await perform({ kind: 'resource', resource, id: resourceID });
    assert.equal(value.id, resourceID);
    artifacts[resource + ':' + resourceID] = value;
  }
  const properties = await perform({ kind: 'properties', resource: 'bead', id: planID });
  assert.deepEqual(properties, plan.properties);
  pass('public client reads Memory, Issue, Link, installed Type and properties');

  for (const collection of ['beads', 'links', 'types']) {
    const result = await collect(collection);
    artifacts[collection] = result;
    const expected = collection === 'beads' ? ['beads/alpha', 'beads/plan', 'beads/prereq', 'beads/work'].map(id)
      : collection === 'links' ? [id('links/back'), contextID, process.env.BDP_DEPENDENCY_ID] : null;
    if (expected) assert.deepEqual(result.items.map(item => item.id).sort(), expected.sort());
    else assert.equal(result.items.length, 4);
  }
  pass('public client traverses complete Bead, Link and Type inventories without omissions or duplicates');

  const owner = client.createContinuationScope();
  try {
    const request = { kind: 'collection', collection: 'beads', type: memoryType, conformsTo: memoryType,
      selector: '$[?@.properties.title]', limit: 1 };
    const first = await perform(request, { continuationScope: owner });
    assert.equal(first.items[0].id, id('beads/alpha'));
    assert.ok(first.next);
    const args = ['update', 'links/context', '--properties', '{"note":"after page"}',
      '--if-revision', plan.ownedLinks[relatedType][0].revision,
      '--if-source-revision', plan.revision, '--json'];
    const beforeBinary = digest(await readFile(process.env.BDP_BD));
    let mutation;
    try {
      const execution = await promisify(execFile)(process.env.BDP_BD, args,
        { timeout: 60_000, maxBuffer: 2 * 1024 * 1024, encoding: 'utf8' });
      await writeFile(join(output, 'client-cli-update.stdout.log'), execution.stdout);
      await writeFile(join(output, 'client-cli-update.stderr.log'), execution.stderr);
      mutation = JSON.parse(execution.stdout);
      assert.equal(mutation.preview, true);
      assert.equal(mutation.result.link.properties.note, 'after page');
    } catch (error) {
      await writeFile(join(output, 'client-cli-update.stdout.log'), error.stdout ?? '');
      await writeFile(join(output, 'client-cli-update.stderr.log'), error.stderr ?? '');
      throw error;
    } finally {
      await writeFile(join(output, 'client-cli-update.json'), JSON.stringify({
        argv: [process.env.BDP_BD, ...args], cwd: process.cwd(), binarySha256: beforeBinary,
        result: mutation ?? null,
      }, null, 2));
    }
    assert.equal(digest(await readFile(process.env.BDP_BD)), beforeBinary, 'installed binary changed');
    const nextRequest = { kind: 'collection', collection: 'beads', continuation: first.next };
    const second = await perform(nextRequest, { continuationScope: owner });
    assert.equal(second.next, null);
    assert.deepEqual(second.items, [plan]);
    // The public client consumes a successful continuation capability. Server
    // replay is a separate real-fetch assertion, using the public parser.
    const replayResponse = await http(first.next);
    assert.equal(replayResponse.response.status, 200);
    const replay = parseBeadCollection(JSON.parse(replayResponse.text));
    assert.deepEqual(replay, second);
    const fresh = await perform({ kind: 'resource', resource: 'bead', id: planID });
    assert.notEqual(fresh.revision, plan.revision);
    assert.equal(fresh.ownedLinks[relatedType][0].properties.note, 'after page');
    const newCollection = await collect('beads', { type: memoryType });
    assert.deepEqual(newCollection.items.find(item => item.id === planID), fresh);
    artifacts.retained = { first, second, replay, fresh };
    pass('second-process guarded CLI update changes fresh reads; public-client continuation and real-fetch replay retain old revision and owned state');
  } finally {
    client.forgetContinuations(owner);
  }

  const selected = await collect('links', { source: planID, target: id('beads/work'), endpoint: planID,
    selector: '$[?@.properties.note == "after page"]' });
  assert.deepEqual(selected.items.map(item => item.id), [contextID]);
  const incidentOwner = client.createContinuationScope();
  try {
    let page = await perform({ kind: 'bead-links', bead: planID, direction: 'both', limit: 1 },
      { continuationScope: incidentOwner });
    const items = [...page.items];
    let pages = 1;
    while (page.next !== null) {
      assert.ok(pages++ < 10);
      page = await perform({ kind: 'bead-links', bead: planID, continuation: page.next },
        { continuationScope: incidentOwner });
      items.push(...page.items);
    }
    assert.deepEqual(items.map(item => item.id).sort(), [id('links/back'), contextID].sort());
    artifacts.incident = items;
  } finally {
    client.forgetContinuations(incidentOwner);
  }
  pass('public client structural and Selector filters plus complete incident Link traversal');

  const aggregate = await http(`${planID}?include=links&limit=1`, { headers: { accept: 'application/json' } });
  assert.equal(aggregate.response.status, 200);
  const { links, ...record } = JSON.parse(aggregate.text);
  parseBeadRecord(record);
  parseLinkCollection(links);
  assert.equal(record.id, planID);
  assert.equal(links.items.length, 1);
  assert.ok(links.next);
  const next = new URL(links.next);
  assert.equal(next.origin + next.pathname, planID);
  assert.equal(next.searchParams.get('view'), 'links');
  assert.equal(next.searchParams.has('include'), false);
  const aggregateNext = await http(links.next);
  assert.equal(aggregateNext.response.status, 200);
  const incidentRemainder = parseLinkCollection(JSON.parse(aggregateNext.text));
  assert.equal(incidentRemainder.next, null);
  assert.deepEqual([...links.items, ...incidentRemainder.items].map(item => item.id).sort(),
    [id('links/back'), contextID].sort());
  artifacts.aggregate = { record, links, incidentRemainder, mechanism: 'real fetch and public subshape parsers; no typed aggregate client API' };
  pass('real-fetch aggregate and incident continuation validated by independent public parsers (outside typed client API)');

  const full = await http(planID);
  assert.equal(full.response.status, 200);
  const etag = full.response.headers.get('etag');
  assert.ok(etag);
  const head = await http(planID, { method: 'HEAD' });
  assert.equal(head.response.status, 200);
  assert.equal(head.bytes.length, 0);
  assert.equal(head.response.headers.get('content-length'), String(full.bytes.length));
  assert.equal(head.response.headers.get('etag'), etag);
  for (const [name, options, status] of [
    ['not modified', { headers: { 'if-none-match': etag } }, 304],
    ['weak match', { headers: { 'if-none-match': `W/${etag}` } }, 304],
    ['precondition failed', { headers: { 'if-match': '"wrong"' } }, 412],
    ['method', { method: 'POST' }, 405],
    ['unacceptable media', { headers: { accept: 'text/plain' } }, 406],
    ['excluded JSON', { headers: { accept: 'application/json;q=0, */*;q=1' } }, 406],
  ]) {
    const result = await http(planID, options);
    assert.equal(result.response.status, status, name);
    assert.equal(result.bytes.length, 0, name);
    if (status === 304) assert.notEqual(result.response.headers.get('content-length'), '0');
    if (status === 405) assert.equal(result.response.headers.get('allow'), 'GET, HEAD');
  }
  const missing = await http('beads/missing', { headers: { accept: 'text/plain', 'if-none-match': '*' } });
  assert.equal(missing.response.status, 404);
  const invalidCursor = await http('beads/?cursor=unknown', { headers: { 'if-none-match': '*' } });
  assert.notEqual(invalidCursor.response.status, 304);
  assert.ok(invalidCursor.response.status >= 400);
  for (const query of ['limit=1&limit=2', 'unknown=true', 'selector=%GG', 'version=unsupported']) {
    const invalid = await http(`beads/?${query}`);
    assert.equal(invalid.response.status, 400, query);
  }
  pass('real HTTP Unicode HEAD, conditional precedence, bodyless 304/412/405/406, missing Resource and invalid query/cursor handling');
} catch (error) {
  failure = { name: error.name, message: error.message, stack: error.stack };
  process.exitCode = 1;
} finally {
  await client.close();
  await writeFile(join(output, 'client-network.json'), JSON.stringify(network, null, 2));
  await writeFile(join(output, 'client-artifacts.json'), JSON.stringify(artifacts, null, 2));
  await writeFile(join(output, 'client-results.json'), JSON.stringify({
    passed: !failure, checks, failure: failure ?? null, requests: network.length,
    transport: 'native fetch; observational receipt wrapper only',
    aggregatePublicClientAPI: false,
    replayMechanism: 'native fetch plus parseBeadCollection; pinned client consumes successful continuation capabilities',
  }, null, 2));
  if (failure) console.error(JSON.stringify(failure));
}
