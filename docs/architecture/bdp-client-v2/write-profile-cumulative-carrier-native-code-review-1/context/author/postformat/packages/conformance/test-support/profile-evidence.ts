/** Private, pure partial-evidence carrier. Never admission or whole-clause proof. */
import { createHash } from "node:crypto";
import { types } from "node:util";
import { canonicalJson } from "../src/canonical-json.js";
import {
  ConformanceArtifactBundleError,
  copySourceBytes,
  parseBoundedJson,
  createCumulativeTransactionalArtifactBundle,
  type ConformanceArtifactSource,
  type ConformanceArtifactBundle,
} from "../src/artifact-bundle.js";
import {
  assertCanonicalProfileContractSelection,
  type ProfileContractSelection,
} from "../src/profile-contract-selection.js";
import { selectApplicableScenariosForProfile } from "../src/selection.js";
import {
  assertSameServedProfileConfiguration,
  ownData,
  planProfileExecution,
  type ServedProfileConfiguration,
} from "./profile-contract-plan.js";
import {
  assertReadUpdateArtifactSet,
  copyReadUpdateArtifactSetSources,
  type ReadUpdateArtifactSet,
} from "./read-update-artifact-set.js";
import { installedContractsDigest } from "../../../apps/bdptest/test-support/installed-contract-identity.js";

type Data = Record<string, unknown>;
type Artifacts =
  | { readonly kind: "read-update"; readonly set: ReadUpdateArtifactSet }
  | {
      readonly kind: "transactional";
      readonly manifest: ConformanceArtifactSource;
      readonly fixture: ConformanceArtifactSource;
      readonly coverage: ConformanceArtifactSource;
    };
type Observation = {
  readonly ownerId: string;
  readonly sourceCommit: string;
  readonly configuration: ServedProfileConfiguration;
  readonly runtimeConfigurationIdentity: string;
  readonly exactConfiguration: unknown;
  readonly ready: readonly unknown[];
  readonly reportSource: ConformanceArtifactSource;
  readonly expectedReportSha256: string;
  readonly lifecycle: { readonly cleanup: "pending" | "reported-complete" | "failed" | "unknown" };
};
type Attempt = {
  readonly outcome: "running" | "failed" | "observed";
  readonly failureStages: readonly ("execution" | "assertion" | "cleanup" | "credential-check")[];
  readonly owners: readonly {
    readonly id: string;
    readonly status: string;
    readonly cleanup: string;
    readonly evidence: string;
  }[];
  readonly sourceDirty: boolean | null;
};
const assembled = new WeakSet<object>();
const MAX_SOURCE = 4194304;
const MAX_SUMMARY = 1048576;
const LIMITATIONS = [
  "source-private carrier only; no public admission or proof token",
  "trusted-source identity equality does not prove run truth",
  "source/historical/data-unit observations never current installed qualification",
  "reported cleanup is not verified child/owner closure",
  "final combined basis unbound; incoming startup/D3a not adopted",
] as const;
const UNRESOLVED = [
  "D2-S1-G2",
  "D3-cold-listener-admission",
  "D4-administration-retained-copies-restore",
  "final-combined-source-configuration-installed-provenance",
  "whole-governing-clause-review",
];
class ProfileEvidenceError extends Error {
  constructor(readonly code: string) {
    super("private profile evidence refused");
    this.name = "ProfileEvidenceError";
  }
}
function refuse(code = "invalid-source"): never {
  throw new ProfileEvidenceError(code);
}
function boundary<T>(work: () => T): T {
  try {
    return work();
  } catch (error) {
    if (error instanceof ProfileEvidenceError) throw error;
    return refuse();
  }
}
function record(value: unknown): Data {
  if (!value || typeof value !== "object" || Array.isArray(value) || types.isProxy(value))
    return refuse();
  return ownData(value, Object.keys(value));
}
function array(value: unknown): unknown[] {
  if (!Array.isArray(value) || types.isProxy(value)) return refuse();
  const ds = Object.getOwnPropertyDescriptors(value);
  if (Reflect.ownKeys(ds).length !== value.length + 1 || value.length > 10000)
    return refuse("capacity");
  return Array.from({ length: value.length }, (_, n) => {
    const d = ds[String(n)];
    if (!d || !("value" in d) || !d.enumerable) return refuse();
    return d.value;
  });
}
function text(value: unknown): string {
  if (typeof value !== "string") return refuse();
  return value;
}
function equal(a: unknown, b: unknown, code = "configuration-mismatch") {
  if (canonicalJson(a) !== canonicalJson(b)) refuse(code);
}
function digest(value: string | Uint8Array) {
  return createHash("sha256").update(value).digest("hex");
}
function frozen<T>(value: T): T {
  if (value && typeof value === "object") {
    for (const x of Object.values(value)) frozen(x);
    Object.freeze(value);
  }
  return value;
}
function safeData(value: unknown, depth = 0, budget = { nodes: 0 }): unknown {
  if (++budget.nodes > 100000 || depth > 128) return refuse("capacity");
  if (value === null || typeof value === "string" || typeof value === "boolean") return value;
  if (typeof value === "number" && Number.isFinite(value)) return value;
  if (Array.isArray(value)) return array(value).map((x) => safeData(x, depth + 1, budget));
  const captured = record(value);
  return Object.fromEntries(
    Object.entries(captured).map(([k, x]) => [k, safeData(x, depth + 1, budget)]),
  );
}
function source(input: unknown) {
  const own = ownData(input, ["bytes", "label"], ["bytes"]);
  try {
    return copySourceBytes({ bytes: own.bytes as Uint8Array }, "manifest", MAX_SOURCE);
  } catch (error) {
    return artifactError(error);
  }
}
function artifactError(error: unknown): never {
  const capacity = [
    "manifest source must not exceed 4194304 bytes",
    "private-profile-report: JSON must not exceed 100000 nodes",
    "private-profile-report: JSON must not exceed depth 128",
    "private-profile-report: a JSON container must not exceed 10000 entries",
  ];
  return refuse(
    error instanceof ConformanceArtifactBundleError && capacity.includes(error.message)
      ? "capacity"
      : "invalid-source",
  );
}
function parsed(bytes: Uint8Array): unknown {
  try {
    return parseBoundedJson(bytes, "manifest", "private-profile-report");
  } catch (error) {
    return artifactError(error);
  }
}
function captureAttempt(input: Attempt): Data {
  const a = ownData(input, ["outcome", "failureStages", "owners", "sourceDirty"]);
  if (
    !["running", "failed", "observed"].includes(text(a.outcome)) ||
    (a.sourceDirty !== null && typeof a.sourceDirty !== "boolean")
  )
    refuse();
  const failureStages = array(a.failureStages).map(text);
  if (
    new Set(failureStages).size !== failureStages.length ||
    failureStages.some(
      (x) => !["execution", "assertion", "cleanup", "credential-check"].includes(x),
    )
  )
    refuse();
  const owners = array(a.owners).map((x) => {
    const r = ownData(x, ["id", "status", "cleanup", "evidence"]);
    if (
      !["unobserved", "running", "returned", "failed", "aborted"].includes(text(r.status)) ||
      !["not-started", "pending", "reported-complete", "failed", "unknown"].includes(
        text(r.cleanup),
      ) ||
      !["absent", "retained", "withheld-credential"].includes(text(r.evidence))
    )
      refuse();
    return { id: text(r.id), status: r.status, cleanup: r.cleanup, evidence: r.evidence };
  });
  return { outcome: a.outcome, failureStages, owners, sourceDirty: a.sourceDirty };
}
function model(
  selection: ProfileContractSelection,
  configuration: ServedProfileConfiguration,
  input: Artifacts,
) {
  assertCanonicalProfileContractSelection(selection);
  if (!["read-update", "transactional"].includes(selection.profile) || selection.history)
    refuse("unsupported-selection");
  assertSameServedProfileConfiguration(configuration, configuration);
  equal([configuration.profile, configuration.history], [selection.profile, selection.history]);
  const capturedArtifacts = ownData(
    input,
    ["kind", "set", "manifest", "fixture", "coverage"],
    ["kind"],
  );
  if (capturedArtifacts.kind !== "read-update" && capturedArtifacts.kind !== "transactional")
    refuse("unsupported-selection");
  let owners: {
    id: string;
    bundle: ConformanceArtifactBundle;
    ids: readonly string[];
    manifest: ConformanceArtifactSource;
  }[];
  let coverage: Data[];
  let artifactDigests: unknown;
  if (capturedArtifacts.kind === "read-update") {
    if (input.kind !== "read-update") return refuse();
    ownData(input, ["kind", "set"]);
    assertReadUpdateArtifactSet(input.set);
    equal(input.set.selection, selection, "invalid-source");
    const sources = copyReadUpdateArtifactSetSources(input.set);
    owners = input.set.shards.map((s) => ({
      id: s.id,
      bundle: s.bundle,
      ids: s.scenarioIds,
      manifest: sources.manifests[s.source]!,
    }));
    coverage = input.set.coverage.rows.map((r) => ({ ...r }));
    artifactDigests = input.set.digests;
  } else {
    if (input.kind !== "transactional") return refuse();
    ownData(input, ["kind", "manifest", "fixture", "coverage"]);
    if (input.kind !== "transactional" || selection.profile !== "transactional")
      refuse("unsupported-selection");
    const manifest = { bytes: source(input.manifest) },
      fixture = { bytes: source(input.fixture) };
    const bundle = createCumulativeTransactionalArtifactBundle(selection, { manifest, fixture });
    const raw = source(input.coverage),
      c = ownData(parsed(raw), [
        "coverageVersion",
        "claimEligible",
        "status",
        "transportAdmission",
        "rows",
      ]);
    if (
      c.coverageVersion !== 1 ||
      c.claimEligible !== false ||
      c.status !== "bounded-observation" ||
      c.transportAdmission !== "private-exact-available"
    )
      refuse("invalid-ownership");
    coverage = array(c.rows).map((x) => ownData(x, ["id", "wholeContract", "reason", "bindings"]));
    equal(
      coverage.map((r) => r.id),
      bundle.manifest.scenarios.map((r) => r.id),
      "invalid-ownership",
    );
    if (coverage.some((r) => r.wholeContract !== false)) refuse("invalid-ownership");
    owners = [
      {
        id: "transactional-baseline",
        bundle,
        ids: bundle.manifest.scenarios.map((r) => r.id),
        manifest,
      },
    ];
    artifactDigests = { ...bundle.digests, coverageDigest: digest(raw) };
  }
  if (owners.length > 2) refuse("capacity");
  const claimed = new Map<string, string>();
  for (const owner of owners) {
    equal(owner.bundle.digests.fixtureDigest, configuration.fixtureDigest);
    for (const id of owner.ids) {
      if (claimed.has(id) || !selection.required.some((r) => r.id === id))
        refuse("invalid-ownership");
      claimed.set(id, owner.id);
    }
    const positions = new Map(owner.ids.map((id, n) => [id, n]));
    for (const row of owner.bundle.manifest.scenarios) {
      const joined = (row.actions ?? row.requests).flatMap((a) =>
        a.assertions.map((x) => `${a.id}${selection.profile === "read-update" ? "." : ":"}${x.id}`),
      );
      if (new Set(joined).size !== joined.length) refuse("invalid-ownership");
      const c = coverage.find((r) => r.id === row.id);
      if (!c) refuse("invalid-ownership");
      equal(c.bindings, [{ scenario: row.id, assertions: joined }], "invalid-ownership");
      for (const a of row.actions ?? row.requests)
        if (
          a.prerequisiteScenario !== undefined &&
          (positions.get(a.prerequisiteScenario) ?? Infinity) >= positions.get(row.id)!
        )
          refuse("invalid-prerequisite");
    }
  }
  return { owners, coverage, claimed, artifactDigests };
}
function joinedOutcome(
  row: Data,
  plan: ConformanceArtifactBundle["manifest"]["scenarios"][number] | undefined,
) {
  if (!plan || ["not-run", "not-implemented"].includes(String(row.category))) return null;
  const lanes = {
    http: array(row.exchanges),
    other: row.actions === undefined ? [] : array(row.actions),
  };
  const authored = plan.actions ?? plan.requests;
  const httpRows = lanes.http.map(record),
    other = lanes.other.map(record);
  // A client action may emit a same-ID incidental exchange; it cannot supply
  // the client action's assertion lane. Truly unknown/duplicate IDs still refuse.
  const httpIds = httpRows.map((x) => text(record(x.request).id));
  if (
    new Set(httpIds).size !== httpIds.length ||
    httpIds.some((id) => !authored.some((a) => a.id === id))
  )
    refuse("invalid-report");
  for (const row of httpRows) {
    const assertions = array(row.assertions).map(record);
    if (
      new Set(assertions.map((x) => x.id)).size !== assertions.length ||
      assertions.some((x) => typeof x.id !== "string" || typeof x.passed !== "boolean")
    )
      refuse("invalid-report");
  }
  if (
    new Set(other.map((x) => record(x.action).id)).size !== other.length ||
    other.some((x) => {
      const actual = record(x.action),
        declared = authored.find((a) => a.id === actual.id);
      return (
        !declared ||
        !("family" in declared) ||
        declared.family === "http" ||
        declared.family !== actual.family
      );
    })
  )
    refuse("invalid-report");
  const values: (boolean | null)[] = [];
  for (const action of plan.actions ?? plan.requests) {
    const http = !("family" in action) || action.family === "http";
    const matching = lanes[http ? "http" : "other"]
      .map(record)
      .filter((x) => (http ? record(x.request).id : record(x.action).id) === action.id);
    if (matching.length > 1) refuse("invalid-report");
    const assertions = matching.length ? array(matching[0]!.assertions).map(record) : [];
    if (
      new Set(assertions.map((x) => x.id)).size !== assertions.length ||
      assertions.some(
        (x) => !action.assertions.some((a) => a.id === x.id) || typeof x.passed !== "boolean",
      )
    )
      refuse("invalid-report");
    for (const authored of action.assertions)
      values.push(
        (assertions.find((x) => x.id === authored.id)?.passed as boolean | undefined) ?? null,
      );
  }
  return values.some((x) => x === false)
    ? false
    : values.some((x) => x === null) || !values.length
      ? null
      : true;
}
/** Complete selected inventory, even when every observation is absent. */
export function assembleProfileEvidence(
  selection: ProfileContractSelection,
  configuration: ServedProfileConfiguration,
  artifacts: Artifacts,
  observations: readonly (Observation | null)[],
  attempt: Attempt,
): Data {
  return boundary(() => {
    const m = model(selection, configuration, artifacts);
    const attempts = captureAttempt(attempt);
    equal(
      array(attempts.owners).map((x) => record(x).id),
      m.owners.map((x) => x.id),
      "invalid-source",
    );
    const supplied = array(observations);
    if (supplied.length !== m.owners.length) refuse("invalid-report");
    const reports: Data[] = [],
      mapped = new Map<string, Data>(),
      nonRequiredReportRows: Data[] = [];
    m.owners.forEach((owner, n) => {
      const observed = supplied[n];
      if (observed === null) {
        reports.push({
          ownerId: owner.id,
          rawWrapperSha256: null,
          canonicalResultSha256: null,
          selectedIds: null,
          resultIds: null,
          ownershipVerification: "unverified",
        });
        return;
      }
      const o = ownData(observed, [
        "ownerId",
        "sourceCommit",
        "configuration",
        "runtimeConfigurationIdentity",
        "exactConfiguration",
        "ready",
        "reportSource",
        "expectedReportSha256",
        "lifecycle",
      ]);
      equal(o.ownerId, owner.id, "invalid-report");
      equal(o.sourceCommit, configuration.sourceCommit);
      assertSameServedProfileConfiguration(
        o.configuration as ServedProfileConfiguration,
        o.configuration as ServedProfileConfiguration,
      );
      equal(configuration, o.configuration);
      if (!/^[a-f0-9]{64}$/.test(text(o.runtimeConfigurationIdentity)))
        refuse("configuration-mismatch");
      const lifecycle = ownData(o.lifecycle, ["cleanup"]);
      if (!["pending", "reported-complete", "failed", "unknown"].includes(text(lifecycle.cleanup)))
        refuse();
      const bytes = source(o.reportSource);
      equal(digest(bytes), o.expectedReportSha256, "invalid-report");
      const wrapper = record(parsed(bytes)),
        result = record(wrapper.result);
      if (wrapper.sourceCommit !== configuration.sourceCommit) refuse("configuration-mismatch");
      if (wrapper.configuration !== undefined) equal(wrapper.configuration, configuration);
      if (
        result.reportVersion !== 4 ||
        result.claimEligible !== false ||
        result.scenarioFilter !== undefined
      )
        refuse("invalid-report");
      equal(
        [result.scope, result.profile, result.seed, result.artifacts],
        [
          configuration.scope,
          configuration.profile,
          owner.bundle.fixture.seed,
          owner.bundle.digests,
        ],
        "invalid-report",
      );
      const selected =
        artifacts.kind === "read-update" && n > 0
          ? [...owner.ids]
          : selectApplicableScenariosForProfile(owner.bundle.catalog, selection.profile).map(
              (r) => r.id,
            );
      equal(
        result.scenarioSelection ?? null,
        artifacts.kind === "read-update" && n > 0 ? owner.ids : null,
        "invalid-report",
      );
      equal(result.selectedScenarioIds, selected, "invalid-report");
      const rows = array(result.scenarios).map(record);
      equal(
        rows.map((r) => r.id),
        selected,
        "invalid-report",
      );
      const exact = record(safeData(o.exactConfiguration));
      equal([exact.scope, exact.profile], [configuration.scope, configuration.profile]);
      const routeDigest = digest(
        canonicalJson({
          scope: exact.scope,
          profile: exact.profile,
          routes: exact.routes,
          maximumRequestHeaderBytes: exact.maximumRequestHeaderBytes,
          maximumRequestBodyBytes: exact.maximumRequestBodyBytes,
        }),
      );
      equal(
        record(result.declarations).exactConfigurationDigest,
        routeDigest,
        "configuration-mismatch",
      );
      const ready = array(o.ready).map(record);
      if (!ready.length) refuse("invalid-report");
      for (const r of ready) {
        if (r.served !== undefined) equal(r.served, configuration);
        else {
          equal(
            [
              r.scope,
              r.profile,
              r.history,
              r.payloadIdentity,
              installedContractsDigest(r.artifactIdentity),
              r.policyIdentity,
              r.authenticationIdentity,
            ],
            [
              configuration.scope,
              configuration.profile,
              configuration.history,
              configuration.installedPayloadDigest,
              configuration.installedContractsDigest,
              configuration.policyDigest,
              configuration.authenticationDigest,
            ],
          );
        }
        equal(
          r.runtimeConfigurationIdentity ?? r.configurationIdentity,
          o.runtimeConfigurationIdentity,
        );
      }
      let stopped = false;
      rows.forEach((r, index) => {
        if (
          !["pass", "fail", "not-applicable", "unsupported-profile", "harness-error"].includes(
            text(r.state),
          )
        )
          refuse("invalid-report");
        if (
          r.category !== undefined &&
          ![
            "aborted",
            "cleanup",
            "deadline",
            "not-implemented",
            "not-run",
            "observation-limit",
            "out-of-scope-target",
            "wire-observation-unavailable",
            "exact-configuration",
            "exact-observation",
          ].includes(text(r.category))
        )
          refuse("invalid-report");
        const metadata = owner.bundle.catalog.scenarios.find((x) => x.id === r.id);
        if (!metadata) refuse("invalid-report");
        equal(r.requiredProfile, metadata.requiredProfile, "invalid-report");
        const plan = owner.bundle.manifest.scenarios.find((p) => p.id === r.id);
        if (plan && r.category === "not-implemented") refuse("invalid-report");
        if (r.category === "not-run") stopped = true;
        if (stopped && r.category !== "not-run") refuse("invalid-report");
        if (
          !plan &&
          (r.state !== "harness-error" ||
            !["not-run", "not-implemented"].includes(text(r.category)))
        )
          refuse("invalid-report");
        if (
          ["not-run", "not-implemented"].includes(String(r.category)) &&
          (array(r.exchanges).length || (r.actions !== undefined && array(r.actions).length))
        )
          refuse("invalid-report");
        const projection = {
          ownerId: owner.id,
          rawWrapperSha256: digest(bytes),
          canonicalResultSha256: digest(canonicalJson(result)),
          rowIndex: index,
          state: r.state,
          category: r.category ?? null,
          authoredAssertionInventoryDigest: digest(
            canonicalJson(
              plan
                ? (plan.actions ?? plan.requests).map((a) => ({
                    id: a.id,
                    family: "family" in a ? a.family : "http",
                    assertions: a.assertions.map((x) => x.id),
                  }))
                : [],
            ),
          ),
          authoredAssertionsAllPassed: joinedOutcome(r, plan),
          reportedCleanupState: lifecycle.cleanup,
          ownershipVerification: "unverified",
        };
        if (m.claimed.get(text(r.id)) === owner.id) mapped.set(text(r.id), projection);
        else if (!selection.required.some((x) => x.id === r.id))
          nonRequiredReportRows.push({ id: r.id, ...projection });
      });
      reports.push({
        ownerId: owner.id,
        rawWrapperSha256: digest(bytes),
        canonicalResultSha256: digest(canonicalJson(result)),
        selectedIds: selected,
        resultIds: rows.map((r) => r.id),
        runtimeConfigurationIdentity: o.runtimeConfigurationIdentity,
        exactConfigurationDigest: routeDigest,
        reportedCleanupState: lifecycle.cleanup,
        ownershipVerification: "unverified",
      });
    });
    const plans = m.owners.map((owner) =>
      planProfileExecution(selection, configuration, { configuration, manifest: owner.manifest }),
    );
    const empty = planProfileExecution(selection, configuration);
    const envelope = frozen({
      carrierVersion: 1,
      kind: "private-profile-evidence",
      claimEligible: false,
      qualificationReady: false,
      profile: selection.profile,
      history: false,
      configuration: safeData(configuration),
      catalogDigests: selection.catalogDigests,
      artifactDigests: m.artifactDigests,
      requiredIds: selection.required.map((r) => r.id),
      retiredIds: selection.retiredIds,
      requirements: selection.required.map((r) => {
        const ownerId = m.claimed.get(r.id) ?? null,
          ownerIndex = m.owners.findIndex((o) => o.id === ownerId);
        const p = (ownerIndex < 0 ? empty : plans[ownerIndex]!).requirements.find(
          (x) => x.id === r.id,
        )!;
        return {
          ...p,
          requiredProfile: r.requiredProfile,
          ownerId,
          declaredWholeContract: m.coverage.find((c) => c.id === r.id)?.wholeContract === true,
          wholeClauseQualification: "unestablished",
          observation: mapped.get(r.id) ?? null,
        };
      }),
      nonRequiredReportRows,
      reports,
      attempt: attempts,
      unresolvedConditions: UNRESOLVED,
      limitations: LIMITATIONS,
    });
    assembled.add(envelope);
    return envelope;
  });
}
/** Reparse independent original sources; never trust a candidate's declared scores. */
export function verifyProfileEvidence(
  selection: ProfileContractSelection,
  configuration: ServedProfileConfiguration,
  artifacts: Artifacts,
  observations: readonly (Observation | null)[],
  attempt: Attempt,
  evidenceSource: ConformanceArtifactSource,
): Data {
  return boundary(() => {
    const bytes = source(evidenceSource);
    if (bytes.byteLength > MAX_SUMMARY) refuse("capacity");
    const candidate = parsed(bytes),
      expected = assembleProfileEvidence(
        selection,
        configuration,
        artifacts,
        observations,
        attempt,
      );
    equal(candidate, expected, "evidence-mismatch");
    equal(Buffer.from(bytes).toString("utf8"), canonicalJson(expected), "evidence-mismatch");
    return expected;
  });
}
/** Closed bound/unbound summary encoder; does not make an unbound attempt a carrier. */
export function serializeProfileEvidence(
  evidence: unknown,
  transientCredentialValues: readonly string[] = [],
  unboundSet?: ReadUpdateArtifactSet,
): Uint8Array {
  return boundary(() => {
    const initial = record(evidence);
    if (initial.kind !== "private-profile-attempt-unbound" && !assembled.has(evidence as object))
      refuse();
    const e = record(safeData(evidence));
    if (e.kind === "private-profile-attempt-unbound") {
      ownData(e, [
        "attemptVersion",
        "kind",
        "bindingStatus",
        "configuration",
        "claimEligible",
        "qualificationReady",
        "requestedProfile",
        "declaredSourceCommit",
        "artifactDigests",
        "requiredIds",
        "owners",
        "observations",
        "attempt",
        "limitations",
      ]);
      if (
        e.attemptVersion !== 1 ||
        e.bindingStatus !== "unbound" ||
        e.configuration !== null ||
        e.claimEligible !== false ||
        e.qualificationReady !== false ||
        e.requestedProfile !== "read-update" ||
        !/^[a-f0-9]{40}$/.test(text(e.declaredSourceCommit))
      )
        refuse();
      const ids = array(e.requiredIds),
        obs = array(e.observations),
        owners = array(e.owners);
      if (
        ids.length !== 130 ||
        new Set(ids).size !== 130 ||
        obs.length !== ids.length ||
        obs.some((x) => x !== null) ||
        owners.length !== ids.length
      )
        refuse();
      owners.forEach((x, n) => {
        const r = ownData(x, ["id", "owner"]);
        if (r.id !== ids[n] || (r.owner !== null && typeof r.owner !== "string")) refuse();
      });
      if (!unboundSet) refuse();
      assertReadUpdateArtifactSet(unboundSet);
      equal(
        e.requiredIds,
        unboundSet.requirements.map((r) => r.id),
        "invalid-source",
      );
      equal(e.owners, unboundSet.requirements, "invalid-source");
      equal(e.artifactDigests, unboundSet.digests, "invalid-source");
      const attempted = captureAttempt(e.attempt as Attempt);
      equal(
        array(attempted.owners).map((x) => record(x).id),
        unboundSet.shards.map((x) => x.id),
        "invalid-source",
      );
      equal(
        e.limitations,
        [
          "no returned served configuration yet",
          "not a bound carrier, target observation, admission or qualification",
        ],
        "invalid-source",
      );
    } else {
      ownData(e, [
        "carrierVersion",
        "kind",
        "claimEligible",
        "qualificationReady",
        "profile",
        "history",
        "configuration",
        "catalogDigests",
        "artifactDigests",
        "requiredIds",
        "retiredIds",
        "requirements",
        "nonRequiredReportRows",
        "reports",
        "attempt",
        "unresolvedConditions",
        "limitations",
      ]);
      if (
        e.carrierVersion !== 1 ||
        e.kind !== "private-profile-evidence" ||
        e.claimEligible !== false ||
        e.qualificationReady !== false ||
        e.history !== false ||
        !["read-update", "transactional"].includes(text(e.profile))
      )
        refuse();
      if (array(e.requirements).some((r) => record(r).wholeClauseQualification !== "unestablished"))
        refuse();
    }
    const secrets = array(transientCredentialValues).map(text);
    if (secrets.some((x) => !x)) refuse("credential-check");
    const forms = secrets.flatMap((s) => [
      s,
      encodeURIComponent(s),
      Buffer.from(s).toString("base64"),
      Buffer.from(s).toString("base64url"),
      Buffer.from(s).toString("hex"),
      [...s].map((x) => `\\u${x.charCodeAt(0).toString(16).padStart(4, "0")}`).join(""),
    ]);
    for (const secret of secrets) {
      const uri = encodeURIComponent(secret);
      for (const encoding of ["base64", "base64url"] as const) {
        forms.push(
          encodeURIComponent(Buffer.from(secret).toString(encoding)),
          Buffer.from(uri).toString(encoding),
        );
      }
    }
    const inspect = (x: unknown): void => {
      if (typeof x === "string") {
        const candidates = new Set([x]);
        for (let layer = 0; layer < 2; layer++)
          for (const value of [...candidates]) {
            try {
              candidates.add(decodeURIComponent(value));
            } catch {
              /* Not URI text. */
            }
            if (/^[A-Za-z0-9+/_-]+={0,2}$/.test(value))
              candidates.add(Buffer.from(value, "base64").toString("utf8"));
          }
        if ([...candidates].some((v) => forms.some((s) => v.includes(s))))
          refuse("credential-check");
      } else if (x !== null && typeof x === "object")
        for (const [k, v] of Object.entries(x)) {
          inspect(k);
          inspect(v);
        }
    };
    inspect(e);
    const bytes = Buffer.from(canonicalJson(e));
    if (bytes.length > MAX_SUMMARY) refuse("capacity");
    return bytes;
  });
}
