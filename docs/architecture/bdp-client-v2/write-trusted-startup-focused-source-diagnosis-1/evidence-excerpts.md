# Exact source/spec evidence

Static inspection only.

### CONTEXT.md:1–22

1: # Bead Protocol
2: 
3: This glossary defines BDP-specific terms used across the repository.
4: 
5: ## Language
6: 
7: **Protocol identifier**:
8: A URI assigned by BDP to identify protocol vocabulary independently of where
9: the corresponding source artifact is hosted. Dereferenceability is not part of
10: its identity.
11: _Avoid_: download URL
12: 
13: **Normative schema bundle**:
14: The single JSON Schema bundle containing BDP v0's normative JSON wire
15: definitions.
16: _Avoid_: schema fragments, implementation schema
17: 
18: **Reference / Pinned Reference**:
19: How anything in BDP points at anything: a URI — or a Pinned Reference,
20: `{ uri, revision }`, the URI plus the revision it was made against. Either
21: reference class may be pinned; the URI alone is identity.
22: _Avoid_: external Type, remote Bead Type

### docs/specs/bdp.md:647–692

647: A Type Descriptor cannot add an operation, query, view, Event Type, or
648: protocol method. Domain-specific behavior belongs in clients that interpret
649: nominal Types and use BDP's generic surface. It is not an extension advertised
650: by a Type or Scope.
651: 
652: ### Scope aggregate constraints
653: 
654: A Type Descriptor owns constraints that can be validated from one Resource
655: and its in-Scope endpoints: properties schemas and endpoint effective-Type
656: requirements. Constraints that inspect other Resources belong to the Scope,
657: not to the globally identified Type contract.
658: 
659: BDP v0 defines one optional aggregate policy shape:
660: 
661: ```text
662: MaximumEndpointMultiplicity {
663:   linkConformsTo: LinkTypeId
664:   endpoint: source | target
665:   max: PositiveInteger
666: }
667: ```
668: 
669: For one in-Scope Bead and one named endpoint, this policy counts the live
670: Links in the Scope whose effective Link Types contain `linkConformsTo` and
671: whose named endpoint equals that Bead. The count includes Links declared as
672: different conforming child Types. If several policies apply, the smallest
673: maximum wins. An out-of-Scope endpoint is never counted or constrained.
674: 
675: Discovery MAY contain `maximumEndpointMultiplicity`, an unordered array of
676: records using the three members above:
677: 
678: ```json
679: {
680:   "maximumEndpointMultiplicity": [
681:     {
682:       "linkConformsTo": "https://work.example/types/parent-child",
683:       "endpoint": "source",
684:       "max": 1
685:     }
686:   ]
687: }
688: ```
689: 
690: An absent member or an empty array means that the Scope defines no such
691: policies. The administrative mechanism remains outside BDP. Minimum multiplicity, tuple uniqueness, acyclicity, and other aggregate
692: graph policies are deferred.

### docs/specs/bdp.md:802–823

802: ### Authorization views
803: 
804: For every request, the authority binds the authenticated principal, or an
805: anonymous principal, to exactly one opaque **Authorization View** of the
806: Scope. The client cannot name, widen, or combine views through request data.
807: Different principals may share a view only when the authority considers
808: their read projections equivalent. The canonical Scope, Bead, and Link URLs
809: remain identity, and they do not vary by view.
810: 
811: An Authorization View is a closed projection of the Scope. Every read,
812: Selector, incident Link view and
813: representation for the request observes that same projection. A Link is
814: visible only when the Link and every in-Scope endpoint Bead are visible. An
815: out-of-Scope endpoint is opaque and does not independently gate Link
816: visibility. A visible Bead need not expose hidden incident Links — a latitude
817: confined to incoming and unowned Links: a visible source's owned Links,
818: covered by its revision, are never withheld, and the view includes their
819: in-Scope targets, under [Owned Links](#owned-links). The
820: authority still evaluates referential integrity, Scope aggregate
821: constraints, deletion safety, and other Scope invariants against its
822: complete authoritative state. A non-disclosing constraint failure may
823: withhold the hidden Resources that caused it.

### docs/design/startup-configuration.md:658–679

658: The separate `@bdp/server/reference-configuration` data entry exposes
659: `loadReferenceDeployment`. This pure preparation step opens no store or listener
660: and grants no profile capability. It accepts an explicit shared Scope, intended
661: listener, profile and Read limits plus the deployment path. It refuses Read or
662: missing profile, fixture/fault conflicts, port0 and unsupported hosts. RU supports
663: loopback `127.0.0.1`/`::1`; TX currently supports `127.0.0.1`. Shared Read settings
664: must match the existing capture tuple: RU page20/100, selector4096/32/256,
665: cursor60000ms; TX page2/32, selector4096/32/256, cursor1000ms. These differ from the
666: ordinary shared Read defaults and must be configured explicitly. The returned
667: `preflightReadLimits` proves a code/capture comparison only.
668: 
669: The internal file has exact keys `format`, `profile`, `configuration`, with format
670: `bdptest-reference-deployment-1` and profile matching shared intent. RU configuration
671: has exact keys `directory`, `create`, `credentials` (token/principal), `policies`
672: (principal to create/write/hidden/readOnly), `contracts` (descriptors/schemas), and
673: `maximumEndpointMultiplicity`. TX has `directory`, `create`, `credentials`
674: (token/principal/canMutate) and `artifacts` (descriptors/schemas). Artifact entries
675: are exact retrievalUri/utf8 pairs. Existing installed captures own all inner
676: validation and identities: RU empty descriptors select the shipped local closure;
677: TX requires descriptors. Tokens must be unique; duplicate principals are allowed.
678: No Scope/listener/limits override, code plugin, grant, evidence or wire installation
679: operation is accepted in the envelope.

### packages/server/src/reference-deployment.test.ts:59–172

59: const scope = "https://deployment.test/acme/";
60: const token = "deployment_writer_sentinel_01234567890123456789";
61: const readToken = "deployment_reader_sentinel_01234567890123456789";
62: const allow = { create: true, write: true, hidden: [], readOnly: [] };
63: const roots: string[] = [];
64: beforeEach(() => {
65:   forbidden.mockClear();
66:   vi.mocked(open).mockClear();
67: });
68: afterEach(() => {
69:   try {
70:     expect(forbidden).not.toHaveBeenCalled();
71:   } finally {
72:     vi.mocked(open).mockReset();
73:     for (const root of roots.splice(0)) rmSync(root, { recursive: true, force: true });
74:   }
75: });
76: // Reset only the mock history/one-shot implementations; the next test reinstalls
77: // the real open so it exercises an actual ordinary-file handle by default.
78: const realFs = await vi.importActual<typeof import("node:fs/promises")>("node:fs/promises");
79: beforeEach(() => {
80:   vi.mocked(open).mockImplementation(realFs.open);
81: });
82: function artifacts() {
83:   const base = `${scope}types/base`,
84:     owner = `${scope}types/owner`,
85:     edge = `${scope}types/edge`,
86:     schema = `${scope}schemas/item`;
87:   return {
88:     descriptors: [
89:       {
90:         retrievalUri: base,
91:         utf8: JSON.stringify({
92:           id: base,
93:           name: "Base",
94:           describes: "bead",
95:           conformsTo: [],
96:           propertiesSchema: schema,
97:         }),
98:       },
99:       {
100:         retrievalUri: owner,
101:         utf8: JSON.stringify({
102:           id: owner,
103:           name: "Owner",
104:           describes: "bead",
105:           conformsTo: [base],
106:           ownsOutgoing: { [edge]: { max: 2 } },
107:         }),
108:       },
109:       {
110:         retrievalUri: edge,
111:         utf8: JSON.stringify({
112:           id: edge,
113:           name: "Edge",
114:           describes: "link",
115:           conformsTo: [],
116:           source: { conformsTo: [base], external: "opaque" },
117:           target: { conformsTo: [base], external: "opaque" },
118:         }),
119:       },
120:     ],
121:     schemas: [
122:       {
123:         retrievalUri: schema,
124:         utf8: JSON.stringify({
125:           $id: schema,
126:           type: "object",
127:           properties: { n: { type: "integer" } },
128:           additionalProperties: false,
129:         }),
130:       },
131:     ],
132:   };
133: }
134: function ru(directory: string) {
135:   return {
136:     format: "bdptest-reference-deployment-1",
137:     profile: "read-update",
138:     configuration: {
139:       directory,
140:       create: true,
141:       credentials: [
142:         { token, principal: "writer" },
143:         { token: readToken, principal: "reader" },
144:       ],
145:       policies: { writer: allow, reader: { ...allow, create: false, write: false } },
146:       contracts: {
147:         descriptors: [] as { retrievalUri: string; utf8: string }[],
148:         schemas: [] as { retrievalUri: string; utf8: string }[],
149:       },
150:       maximumEndpointMultiplicity: [] as {
151:         endpoint: string;
152:         linkConformsTo: string;
153:         max: number;
154:       }[],
155:     },
156:   };
157: }
158: function tx(directory: string) {
159:   return {
160:     format: "bdptest-reference-deployment-1",
161:     profile: "transactional",
162:     configuration: {
163:       directory,
164:       create: true,
165:       credentials: [
166:         { token, principal: "same", canMutate: true },
167:         { token: readToken, principal: "same", canMutate: false },
168:       ],
169:       artifacts: artifacts(),
170:     },
171:   };
172: }

### packages/server/src/reference-deployment.test.ts:204–216

204: async function failure(input: ReferenceDeploymentInput, code: string) {
205:   try {
206:     await loadReferenceDeployment(input);
207:   } catch (error) {
208:     expect(error).toBeInstanceOf(ReferenceDeploymentError);
209:     expect(error).toMatchObject({ code });
210:     expect(error).not.toHaveProperty("cause");
211:     const text = `${String(error)} ${JSON.stringify(error)}`;
212:     for (const secret of [token, readToken, "payload_sentinel", "unsafe_key_sentinel"])
213:       expect(text).not.toContain(secret);
214:     return;
215:   }
216:   throw new Error("expected local preparation failure");

### packages/server/src/reference-deployment.test.ts:237–269

237: it("captures the RU default closure through the existing capture and exposes only an inert handle", async () => {
238:   const f = fixture(),
239:     body = ru(f.directory);
240:   const prepared = await loadReferenceDeployment(f.input);
241:   expect(open).toHaveBeenCalledTimes(1);
242:   const direct = captureInstalledReferenceConfiguration({
243:     scope,
244:     listener: { host: "127.0.0.1", port: 43210 },
245:     ...body.configuration,
246:   });
247:   expect(prepared).toMatchObject({
248:     profile: "read-update",
249:     canonicalScope: scope,
250:     listener: { host: "127.0.0.1", port: 43210 },
251:     preflightReadLimits: DEVELOPMENT_READ_LIMITS,
252:     artifactIdentity: direct.artifactIdentity,
253:     configurationIdentity: direct.identity,
254:     policyIdentity: direct.policyIdentity,
255:     authenticationIdentity: direct.authenticationIdentity,
256:     fileDigest: createHash("sha256").update(readFileSync(f.input.deploymentFile)).digest("hex"),
257:   });
258:   inert(prepared);
259:   expect(JSON.stringify(prepared)).not.toContain(token);
260:   const state = referenceDeploymentState(prepared);
261:   expect(state.profile).toBe("read-update");
262:   if (state.profile !== "read-update") throw new Error("wrong capture");
263:   expect(state.installed.identity).toBe(direct.identity);
264:   expect(state.installed.credentials).toEqual(direct.credentials);
265:   expect(state.installed.configuration.contracts.identity).toBe(
266:     direct.configuration.contracts.identity,
267:   );
268:   expect(state.installed.configuration.access.capture("reader").write).toBe(false);
269:   expect(Object.isFrozen(state.projectedInput)).toBe(true);

### packages/server/src/reference-deployment.test.ts:304–325

304: it("delegates explicit RU schemas, ownership and endpoint multiplicity without an alternate compiler", async () => {
305:   const f = fixture(),
306:     body = ru(f.directory);
307:   body.configuration.contracts = artifacts();
308:   body.configuration.maximumEndpointMultiplicity = [
309:     { endpoint: "target", linkConformsTo: `${scope}types/edge`, max: 1 },
310:   ];
311:   f.write(body);
312:   const prepared = await loadReferenceDeployment(f.input),
313:     state = referenceDeploymentState(prepared);
314:   const direct = captureInstalledReferenceConfiguration({
315:     scope,
316:     listener: prepared.listener,
317:     ...body.configuration,
318:   });
319:   expect(prepared.artifactIdentity).toBe(direct.artifactIdentity);
320:   if (state.profile !== "read-update") throw new Error("wrong capture");
321:   expect(state.installed.identity).toBe(direct.identity);
322:   expect(state.installed.configuration.maximumEndpointMultiplicity).toEqual(
323:     direct.configuration.maximumEndpointMultiplicity,
324:   );
325:   expect(existsSync(f.directory)).toBe(false);

### packages/server/src/reference-deployment.test.ts:487–553

487: it("delegates RU credential, policy, provisioning, artifact, locality and aggregate refusals", async () => {
488:   const f = fixture(),
489:     configuration = ru(f.directory).configuration;
490:   for (const patch of [
491:     { directory: "relative" },
492:     { create: "true" },
493:     { credentials: [] },
494:     {
495:       credentials: [
496:         { token, principal: "writer" },
497:         { token, principal: "reader" },
498:       ],
499:     },
500:     { credentials: [{ token: "short", principal: "writer" }] },
501:     { credentials: [{ token, principal: "writer", canMutate: true }] },
502:     { policies: { writer: allow } },
503:     { policies: { writer: { ...allow, unknown: true }, reader: allow } },
504:     {
505:       contracts: {
506:         descriptors: [{ retrievalUri: `${scope}types/bad`, utf8: "payload_sentinel" }],
507:         schemas: [],
508:       },
509:     },
510:     {
511:       contracts: {
512:         descriptors: [],
513:         schemas: [
514:           { retrievalUri: `${scope}schemas/bad`, utf8: '{"$ref":"https://missing.test/schema"}' },
515:         ],
516:       },
517:     },
518:     {
519:       contracts: {
520:         descriptors: [
521:           {
522:             retrievalUri: "https://foreign.test/type",
523:             utf8: JSON.stringify({
524:               id: "https://foreign.test/type",
525:               name: "Foreign",
526:               describes: "bead",
527:               conformsTo: [],
528:             }),
529:           },
530:         ],
531:         schemas: [],
532:       },
533:     },
534:     {
535:       maximumEndpointMultiplicity: [
536:         { endpoint: "target", linkConformsTo: `${scope}types/absent`, max: 1 },
537:       ],
538:     },
539:   ]) {
540:     f.write({ ...ru(f.directory), configuration: { ...configuration, ...patch } });
541:     await failure(f.input, "contract");
542:   }
543:   const invalid = artifacts();
544:   invalid.descriptors[1]!.utf8 = JSON.stringify({
545:     id: `${scope}types/owner`,
546:     name: "Owner",
547:     describes: "bead",
548:     conformsTo: [],
549:     ownsOutgoing: { [`${scope}types/missing`]: { max: 2 } },
550:   });
551:   f.write({ ...ru(f.directory), configuration: { ...configuration, contracts: invalid } });
552:   await failure(f.input, "contract");
553: });

### packages/server/src/reference-deployment.ts:254–314

254: export async function loadReferenceDeployment(
255:   input: ReferenceDeploymentInput,
256: ): Promise<PreparedReferenceDeployment> {
257:   let phase: ReferenceDeploymentErrorCode = "configuration";
258:   try {
259:     const startup = snapshot(input);
260:     phase = "file";
261:     const bytes = await readDeployment(startup.deploymentFile);
262:     phase = "syntax";
263:     // Keep BOM visible to the existing strict JSON decoder.
264:     const lossless = decodeJsonDocument(
265:       new TextDecoder("utf-8", { fatal: true, ignoreBOM: true }).decode(bytes),
266:     );
267:     boundLossless(lossless);
268:     const admitted = admitJsonNumbers(lossless, {
269:       diagnostics: 1,
270:       diagnostic: () => ({ message: "unsupported reference deployment number" }),
271:     });
272:     if (!admitted.ok) refuse("syntax");
273:     phase = "contract";
274:     const envelope = administrativeData(admitted.value, ["format", "profile", "configuration"]);
275:     if (
276:       envelope.format !== "bdptest-reference-deployment-1" ||
277:       envelope.profile !== startup.profile
278:     )
279:       refuse("contract");
280:     const fileDigest = createHash("sha256").update(bytes).digest("hex");
281:     const summary = {
282:       format: "bdptest-reference-deployment-1" as const,
283:       canonicalScope: startup.canonicalScope,
284:       listener: startup.listener,
285:       preflightReadLimits: startup.limits,
286:       fileDigest,
287:     };
288:     if (startup.profile === "read-update") {
289:       const configuration = administrativeData(envelope.configuration, [
290:         "directory",
291:         "create",
292:         "credentials",
293:         "policies",
294:         "contracts",
295:         "maximumEndpointMultiplicity",
296:       ]);
297:       const projectedInput = immutable({
298:         scope: startup.canonicalScope,
299:         listener: startup.listener,
300:         ...configuration,
301:       });
302:       const installed = captureInstalledReferenceConfiguration(projectedInput);
303:       const handle: PreparedReadUpdateDeployment = Object.freeze({
304:         ...summary,
305:         profile: "read-update",
306:         artifactIdentity: installed.artifactIdentity,
307:         configurationIdentity: installed.identity,
308:         policyIdentity: installed.policyIdentity,
309:         authenticationIdentity: installed.authenticationIdentity,
310:       });
311:       phase = "internal";
312:       registerReferenceDeployment(handle, { profile: "read-update", projectedInput, installed });
313:       return handle;
314:     }

### packages/server/src/reference-deployment.ts:351–355

351:   } catch (error) {
352:     if (error instanceof ReferenceDeploymentError) throw error;
353:     // Neither causes nor parser/filesystem/compiler details cross this local boundary.
354:     throw new ReferenceDeploymentError(phase);
355:   }

### packages/server/src/installed-reference-configuration.ts:85–123

85: export function captureInstalledReferenceConfiguration(input: unknown) {
86:   const fields = administrativeData(input, [
87:     "scope",
88:     "listener",
89:     "directory",
90:     "create",
91:     "credentials",
92:     "policies",
93:     "contracts",
94:     "maximumEndpointMultiplicity",
95:   ]);
96:   const scope = parseCanonicalScope(string(fields.scope));
97:   const listener = administrativeData(fields.listener, ["host", "port"]);
98:   const host = string(listener.host, 128);
99:   if (
100:     !["127.0.0.1", "::1"].includes(host) ||
101:     !Number.isSafeInteger(listener.port) ||
102:     (listener.port as number) < 1 ||
103:     (listener.port as number) > 65535
104:   )
105:     throw new TypeError("explicit loopback listener required");
106:   const directory = string(fields.directory);
107:   if (!path.isAbsolute(directory) || typeof fields.create !== "boolean")
108:     throw new TypeError("explicit provisioning required");
109:   const credentials = administrativeArray(fields.credentials, 32).map((value) => {
110:     const entry = administrativeData(value, ["token", "principal"]);
111:     const token = string(entry.token, 256),
112:       principal = string(entry.principal, 128);
113:     if (!/^[A-Za-z0-9_-]{32,256}$/.test(token) || !/^[A-Za-z0-9_-]{1,128}$/.test(principal))
114:       throw new TypeError("invalid trusted credential");
115:     return Object.freeze({ token, principal });
116:   });
117:   if (!credentials.length || new Set(credentials.map((c) => c.token)).size !== credentials.length)
118:     throw new TypeError("unique credentials required");
119:   const access = createDevelopmentAccessPolicy(
120:     fields.policies as Record<string, DevelopmentAccess>,
121:   );
122:   for (const c of credentials) access.policy.capture(c.principal);
123:   const source = administrativeData(fields.contracts, ["descriptors", "schemas"]);

### packages/server/src/installed-reference-configuration.ts:135–168

135:   // Empty descriptors select the shipped, complete local reference closure.
136:   const descriptors = administrativeArray(source.descriptors, 4096);
137:   const bundle = {
138:     descriptors: descriptors.length
139:       ? artifacts(descriptors)
140:       : referenceTypeArtifacts(scope).map((d) => ({
141:           retrievalUri: d.id,
142:           utf8Bytes: Buffer.from(stringifyJsonValue(d)),
143:         })),
144:     schemas: artifacts(source.schemas),
145:     limits: CONTRACT_ARTIFACT_CEILINGS,
146:   };
147:   const contracts = compileDevelopmentContracts(bundle);
148:   // Hosted artifacts are a closed manifest under this configured Scope only.
149:   if (contracts.descriptors.some((d) => !d.id.startsWith(`${scope}types/`)))
150:     throw new TypeError("installed candidate Types must be Scope-local");
151:   const schemas = bundle.schemas.map((artifact) => {
152:     if (
153:       !artifact.retrievalUri.startsWith(`${scope}schemas/`) ||
154:       new URL(artifact.retrievalUri).search ||
155:       new URL(artifact.retrievalUri).hash
156:     )
157:       throw new TypeError("installed schema hosting must be Scope-local");
158:     return Object.freeze({ uri: artifact.retrievalUri, utf8: artifact.utf8Bytes.toString("utf8") });
159:   });
160:   const aggregate = captureDevelopmentAggregatePolicy(
161:     fields.maximumEndpointMultiplicity,
162:     contracts.descriptors,
163:   );
164:   const configuration = Object.freeze({
165:     contracts,
166:     access: access.policy,
167:     maximumEndpointMultiplicity: Object.freeze(aggregate),
168:   });

### packages/server/src/development-contracts.ts:23–53

23: export function administrativeData(
24:   value: unknown,
25:   keys: readonly string[],
26: ): Record<string, unknown> {
27:   if (
28:     !value ||
29:     typeof value !== "object" ||
30:     isProxy(value) ||
31:     ![Object.prototype, null].includes(Object.getPrototypeOf(value))
32:   )
33:     throw new TypeError("plain administrative data required");
34:   const descriptors = Object.getOwnPropertyDescriptors(value);
35:   if (
36:     Reflect.ownKeys(descriptors).length !== keys.length ||
37:     keys.some((k) => !descriptors[k] || !Object.hasOwn(descriptors[k], "value"))
38:   )
39:     throw new TypeError("exact administrative data fields required");
40:   return Object.fromEntries(keys.map((k) => [k, descriptors[k]?.value]));
41: }
42: export function administrativeArray(value: unknown, maximum: number): readonly unknown[] {
43:   if (!Array.isArray(value) || isProxy(value) || value.length > maximum)
44:     throw new TypeError("bounded administrative array required");
45:   const fields = Object.getOwnPropertyDescriptors(value);
46:   if (Reflect.ownKeys(fields).length !== value.length + 1)
47:     throw new TypeError("dense administrative array required");
48:   return Array.from({ length: value.length }, (_, i) => {
49:     const field = fields[String(i)];
50:     if (!field || !Object.hasOwn(field, "value"))
51:       throw new TypeError("administrative accessor refused");
52:     return field.value;
53:   });

### packages/server/src/development-contracts.ts:74–108

74: export function compileDevelopmentContracts(
75:   input: ContractArtifactBundleInput,
76: ): DevelopmentContracts {
77:   const captured = ingestContractArtifacts(input);
78:   // Reconstruct solely from captured bytes: caller mutation/accessors cannot alter compilation.
79:   const artifact = (kind: "descriptor" | "schema") =>
80:     captured.artifacts
81:       .filter((a) => a.kind === kind)
82:       .map((a) => ({
83:         retrievalUri: a.retrievalUri,
84:         utf8Bytes: Buffer.from(a.bytesBase64, "base64"),
85:       }));
86:   const bundle = {
87:     descriptors: artifact("descriptor"),
88:     schemas: artifact("schema"),
89:     limits: captured.receipt.limits,
90:   };
91:   const types: Record<string, string> = Object.create(null);
92:   const contracts = new Map<string, InstalledResourceContract>();
93:   for (const descriptor of captured.descriptors) {
94:     if (Buffer.byteLength(JSON.stringify(descriptor.id)) > 4096)
95:       throw new RangeError("Type diagnostic identifier bound");
96:     types[descriptor.id] = stringifyJsonValue(descriptor);
97:     let validateProperties: PropertyValidator | undefined;
98:     if (descriptor.propertiesSchema !== undefined) {
99:       const compiled = compilePrivateSchemaEvaluator({
100:         bundle,
101:         entry: { kind: "descriptor", typeId: descriptor.id },
102:         graphLimits: SCHEMA_GRAPH_CEILINGS,
103:         limits: EVALUATOR_CEILINGS,
104:       });
105:       if (compiled.kind !== "compiled")
106:         throw new TypeError(`administrative schema refused: ${compiled.reason}`);
107:       validateProperties = (properties, emitter) => {
108:         const outcome = compiled.evaluateUtf8(Buffer.from(stringifyJsonValue(properties)));

### packages/server/src/development-contracts.ts:181–243

181: function access(value: DevelopmentAccess): DevelopmentAccessSnapshot {
182:   if (!value || typeof value !== "object" || isProxy(value))
183:     throw new TypeError("invalid access policy");
184:   const fields = Object.getOwnPropertyDescriptors(value);
185:   if (
186:     Reflect.ownKeys(fields).length !== 4 ||
187:     ["create", "write", "hidden", "readOnly"].some(
188:       (k) => !fields[k] || !Object.hasOwn(fields[k], "value"),
189:     )
190:   )
191:     throw new TypeError("invalid access policy fields");
192:   if (typeof fields.create?.value !== "boolean" || typeof fields.write?.value !== "boolean")
193:     throw new TypeError("invalid access flags");
194:   const ids = (v: unknown): readonly string[] => {
195:     const values = administrativeArray(v, 1000).map((id) => {
196:       if (typeof id !== "string" || (!id.startsWith("beads/") && !id.startsWith("links/")))
197:         throw new TypeError("invalid access resource identity");
198:       assertCanonicalPathSegments(id, "access resource identity");
199:       return id;
200:     });
201:     return Object.freeze([...new Set(values)].sort());
202:   };
203:   const snapshot = {
204:     create: fields.create.value as boolean,
205:     write: fields.write.value as boolean,
206:     hidden: ids(fields.hidden?.value),
207:     readOnly: ids(fields.readOnly?.value),
208:   };
209:   return Object.freeze({
210:     ...snapshot,
211:     identity: createHash("sha256").update(stringifyJsonValue(snapshot)).digest("hex"),
212:   });
213: }
214: /** Private operator handle replaces access atomically between synchronous owner turns.
215:  * Existing captures retain immutable values. No mutation request can reach replace.
216:  */
217: export function createDevelopmentAccessPolicy(
218:   initial: Readonly<Record<string, DevelopmentAccess>>,
219: ) {
220:   const snapshots = new Map<string, DevelopmentAccessSnapshot>();
221:   function replace(principal: string, value: DevelopmentAccess) {
222:     if (!/^[A-Za-z0-9_-]{1,128}$/.test(principal)) throw new TypeError("invalid stable principal");
223:     snapshots.set(principal, access(value));
224:   }
225:   if (!initial || typeof initial !== "object" || isProxy(initial))
226:     throw new TypeError("principal policies required");
227:   for (const [principal, value] of Object.entries(
228:     administrativeData(initial, Object.keys(initial)),
229:   ))
230:     replace(principal, value as DevelopmentAccess);
231:   const policy: DevelopmentAccessPolicy = Object.freeze({
232:     identity: () =>
233:       createHash("sha256")
234:         .update(stringifyJsonValue([...snapshots.entries()].sort()))
235:         .digest("hex"),
236:     capture: (principal: string) => {
237:       const current = snapshots.get(principal);
238:       if (!current) throw new TypeError("principal lacks administrative policy");
239:       return current;
240:     },
241:   });
242:   policies.add(policy);
243:   return Object.freeze({ policy, replace });

### packages/server/src/development-contracts.ts:247–273

247: export function captureDevelopmentAggregatePolicy(
248:   input: unknown,
249:   descriptors: readonly TypeDescriptor[],
250: ): readonly MaximumEndpointMultiplicityPolicy[] {
251:   const entries = administrativeArray(input, 1000).map((item) => {
252:     const value = administrativeData(item, ["endpoint", "linkConformsTo", "max"]);
253:     if (
254:       (value.endpoint !== "source" && value.endpoint !== "target") ||
255:       typeof value.linkConformsTo !== "string" ||
256:       !descriptors.some((d) => d.id === value.linkConformsTo && d.describes === "link") ||
257:       typeof value.max !== "number" ||
258:       !Number.isSafeInteger(value.max) ||
259:       value.max < 1
260:     )
261:       throw new TypeError("invalid installed aggregate policy");
262:     return Object.freeze({
263:       endpoint: value.endpoint,
264:       linkConformsTo: value.linkConformsTo,
265:       max: value.max,
266:     }) as MaximumEndpointMultiplicityPolicy;
267:   });
268:   const canonical = new Map(entries.map((entry) => [stringifyJsonValue(entry), entry]));
269:   return Object.freeze(
270:     [...canonical.entries()]
271:       .sort(([a], [b]) => (a < b ? -1 : a > b ? 1 : 0))
272:       .map(([, entry]) => entry),
273:   );

### packages/server/src/installed-contract-artifact.ts:265–283

265: function checkDependencies(descriptors: readonly TypeDescriptor[], maximum: number): number {
266:   const byId = new Map(descriptors.map((descriptor) => [descriptor.id, descriptor]));
267:   let count = 0;
268:   const requireType = (id: string, category: "bead" | "link"): void => {
269:     if (count >= maximum) refuse("limit");
270:     count++;
271:     if (byId.get(id)?.describes !== category) refuse("descriptor-dependency");
272:   };
273:   for (const descriptor of descriptors) {
274:     for (const parent of descriptor.conformsTo) requireType(parent, descriptor.describes);
275:     if (descriptor.describes === "link") {
276:       for (const requirement of descriptor.source.conformsTo) requireType(requirement, "bead");
277:       for (const requirement of descriptor.target.conformsTo) requireType(requirement, "bead");
278:     } else if (descriptor.ownsOutgoing !== undefined) {
279:       for (const id of Object.keys(descriptor.ownsOutgoing))
280:         if (id !== "*") requireType(id, "link");
281:     }
282:   }
283:   return count;

### packages/server/src/installed-contract-artifact.ts:296–310

296: function captureContractArtifacts(input: ContractArtifactBundleInput): ContractArtifactCandidate {
297:   const fields = record(input, ["descriptors", "schemas", "limits"]);
298:   const limits = snapshotLimits(fields.limits);
299:   const descriptors = array(fields.descriptors, limits.descriptors);
300:   const schemas = array(fields.schemas, limits.schemas);
301:   const submittedArtifacts = descriptors.length + schemas.length;
302:   if (submittedArtifacts > limits.submittedArtifacts) refuse("limit");
303:   let submittedBytes = 0;
304:   let submittedUriBytes = 0;
305:   let syntaxNodes = 0;
306:   const captured = new Map<string, CapturedArtifact>();
307:   for (const [kind, entries] of [
308:     ["descriptor", descriptors],
309:     ["schema", schemas],
310:   ] as const) {

### packages/server/src/installed-contract-artifact.ts:357–383

357:       let syntax: LosslessJsonValue;
358:       try {
359:         syntax = decodeJsonDocument(text);
360:       } catch (error) {
361:         if (error instanceof JsonSyntaxError) refuse("json-syntax");
362:         throw error;
363:       }
364:       syntaxNodes += countSyntax(
365:         syntax,
366:         kind === "descriptor" ? limits.descriptorDepth : limits.schemaDepth,
367:         limits.syntaxNodes - syntaxNodes,
368:       );
369:       let descriptor: TypeDescriptor | undefined;
370:       if (kind === "descriptor") {
371:         const value = descriptorValue(syntax);
372:         try {
373:           descriptor = parseTypeDescriptor(value);
374:         } catch (error) {
375:           if (error instanceof ProtocolArtifactValidationError) refuse("descriptor");
376:           throw error;
377:         }
378:         if (descriptor.id !== uri) refuse("descriptor");
379:         if (descriptor.describes === "bead" && descriptor.ownsOutgoing !== undefined) {
380:           for (const declaration of Object.values(descriptor.ownsOutgoing))
381:             if (declaration.max > limits.maxOwnershipMax) refuse("limit");
382:         }
383:       }

### packages/server/src/installed-contract-artifact.ts:405–440

405:   const artifacts = [...captured.values()].sort((a, b) => Buffer.compare(a.uriBytes, b.uriBytes));
406:   const parsed = Object.freeze(
407:     artifacts.flatMap((artifact) =>
408:       artifact.descriptor === undefined ? [] : [artifact.descriptor],
409:     ),
410:   );
411:   const dependencyReferences = checkDependencies(parsed, limits.dependencyReferences);
412:   let conformance: MeasuredTypeConformanceIndex;
413:   try {
414:     conformance = createTypeConformanceIndex(parsed, {
415:       nodes: limits.descriptors,
416:       edges: limits.conformanceEdges,
417:       depth: limits.conformanceDepth,
418:       memberships: limits.effectiveMemberships,
419:       unionAttempts: limits.unionAttempts,
420:     });
421:   } catch (error) {
422:     if (error instanceof ProtocolArtifactValidationError) refuse("conformance");
423:     throw error;
424:   }
425:   return Object.freeze({
426:     format: CONTRACT_ARTIFACT_FORMAT,
427:     sha256: digest(artifacts),
428:     artifacts: Object.freeze(artifacts.map((artifact) => artifact.entry)),
429:     descriptors: parsed,
430:     conformance,
431:     receipt: Object.freeze({
432:       limits,
433:       submittedArtifacts,
434:       submittedBytes,
435:       submittedUriBytes,
436:       syntaxNodes,
437:       dependencyReferences,
438:       conformance: conformance.statistics,
439:     }),
440:   });

### packages/server/src/installed-schema-evaluator.ts:157–173

157: function refusal(
158:   error: unknown,
159:   phase: "compile" | "instance",
160: ): Extract<EvaluationOutcome, { kind: "refused" }> {
161:   const reason =
162:     error instanceof EvaluationRefusal
163:       ? error.code
164:       : error instanceof DecimalLimitError
165:         ? "limit-coefficientDigits"
166:         : error instanceof SchemaGraphError
167:           ? `graph:${error.code}`
168:           : error instanceof ContractArtifactError
169:             ? `artifact:${error.code}`
170:             : error instanceof ContractArtifactLocalError
171:               ? "local-failure"
172:               : "local-failure";
173:   return Object.freeze({ kind: "refused", stage, phase, reason });

### packages/server/src/installed-schema-evaluator.ts:180–201

180: export function compilePrivateSchemaEvaluator(input: Compilation): CompilationOutcome {
181:   try {
182:     const fields = data(input, ["bundle", "graphLimits", "entry", "limits"]);
183:     const budget = new EvaluationBudget(fields.limits as EvaluatorLimits);
184:     // Select the entry via data descriptors, without invoking a supplied getter.
185:     const raw = fields.entry;
186:     if (!raw || typeof raw !== "object" || isProxy(raw)) throw new EvaluationRefusal("entry");
187:     const kind = Object.getOwnPropertyDescriptor(raw, "kind")?.value;
188:     const entry = data(raw, kind === "descriptor" ? ["kind", "typeId"] : ["kind", "retrievalUri"]);
189:     if (kind !== "descriptor" && kind !== "schema") throw new EvaluationRefusal("entry");
190:     const identifier = kind === "descriptor" ? entry.typeId : entry.retrievalUri;
191:     if (typeof identifier !== "string") throw new EvaluationRefusal("entry");
192:     const graph = buildInstalledSchemaGraph(
193:       fields.bundle as ContractArtifactBundleInput,
194:       fields.graphLimits as SchemaGraphLimits,
195:     );
196:     budget.work(graph.descriptorRoots.length + graph.resources.length);
197:     const root =
198:       kind === "descriptor"
199:         ? graph.descriptorRoots.find((r) => r.descriptor === identifier)?.node
200:         : graph.resources.find((r) => r.retrievalAlias === identifier)?.node;
201:     if (root === undefined) throw new EvaluationRefusal("entry");

### packages/server/src/installed-schema-graph.ts:176–191

176: function build(input: ContractArtifactBundleInput, b: GraphBudget): SchemaGraphCandidate {
177:   const artifact = ingestContractArtifacts(input);
178:   // Verify the aggregate against the same immutable rows indexed below. Keep
179:   // this per-build and charged; no shared parsed prelude or readiness cache.
180:   const builtinHash = createHash("sha256");
181:   for (const doc of BUILTIN_SCHEMA_DOCUMENTS) {
182:     b.charge("scanWork", doc.uri.length + doc.sha256.length + 2);
183:     builtinHash.update(`${doc.uri}\0${doc.sha256}\n`);
184:   }
185:   if (builtinHash.digest("hex") !== BUILTIN_SCHEMA_DIGEST) refuseGraph("internal-invariant");
186:   const documents = [
187:     ...artifact.artifacts
188:       .filter((a) => a.kind === "schema")
189:       .map((a) => ({ uri: a.retrievalUri, bytesBase64: a.bytesBase64, sha256: a.sha256 })),
190:     ...BUILTIN_SCHEMA_DOCUMENTS,
191:   ];

### packages/server/src/installed-schema-graph.ts:436–446

436:   const resolve = (
437:     base: string,
438:     raw: string,
439:     location: number | string,
440:   ): { resolved: ResolvedSchemaUri; target: number; dynamic: boolean } => {
441:     try {
442:       const resolved = resolveSchemaUri(base, raw, b);
443:       retain(resolved.uri);
444:       b.charge("scanWork", resolved.resource.length);
445:       const rid = identities.get(resolved.resource);
446:       if (rid === undefined) refuseGraph("missing-resource");

### packages/server/src/installed-schema-graph.ts:498–525

498:   for (const ref of unresolved) {
499:     const node = nodes[ref.node] as SchemaNode;
500:     const r = resolve((resources[node.resource] as SchemaResource).canonicalUri, ref.raw, ref.node);
501:     const dynamic = ref.keyword === "$dynamicRef" && r.dynamic;
502:     references.push(
503:       Object.freeze({
504:         ...ref,
505:         resolvedUri: r.resolved.uri,
506:         target: r.target,
507:         fragmentKind: r.resolved.fragmentKind,
508:         kind: dynamic ? "dynamic-anchor" : "static",
509:         ...(dynamic ? { anchor: r.resolved.fragment } : {}),
510:       }),
511:     );
512:   }
513:   const descriptorRoots = [];
514:   for (const d of artifact.descriptors)
515:     if (d.propertiesSchema !== undefined) {
516:       b.charge("descriptorRoots");
517:       const r = resolve(d.id, d.propertiesSchema, d.id);
518:       descriptorRoots.push(
519:         Object.freeze({
520:           descriptor: d.id,
521:           originalUri: d.propertiesSchema,
522:           resolvedUri: r.resolved.uri,
523:           node: r.target,
524:           resource: (nodes[r.target] as SchemaNode).resource,
525:         }),

### packages/protocol/src/reference-domain.ts:1–73

1: import type { TypeDescriptor, TypeSummary } from "./index.js";
2: 
3: export const REFERENCE_BEAD_TYPES = [
4:   ["https://work.example/types/tracked-item", "Tracked Item"],
5:   ["https://work.example/types/work-item", "Work Item"],
6:   ["https://work.example/types/task", "Task"],
7:   ["https://work.example/types/bug", "Bug"],
8:   ["https://work.example/types/feature", "Feature"],
9:   ["https://work.example/types/chore", "Chore"],
10:   ["https://work.example/types/epic", "Epic"],
11:   ["https://work.example/types/decision", "Decision"],
12: ] as const;
13: 
14: export const REFERENCE_BLOCKING_LINK_TYPE_ID = "https://work.example/types/blocks";
15: 
16: export const REFERENCE_LINK_TYPES = [
17:   ["https://work.example/types/relationship", "Relationship"],
18:   ["https://work.example/types/dependency", "Dependency"],
19:   [REFERENCE_BLOCKING_LINK_TYPE_ID, "Blocks"],
20:   ["https://work.example/types/relates", "Relates"],
21:   ["https://work.example/types/cites", "Cites"],
22: ] as const;
23: 
24: export const REFERENCE_TYPE_SUMMARIES: readonly TypeSummary[] = [
25:   ...REFERENCE_BEAD_TYPES.map(([id, name]) => ({ id, name, describes: "bead" as const })),
26:   ...REFERENCE_LINK_TYPES.map(([id, name]) => ({ id, name, describes: "link" as const })),
27: ];
28: 
29: export const REFERENCE_TYPE_DESCRIPTORS: readonly TypeDescriptor[] = REFERENCE_TYPE_SUMMARIES.map(
30:   (summary): TypeDescriptor =>
31:     summary.describes === "link"
32:       ? {
33:           id: summary.id,
34:           name: summary.name,
35:           describes: "link",
36:           conformsTo:
37:             summary.id === REFERENCE_BLOCKING_LINK_TYPE_ID
38:               ? ["https://work.example/types/dependency"]
39:               : summary.id === "https://work.example/types/dependency"
40:                 ? ["https://work.example/types/relationship"]
41:                 : [],
42:           source: { conformsTo: [] },
43:           target: { conformsTo: [] },
44:         }
45:       : {
46:           id: summary.id,
47:           name: summary.name,
48:           describes: "bead",
49:           conformsTo:
50:             summary.id === "https://work.example/types/work-item"
51:               ? ["https://work.example/types/tracked-item"]
52:               : summary.id !== "https://work.example/types/tracked-item" &&
53:                   summary.id !== "https://work.example/types/decision"
54:                 ? ["https://work.example/types/work-item"]
55:                 : [],
56:           // Decision owns every outgoing Link, with explicit citations and an empty
57:           // relates group; Feature witnesses empty wildcard-only ownership.
58:           // The reference-domain bounds keep the plane always
59:           // servable inline; the label is descriptor documentation and
60:           // never appears in Resource records.
61:           ...(summary.id === "https://work.example/types/decision"
62:             ? {
63:                 ownsOutgoing: {
64:                   "*": { max: 8 },
65:                   "https://work.example/types/relates": { max: 1 },
66:                   "https://work.example/types/cites": { label: "cites", max: 8 },
67:                 },
68:               }
69:             : summary.id === "https://work.example/types/feature"
70:               ? { ownsOutgoing: { "*": { max: 8 } } }
71:               : {}),
72:         },
73: );

### packages/server/src/reference-deployment-state.ts:1–32

1: /** Private identity registry. Never exported by a package entry. */
2: import type { PreparedReferenceDeployment } from "./reference-deployment.js";
3: import type { captureInstalledReferenceConfiguration } from "./installed-reference-configuration.js";
4: import type { captureInstalledTransactionConfiguration } from "./transaction-installed-configuration.js";
5: import type { captureTransactionRuntimeConfiguration } from "./transaction-runtime-configuration.js";
6: 
7: export type ReferenceDeploymentState =
8:   | {
9:       readonly profile: "read-update";
10:       readonly projectedInput: Readonly<Record<string, unknown>>;
11:       readonly installed: ReturnType<typeof captureInstalledReferenceConfiguration>;
12:     }
13:   | {
14:       readonly profile: "transactional";
15:       readonly projectedInput: Readonly<Record<string, unknown>>;
16:       readonly installed: ReturnType<typeof captureInstalledTransactionConfiguration>;
17:       readonly preflight: ReturnType<typeof captureTransactionRuntimeConfiguration>;
18:     };
19: 
20: const states = new WeakMap<object, ReferenceDeploymentState>();
21: export function registerReferenceDeployment(
22:   handle: PreparedReferenceDeployment,
23:   state: ReferenceDeploymentState,
24: ): void {
25:   if (states.has(handle)) throw new TypeError("reference deployment already registered");
26:   states.set(handle, Object.freeze(state));
27: }
28: export function referenceDeploymentState(handle: unknown): ReferenceDeploymentState {
29:   const state = handle !== null && typeof handle === "object" ? states.get(handle) : undefined;
30:   if (state === undefined) throw new TypeError("registered reference deployment required");
31:   return state;
32: }
