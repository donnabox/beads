// Harness adapted from c92114ec68f92871fc296dc2fcc33bba78229065; public client stays pinned and unchanged.
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
assert.ok(scope && checkout && output && process.env.BDP_BD && process.env.BDP_TOKEN);
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
  // This capture exercises the authenticated public interface. Never forward
  // its credential outside the exact initialized origin/Scope, or on redirect.
  const destination = new URL(url);
  const authority = new URL(scope);
  assert.equal(destination.origin, authority.origin);
  assert.ok(destination.pathname.startsWith(authority.pathname));
  assert.equal(destination.username + destination.password, '');
  const headers = new Headers(input instanceof Request ? input.headers : undefined);
  new Headers(init.headers).forEach((value, key) => headers.set(key, value));
  headers.set('Authorization', `Bearer ${process.env.BDP_TOKEN}`);
  const response = await fetch(input, { ...init, headers, redirect: 'manual',
    signal: init.signal ?? AbortSignal.timeout(15_000) });
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

  // The installed CLI is the writer; the independent public client observes
  // the resulting Issue state through the read-only HTTP surface.
  const workID = id('beads/work');
  const targetID = id('beads/prereq');
  const dependencyID = process.env.BDP_DEPENDENCY_ID;
  assert.ok(dependencyID);
  const workBefore = artifacts['bead:' + workID];
  const targetBefore = await perform({ kind: 'resource', resource: 'bead', id: targetID });
  const dependencyBefore = await perform({ kind: 'resource', resource: 'link', id: dependencyID });
  assert.equal(dependencyBefore.source, workID);
  assert.equal(dependencyBefore.target, targetID);
  assert.deepEqual(Object.values(workBefore.ownedLinks).flat(), [dependencyBefore]);
  const incidentBefore = await perform({ kind: 'bead-links', bead: workID, direction: 'both', limit: 100 });
  assert.equal(incidentBefore.next, null);
  assert.deepEqual(incidentBefore.items.map(link => link.id).sort(),
    [contextID, id('links/back'), dependencyID].sort());
  const workHTTPBefore = await http(workID);
  assert.equal(workHTTPBefore.response.status, 200);
  const workETagBefore = workHTTPBefore.response.headers.get('etag');
  assert.ok(workETagBefore);
  assert.deepEqual(parseBeadRecord(JSON.parse(workHTTPBefore.text)), workBefore);

  const issueFields = { title: 'Edited work — 雪', description: 'CLI Issue description.\n',
    design: '# Design\nPreserve Dependencies and contextual Links.\n',
    acceptance_criteria: 'The public BDP client sees all four edited fields.\n' };
  const issueArgs = ['update', workID, '--title', issueFields.title,
    '--description', issueFields.description, '--design', issueFields.design,
    '--acceptance', issueFields.acceptance_criteria, '--if-revision', workBefore.revision, '--json'];
  const issueBinary = digest(await readFile(process.env.BDP_BD));
  let issueExecution;
  let issueExitCode = null;
  try {
    issueExecution = await promisify(execFile)(process.env.BDP_BD, issueArgs,
      { timeout: 60_000, maxBuffer: 2 * 1024 * 1024, encoding: 'utf8' });
    issueExitCode = 0;
  } catch (error) {
    issueExecution = error;
    issueExitCode = typeof error.code === 'number' ? error.code : null;
    throw error;
  } finally {
    await writeFile(join(output, 'client-cli-issue-update.stdout.log'), issueExecution?.stdout ?? '');
    await writeFile(join(output, 'client-cli-issue-update.stderr.log'), issueExecution?.stderr ?? '');
    await writeFile(join(output, 'client-cli-issue-update.json'), JSON.stringify({
      argv: [process.env.BDP_BD, ...issueArgs], cwd: process.cwd(), binarySha256: issueBinary,
      exitCode: issueExitCode, signal: issueExecution?.signal ?? null,
      stdoutSha256: digest(issueExecution?.stdout ?? ''), stderrSha256: digest(issueExecution?.stderr ?? ''),
    }, null, 2));
  }
  assert.equal(digest(await readFile(process.env.BDP_BD)), issueBinary, 'installed binary changed');
  const issueMutation = JSON.parse(issueExecution.stdout);
  assert.equal(issueMutation.preview, true);
  assert.equal(issueMutation.result.changed, true);
  const workAfter = await perform({ kind: 'resource', resource: 'bead', id: workID });
  assert.equal(workAfter.revision, issueMutation.result.issue.revision);
  assert.notEqual(workAfter.revision, workBefore.revision);
  assert.equal(workAfter.type, workBefore.type);
  assert.deepEqual(workAfter.properties, {
    ...workBefore.properties, ...issueFields, updated_at: workAfter.properties.updated_at,
  });
  assert.deepEqual(workAfter.ownedLinks, workBefore.ownedLinks);
  assert.deepEqual(await perform({ kind: 'properties', resource: 'bead', id: workID }), workAfter.properties);
  const targetAfter = await perform({ kind: 'resource', resource: 'bead', id: targetID });
  assert.deepEqual(targetAfter, targetBefore);
  const dependencyAfter = await perform({ kind: 'resource', resource: 'link', id: dependencyID });
  assert.deepEqual(dependencyAfter, dependencyBefore);
  const incidentAfter = await perform({ kind: 'bead-links', bead: workID, direction: 'both', limit: 100 });
  assert.deepEqual(incidentAfter, incidentBefore);
  const issueInventory = await perform({ kind: 'collection', collection: 'beads', limit: 100 });
  assert.equal(issueInventory.next, null);
  assert.equal(issueInventory.items.length, 4);
  assert.deepEqual(issueInventory.items.find(bead => bead.id === workID), workAfter);
  const workHTTPAfter = await http(workID, { headers: { 'if-none-match': workETagBefore } });
  assert.equal(workHTTPAfter.response.status, 200);
  assert.ok(workHTTPAfter.response.headers.get('etag'));
  assert.notEqual(workHTTPAfter.response.headers.get('etag'), workETagBefore);
  assert.deepEqual(parseBeadRecord(JSON.parse(workHTTPAfter.text)), workAfter);
  artifacts.issueEdit = { before: workBefore, after: workAfter, mutation: issueMutation,
    targetBefore, targetAfter, dependencyBefore, dependencyAfter, incidentBefore, incidentAfter,
    inventory: issueInventory, etagBefore: workETagBefore, etagAfter: workHTTPAfter.response.headers.get('etag'),
    writer: 'installed CLI; HTTP surface remains read-only' };
  pass('public client sees guarded CLI Issue text edit in resource, properties and inventory; owned Dependencies, incident Links and target remain unchanged; old ETag yields fresh 200');

  // Extend the same real workspace with claim -> append. The helper records
  // bounded installed writes; every observed record still comes from public BDP.
  async function notesCLI(name, args) {
    const binarySha256 = digest(await readFile(process.env.BDP_BD));
    let execution;
    let exitCode = null;
    try {
      execution = await promisify(execFile)(process.env.BDP_BD, args,
        { timeout: 60_000, maxBuffer: 2 * 1024 * 1024, encoding: 'utf8' });
      exitCode = 0;
    } catch (error) {
      execution = error;
      exitCode = typeof error.code === 'number' ? error.code : null;
      throw error;
    } finally {
      await writeFile(join(output, name + '.stdout.log'), execution?.stdout ?? '');
      await writeFile(join(output, name + '.stderr.log'), execution?.stderr ?? '');
      await writeFile(join(output, name + '.json'), JSON.stringify({
        argv: [process.env.BDP_BD, ...args], cwd: process.cwd(), binarySha256,
        exitCode, signal: execution?.signal ?? null,
        stdoutSha256: digest(execution?.stdout ?? ''), stderrSha256: digest(execution?.stderr ?? ''),
      }, null, 2));
    }
    assert.equal(digest(await readFile(process.env.BDP_BD)), binarySha256, 'installed binary changed');
    const mutation = JSON.parse(execution.stdout);
    assert.equal(mutation.schemaVersion, 1);
    assert.equal(mutation.preview, true);
    return mutation;
  }
  const notesActor = 'bdp-read-note-holder';
  const claimMutation = await notesCLI('client-cli-issue-claim',
    ['update', workID, '--claim', '--actor', notesActor, '--json']);
  assert.equal(claimMutation.result.changed, true);
  const claimed = await perform({ kind: 'resource', resource: 'bead', id: workID });
  assert.equal(claimed.revision, claimMutation.result.issue.revision);
  assert.notEqual(claimed.revision, workAfter.revision);
  assert.equal(claimed.properties.status, 'in_progress');
  assert.equal(claimed.properties.assignee, notesActor);
  assert.equal(claimed.properties.notes ?? '', '');
  assert.equal(Date.parse(claimed.properties.lease_expires_at) - Date.parse(claimed.properties.heartbeat_at), 300_000);
  assert.ok(claimed.properties.started_at);
  const claimFields = ['status', 'assignee', 'started_at', 'updated_at', 'lease_expires_at', 'heartbeat_at', 'lease_granted_node'];
  const expectedClaimProperties = { ...workAfter.properties };
  for (const field of claimFields) {
    if (Object.hasOwn(claimed.properties, field)) expectedClaimProperties[field] = claimed.properties[field];
    else delete expectedClaimProperties[field];
  }
  assert.deepEqual(claimed.properties, expectedClaimProperties);
  assert.deepEqual(claimed.ownedLinks, workAfter.ownedLinks);
  const claimedHTTP = await http(workID);
  assert.equal(claimedHTTP.response.status, 200);
  const notesETagBefore = claimedHTTP.response.headers.get('etag');
  assert.ok(notesETagBefore);
  assert.deepEqual(parseBeadRecord(JSON.parse(claimedHTTP.text)), claimed);
  const claimNetworkIndex = network.length - 1;
  await writeFile(join(output, "client-issue-notes-before.body"), claimedHTTP.bytes);
  const noopMutation = await notesCLI('client-cli-issue-append-noop',
    ['update', workID, '--append-notes=', '--if-revision', claimed.revision, '--actor', notesActor, '--json']);
  assert.equal(noopMutation.result.changed, false);
  assert.deepEqual(noopMutation.result.issue, claimMutation.result.issue);
  const noopRecord = await perform({ kind: 'resource', resource: 'bead', id: workID });
  assert.deepEqual(noopRecord, claimed);
  const noopHTTP = await http(workID, { headers: { 'if-none-match': notesETagBefore } });
  assert.equal(noopHTTP.response.status, 304);
  assert.equal(noopHTTP.bytes.length, 0);
  assert.equal(noopHTTP.response.headers.get('etag'), notesETagBefore);
  const noopNetworkIndex = network.length - 1;
  await writeFile(join(output, "client-issue-notes-noop.body"), noopHTTP.bytes);
  pass('public client sees installed claim and empty append no-op without changing notes, lease, revision or ETag');

  const appendText = '  Progress — 雪\r\nsecond line\t  ';
  const appendMutation = await notesCLI('client-cli-issue-append',
    ['update', workID, '--append-notes', appendText, '--if-revision', claimed.revision, '--actor', notesActor, '--json']);
  assert.equal(appendMutation.result.changed, true);
  const appended = await perform({ kind: 'resource', resource: 'bead', id: workID });
  assert.equal(appended.revision, appendMutation.result.issue.revision);
  assert.notEqual(appended.revision, claimed.revision);
  assert.equal(appended.type, claimed.type);
  assert.deepEqual(appended.properties, { ...claimed.properties, notes: appendText, updated_at: appended.properties.updated_at });
  assert.deepEqual(appended.ownedLinks, claimed.ownedLinks);
  const appendedProperties = await perform({ kind: 'properties', resource: 'bead', id: workID });
  assert.deepEqual(appendedProperties, appended.properties);
  const appendedInventory = await perform({ kind: 'collection', collection: 'beads', limit: 100 });
  assert.equal(appendedInventory.next, null);
  assert.equal(appendedInventory.items.length, 4);
  assert.deepEqual(appendedInventory.items.find(bead => bead.id === workID), appended);
  assert.deepEqual(await perform({ kind: 'resource', resource: 'bead', id: targetID }), targetBefore);
  assert.deepEqual(await perform({ kind: 'resource', resource: 'link', id: dependencyID }), dependencyBefore);
  const appendedIncident = await perform({ kind: 'bead-links', bead: workID, direction: 'both', limit: 100 });
  assert.deepEqual(appendedIncident, incidentBefore);
  const appendedHTTP = await http(workID, { headers: { 'if-none-match': notesETagBefore } });
  assert.equal(appendedHTTP.response.status, 200);
  assert.ok(appendedHTTP.response.headers.get('etag'));
  assert.notEqual(appendedHTTP.response.headers.get('etag'), notesETagBefore);
  assert.deepEqual(parseBeadRecord(JSON.parse(appendedHTTP.text)), appended);
  await writeFile(join(output, "client-issue-notes-after.body"), appendedHTTP.bytes);
  artifacts.issueNotes = { before: workAfter, claimed, claimMutation, noopMutation, noopRecord,
    appendText, appendMutation, after: appended, properties: appendedProperties,
    inventory: appendedInventory, incident: appendedIncident,
    etagBefore: notesETagBefore, etagAfter: appendedHTTP.response.headers.get('etag'),
    claimNetworkIndex, noopNetworkIndex, appendNetworkIndex: network.length - 1,
    writer: 'installed CLI; unchanged independent public client over authenticated BDP Read' };
  pass('public client sees exact appended notes through resource, properties and inventory with preserved lease/owned Links; predecessor ETag yields fresh 200');

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

  const memoryProperties = { title: 'Edited plan — 雪', body: 'Updated through the installed CLI.\n' };
  const memoryArgs = ['update', planID, '--properties', JSON.stringify(memoryProperties),
    '--if-revision', record.revision, '--json'];
  const memoryBinary = digest(await readFile(process.env.BDP_BD));
  let memoryExecution;
  try {
    memoryExecution = await promisify(execFile)(process.env.BDP_BD, memoryArgs,
      { timeout: 60_000, maxBuffer: 2 * 1024 * 1024, encoding: 'utf8' });
  } catch (error) {
    memoryExecution = error;
    throw error;
  } finally {
    await writeFile(join(output, 'client-cli-memory-update.stdout.log'), memoryExecution?.stdout ?? '');
    await writeFile(join(output, 'client-cli-memory-update.stderr.log'), memoryExecution?.stderr ?? '');
    await writeFile(join(output, 'client-cli-memory-update.json'), JSON.stringify({
      argv: [process.env.BDP_BD, ...memoryArgs], cwd: process.cwd(), binarySha256: memoryBinary,
    }, null, 2));
  }
  assert.equal(digest(await readFile(process.env.BDP_BD)), memoryBinary);
  const memoryMutation = JSON.parse(memoryExecution.stdout);
  assert.equal(memoryMutation.result.changed, true);
  assert.deepEqual(memoryMutation.result.memory.properties, memoryProperties);
  const edited = await perform({ kind: 'resource', resource: 'bead', id: planID });
  assert.deepEqual(edited.properties, memoryProperties);
  assert.equal(edited.revision, memoryMutation.result.memory.revision);
  assert.notEqual(edited.revision, record.revision);
  assert.deepEqual(edited.ownedLinks, record.ownedLinks);
  const unchangedLink = await perform({ kind: 'resource', resource: 'link', id: contextID });
  assert.deepEqual(unchangedLink, record.ownedLinks[relatedType][0]);
  assert.deepEqual(await perform({ kind: 'properties', resource: 'bead', id: planID }), memoryProperties);
  artifacts.memoryEdit = { record: edited, mutation: memoryMutation, unchangedLink };
  pass('public client sees CLI Memory content edit with unchanged owned Link and consistent properties');

  // A retained current-read cursor is a snapshot capability, not public History.
  // Delete only after all earlier fixtures/checks have completed.
  const deletionOwner = client.createContinuationScope();
  try {
    const deletionSourceBefore = await perform({ kind: 'resource', resource: 'bead', id: workID });
    const deletionLinkBefore = await perform({ kind: 'resource', resource: 'link', id: dependencyID });
    assert.deepEqual(deletionSourceBefore, artifacts.issueNotes.after);
    assert.deepEqual(deletionLinkBefore, dependencyBefore);
    const deletionLinksBefore = await perform({ kind: 'collection', collection: 'links', limit: 100 });
    assert.equal(deletionLinksBefore.next, null);
    assert.equal(deletionLinksBefore.items.length, 3);
    const retainedFirst = await perform({ kind: 'collection', collection: 'beads', limit: 1 },
      { continuationScope: deletionOwner });
    assert.equal(retainedFirst.items[0].id, id('beads/alpha'));
    assert.ok(retainedFirst.next);

    const deletionArgs = ['unlink', dependencyID, '--if-revision', deletionLinkBefore.revision,
      '--if-source-revision', deletionSourceBefore.revision, '--actor', 'bdp-read-dependency-remover', '--json'];
    const deletionBinary = digest(await readFile(process.env.BDP_BD));
    let deletionExecution;
    let deletionExitCode = null;
    try {
      deletionExecution = await promisify(execFile)(process.env.BDP_BD, deletionArgs,
        { timeout: 60_000, maxBuffer: 2 * 1024 * 1024, encoding: 'utf8' });
      deletionExitCode = 0;
    } catch (error) {
      deletionExecution = error;
      deletionExitCode = typeof error.code === 'number' ? error.code : null;
      throw error;
    } finally {
      await writeFile(join(output, 'client-cli-dependency-unlink.stdout.log'), deletionExecution?.stdout ?? '');
      await writeFile(join(output, 'client-cli-dependency-unlink.stderr.log'), deletionExecution?.stderr ?? '');
      await writeFile(join(output, 'client-cli-dependency-unlink.json'), JSON.stringify({
        argv: [process.env.BDP_BD, ...deletionArgs], cwd: process.cwd(), binarySha256: deletionBinary,
        exitCode: deletionExitCode, signal: deletionExecution?.signal ?? null,
        stdoutSha256: digest(deletionExecution?.stdout ?? ''), stderrSha256: digest(deletionExecution?.stderr ?? ''),
      }, null, 2));
    }
    assert.equal(digest(await readFile(process.env.BDP_BD)), deletionBinary, 'installed binary changed');
    const deletionMutation = JSON.parse(deletionExecution.stdout);
    assert.equal(deletionMutation.preview, true);
    assert.equal(deletionMutation.result.changed, true);
    assert.equal(deletionMutation.result.link.id, dependencyID);
    assert.equal(deletionMutation.result.link.state, 'deleted');
    assert.equal(deletionMutation.result.link.previousVersion, deletionLinkBefore.revision);

    const retainedPages = [retainedFirst];
    while (retainedPages.at(-1).next !== null) {
      assert.ok(retainedPages.length < 10, 'retained deletion page cap');
      retainedPages.push(await perform({ kind: 'collection', collection: 'beads',
        continuation: retainedPages.at(-1).next }, { continuationScope: deletionOwner }));
    }
    const retainedItems = retainedPages.flatMap(page => page.items);
    assert.equal(retainedItems.length, 4);
    assert.equal(new Set(retainedItems.map(item => item.id)).size, 4);
    const retainedSource = retainedItems.find(bead => bead.id === workID);
    assert.deepEqual(retainedSource, deletionSourceBefore);
    assert.deepEqual(Object.values(retainedSource.ownedLinks).flat(), [deletionLinkBefore]);

    const deletionSourceAfter = await perform({ kind: 'resource', resource: 'bead', id: workID });
    assert.equal(deletionSourceAfter.revision, deletionMutation.result.source.revision);
    assert.notEqual(deletionSourceAfter.revision, deletionSourceBefore.revision);
    assert.deepEqual(Object.values(deletionSourceAfter.ownedLinks).flat(), []);
    assert.deepEqual(deletionSourceAfter.properties, {
      ...deletionSourceBefore.properties, updated_at: deletionSourceAfter.properties.updated_at,
    });
    const deletionLinksAfter = await perform({ kind: 'collection', collection: 'links', limit: 100 });
    assert.equal(deletionLinksAfter.next, null);
    const survivingLinks = deletionLinksBefore.items.filter(link => link.id !== dependencyID);
    assert.equal(survivingLinks.length, 2);
    assert.deepEqual(deletionLinksAfter.items, survivingLinks);
    const deletionIncident = await perform({ kind: 'bead-links', bead: workID, direction: 'both', limit: 100 });
    assert.equal(deletionIncident.next, null);
    assert.deepEqual(deletionIncident.items, survivingLinks);
    const deletionTargetAfter = await perform({ kind: 'resource', resource: 'bead', id: targetID });
    assert.deepEqual(deletionTargetAfter, targetBefore);
    const gone = await client.perform({ kind: 'resource', resource: 'link', id: dependencyID });
    assert.equal(isBdpClientProblem(gone), true);
    assert.equal(gone.code, 'resource-not-found');
    assert.equal(gone.status, 404);
    const goneHTTP = await http(dependencyID);
    assert.equal(goneHTTP.response.status, 404);
    assert.equal(JSON.parse(goneHTTP.text).code, 'resource-not-found');
    artifacts.dependencyUnlink = { mutation: deletionMutation, sourceBefore: deletionSourceBefore,
      sourceAfter: deletionSourceAfter, linkBefore: deletionLinkBefore, retainedPages,
      linksBefore: deletionLinksBefore, linksAfter: deletionLinksAfter, incidentAfter: deletionIncident,
      targetAfter: deletionTargetAfter, gone, mechanism: 'CLI deletion; current BDP Read and retained collection snapshot, not public History' };
    pass('guarded CLI Dependency unlink updates fresh public-client ownership/inventory, preserves target/context, returns current 404 resource-not-found, and retains old source/Link through an existing cursor');
  } finally {
    client.forgetContinuations(deletionOwner);
  }

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
