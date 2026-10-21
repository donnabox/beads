/** Private installed-test association only. Never a ConformanceRunResult or admission artifact. */
import { createHash, randomUUID } from "node:crypto";
import { existsSync, renameSync, rmSync, writeFileSync, openSync, closeSync, fstatSync, lstatSync, readFileSync } from "node:fs";
import { isDeepStrictEqual } from "node:util";
import {
  canonicalJson,
  selectApplicableScenariosForProfile,
  type ConformanceRunResult,
  type ExactHttpConfiguration,
} from "@bdp/conformance";
import {
  assertReadUpdateArtifactSet,
  assertSameServedProfileConfiguration,
  bindServedProfileConfiguration,
  type ReadUpdateArtifactSet,
} from "@bdp/conformance/testing";
import { installedContractsDigest } from "./installed-contract-identity.js";
import { assembleProfileEvidence, serializeProfileEvidence } from "../../../packages/conformance/test-support/profile-evidence.js";

/** Separate private sidecars. First creation is exclusive; ownership is never inferred. */
export function createPrivateProfileSidecarWriter(options: {
  readonly base: string;
  readonly profile: "read-update" | "transactional";
  readonly credentialValues: readonly string[];
  readonly unboundSet?: ReadUpdateArtifactSet;
}) {
  const paths = { attempt: `${options.base}.${options.profile}.profile.attempt.json`,
    summary: `${options.base}.${options.profile}.profile.json` };
  const owned = new Map<string, { dev: number; ino: number; bytes: number; digest: string }>();
  const failure = () => Object.assign(new Error("private profile evidence persistence refused"), { code: "persistence" });
  for (const destination of Object.values(paths)) {
    try { lstatSync(destination); throw failure(); }
    catch (error) { if ((error as NodeJS.ErrnoException).code !== "ENOENT") throw failure(); }
  }
  function write(destination: string, envelope: unknown) {
    const bytes = serializeProfileEvidence(envelope, options.credentialValues, options.unboundSet);
    const captured = JSON.parse(Buffer.from(bytes).toString("utf8")) as { profile?: unknown; requestedProfile?: unknown };
    if ((captured.profile ?? captured.requestedProfile) !== options.profile) throw failure();
    const hash = createHash("sha256").update(bytes).digest("hex");
    const previous = owned.get(destination);
    if (!previous) {
      let fd: number | undefined;
      try {
        fd = openSync(destination, "wx", 0o600);
        writeFileSync(fd, bytes);
        const stat = fstatSync(fd);
        owned.set(destination, { dev: stat.dev, ino: stat.ino, bytes: bytes.byteLength, digest: hash });
      } catch { throw failure(); }
      finally { if (fd !== undefined) { try { closeSync(fd); } catch { throw failure(); } } }
      return;
    }
    const temporary = `${destination}.pending-${randomUUID()}`;
    let created = false;
    try {
      const stat = lstatSync(destination);
      if (!stat.isFile() || stat.dev !== previous.dev || stat.ino !== previous.ino || stat.size !== previous.bytes ||
        createHash("sha256").update(readFileSync(destination)).digest("hex") !== previous.digest) throw failure();
      const fd = openSync(temporary, "wx", 0o600); created = true;
      try { writeFileSync(fd, bytes); } finally { closeSync(fd); }
      renameSync(temporary, destination); created = false;
      const current = lstatSync(destination);
      owned.set(destination, { dev: current.dev, ino: current.ino, bytes: bytes.byteLength, digest: hash });
    } catch {
      if (created) { try { rmSync(temporary); } catch { throw new AggregateError([failure(), failure()], "private profile evidence persistence refused"); } }
      throw failure();
    }
  }
  return Object.freeze({ writeAttempt: (envelope: unknown) => write(paths.attempt, envelope),
    writeSummary: (envelope: unknown) => write(paths.summary, envelope) });
}

type Bundle = ReadUpdateArtifactSet["shards"][number]["bundle"];
export function artifactRoutes(bundle: Bundle, scope: string) {
  const known: Record<string, string> = {
    ...Object.fromEntries(
      Object.entries(bundle.fixture.bindings).map(([key, value]) => [
        key,
        new URL(value, scope).href,
      ]),
    ),
    scope,
    "service-desc": `${scope}bdp.json`,
    beads: `${scope}beads/`,
    links: `${scope}links/`,
    types: `${scope}types/`,
    operations: `${scope}operations/`,
  };
  const reads = bundle.manifest.scenarios.flatMap((row) =>
    (row.actions ?? row.requests).flatMap((action) => {
      if ("family" in action && action.family !== "http") return [];
      const base = known[action.target.binding];
      if (!base) throw Error("Read target binding is not in the fixed fixture");
      const url = new URL(action.target.path ?? "", base);
      for (const [key, input] of Object.entries(action.target.query ?? {}))
        for (const value of Array.isArray(input) ? input : [input]) {
          if (typeof value !== "string") throw Error("Read fixture query requires literal strings");
          url.searchParams.append(key, value);
        }
      return [{ method: action.method, url: url.href }];
    }),
  );
  const clients = bundle.manifest.scenarios.flatMap((row) =>
    (row.actions ?? []).flatMap((action) => {
      if (action.family !== "client" || action.operation !== "installed.exact-request") return [];
      const input = action.input as Record<string, unknown>;
      if (input.method !== "GET" || typeof input.targetBinding !== "string") return [];
      const url = known[input.targetBinding];
      return url === undefined ? [] : [{ method: "GET" as const, url }];
    }),
  );
  const routes = [
    ...clients,
    ...reads,
    ...[
      "sequence",
      "create-bead",
      "update-bead",
      "delete-bead",
      "create-link",
      "update-link",
      "delete-link",
      "put-alias",
      "delete-alias",
    ].map((suffix) => ({ method: "POST" as const, url: `${scope}operations/${suffix}` })),
  ];
  return routes.filter(
    (r, n) => routes.findIndex((other) => r.method === other.method && r.url === other.url) === n,
  );
}
export type ArtifactObservation = {
  readonly sourceCommit: string;
  readonly configuration: ReturnType<typeof bindServedProfileConfiguration>;
  readonly baselineReceipt: Record<string, unknown>;
  readonly sessions: readonly Record<string, unknown>[];
  readonly result: ConformanceRunResult;
  readonly exactConfiguration: ExactHttpConfiguration;
  readonly plannedManifestDigest: string | undefined;
  readonly catalogSources: Readonly<Record<string, string>>;
  readonly plannedIds: readonly string[];
  readonly requiredIds: readonly string[];
  readonly cleanupComplete: boolean;
};
function equal(actual: unknown, expected: unknown, label: string) {
  if (!isDeepStrictEqual(actual, expected)) throw Error(`artifact-set ${label} mismatch`);
}
function readySummary(receipt: Record<string, unknown>) {
  return Object.fromEntries(
    [
      "scope",
      "profile",
      "history",
      "payloadIdentity",
      "artifactIdentity",
      "policyIdentity",
      "authenticationIdentity",
      "configurationIdentity",
      "listenUrl",
    ]
      .filter((key) => receipt[key] !== undefined)
      .map((key) => [key, receipt[key]]),
  );
}
function digest(value: unknown) {
  return createHash("sha256").update(canonicalJson(value)).digest("hex");
}

/** Maps actual returned rows. Missing observations stay null; no runner result is invented. */
function requiredObservations(
  set: ReadUpdateArtifactSet,
  results: readonly (ConformanceRunResult | undefined)[],
) {
  const resultDigests = results.map((result) =>
    result === undefined ? undefined : digest(result),
  );
  const required = set.requirements.map(({ id, owner }) => {
    const index = owner === null ? 0 : set.shards.findIndex((s) => s.id === owner);
    const result = results[index];
    const rowIndex = result?.scenarios.findIndex((r) => r.id === id) ?? -1;
    const row = rowIndex < 0 ? undefined : result?.scenarios[rowIndex];
    return {
      id,
      owner,
      observation:
        row === undefined
          ? null
          : {
              shardId: set.shards[index]!.id,
              resultDigest: resultDigests[index],
              rowIndex,
              state: row.state,
              ...(row.category ? { category: row.category } : {}),
            },
      ...(owner === null
        ? { reason: "not-planned" as const }
        : row === undefined
          ? { reason: "unobserved" as const }
          : {}),
    };
  });
  const partitions = {
    pass: required.filter((r) => r.observation?.state === "pass").map((r) => r.id),
    fail: required.filter((r) => r.observation?.state === "fail").map((r) => r.id),
    notRun: required.filter((r) => r.observation?.category === "not-run").map((r) => r.id),
    notImplemented: required
      .filter((r) => r.observation?.category === "not-implemented")
      .map((r) => r.id),
    other: required
      .filter(
        (r) =>
          r.observation &&
          !["pass", "fail"].includes(r.observation.state) &&
          !["not-run", "not-implemented"].includes(r.observation.category ?? ""),
      )
      .map((r) => r.id),
    unobserved: required.filter((r) => r.observation === null).map((r) => r.id),
  };
  return {
    requiredObservations: required,
    // Returned owner outcomes, excluding missing-plan/not-run placeholders. State
    // partitions expose aborted/skipped outcomes without calling them passes.
    observedIds: required
      .filter(
        (r) =>
          r.owner !== null &&
          r.observation !== null &&
          !["not-run", "not-implemented"].includes(r.observation.category ?? ""),
      )
      .map((r) => r.id),
    observedStatePartitions: partitions,
    uncoveredIds: set.coverage.rows.filter((r) => !r.wholeContract).map((r) => r.id),
  };
}

export function associateArtifactObservations(
  set: ReadUpdateArtifactSet,
  sourceCommit: string,
  observations: readonly ArtifactObservation[],
) {
  assertReadUpdateArtifactSet(set);
  if (observations.length !== set.shards.length)
    throw Error("artifact-set missing or extra report");
  const first = observations[0]!;
  const summary = set.shards.map((shard, index) => {
    const report = observations[index]!;
    const { result, configuration } = report;
    if (!report.cleanupComplete) throw Error("artifact-set cleanup incomplete");
    equal(report.sourceCommit, sourceCommit, "source");
    equal(configuration.sourceCommit, sourceCommit, "configuration source");
    assertSameServedProfileConfiguration(first.configuration, configuration);
    equal(configuration.fixtureDigest, set.digests.fixtureDigest, "configuration fixture");
    equal(configuration.profile, "read-update", "profile");
    equal(configuration.history, false, "History");
    equal(report.catalogSources, set.digests.catalogDigests, "catalog source map");
    equal(report.plannedIds, shard.scenarioIds, "owned plans");
    equal(
      report.requiredIds,
      set.requirements.map((r) => r.id),
      "required inventory",
    );
    equal(report.plannedManifestDigest, shard.bundle.digests.manifestDigest, "planned manifest");
    equal(result.artifacts, shard.bundle.digests, "report artifacts");
    equal(result.profile, configuration.profile, "report profile");
    equal(result.scope, configuration.scope, "report Scope");
    equal(result.claimEligible, false, "claim");
    equal(result.reportVersion, 4, "report version");
    equal(result.seed, shard.bundle.fixture.seed, "seed");
    equal(result.scenarioFilter, undefined, "filter");
    const selected =
      index === 0
        ? selectApplicableScenariosForProfile(shard.bundle.catalog, "read-update").map((r) => r.id)
        : shard.scenarioIds;
    equal(
      result.scenarioSelection,
      index === 0 ? undefined : shard.scenarioIds,
      "requested selection",
    );
    equal(result.selectedScenarioIds, selected, "selected IDs");
    equal(
      result.scenarios.map((r) => r.id),
      selected,
      "result IDs",
    );
    const owned = new Set(shard.scenarioIds);
    for (const row of result.scenarios) {
      if (row.cleanupError !== undefined) throw Error(`artifact-set row cleanup failure ${row.id}`);
      if (owned.has(row.id)) equal(row.state, "pass", `owned outcome ${row.id}`);
      else {
        equal(row.state, "harness-error", `unplanned outcome ${row.id}`);
        equal(row.category, "not-implemented", `unplanned category ${row.id}`);
        equal(row.exchanges, [], `unplanned exchanges ${row.id}`);
        equal(row.actions?.length ?? 0, 0, `unplanned actions ${row.id}`);
      }
    }
    const exact = report.exactConfiguration;
    equal(exact.scope, configuration.scope, "route Scope");
    equal(exact.profile, configuration.profile, "route profile");
    equal(exact.routes, artifactRoutes(shard.bundle, configuration.scope), "routes");
    equal(exact.defaultCredentialRef, "operator", "default credential handle");
    equal(
      exact.credentialHandles,
      [
        { id: "operator", headerNames: ["authorization"] },
        { id: "anonymous", headerNames: [] },
        { id: "unknown", headerNames: ["authorization"] },
      ],
      "credential handles",
    );
    equal(exact.maximumRequestHeaderBytes, 16384, "header bound");
    equal(exact.maximumRequestBodyBytes, 131072, "body bound");
    const routeDigest = digest({
      scope: exact.scope,
      profile: exact.profile,
      routes: exact.routes,
      maximumRequestHeaderBytes: exact.maximumRequestHeaderBytes,
      maximumRequestBodyBytes: exact.maximumRequestBodyBytes,
    });
    equal(
      "exactConfigurationDigest" in result.declarations
        ? result.declarations.exactConfigurationDigest
        : undefined,
      routeDigest,
      "route digest",
    );
    equal(
      report.sessions.map((s) => s.scenario),
      shard.scenarioIds,
      "actual sessions",
    );
    for (const receipt of [report.baselineReceipt, ...report.sessions]) {
      equal(
        receipt.configurationIdentity,
        first.baselineReceipt.configurationIdentity,
        "actual runtime configuration",
      );
      equal(receipt.scope, configuration.scope, "actual Scope");
      equal(receipt.profile, configuration.profile, "actual profile");
      equal(receipt.history, configuration.history, "actual History");
      equal(receipt.payloadIdentity, configuration.installedPayloadDigest, "actual payload");
      equal(
        installedContractsDigest(receipt.artifactIdentity),
        configuration.installedContractsDigest,
        "actual contracts",
      );
      equal(receipt.policyIdentity, configuration.policyDigest, "actual policy");
      equal(
        receipt.authenticationIdentity,
        configuration.authenticationDigest,
        "actual authentication",
      );
    }
    return {
      id: shard.id,
      source: shard.source,
      selectionMode: index === 0 ? "unfiltered" : "owned-exact",
      requestedIds: result.scenarioSelection ?? null,
      selectedIds: result.selectedScenarioIds,
      resultIds: result.scenarios.map((r) => r.id),
      manifestDigest: shard.bundle.digests.manifestDigest,
      routeDigest,
      routes: exact.routes,
      resultDigest: digest(result),
      resultBytes: Buffer.byteLength(canonicalJson(result)),
      actualBootstrap: readySummary(report.baselineReceipt),
      actualSessions: report.sessions.map((s) => ({ scenario: s.scenario, ...readySummary(s) })),
      cleanupComplete: report.cleanupComplete,
    };
  });
  return {
    claimEligible: false as const,
    sourceCommit,
    ...set.digests,
    sourceBytes: set.sourceBytes,
    configuration: first.configuration,
    runtimeConfigurationIdentity: first.baselineReceipt.configurationIdentity,
    requiredOwnership: set.requirements,
    authoredIds: set.requirements.filter((r) => r.owner !== null).map((r) => r.id),
    ...requiredObservations(
      set,
      observations.map((r) => r.result),
    ),
    reviewedWholeIds: set.coverage.rows.filter((r) => r.wholeContract).map((r) => r.id),
    omittedRequiredIds: set.requirements.filter((r) => r.owner === null).map((r) => r.id),
    optionalCatalogIds: first.result.selectedScenarioIds.filter(
      (id) => !set.requirements.some((r) => r.id === id),
    ),
    shards: summary,
  };
}

/** Safe attempt evidence is separate from the all-checks-passed set summary. */
export function createArtifactEvidenceWriter(
  set: ReadUpdateArtifactSet,
  sourceCommit: string,
  filename: string | undefined,
  credentialValues: readonly string[],
) {
  assertReadUpdateArtifactSet(set);
  const credentials = [...credentialValues];
  if (!/^[a-f0-9]{40}$/.test(sourceCommit) || credentials.some((s) => !s))
    throw Error("invalid evidence writer identity");
  if (
    filename &&
    [
      filename,
      `${filename}.attempt.json`,
      ...set.shards.map((s) => `${filename}.${s.id}.json`),
    ].some(existsSync)
  )
    throw Error("evidence output already exists; choose a fresh attempt path");
  type Cleanup = "not-started" | "pending" | "complete" | "failed";
  const rows = set.shards.map((s) => ({
    id: s.id,
    source: s.source,
    manifestDigest: s.bundle.digests.manifestDigest,
    status: "unobserved" as "unobserved" | "running" | "returned" | "aborted" | "failed",
    cleanup: "not-started" as Cleanup,
    evidence: "absent" as "absent" | "retained" | "withheld-credential",
    resultDigest: undefined as string | undefined,
    writtenReportDigest: undefined as string | undefined,
    writtenReportBytes: undefined as number | undefined,
    timing: undefined as
      | {
          budgetRemainingMsAtStart: number;
          elapsedMs: number;
          outcome: "running" | "pass" | "aborted" | "failed";
        }
      | undefined,
  }));
  const snapshots = new Map<number, Record<string, unknown>>();
  const profilePath = filename && filename === process.env.BDP_CUMULATIVE_REPORT_PATH
    ? process.env.BDP_PRIVATE_PROFILE_EVIDENCE_PATH : undefined;
  const sidecar = profilePath ? createPrivateProfileSidecarWriter({ base: profilePath,
    profile: "read-update", credentialValues: credentials, unboundSet: set }) : undefined;
  let profileConfiguration: ArtifactObservation["configuration"] | undefined;
  const profileBindings = new Map<number, ArtifactObservation["configuration"]>();
  const profileBytes = new Map<number, Uint8Array>();
  let failed = false;
  type FailureStage = "execution" | "assertion" | "cleanup" | "credential-check";
  const failureStages: FailureStage[] = [];
  function safeJson(value: unknown) {
    const text = JSON.stringify(value, null, 2);
    function inspect(v: unknown): boolean {
      if (typeof v === "string") {
        const candidates = [v];
        if (v.length % 4 === 0 && /^[A-Za-z0-9+/]+={0,2}$/.test(v))
          candidates.push(Buffer.from(v, "base64").toString("utf8"));
        try {
          candidates.push(decodeURIComponent(v));
        } catch {
          /* Non-URI strings remain checked literally. */
        }
        return candidates.some((candidate) =>
          credentials.some((secret) => candidate.includes(secret)),
        );
      }
      return (
        v !== null &&
        typeof v === "object" &&
        Object.entries(v).some(([key, value]) => inspect(key) || inspect(value))
      );
    }
    if (inspect(JSON.parse(text))) throw Error("credential-bearing evidence withheld");
    return text;
  }
  // Cleanup metadata replaces a complete file atomically. A failed replacement
  // leaves previously returned genuine evidence intact (not a power-loss claim).
  function replaceFile(destination: string, encoded: string) {
    const temporary = `${destination}.pending-${randomUUID()}`;
    try {
      writeFileSync(temporary, encoded, { flag: "wx", mode: 0o600 });
      renameSync(temporary, destination);
    } catch (error) {
      try {
        rmSync(temporary, { force: true });
      } catch (cleanup) {
        throw new AggregateError([error, cleanup], "evidence replacement and cleanup failed", {
          cause: error,
        });
      }
      throw error;
    }
  }
  function ledger(outcome: "running" | "failed" | "observed", stage?: FailureStage) {
    if (stage && !failureStages.includes(stage)) failureStages.push(stage);
    const value = {
      claimEligible: false,
      outcome,
      ...(failureStages.length
        ? { failureStage: failureStages[0], failureStages: [...failureStages] }
        : {}),
      sourceCommit,
      ...set.digests,
      ...requiredObservations(
        set,
        set.shards.map(
          (_, index) => snapshots.get(index)?.result as ConformanceRunResult | undefined,
        ),
      ),
      shards: rows.map((r) => ({
        ...r,
        ...(r.evidence === "retained" && filename ? { report: `${filename}.${r.id}.json` } : {}),
      })),
    };
    const encoded = safeJson(value);
    if (filename) replaceFile(`${filename}.attempt.json`, encoded);
    if (sidecar) sidecar.writeAttempt(profileEvidence(value.outcome));
    return value;
  }
  function persist(index: number) {
    const snapshot = snapshots.get(index);
    if (!snapshot) return;
    snapshot.cleanupStatus = rows[index]!.cleanup;
    snapshot.cleanupComplete = rows[index]!.cleanup === "complete";
    const encoded = safeJson(snapshot);
    if (filename) replaceFile(`${filename}.${rows[index]!.id}.json`, encoded);
    if(sidecar) profileBytes.set(index, Buffer.from(encoded));
    rows[index]!.evidence = "retained";
    rows[index]!.writtenReportDigest = createHash("sha256").update(encoded).digest("hex");
    rows[index]!.writtenReportBytes = Buffer.byteLength(encoded);
  }
  function profileEvidence(outcome: "running" | "failed" | "observed") {
    const attempt = { outcome, failureStages: [...failureStages],
      owners: rows.map(r => ({ id: r.id, status: r.status,
        cleanup: r.cleanup === "complete" ? "reported-complete" : r.cleanup, evidence: r.evidence })),
      sourceDirty: typeof snapshots.get(0)?.sourceDirty === "boolean" ? snapshots.get(0)!.sourceDirty as boolean : null };
    if (!profileConfiguration) return {
      attemptVersion: 1, kind: "private-profile-attempt-unbound", bindingStatus: "unbound", configuration: null,
      claimEligible: false, qualificationReady: false, requestedProfile: "read-update", declaredSourceCommit: sourceCommit,
      artifactDigests: set.digests, requiredIds: set.requirements.map(r=>r.id), owners: set.requirements,
      observations: set.requirements.map(()=>null), attempt,
      limitations: ["no returned served configuration yet", "not a bound carrier, target observation, admission or qualification"],
    };
    return assembleProfileEvidence(set.selection, profileConfiguration, { kind: "read-update", set },
      set.shards.map((shard,index) => {
        const bytes = profileBytes.get(index), snapshot = snapshots.get(index), configuration = profileBindings.get(index);
        if(!bytes || !snapshot || !configuration) return null;
        return { ownerId: shard.id, sourceCommit, configuration,
          runtimeConfigurationIdentity: String((snapshot.baselineReceipt as Record<string,unknown>).configurationIdentity),
          exactConfiguration: snapshot.exactConfiguration,
          ready: [snapshot.baselineReceipt, ...snapshot.sessions as unknown[]],
          reportSource: { bytes }, expectedReportSha256: createHash("sha256").update(bytes).digest("hex"),
          lifecycle: { cleanup: rows[index]!.cleanup === "complete" ? "reported-complete" as const : rows[index]!.cleanup === "failed" ? "failed" as const : "pending" as const },
        };
      }), attempt);
  }
  const api = {
    start(index: number) {
      rows[index]!.status = "running";
      ledger("running");
    },
    acquired(index: number) {
      rows[index]!.cleanup = "pending";
      ledger("running");
    },
    returned(index: number, report: ArtifactObservation) {
      // Only known wrapper fields accompany the genuine runner result. Never retain
      // thrown errors, child stderr, resolver functions or arbitrary wrapper keys.
      const extra = report as ArtifactObservation & {
        observations?: unknown;
        setupObservations?: unknown;
        nativeReadObservations?: unknown;
        sourceDirty?: boolean;
      };
      const snapshot = {
        claimEligible: false,
        sourceCommit: report.sourceCommit,
        configuration: report.configuration,
        baselineReceipt: readySummary(report.baselineReceipt),
        sessions: report.sessions.map((s) => ({ scenario: s.scenario, ...readySummary(s) })),
        catalogSources: report.catalogSources,
        plannedManifestDigest: report.plannedManifestDigest,
        plannedIds: report.plannedIds,
        requiredIds: report.requiredIds,
        exactConfiguration: {
          scope: report.exactConfiguration.scope,
          profile: report.exactConfiguration.profile,
          routes: report.exactConfiguration.routes,
          maximumRequestHeaderBytes: report.exactConfiguration.maximumRequestHeaderBytes,
          maximumRequestBodyBytes: report.exactConfiguration.maximumRequestBodyBytes,
          credentialHandles: report.exactConfiguration.credentialHandles,
          defaultCredentialRef: report.exactConfiguration.defaultCredentialRef,
        },
        result: report.result,
        observations: extra.observations,
        setupObservations: extra.setupObservations,
        nativeReadObservations: extra.nativeReadObservations,
        sourceDirty: extra.sourceDirty,
      };
      let captured: Record<string, unknown>;
      try {
        // Capture once: later caller mutation cannot corrupt a retained child report.
        captured = JSON.parse(safeJson(snapshot));
      } catch (error) {
        rows[index]!.status = "failed";
        rows[index]!.evidence = "withheld-credential";
        failed = true;
        ledger("failed", "credential-check");
        throw error;
      }
      if(sidecar) {
        assertSameServedProfileConfiguration(report.configuration, report.configuration);
        if(profileConfiguration) assertSameServedProfileConfiguration(profileConfiguration, report.configuration);
        else profileConfiguration = report.configuration;
        profileBindings.set(index, report.configuration);
      }
      snapshots.set(index, captured);
      rows[index]!.status = "returned";
      rows[index]!.resultDigest = digest(report.result);
      persist(index);
      ledger("running");
    },
    cleaned(index: number, status: "complete" | "failed") {
      rows[index]!.cleanup = status;
      if (status === "failed") {
        rows[index]!.status = "failed";
        failed = true;
      }
      persist(index);
      ledger(failed ? "failed" : "running", status === "failed" ? "cleanup" : undefined);
    },
    failed(index: number, aborted: boolean) {
      rows[index]!.status = aborted ? "aborted" : "failed";
      failed = true;
      ledger("failed", "execution");
    },
    timed(index: number, timing: NonNullable<(typeof rows)[number]["timing"]>) {
      rows[index]!.timing = {
        budgetRemainingMsAtStart: timing.budgetRemainingMsAtStart,
        elapsedMs: timing.elapsedMs,
        outcome: timing.outcome,
      };
      ledger(failed ? "failed" : "running");
    },
    failedSet() {
      failed = true;
      ledger("failed", "assertion");
    },
    succeeded(summary: Record<string, unknown>) {
      if (
        failed ||
        rows.some(
          (r) => r.status !== "returned" || r.cleanup !== "complete" || r.evidence !== "retained",
        )
      )
        throw Error("cannot publish incomplete artifact evidence");
      const encoded = safeJson({
        ...summary,
        writtenReports: rows.map((r) => ({
          id: r.id,
          resultDigest: r.resultDigest,
          ...(filename
            ? {
                report: `${filename}.${r.id}.json`,
                writtenReportDigest: r.writtenReportDigest,
                writtenReportBytes: r.writtenReportBytes,
              }
            : {}),
        })),
      });
      ledger("observed");
      if (filename) replaceFile(filename, encoded);
      if(sidecar) sidecar.writeSummary(profileEvidence("observed"));
    },
  };
  ledger("running");
  return api;
}
export type ArtifactEvidenceWriter = ReturnType<typeof createArtifactEvidenceWriter>;
