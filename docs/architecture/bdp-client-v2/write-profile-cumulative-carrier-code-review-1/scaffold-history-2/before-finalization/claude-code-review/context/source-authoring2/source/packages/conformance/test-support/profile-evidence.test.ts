/** Forty-eight finite DATA callbacks. No service, child or HTTP executor is launched. */
import { createHash } from "node:crypto";
import {
  readFileSync,
  mkdtempSync,
  rmSync,
  writeFileSync,
  lstatSync,
  existsSync,
  mkdirSync,
  symlinkSync,
} from "node:fs";
import { tmpdir } from "node:os";
import path from "node:path";
import { expect, it, vi } from "vitest";
const ioFaults = vi.hoisted(() => ({ write: false, rename: false, cleanup: false }));
vi.mock("node:fs", async (importOriginal) => {
  const actual = await importOriginal<typeof import("node:fs")>();
  return {
    ...actual,
    writeFileSync: vi.fn((...args: Parameters<typeof actual.writeFileSync>) => {
      if (ioFaults.write && typeof args[0] === "number") {
        ioFaults.write = false;
        throw new Error("unit-write-private");
      }
      return actual.writeFileSync(...args);
    }),
    renameSync: vi.fn((...args: Parameters<typeof actual.renameSync>) => {
      if (ioFaults.rename) {
        ioFaults.rename = false;
        throw new Error("unit-rename-private");
      }
      return actual.renameSync(...args);
    }),
    rmSync: vi.fn((...args: Parameters<typeof actual.rmSync>) => {
      if (ioFaults.cleanup) {
        ioFaults.cleanup = false;
        throw new Error("unit-cleanup-private");
      }
      return actual.rmSync(...args);
    }),
  };
});
import {
  assembleProfileEvidence,
  verifyProfileEvidence,
  serializeProfileEvidence,
} from "./profile-evidence.js";
import {
  createReadUpdateArtifactSet,
  copyReadUpdateArtifactSetSources,
} from "./read-update-artifact-set.js";
import {
  bindServedProfileConfiguration,
  loadProfileContractSelection,
} from "./profile-contract-plan.js";
import { createCumulativeTransactionalArtifactBundle } from "../src/artifact-bundle.js";
import { canonicalJson } from "../src/canonical-json.js";
import { createJsonSchemaValidator } from "../src/schema-validator.js";
import schema from "../../../schemas/bdp-v0.schema.json" with { type: "json" };
import { selectApplicableScenariosForProfile } from "../src/selection.js";
import { runConformanceMatrix, runCumulativeTransactionalMatrix } from "../src/runner.js";
import {
  createPrivateProfileSidecarWriter,
  createArtifactEvidenceWriter,
  artifactRoutes,
} from "../../../apps/bdptest/test-support/read-update-artifact-execution.js";

type Data = Record<string, unknown>;
type Artifacts = Parameters<typeof assembleProfileEvidence>[2];
type Observation = NonNullable<Parameters<typeof assembleProfileEvidence>[3][number]>;
const rec = (v: unknown) => v as Data;
const json = (v: unknown) => ({ bytes: Buffer.from(JSON.stringify(v)) });
const sha = (v: Uint8Array | string) => createHash("sha256").update(v).digest("hex");
const read = (p: string) => ({ bytes: readFileSync(new URL(p, import.meta.url)) });
const scope = "https://reference.invalid/acme/";
const ruSelection = loadProfileContractSelection("read-update"),
  txSelection = loadProfileContractSelection("transactional");
const ruIndex = read("./read-update-artifact-set-v1.json"),
  ruFixture = read("../fixtures/read-update-reference-v1.json"),
  ruCoverage = read("./read-update-coverage-v1.json");
const index = JSON.parse(ruIndex.bytes.toString()) as { shards: { id: string; source: string }[] };
const ruSources = {
  index: ruIndex,
  fixture: ruFixture,
  coverage: ruCoverage,
  manifests: Object.fromEntries(index.shards.map((x) => [x.source, read(`../${x.source}`)])),
};
let set: ReturnType<typeof createReadUpdateArtifactSet> | undefined;
const ruSet = () => (set ??= createReadUpdateArtifactSet(ruSelection, ruSources));
const txSources = {
  manifest: read("./transactional/manifest-v1.json"),
  fixture: read("./transactional/fixture-v1.json"),
  coverage: read("./transactional/coverage-v1.json"),
};
const txBundle = () => createCumulativeTransactionalArtifactBundle(txSelection, txSources);
const schemaValidator = createJsonSchemaValidator(schema);
function configuration(
  profile: "read-update" | "transactional",
  fixture = profile === "read-update" ? ruFixture : txSources.fixture,
  overrides: Partial<Parameters<typeof bindServedProfileConfiguration>[0]> = {},
) {
  return bindServedProfileConfiguration({
    scope,
    profile,
    history: false,
    sourceCommit: "a".repeat(40),
    installedPayloadDigest: "b".repeat(64),
    installedContractsDigest: "c".repeat(64),
    policyDigest: "d".repeat(64),
    authenticationDigest: "e".repeat(64),
    fixture,
    ...overrides,
  });
}
function context(profile: "read-update" | "transactional" = "read-update") {
  const selection = profile === "read-update" ? ruSelection : txSelection;
  const artifacts: Artifacts =
    profile === "read-update"
      ? { kind: "read-update", set: ruSet() }
      : { kind: "transactional", ...txSources };
  const owners =
    profile === "read-update"
      ? ruSet().shards.map((s) => ({ id: s.id, bundle: s.bundle, ids: s.scenarioIds }))
      : [
          {
            id: "transactional-baseline",
            bundle: txBundle(),
            ids: txBundle().manifest.scenarios.map((x) => x.id),
          },
        ];
  return { selection, artifacts, owners, configuration: configuration(profile) };
}
type Context = ReturnType<typeof context>;
function attempt(c: Context, outcome: "running" | "failed" | "observed" = "running") {
  return {
    outcome,
    failureStages: [] as ("execution" | "assertion" | "cleanup" | "credential-check")[],
    owners: c.owners.map((x) => ({
      id: x.id,
      status: "unobserved",
      cleanup: "not-started",
      evidence: "absent",
    })),
    sourceDirty: null,
  };
}
const absent = (c: Context) => c.owners.map(() => null);
function assemble(
  c: Context,
  observations: Parameters<typeof assembleProfileEvidence>[3] = absent(c),
  a = attempt(c),
) {
  return assembleProfileEvidence(c.selection, c.configuration, c.artifacts, observations, a);
}
function exact(c: Context) {
  return {
    scope: c.configuration.scope,
    profile: c.selection.profile as "read-update" | "transactional",
    routes: artifactRoutes(c.owners[0]!.bundle, c.configuration.scope),
    maximumRequestHeaderBytes: 16384,
    maximumRequestBodyBytes: c.selection.profile === "read-update" ? 131072 : 65536,
    defaultCredentialRef: "operator",
    credentialHandles: [
      ...new Set([
        "operator",
        ...c.owners[0]!.bundle.manifest.scenarios.flatMap((row) =>
          (row.actions ?? row.requests).flatMap((action) =>
            "credentialRef" in action && action.credentialRef !== undefined
              ? [action.credentialRef]
              : [],
          ),
        ),
      ]),
    ].map((id) => ({ id, headerNames: id === "anonymous" ? [] : ["authorization"] })),
  };
}
function observation(c: Context, n = 0, resultOverride?: unknown): Observation {
  const owner = c.owners[n]!,
    cfg = c.configuration,
    mode = exact(c),
    runtime = "f".repeat(64);
  const selected =
    c.selection.profile === "read-update" && n > 0
      ? [...owner.ids]
      : selectApplicableScenariosForProfile(owner.bundle.catalog, c.selection.profile).map(
          (r) => r.id,
        );
  const result = resultOverride ?? {
    reportVersion: 4,
    claimEligible: false,
    scope: cfg.scope,
    profile: cfg.profile,
    seed: owner.bundle.fixture.seed,
    artifacts: owner.bundle.digests,
    declarations: {
      targetLabel: "synthetic-unit-data-not-authority",
      exactConfigurationDigest: sha(
        canonicalJson({
          scope: mode.scope,
          profile: mode.profile,
          routes: mode.routes,
          maximumRequestHeaderBytes: mode.maximumRequestHeaderBytes,
          maximumRequestBodyBytes: mode.maximumRequestBodyBytes,
        }),
      ),
    },
    selectedScenarioIds: selected,
    ...(c.selection.profile === "read-update" && n > 0 ? { scenarioSelection: owner.ids } : {}),
    scenarios: selected.map((id) => ({
      id,
      requiredProfile: owner.bundle.catalog.scenarios.find((r) => r.id === id)!.requiredProfile,
      state: "harness-error",
      category: "not-run",
      requirements: [],
      exchanges: [],
    })),
  };
  const ready = {
    scope: cfg.scope,
    profile: cfg.profile,
    history: false,
    payloadIdentity: cfg.installedPayloadDigest,
    artifactIdentity: `development-schema5-${cfg.installedContractsDigest}`,
    policyIdentity: cfg.policyDigest,
    authenticationIdentity: cfg.authenticationDigest,
    configurationIdentity: runtime,
  };
  const baseline =
    c.selection.profile === "read-update"
      ? ready
      : {
          served: cfg,
          operationTargets: unitTargets(c),
          runtimeConfigurationIdentity: runtime,
          authorityMode: "canonical",
        };
  const reportSource = json(
    c.selection.profile === "read-update"
      ? {
          sourceCommit: cfg.sourceCommit,
          configuration: cfg,
          baselineReceipt: baseline,
          sessions: [],
          exactConfiguration: mode,
          result,
        }
      : {
          sourceCommit: cfg.sourceCommit,
          entry: "synthetic-unit-data-not-installed",
          baseline,
          sessions: [],
          result,
        },
  );
  return {
    ownerId: owner.id,
    sourceCommit: cfg.sourceCommit,
    configuration: cfg,
    runtimeConfigurationIdentity: runtime,
    exactConfiguration: mode,
    ready: [baseline],
    reportSource,
    expectedReportSha256: sha(reportSource.bytes),
    lifecycle: { cleanup: "pending" },
  };
}
function withObservation(c: Context, o = observation(c), n = 0) {
  const rows: (Observation | null)[] = absent(c);
  rows[n] = o;
  return rows;
}
function changeReport(o: Observation, edit: (wrapper: Data) => void): Observation {
  const wrapper = JSON.parse(Buffer.from(o.reportSource.bytes).toString()) as Data;
  edit(wrapper);
  const reportSource = json(wrapper);
  return { ...o, reportSource, expectedReportSha256: sha(reportSource.bytes) };
}
// Synthetic source-data metadata only: these fixtures never claim an installed child/listener.
function unitTargets(c: Context) {
  const suffixes = {
    batch: "batch",
    createBead: "create-bead",
    createLink: "create-link",
    deleteAlias: "delete-alias",
    deleteBead: "delete-bead",
    deleteLink: "delete-link",
    deleteWhere: "delete-where",
    putAlias: "put-alias",
    sequence: "sequence",
    updateBead: "update-bead",
    updateLink: "update-link",
    updateWhere: "update-where",
  };
  return Object.fromEntries(
    Object.entries(suffixes).map(([k, v]) => [k, c.configuration.scope + "operations/" + v]),
  );
}
function unitSession(c: Context, scenario: string, ordinal = 0, reopen = false): Data {
  const cfg = c.configuration;
  const common = {
    scope: cfg.scope,
    profile: cfg.profile,
    history: false,
    payloadIdentity: cfg.installedPayloadDigest,
    artifactIdentity: `development-schema5-${cfg.installedContractsDigest}`,
    policyIdentity: cfg.policyDigest,
    authenticationIdentity: cfg.authenticationDigest,
  };
  return c.selection.profile === "read-update"
    ? { scenario, ...common, configurationIdentity: "f".repeat(64) }
    : {
        event: "qualification.ready",
        authorityMode: "canonical",
        runtimeConfigurationIdentity: "f".repeat(64),
        ...common,
        claimEligible: false,
        listenUrl: `http://127.0.0.1:${41000 + ordinal}/`,
        targets: unitTargets(c),
        scenario,
        reopen,
        sessionOrdinal: ordinal,
      };
}
function withUnitSessions(c: Context, o: Observation, sessions: Data[]): Observation {
  const updated = changeReport(o, (w) => {
    w.sessions = sessions;
  });
  return { ...updated, ready: [o.ready[0], ...sessions] };
}
function reachedThrough(c: Context, rows: Data[], index: number) {
  for (let i = 0; i <= index; i++) {
    const row = rows[i]!,
      plan = c.owners[0]!.bundle.manifest.scenarios.find((p) => p.id === row.id);
    if (plan) {
      delete row.category;
      row.state = "pass";
    } else {
      row.state = "harness-error";
      row.category = "not-implemented";
    }
  }
}
function prerequisiteData(
  phase: "initial" | "recheck",
  stop?: "category" | "cleanup",
  retainedActions = false,
) {
  let c = context();
  if (retainedActions) {
    // Existing action/IDs only; an in-memory source-data prerequisite exercises the real generic schema.
    const sources = copyReadUpdateArtifactSetSources(ruSet()),
      sourceName = index.shards[0]!.source;
    const manifest = JSON.parse(Buffer.from(sources.manifests[sourceName]!.bytes).toString());
    const plan = manifest.scenarios.find((p: { actions?: Data[] }) =>
      p.actions?.some((a) => a.family !== "http"),
    );
    plan.actions.at(-1).prerequisiteScenario = manifest.scenarios[0].id;
    (sources.manifests as Record<string, { bytes: Uint8Array }>)[sourceName] = json(manifest);
    const unitSet = createReadUpdateArtifactSet(ruSelection, sources);
    c = {
      ...c,
      artifacts: { kind: "read-update", set: unitSet },
      owners: unitSet.shards.map((s) => ({ id: s.id, bundle: s.bundle, ids: s.scenarioIds })),
    };
  }
  const owner = c.owners[0]!,
    plan = owner.bundle.manifest.scenarios.find(
      (p) =>
        (p.actions ?? p.requests).some((a) => a.prerequisiteScenario !== undefined) &&
        (!retainedActions ||
          (p.actions ?? p.requests).some((a) => "family" in a && a.family !== "http")),
    )!;
  const prerequisiteId = (plan.actions ?? plan.requests).find(
    (a) => a.prerequisiteScenario !== undefined,
  )!.prerequisiteScenario!;
  const o = changeReport(observation(c), (w) => {
    const rows = rec(w.result).scenarios as Data[],
      row = rows.find((r) => r.id === plan.id)!;
    reachedThrough(c, rows, rows.indexOf(row));
    const prior = rows.find((r) => r.id === prerequisiteId)!;
    prior.state = phase === "initial" ? "fail" : "pass";
    row.state = "harness-error";
    row.category = "not-run";
    const metadata = owner.bundle.catalog.scenarios.find((r) => r.id === prerequisiteId)!;
    row.prerequisiteFailure = {
      id: prerequisiteId,
      requirements: metadata.requirements.map(({ source, anchor }) => `${source}${anchor}`),
      reason: "synthetic source-data prerequisite",
      phase,
      ...(stop === "category" ? { category: "observation-limit" } : {}),
    };
    if (phase === "recheck") {
      const declared = plan.actions ?? plan.requests;
      const retained = retainedActions ? declared : declared.slice(0, 1);
      row.exchanges = retained
        .filter((a) => !("family" in a) || a.family === "http")
        .map((a) => ({ request: { id: a.id }, assertions: [] }));
      row.actions = retained
        .filter((a) => "family" in a && a.family !== "http")
        .map((a) => ({
          action: {
            id: a.id,
            family: "family" in a ? a.family : "http",
            operation: "operation" in a ? a.operation : undefined,
          },
          assertions: [],
        }));
    }
    if (stop === "cleanup") row.cleanupError = "synthetic source-data cleanup";
    if (!stop) {
      // One following actual selected row demonstrates that a local skip is not a matrix stop.
      const next = rows[rows.indexOf(row) + 1]!;
      reachedThrough(c, [next], 0);
    }
  });
  return { c, o, plan, prerequisiteId };
}
function missingExecutorData() {
  const c = context(),
    owner = c.owners[0]!,
    plan = owner.bundle.manifest.scenarios.find((p) =>
      (p.actions ?? p.requests).some((a) => "family" in a && a.family !== "http"),
    )!;
  const o = changeReport(observation(c), (w) => {
    const rows = rec(w.result).scenarios as Data[],
      row = rows.find((r) => r.id === plan.id)!;
    reachedThrough(c, rows, rows.indexOf(row) + 1);
    row.state = "harness-error";
    row.category = "not-implemented";
    const declared = plan.actions ?? plan.requests,
      missing = declared.findIndex((a) => "family" in a && a.family !== "http");
    const prefix = declared.slice(0, missing + 1);
    row.exchanges = prefix
      .filter((a) => !("family" in a) || a.family === "http")
      .map((a) => ({ request: { id: a.id }, assertions: [] }));
    row.actions = prefix
      .filter((a) => "family" in a && a.family !== "http")
      .map((a) => ({
        action: {
          id: a.id,
          family: "family" in a ? a.family : "http",
          operation: "operation" in a ? a.operation : undefined,
        },
        assertions: [],
        executionError: {
          category: "not-implemented",
          message: "synthetic source-data missing executor",
        },
      }));
  });
  return { c, o, plan };
}

function code(work: () => unknown, expected: string) {
  let caught: unknown;
  try {
    work();
  } catch (error) {
    caught = error;
  }
  expect(caught).toBeInstanceOf(Error);
  expect(rec(caught).code).toBe(expected);
  expect((caught as Error).message).toBe("private profile evidence refused");
}
function unbound(outcome: "running" | "failed" = "running") {
  const c = context();
  return {
    attemptVersion: 1,
    kind: "private-profile-attempt-unbound",
    bindingStatus: "unbound",
    configuration: null,
    claimEligible: false,
    qualificationReady: false,
    requestedProfile: "read-update",
    declaredSourceCommit: c.configuration.sourceCommit,
    artifactDigests: ruSet().digests,
    requiredIds: ruSet().requirements.map((r) => r.id),
    owners: ruSet().requirements,
    observations: ruSet().requirements.map(() => null),
    attempt: attempt(c, outcome),
    limitations: [
      "no returned served configuration yet",
      "not a bound carrier, target observation, admission or qualification",
    ],
  };
}
function sourceMutation(
  edit: (sources: ReturnType<typeof copyReadUpdateArtifactSetSources>) => void,
) {
  const sources = copyReadUpdateArtifactSetSources(ruSet());
  edit(sources);
  expect(() => createReadUpdateArtifactSet(ruSelection, sources)).toThrow();
  code(
    () =>
      assembleProfileEvidence(
        ruSelection,
        configuration("read-update"),
        { kind: "read-update", set: { ...ruSet() } },
        [null, null],
        attempt(context()),
      ),
    "invalid-source",
  );
}
const dataCases: Record<string, () => void | Promise<void>> = {
  V01() {
    const c = context(),
      e = assemble(c),
      rs = e.requirements as Data[];
    expect(rs).toHaveLength(130);
    expect(rs.filter((r) => r.ownerId !== null)).toHaveLength(56);
    expect(rs.filter((r) => r.ownerId === null)).toHaveLength(74);
    expect(rs.filter((r) => r.declaredWholeContract)).toHaveLength(34);
    expect(
      rs.every((r) => r.observation === null && r.wholeClauseQualification === "unestablished"),
    ).toBe(true);
    expect(e.qualificationReady).toBe(false);
  },
  V02() {
    const c = context("transactional"),
      e = assemble(c),
      rs = e.requirements as Data[];
    expect(rs).toHaveLength(251);
    expect(rs.filter((r) => r.ownerId !== null)).toHaveLength(17);
    expect(rs.filter((r) => r.ownerId === null)).toHaveLength(234);
    expect(rs.some((r) => r.declaredWholeContract)).toBe(false);
    expect(e.claimEligible).toBe(false);
  },
  async V03() {
    await aborted("read-update");
  },
  async V04() {
    await aborted("transactional");
  },
  V05() {
    for (const profile of ["read-update", "transactional"] as const) {
      const c = context(profile),
        o = observation(c),
        owner = c.owners[0]!;
      const plan = owner.bundle.manifest.scenarios.find((r) =>
        (r.actions ?? r.requests).some((a) => "family" in a && a.family !== "http"),
      )!;
      expect(plan).toBeDefined();
      for (const vector of ["all-true", "missing", "false-missing", "false-true"] as const) {
        const report = changeReport(o, (w) => {
          const result = rec(w.result),
            rows = result.scenarios as Data[];
          const row = rows.find((r) => r.id === plan.id)!;
          delete row.category;
          row.state = vector === "all-true" ? "pass" : "fail";
          const actions = plan.actions ?? plan.requests;
          row.exchanges = actions
            .filter((a) => !("family" in a) || a.family === "http")
            .map((a) => ({
              request: { id: a.id },
              assertions: a.assertions.map((x) => ({ id: x.id, passed: true })),
            }));
          row.actions = actions
            .filter((a) => "family" in a && a.family !== "http")
            .map((a) => ({
              action: { id: a.id, family: "family" in a ? a.family : "http" },
              assertions: a.assertions.map((x) => ({ id: x.id, passed: true })),
            }));
          const lanes = [...(row.exchanges as Data[]), ...(row.actions as Data[])];
          const first = lanes[0]!.assertions as Data[];
          expect(lanes.flatMap((x) => x.assertions as Data[]).length).toBeGreaterThan(1);
          if (vector.startsWith("false")) first[0]!.passed = false;
          if (vector.includes("missing")) {
            const last = lanes.at(-1)!.assertions as Data[];
            last.pop();
          }
          // Same-ID incidental exchange cannot replace a client action's authored assertions.
          for (const a of row.actions as Data[])
            (row.exchanges as Data[]).push({
              request: { id: rec(a.action).id },
              assertions: [{ id: "unrelated", passed: false }],
            });
          (row.exchanges as Data[]).sort(
            (a, b) =>
              actions.findIndex((x) => x.id === rec(a.request).id) -
              actions.findIndex((x) => x.id === rec(b.request).id),
          );
          const at = rows.indexOf(row);
          for (let i = 0; i < at; i++) {
            if (rows[i]!.category === "not-run") rows[i]!.category = "not-implemented";
          }
          // Only genuinely absent plans may be not-implemented; other preceding rows remain explicit unexecuted data.
          for (let i = 0; i < at; i++)
            if (owner.ids.includes(String(rows[i]!.id))) {
              delete rows[i]!.category;
              rows[i]!.state = "fail";
            }
        });
        const evidence = assemble(c, withObservation(c, report));
        const row = (evidence.requirements as Data[]).find((r) => r.id === plan.id)!;
        const value = rec(row.observation).authoredAssertionsAllPassed;
        expect(value).toBe(vector.startsWith("false") ? false : vector === "missing" ? null : true);
        expect(row.wholeClauseQualification).toBe("unestablished");
      }
    }
  },
  V06() {
    const c = context(),
      o = observation(c),
      e = assemble(c, withObservation(c, o)),
      bytes = serializeProfileEvidence(e);
    expect(
      verifyProfileEvidence(
        c.selection,
        c.configuration,
        c.artifacts,
        withObservation(c, o),
        attempt(c),
        { bytes },
      ),
    ).toEqual(e);
    const original = o.reportSource.bytes;
    const raw = JSON.parse(Buffer.from(original).toString());
    for (const row of (e.requirements as Data[]).filter((r) => r.observation !== null)) {
      expect(rec(row.observation).rawWrapperSha256).toBe(sha(original));
      expect(rec(row.observation).canonicalResultSha256).toBe(sha(canonicalJson(raw.result)));
    }
    expect(rec((e.reports as Data[])[0]).canonicalResultSha256).toBe(
      sha(canonicalJson(raw.result)),
    );
    original[0] = 0;
    expect(
      rec((e.requirements as Data[]).find((r) => r.observation !== null)!.observation)
        .rawWrapperSha256,
    ).toBe(o.expectedReportSha256);
    for (const field of ["result", "sessions", "baseline"]) {
      const value = { field, omitted: undefined, array: [undefined, 1] };
      const actual = JSON.parse(JSON.stringify(value));
      expect(canonicalJson(JSON.parse(JSON.stringify(value)))).toBe(canonicalJson(actual));
    }
    const wrapper = {
      sourceCommit: "a".repeat(40),
      entry: "unit-only",
      result: { unit: true },
      sessions: [],
      baseline: {},
      supplemental: {},
      sequenceMemberSupplemental: {},
      principalSupplemental: {},
      conditionalSupplemental: {},
    };
    const handed = Buffer.from(JSON.stringify(wrapper, null, 2) + "\n");
    expect(sha(handed)).not.toBe(sha(canonicalJson(wrapper)));
    expect(handed.toString()).toBe(JSON.stringify(wrapper, null, 2) + "\n");
  },
  R10() {
    const c = context();
    code(
      () =>
        assembleProfileEvidence(
          { ...c.selection },
          c.configuration,
          c.artifacts,
          absent(c),
          attempt(c),
        ),
      "invalid-source",
    );
  },
  R11() {
    const c = context();
    code(
      () =>
        assembleProfileEvidence(
          c.selection,
          { ...c.configuration },
          c.artifacts,
          absent(c),
          attempt(c),
        ),
      "invalid-source",
    );
  },
  R12() {
    const c = context();
    let calls = 0;
    const bad = Object.defineProperty({}, "kind", {
      enumerable: true,
      get() {
        calls++;
        throw Error("tripwire");
      },
    });
    code(
      () =>
        assembleProfileEvidence(
          c.selection,
          c.configuration,
          bad as Artifacts,
          absent(c),
          attempt(c),
        ),
      "invalid-source",
    );
    expect(calls).toBe(0);
    for (const profile of ["read-update", "transactional"] as const) {
      const unit = context(profile),
        o = observation(unit);
      const tripwire = () =>
        Object.defineProperty({}, "secret", {
          enumerable: true,
          get() {
            calls++;
            throw Error("tripwire");
          },
        });
      for (const field of [
        "ownerId",
        "sourceCommit",
        "expectedReportSha256",
        "runtimeConfigurationIdentity",
      ])
        code(
          () => assemble(unit, withObservation(unit, { ...o, [field]: tripwire() } as Observation)),
          "invalid-source",
        );
      const accessor = Object.defineProperty([o.ready[0]], "0", {
        enumerable: true,
        get() {
          calls++;
          throw Error("tripwire");
        },
      });
      const nested =
        profile === "transactional"
          ? { ...rec(o.ready[0]), served: tripwire() }
          : { ...rec(o.ready[0]), scope: tripwire() };
      for (const ready of [
        accessor,
        [nested],
        [
          () => {
            calls++;
          },
        ],
        [o.ready[0], { scenario: "unit", metadata: tripwire() }],
      ])
        code(() => assemble(unit, withObservation(unit, { ...o, ready })), "invalid-source");
    }
    expect(calls).toBe(0);
  },
  R13() {
    const c = context();
    let calls = 0;
    const bad = new Proxy(c.artifacts, {
      ownKeys() {
        calls++;
        throw Error("tripwire");
      },
    });
    code(
      () => assembleProfileEvidence(c.selection, c.configuration, bad, absent(c), attempt(c)),
      "invalid-source",
    );
    expect(calls).toBe(0);
    for (const profile of ["read-update", "transactional"] as const) {
      const unit = context(profile),
        o = observation(unit);
      const proxy = (value: object) =>
        new Proxy(value, {
          ownKeys() {
            calls++;
            throw Error("tripwire");
          },
          get() {
            calls++;
            throw Error("tripwire");
          },
        });
      for (const field of [
        "ownerId",
        "sourceCommit",
        "expectedReportSha256",
        "runtimeConfigurationIdentity",
      ])
        code(
          () => assemble(unit, withObservation(unit, { ...o, [field]: proxy({}) } as Observation)),
          "invalid-source",
        );
      for (const ready of [
        proxy([...o.ready]),
        [proxy(rec(o.ready[0]))],
        [{ ...rec(o.ready[0]), served: proxy({}) }],
        [o.ready[0], proxy({})],
      ])
        code(
          () => assemble(unit, withObservation(unit, { ...o, ready } as Observation)),
          "invalid-source",
        );
    }
    expect(calls).toBe(0);
  },
  R14() {
    sourceMutation((s) => {
      const a = JSON.parse(Buffer.from(s.manifests[index.shards[0]!.source]!.bytes).toString());
      const b = JSON.parse(Buffer.from(s.manifests[index.shards[1]!.source]!.bytes).toString());
      b.scenarios.unshift(a.scenarios[0]);
      (s.manifests as Record<string, { bytes: Uint8Array }>)[index.shards[1]!.source] = json(b);
    });
  },
  R15() {
    sourceMutation((s) => {
      const p = JSON.parse(Buffer.from(s.manifests[index.shards[0]!.source]!.bytes).toString());
      expect(ruSelection.retiredIds.length).toBeGreaterThan(0);
      p.scenarios[0].id = ruSelection.retiredIds[0];
      (s.manifests as Record<string, { bytes: Uint8Array }>)[index.shards[0]!.source] = json(p);
    });
  },
  R16() {
    sourceMutation((s) => {
      const p = JSON.parse(Buffer.from(s.manifests[index.shards[0]!.source]!.bytes).toString());
      p.scenarios[0].id = "unknown.requirement";
      (s.manifests as Record<string, { bytes: Uint8Array }>)[index.shards[0]!.source] = json(p);
    });
  },
  R17() {
    sourceMutation((s) => {
      const p = JSON.parse(Buffer.from(s.manifests[index.shards[0]!.source]!.bytes).toString());
      p.scenarios.reverse();
      (s.manifests as Record<string, { bytes: Uint8Array }>)[index.shards[0]!.source] = json(p);
    });
  },
  R18() {
    for (const cross of [false, true])
      sourceMutation((s) => {
        const p = JSON.parse(Buffer.from(s.manifests[index.shards[0]!.source]!.bytes).toString()),
          other = JSON.parse(Buffer.from(s.manifests[index.shards[1]!.source]!.bytes).toString());
        (p.scenarios[0].actions ?? p.scenarios[0].requests)[0].prerequisiteScenario = cross
          ? other.scenarios[0].id
          : p.scenarios.at(-1).id;
        (s.manifests as Record<string, { bytes: Uint8Array }>)[index.shards[0]!.source] = json(p);
      });
  },
  R19() {
    const c = context(),
      o = changeReport(observation(c), (w) => {
        rec(w.result).reportVersion = 3;
      });
    code(() => assemble(c, withObservation(c, o)), "invalid-report");
  },
  R20() {
    const c = context("transactional"),
      o = observation(c);
    code(
      () => assemble(c, withObservation(c, { ...o, expectedReportSha256: "0".repeat(64) })),
      "invalid-report",
    );
    const artifactMismatch = changeReport(o, (wrapper) => {
      rec(rec(wrapper.result).artifacts).manifestDigest = "0".repeat(64);
    });
    code(() => assemble(c, withObservation(c, artifactMismatch)), "invalid-report");
    for (const field of ["result", "sessions", "baseline"]) {
      const genuine = {
        result: { n: 1, omit: undefined },
        sessions: [{ n: 2 }],
        baseline: { n: 3 },
      };
      const actual = JSON.parse(JSON.stringify(genuine));
      actual[field] = { changed: true };
      expect(canonicalJson(rec(actual)[field])).not.toBe(
        canonicalJson(JSON.parse(JSON.stringify(rec(genuine)[field]))),
      );
    }
    const bytes = serializeProfileEvidence(assemble(c, withObservation(c, o)));
    const changed = changeReport(o, (w) => {
      rec(rec(w.result).declarations).targetLabel = "different-unit-data";
    });
    code(
      () =>
        verifyProfileEvidence(
          c.selection,
          c.configuration,
          c.artifacts,
          withObservation(c, changed),
          attempt(c),
          { bytes },
        ),
      "evidence-mismatch",
    );
    for (const profile of ["read-update", "transactional"] as const) {
      const unit = context(profile),
        plain = observation(unit),
        ids = unit.owners[0]!.ids;
      const full = withUnitSessions(unit, plain, [
        unitSession(unit, ids[0]!, 0),
        unitSession(unit, ids[1]!, 1),
      ]);
      expect(assemble(unit, withObservation(unit, full)).qualificationReady).toBe(false);
      const baselineKey = profile === "read-update" ? "baselineReceipt" : "baseline";
      const variants: [string, (w: Data) => void][] = [
        [
          "invalid-report",
          (w) => {
            delete w[baselineKey];
          },
        ],
        [
          "invalid-report",
          (w) => {
            delete w.sessions;
          },
        ],
        [
          "invalid-report",
          (w) => {
            w.sessions = {};
          },
        ],
        [
          "configuration-mismatch",
          (w) => {
            w.sourceCommit = "0".repeat(40);
          },
        ],
        [
          "invalid-report",
          (w) => {
            rec(w[baselineKey]).extra = true;
          },
        ],
        [
          "configuration-mismatch",
          (w) => {
            (w.sessions as Data[]).pop();
          },
        ],
        [
          "invalid-report",
          (w) => {
            (w.sessions as Data[]).reverse();
          },
        ],
        [
          "invalid-report",
          (w) => {
            (w.sessions as Data[])[1] = { ...(w.sessions as Data[])[0] };
          },
        ],
        [
          "configuration-mismatch",
          (w) => {
            (w.sessions as Data[])[0]!.payloadIdentity = "0".repeat(64);
          },
        ],
      ];
      if (profile === "read-update")
        variants.push(
          [
            "configuration-mismatch",
            (w) => {
              rec(w.configuration).scope = "https://other.invalid/";
            },
          ],
          [
            "configuration-mismatch",
            (w) => {
              rec(w.baselineReceipt).configurationIdentity = "0".repeat(64);
            },
          ],
          [
            "configuration-mismatch",
            (w) => {
              rec(w.exactConfiguration).defaultCredentialRef = "other";
            },
          ],
          [
            "invalid-report",
            (w) => {
              rec(w.baselineReceipt).listenUrl = "http://127.0.0.1:41000/";
            },
          ],
        );
      else
        variants.push(
          [
            "configuration-mismatch",
            (w) => {
              rec(rec(w.baseline).served).sourceCommit = "0".repeat(40);
            },
          ],
          [
            "configuration-mismatch",
            (w) => {
              rec(w.baseline).runtimeConfigurationIdentity = "0".repeat(64);
            },
          ],
          [
            "configuration-mismatch",
            (w) => {
              rec(rec(w.baseline).operationTargets).batch = scope + "operations/other";
            },
          ],
          [
            "invalid-report",
            (w) => {
              (w.sessions as Data[])[0]!.listenUrl = "http://127.0.0.1:0/";
            },
          ],
          [
            "invalid-report",
            (w) => {
              (w.sessions as Data[])[0]!.sessionOrdinal = 1;
            },
          ],
          [
            "invalid-report",
            (w) => {
              (w.sessions as Data[])[0]!.reopen = true;
            },
          ],
        );
      for (const [expected, edit] of variants) {
        const changed = changeReport(full, edit);
        // Rehashed bytes force metadata validation rather than a stale expected-SHA failure.
        code(() => assemble(unit, withObservation(unit, changed)), expected);
      }
      const mismatchedReady = [...full.ready].reverse();
      code(
        () => assemble(unit, withObservation(unit, { ...full, ready: mismatchedReady })),
        "configuration-mismatch",
      );
      if (profile === "transactional") {
        const reopenPlan = unit.owners[0]!.bundle.manifest.scenarios.find((p) =>
          (p.actions ?? p.requests).some(
            (a) => "operation" in a && a.operation === "transactional.reopen",
          ),
        )!;
        const reopening = withUnitSessions(unit, plain, [
          unitSession(unit, reopenPlan.id, 0),
          unitSession(unit, reopenPlan.id, 1, true),
        ]);
        expect(assemble(unit, withObservation(unit, reopening)).claimEligible).toBe(false);
      }
    }
  },
  R21() {
    const c = context();
    for (const field of ["scenarioFilter", "scenarioSelection"]) {
      const o = changeReport(observation(c), (w) => {
        rec(w.result)[field] = field === "scenarioFilter" ? "partial" : ["partial"];
      });
      code(() => assemble(c, withObservation(c, o)), "invalid-report");
    }
  },
  R22() {
    const c = context();
    for (const extra of [false, true]) {
      const o = changeReport(observation(c), (w) => {
        const rs = rec(w.result).scenarios as Data[];
        if (extra) rs.push({ ...rs[0] });
        else rs.pop();
      });
      code(() => assemble(c, withObservation(c, o)), "invalid-report");
    }
    const o = changeReport(observation(c), (w) => {
      const rs = rec(w.result).scenarios as Data[],
        row = rs[0]!;
      delete row.category;
      row.state = "fail";
      row.actions = [{ action: { id: "unknown-action", family: "client" }, assertions: [] }];
    });
    code(() => assemble(c, withObservation(c, o)), "invalid-report");
    for (const variant of [
      "unknown-http",
      "duplicate-unknown-http",
      "duplicate-http",
      "duplicate-assertion",
      "wrong-assertion",
      "wrong-family",
      "duplicate-action",
    ]) {
      const bad = changeReport(observation(c), (w) => {
        const row = (rec(w.result).scenarios as Data[]).find((r) => r.id === c.owners[0]!.ids[0])!,
          plan = c.owners[0]!.bundle.manifest.scenarios[0]!;
        const rows = rec(w.result).scenarios as Data[],
          at = rows.indexOf(row);
        for (let i = 0; i < at; i++) {
          rows[i]!.category = "not-implemented";
        }
        delete row.category;
        row.state = "fail";
        const action = (plan.actions ?? plan.requests)[0]!;
        if (variant.includes("unknown-http"))
          row.exchanges = Array.from({ length: variant.startsWith("duplicate") ? 2 : 1 }, () => ({
            request: { id: "unknown-action" },
            assertions: [],
          }));
        else if (variant === "duplicate-http")
          row.exchanges = [0, 1].map(() => ({ request: { id: action.id }, assertions: [] }));
        else if (variant === "wrong-family" || variant === "duplicate-action")
          row.actions = Array.from({ length: variant === "duplicate-action" ? 2 : 1 }, () => ({
            action: { id: action.id, family: "invented-family" },
            assertions: [],
          }));
        else {
          const assertion = {
            id: variant === "wrong-assertion" ? "unknown-assertion" : action.assertions[0]!.id,
            passed: true,
          };
          if ("family" in action && action.family !== "http")
            row.actions = [
              {
                action: { id: action.id, family: action.family },
                assertions:
                  variant === "duplicate-assertion" ? [assertion, assertion] : [assertion],
              },
            ];
          else
            row.exchanges = [
              {
                request: { id: action.id },
                assertions:
                  variant === "duplicate-assertion" ? [assertion, assertion] : [assertion],
              },
            ];
        }
      });
      code(() => assemble(c, withObservation(c, bad)), "invalid-report");
    }
    const tx = context("transactional"),
      diagnostic = changeReport(observation(tx), (w) => {
        (rec(w.result).scenarios as Data[]).pop();
      });
    code(() => assemble(tx, withObservation(tx, diagnostic)), "invalid-report");
    const initial = prerequisiteData("initial"),
      recheck = prerequisiteData("recheck", undefined, true);
    for (const [unit, edits] of [
      [
        initial,
        [
          (r: Data) => {
            rec(r.prerequisiteFailure).id = "unknown-id";
          },
          (r: Data) => {
            rec(r.prerequisiteFailure).id = initial.c.owners[1]!.ids[0];
          },
          (r: Data) => {
            rec(r.prerequisiteFailure).id = initial.c.owners[0]!.ids.at(-1);
          },
          (r: Data) => {
            rec(r.prerequisiteFailure).phase = "other";
          },
          (r: Data) => {
            rec(r.prerequisiteFailure).phase = false;
          },
          (r: Data) => {
            rec(r.prerequisiteFailure).category = "deadline";
          },
          (r: Data) => {
            rec(r.prerequisiteFailure).requirements = [];
          },
          (r: Data) => {
            r.exchanges = [{ request: { id: "invented" }, assertions: [] }];
          },
        ],
      ],
      [
        recheck,
        [
          (r: Data) => {
            rec(r.prerequisiteFailure).category = "not-implemented";
          },
          (r: Data) => {
            (r.actions as Data[])[0]!.action = { id: "unknown", family: "client" };
          },
          (r: Data) => {
            (r.actions as Data[]).push((r.actions as Data[])[0]!);
          },
          (r: Data) => {
            (r.exchanges as Data[]).push((r.exchanges as Data[])[0]!);
          },
          (r: Data) => {
            (r.actions as Data[])[0]!.assertions = [{ id: "unknown", passed: true }];
          },
          (r: Data) => {
            (r.actions as Data[])[0]!.assertions = [{ id: "unknown", passed: "yes" }];
          },
        ],
      ],
    ] as const) {
      for (const edit of edits) {
        const bad = changeReport(unit.o, (w) =>
          edit((rec(w.result).scenarios as Data[]).find((r) => r.id === unit.plan.id)!),
        );
        code(() => assemble(unit.c, withObservation(unit.c, bad)), "invalid-report");
      }
    }
    for (const stop of ["category", "cleanup"] as const) {
      const unit = prerequisiteData("recheck", stop);
      const bad = changeReport(unit.o, (w) => {
        const rows = rec(w.result).scenarios as Data[],
          at = rows.findIndex((r) => r.id === unit.plan.id);
        reachedThrough(unit.c, [rows[at + 1]!], 0);
      });
      code(() => assemble(unit.c, withObservation(unit.c, bad)), "invalid-report");
    }
    const missing = missingExecutorData();
    for (const edit of [
      (r: Data) => {
        r.actions = [];
      },
      (r: Data) => {
        (r.exchanges as Data[])[0]!.transportError = { category: "deadline" };
      },
      (r: Data) => {
        delete (r.actions as Data[]).at(-1)!.executionError;
      },
      (r: Data) => {
        rec((r.actions as Data[]).at(-1)!.executionError).category = false;
      },
      (r: Data) => {
        rec((r.actions as Data[]).at(-1)!.executionError).message = {};
      },
      (r: Data) => {
        rec((r.actions as Data[]).at(-1)!.action).operation = "invented";
      },
      (r: Data) => {
        rec((r.actions as Data[]).at(-1)!.action).family = "invented";
      },
      (r: Data) => {
        rec((r.actions as Data[]).at(-1)!.action).id = "unknown";
      },
      (r: Data) => {
        (r.actions as Data[]).push({
          action: { id: "after-failure", family: "client", operation: "invented" },
          assertions: [],
        });
      },
    ]) {
      const bad = changeReport(missing.o, (w) =>
        edit((rec(w.result).scenarios as Data[]).find((r) => r.id === missing.plan.id)!),
      );
      code(() => assemble(missing.c, withObservation(missing.c, bad)), "invalid-report");
    }
  },
  R23() {
    const c = context(),
      e = JSON.parse(Buffer.from(serializeProfileEvidence(assemble(c))).toString());
    e.requirements[0].declaredWholeContract = !e.requirements[0].declaredWholeContract;
    code(
      () =>
        verifyProfileEvidence(
          c.selection,
          c.configuration,
          c.artifacts,
          absent(c),
          attempt(c),
          json(e),
        ),
      "evidence-mismatch",
    );
  },
  R24() {
    const c = context(),
      o = observation(c);
    code(
      () =>
        assemble(c, withObservation(c, { ...o, reportSource: { bytes: Buffer.alloc(4194305) } })),
      "capacity",
    );
    for (const junk of [
      Array.from({ length: 100 }, () => Array.from({ length: 1000 }, () => 0)),
      Array.from({ length: 10001 }, () => 0),
      nested(130),
    ]) {
      const bad = changeReport(o, (w) => {
        w.junk = junk;
      });
      code(() => assemble(c, withObservation(c, bad)), "capacity");
    }
    code(
      () =>
        assemble(
          c,
          withObservation(c, {
            ...o,
            reportSource: { bytes: Buffer.from([255]), label: "secret-label" },
            expectedReportSha256: sha(Buffer.from([255])),
          }),
        ),
      "invalid-source",
    );
    const large = {
      ...c,
      configuration: configuration("read-update", ruFixture, {
        scope: `https://reference.invalid/${"x".repeat(1048576)}/`,
      }),
    };
    code(() => serializeProfileEvidence(assemble(large)), "capacity");
  },
  L01() {
    const valid = unbound();
    expect(
      JSON.parse(Buffer.from(serializeProfileEvidence(valid, [], ruSet())).toString())
        .configuration,
    ).toBeNull();
    for (const edit of [
      (v: Data) => {
        delete v.configuration;
      },
      (v: Data) => {
        v.extra = true;
      },
      (v: Data) => {
        v.configuration = {};
      },
      (v: Data) => {
        v.kind = "unknown-kind";
      },
      (v: Data) => {
        v.attemptVersion = 2;
      },
      (v: Data) => {
        v.bindingStatus = "bound";
      },
      (v: Data) => {
        v.claimEligible = true;
      },
      (v: Data) => {
        v.qualificationReady = true;
      },
      (v: Data) => {
        v.requestedProfile = "transactional";
      },
      (v: Data) => {
        (v.observations as unknown[]).pop();
      },
      (v: Data) => {
        (v.observations as unknown[]).push(null);
      },
      (v: Data) => {
        (v.observations as unknown[])[0] = {};
      },
      (v: Data) => {
        (v.requiredIds as string[])[0] = "unknown.requirement";
      },
    ]) {
      const v = JSON.parse(JSON.stringify(valid));
      edit(v);
      code(() => serializeProfileEvidence(v, [], ruSet()), "invalid-source");
    }
    let calls = 0;
    const accessor = Object.defineProperty({ ...valid }, "configuration", {
      enumerable: true,
      get() {
        calls++;
        throw new Error("unit-unbound-private");
      },
    });
    code(() => serializeProfileEvidence(accessor, [], ruSet()), "invalid-source");
    expect(calls).toBe(0);
  },
  L02() {
    const v = unbound();
    v.attempt.owners[0]!.status = "running";
    v.attempt.owners[0]!.cleanup = "pending";
    const decoded = JSON.parse(Buffer.from(serializeProfileEvidence(v, [], ruSet())).toString());
    expect(decoded.observations.every((x: unknown) => x === null)).toBe(true);
    expect(decoded.configuration).toBeNull();
  },
  L03() {
    const c = context("transactional"),
      genuine = { result: { real: true }, sessions: [1], baseline: { real: true } };
    for (const input of [
      undefined,
      { bytes: Buffer.from("bad"), rawWrapperSha256: "0".repeat(64) },
    ]) {
      let failed = false;
      try {
        if (!input) throw Error();
        const o = observation(c);
        assemble(
          c,
          withObservation(c, {
            ...o,
            reportSource: { bytes: input.bytes },
            expectedReportSha256: input.rawWrapperSha256,
          }),
        );
      } catch {
        failed = true;
      }
      expect(failed).toBe(true);
      expect(genuine.result.real).toBe(true);
      const a = attempt(c, "failed");
      a.failureStages.push("credential-check");
      const e = assemble(c, absent(c), a);
      expect(rec((e.reports as unknown[])[0]).rawWrapperSha256).toBeNull();
      expect(e.qualificationReady).toBe(false);
    }
  },
  L04() {
    const c = context(),
      o = { ...observation(c), lifecycle: { cleanup: "reported-complete" as const } },
      e = assemble(c, withObservation(c, o));
    expect(
      rec((e.requirements as Data[]).find((r) => r.observation !== null)!.observation)
        .ownershipVerification,
    ).toBe("unverified");
    expect(e.qualificationReady).toBe(false);
  },
  L05() {
    const c = context(),
      o = observation(c),
      a = attempt(c, "failed");
    a.failureStages.push("cleanup");
    const e = assemble(c, withObservation(c, { ...o, lifecycle: { cleanup: "failed" } }), a);
    expect(rec(e.attempt).outcome).toBe("failed");
    expect(
      rec((e.requirements as Data[]).find((r) => r.observation !== null)!.observation)
        .rawWrapperSha256,
    ).toBe(o.expectedReportSha256);
    const asserted = attempt(c, "failed");
    asserted.failureStages.push("assertion");
    asserted.owners[0]!.status = "failed";
    asserted.owners[0]!.cleanup = "reported-complete";
    const assertionEvidence = assemble(
      c,
      withObservation(c, { ...o, lifecycle: { cleanup: "reported-complete" } }),
      asserted,
    );
    expect(rec(assertionEvidence.attempt).failureStages).toEqual(["assertion"]);
    expect(rec((assertionEvidence.reports as Data[])[0]).reportedCleanupState).toBe(
      "reported-complete",
    );
    const original = new Error("unit-original-legacy");
    let observed: unknown;
    try {
      try {
        throw original;
      } finally {
        /* Actual original hook errors are not carrier refusals. */
      }
    } catch (x) {
      observed = x;
    }
    expect(observed).toBe(original);
  },
  L06() {
    const v = unbound("failed");
    v.attempt.failureStages.push("execution");
    const e = JSON.parse(Buffer.from(serializeProfileEvidence(v, [], ruSet())).toString());
    expect(e.configuration).toBeNull();
    expect(e.observations).toEqual(Array(130).fill(null));
    expect(e.attempt.outcome).toBe("failed");
  },
  async L07() {
    const { c, o } = await aborted("transactional"),
      a = attempt(c, "failed");
    a.failureStages.push("execution");
    a.owners[0]!.status = "aborted";
    const e = assemble(c, withObservation(c, o), a);
    expect(rec(e.attempt).outcome).toBe("failed");
    expect(
      (e.requirements as Data[])
        .filter((r) => r.observation !== null)
        .every((r) => rec(r.observation).authoredAssertionsAllPassed === null),
    ).toBe(true);
  },
  L08() {
    const c = context(),
      e = assemble(c, withObservation(c, observation(c)), attempt(c, "failed"));
    expect((e.reports as Data[])[0]!.rawWrapperSha256).not.toBeNull();
    expect((e.reports as Data[])[1]!.rawWrapperSha256).toBeNull();
    expect(e.qualificationReady).toBe(false);
    for (const [phase, stop, retain] of [
      ["initial", undefined, false],
      ["recheck", undefined, false],
      ["recheck", undefined, true],
      ["recheck", "category", false],
      ["recheck", "cleanup", true],
    ] as const) {
      const unit = prerequisiteData(phase, stop, retain),
        projected = assemble(unit.c, withObservation(unit.c, unit.o));
      const row = (projected.requirements as Data[]).find((r) => r.id === unit.plan.id)!;
      expect(rec(row.observation).authoredAssertionsAllPassed).toBeNull();
      expect(rec(row.observation).state).toBe("harness-error");
      expect(projected.qualificationReady).toBe(false);
      const original = JSON.parse(Buffer.from(unit.o.reportSource.bytes).toString());
      const retainedRow = rec(original.result).scenarios as Data[];
      if (retain)
        expect(
          (retainedRow.find((r) => r.id === unit.plan.id)!.actions as Data[]).length,
        ).toBeGreaterThan(0);
    }
    const missing = missingExecutorData(),
      retained = assemble(missing.c, withObservation(missing.c, missing.o));
    expect(
      rec((retained.requirements as Data[]).find((r) => r.id === missing.plan.id)!.observation)
        .authoredAssertionsAllPassed,
    ).toBeNull();
  },
  L09() {
    const c = context(),
      o = observation(c),
      bad = { ...o, runtimeConfigurationIdentity: "0".repeat(64) };
    code(() => assemble(c, withObservation(c, bad)), "configuration-mismatch");
    expect(o.runtimeConfigurationIdentity).toBe("f".repeat(64));
  },
  L10() {
    const directory = mkdtempSync(path.join(tmpdir(), "profile-data-"));
    try {
      const base = path.join(directory, "fresh");
      for (const profile of ["read-update", "transactional"] as const) {
        const c = context(profile),
          e = assemble(c);
        const writer = createPrivateProfileSidecarWriter({ base, profile, credentialValues: [] });
        writer.writeAttempt(e);
        const file = `${base}.${profile}.profile.attempt.json`;
        expect(lstatSync(file).mode & 0o777).toBe(0o600);
        writer.writeAttempt(e);
        writer.writeSummary(e);
        expect(() =>
          createPrivateProfileSidecarWriter({ base, profile, credentialValues: [] }),
        ).toThrow("persistence refused");
        writeFileSync(file, "foreign");
        expect(() => writer.writeAttempt(e)).toThrow("persistence refused");
        expect(readFileSync(file, "utf8")).toBe("foreign");
        for (const suffix of ["profile.attempt.json", "profile.json"]) {
          const stale = path.join(directory, `${profile}-${suffix}`),
            destination = `${stale}.${profile}.${suffix}`;
          writeFileSync(destination, "stale");
          expect(() =>
            createPrivateProfileSidecarWriter({ base: stale, profile, credentialValues: [] }),
          ).toThrow();
          expect(readFileSync(destination, "utf8")).toBe("stale");
        }
        for (const kind of ["directory", "symlink"] as const) {
          const blocked = path.join(directory, `${profile}-${kind}`);
          const destination = `${blocked}.${profile}.profile.attempt.json`;
          if (kind === "directory") mkdirSync(destination);
          else symlinkSync(file, destination);
          expect(() =>
            createPrivateProfileSidecarWriter({ base: blocked, profile, credentialValues: [] }),
          ).toThrow("persistence refused");
          expect(lstatSync(destination).isSymbolicLink()).toBe(kind === "symlink");
        }
        const replacementBase = path.join(directory, `${profile}-write-replacement`);
        const replacement = createPrivateProfileSidecarWriter({
          base: replacementBase,
          profile,
          credentialValues: [],
        });
        replacement.writeAttempt(e);
        const replacementFile = `${replacementBase}.${profile}.profile.attempt.json`;
        const replacementBytes = readFileSync(replacementFile);
        ioFaults.write = true;
        expect(() => replacement.writeAttempt(e)).toThrow("persistence refused");
        expect(readFileSync(replacementFile)).toEqual(replacementBytes);
        const race = path.join(directory, `${profile}-race`),
          first = createPrivateProfileSidecarWriter({ base: race, profile, credentialValues: [] }),
          second = createPrivateProfileSidecarWriter({ base: race, profile, credentialValues: [] });
        first.writeAttempt(e);
        expect(() => second.writeAttempt(e)).toThrow();
        const partial = path.join(directory, `${profile}-partial`),
          own = createPrivateProfileSidecarWriter({ base: partial, profile, credentialValues: [] });
        own.writeAttempt(e);
        rmSync(`${partial}.${profile}.profile.attempt.json`);
        expect(() => own.writeAttempt(e)).toThrow();
        const writeBase = path.join(directory, `${profile}-write`),
          broken = createPrivateProfileSidecarWriter({
            base: writeBase,
            profile,
            credentialValues: [],
          });
        ioFaults.write = true;
        expect(() => broken.writeAttempt(e)).toThrow("persistence refused");
        expect(existsSync(`${writeBase}.${profile}.profile.attempt.json`)).toBe(true);
        expect(() =>
          createPrivateProfileSidecarWriter({ base: writeBase, profile, credentialValues: [] }),
        ).toThrow();
        for (const cleanup of [false, true]) {
          const renameBase = path.join(directory, `${profile}-rename-${cleanup}`),
            rename = createPrivateProfileSidecarWriter({
              base: renameBase,
              profile,
              credentialValues: [],
            });
          rename.writeAttempt(e);
          const destination = `${renameBase}.${profile}.profile.attempt.json`,
            previous = readFileSync(destination);
          ioFaults.rename = true;
          ioFaults.cleanup = cleanup;
          let caught: unknown;
          try {
            rename.writeAttempt(e);
          } catch (error) {
            caught = error;
          }
          expect(caught).toBeInstanceOf(cleanup ? AggregateError : Error);
          expect((caught as Error).message).toBe("private profile evidence persistence refused");
          expect(readFileSync(destination)).toEqual(previous);
          if (cleanup)
            expect(
              (caught as AggregateError).errors.every(
                (error: Error) => error.message === "private profile evidence persistence refused",
              ),
            ).toBe(true);
        }
      }
      const legacy = path.join(directory, "legacy");
      writeFileSync(legacy, "original");
      expect(() => createArtifactEvidenceWriter(ruSet(), "a".repeat(40), legacy, [])).toThrow();
      expect(readFileSync(legacy, "utf8")).toBe("original");
      expect(existsSync(`${base}.profile.json`)).toBe(false);
    } finally {
      ioFaults.write = false;
      ioFaults.rename = false;
      ioFaults.cleanup = false;
      rmSync(directory, { recursive: true, force: true });
    }
  },
};
function nested(n: number): unknown {
  let value: unknown = 0;
  for (let i = 0; i < n; i++) value = { next: value };
  return value;
}
const mismatchFields: Record<
  string,
  Partial<Parameters<typeof bindServedProfileConfiguration>[0]>
> = {
  R01: { scope: "https://other.invalid/acme/" },
  R02: { profile: "transactional" },
  R04: { sourceCommit: "0".repeat(40) },
  R05: { installedPayloadDigest: "0".repeat(64) },
  R06: { installedContractsDigest: "0".repeat(64) },
  R07: { policyDigest: "0".repeat(64) },
  R08: { authenticationDigest: "0".repeat(64) },
  R09: { fixture: json({ ...JSON.parse(ruFixture.bytes.toString()), id: "different-fixture" }) },
};
for (const [id, fields] of Object.entries(mismatchFields))
  dataCases[id] = () => {
    const c = context(),
      o = { ...observation(c), configuration: configuration("read-update", ruFixture, fields) };
    code(() => assemble(c, withObservation(c, o)), "configuration-mismatch");
  };
dataCases.R03 = () => {
  const c = context(),
    selection = loadProfileContractSelection("read-update", true);
  code(
    () => assembleProfileEvidence(selection, c.configuration, c.artifacts, absent(c), attempt(c)),
    "unsupported-selection",
  );
};
const credential = "private?credential-tripwire-0123456789";
const base64 = Buffer.from(credential).toString("base64");
const encodings = [
  credential,
  encodeURIComponent(credential),
  base64,
  Buffer.from(credential).toString("base64url"),
  Buffer.from(credential).toString("hex"),
  [...credential].map((c) => `\\u${c.charCodeAt(0).toString(16).padStart(4, "0")}`).join(""),
  encodeURIComponent(base64),
  Buffer.from(encodeURIComponent(credential)).toString("base64"),
];
encodings.forEach((encoding, n) => {
  dataCases[`S${String(n + 1).padStart(2, "0")}`] = () => {
    const c = context(),
      safe = encodeURIComponent(encodeURIComponent(encoding));
    c.configuration = configuration("read-update", ruFixture, {
      scope: `https://reference.invalid/${safe}/`,
    });
    const e = assemble(c);
    code(() => serializeProfileEvidence(e, [credential]), "credential-check");
  };
});
async function aborted(profile: "read-update" | "transactional") {
  const c = context(profile),
    owner = c.owners[0]!,
    mode = exact(c);
  let calls = 0;
  const forbidden = async () => {
    calls++;
    throw Error("no product execution");
  };
  const options = {
    scope,
    profile,
    seed: owner.bundle.fixture.seed,
    artifactBundle: owner.bundle,
    signal: AbortSignal.abort(),
    declaredTargetLabel: "already-aborted-pure-data",
    schemaValidator,
    execute: forbidden,
    exactExecution: { execute: forbidden, configuration: mode },
    harness: { prepare: forbidden, cleanup: async () => {} },
  };
  const result =
    profile === "transactional"
      ? await runCumulativeTransactionalMatrix(options, {
          current() {
            throw Error("no session");
          },
          reopen: forbidden,
        })
      : await runConformanceMatrix(options);
  expect(calls).toBe(0);
  expect(
    result.scenarios.every(
      (r) => r.state === "harness-error" && r.category === "not-run" && r.exchanges.length === 0,
    ),
  ).toBe(true);
  const o = observation(c, 0, result),
    e = assemble(c, withObservation(c, o));
  expect(
    (e.requirements as Data[])
      .filter((r) => r.observation !== null)
      .every((r) => rec(r.observation).authoredAssertionsAllPassed === null),
  ).toBe(true);
  if (profile === "transactional") {
    expect(result.scenarios).toHaveLength(256);
    expect(e.nonRequiredReportRows as unknown[]).toHaveLength(5);
  }
  return { c, o, result };
}
// Exactly48 IDs, one Vitest callback apiece; repository default5s, CLI workers chosen by the release.
for (const [id, work] of Object.entries(dataCases).sort(([a], [b]) => a.localeCompare(b)))
  it(`private profile data ${id}`, work);
