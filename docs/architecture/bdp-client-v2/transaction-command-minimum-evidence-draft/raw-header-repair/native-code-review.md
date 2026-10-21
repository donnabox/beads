# Native code review — raw authorization-view admission repair

No actionable findings in the frozen five-path correction. The implementation matches the accepted narrow plan and closes both independent authored-admission false positives while preventing the same field from becoming a credential handle. This is a clean source-review conclusion, not a test result or installed qualification.

## Exact reviewed binding

Worktree /Users/dbox/repos/wt/vickie/vickie-command-minimum-evidence; ordinary HEAD453b6dab386746bb03aac2c0a293572b8a71c969/tree40782d97c9764bc19ae78c46d1d39475e3db317a.
- source-binding.json:612cbb471df504b3a0440e11f15a29911cf59aeeae89247836f04d5b19dc580b
- Five-path correction.patch:4b7a84fb9910a686a6c154b86eb6bf1addc932787ca47c9c11598132215209be
- Complete ten-path draft.patch:fc4fd747379917531c4d64d3fe52603337e72ec3f78e09aedcab8e9a8cc0b4b7
- selection.json:17859e4aff9ab0863077df90dab022dfeeef9a80f6625a113398d47b68ac8865
- Author report.md:5e44e62522ee7ab1acd2d3a31516f3407b2e357dd1fe2a65cc3cbda29bb4fd3c
- Reviewed plan:db1599a1f1b38f437722b8abae8acda1847c1b1e4fd8bb2c983b118d32b2acb6

Independently recomputed all ten file hashes, compared actual git diff bytes to both retained patches, and checked the exact ten modified paths. The original five files and their restricted diff remain byte-identical to event-final/source-binding.json and ea04e6ba34c9e1d9b724348f32407c6fd6bd78b795cad0481e7f4093b0f21762. The new correction is252 insertions/12 deletions across exactly two production files and three test files.

## Production and security review

executable-manifest.ts:421–429 adds only a module-private predicate. It lowercases for classification, returns false only for exact bdp-authorization-view, and otherwise retains both original explicit and broad sensitive-name patterns. Its raw use at984 still combines with the unchanged executor-owned set. Its normalized use at1477 still follows the unchanged lowercase-name and HTTP name/value checks. Raw casing, repeated lines and empty field values are preserved; no prefix/family exception is introduced.

http-executor.ts:343–345 applies the same exact exception around the unchanged SECRET_FIELD. The authored branch at675 retains reserved-field rejection, own-data checks, bounds and byte-preserving append. The inverse credential-handle use at479 now rejects the exact context field before an executor is returned. Lowercase handle names, sensitive-name eligibility, reserved names and duplicate checks remain. Resolver membership/cardinality and route safety remain unchanged; a resolver cannot add an undeclared context field.

Actual Authorization/Cookie and token/secret/password/credential/auth lookalikes remain protected. Unicode or malformed-name lookalikes cannot exploit lowercasing because the ASCII HTTP field-name validation remains in place. No new export, transport seam, report permission, runtime/fixture/API or public profile admission is added.

runner.ts and all report allowlists/projections are unchanged. The exact report retains Authorization only as an allowlisted redacted name; context fields are omitted. The app’s names-first diagnostics and complete runner exact-header witness remain unchanged.

## Regression oracle and lifecycle review

- executable-manifest.test.ts:607ff covers both supported versions, both normalized request/action forms, empty/nonempty values and continued mixed-case normalized rejection. :1536ff covers raw request/action forms with exact lowercase and mixed-case repeated view lines, including an empty value, plus the other two context fields. Parsed line arrays and JSON reload equality are asserted.
- http-executor.test.ts:730–788 makes one real loopback exchange. It accumulates TCP bytes until complete headers plus the two-byte body are present, then responds once. The entire expected raw head includes authored context lines, Authorization, defaults and framing in exact order; body and exactRequest are independently checked. Resolver runs once and legacy resolver remains unused. Existing tracked sockets/listeners and afterEach destruction/close remain, with no new child/store lifecycle.
- :790–803 checks synchronous refusal of exact lowercase context, mixed-case context and ordinary nonsecret credential handles. The constructor failure occurs before an executor is available; no fabricated request-write-state assertion is needed.
- :806–861 checks all six proposed prefix/suffix lookalikes through both raw parser forms and the direct serializer. Parser assertions require the specific source-header refusal issue, not any unrelated parse failure. Serializer assertions require configuration/not-started plus no resolver calls.
- runner.test.ts:5575–5618 uses the existing mocked exact-transport test seam, not installed evidence. It inserts a synthetic credential at the correct authored-prefix boundary, requires scenario pass and authored-head pass, and then checks context omission plus both sentinel absences. Four elisions are correct: two context lines, Host and Connection. Authorization is retained only with redacted value. Assertions are not vacuous on rejected admission; exact executor is called once and ordinary executor is not called.

No static type/API mismatch is apparent in the added tests: the parser accepts unknown source input, ManifestValidationError.issues supplies the inspected message records, and exact-mode credential configuration follows existing test shapes. This statement does not substitute for compiler execution.

## Selection and limits

Independently checked selection.json has30 unique path/full-name pairs and that every pair matches its anchored regex. The seven expanded new cases and fifteen existing focused cases agree with the source titles/parameter lists. Eight existing evidence callbacks include the historical preservation/fence callback; no whole-file selection is requested.

Source-modelled workload remains one loopback request/listener, zero stores or product/installed children for this30-case selection. Worker and build processes are separate. Existing installed evidence remains208 expected/225 maximum HTTP and26children, with209 expected HTTP only when adding the separate loopback test. No runtime/test deadline or size/route limit is raised.

The original five-file static failure is preserved and must not be relabeled as a pass. Fresh type/build/format/lint/boundary/Read checks and the frozen30-case execution remain separate root-owned validation. The changed production conformance source must be rebuilt before emitted-app/installed observations. Actual Claude final-code review, exact-head CI, source/package provenance and root publication acceptance remain outstanding gates; this report claims none of them.

## Final correction file hashes

- packages/conformance/src/executable-manifest.ts:65af0451cd046d57c10b0666e54efbaaecc2225551d3accc5cc1d032322913fa
- packages/conformance/src/http-executor.ts:770a8405ae1532484b5a2658619c067832644947888f95e0c6575002376f1b5d
- packages/conformance/src/executable-manifest.test.ts:2e548c9e7ef8ff01e4933ce33f48bd97e91e277aebdd70c658c350c081e9ba7c
- packages/conformance/src/http-executor.test.ts:4255422c6b7c5b0a0cadd71c73a045a3a2ae12b5e17708c14afc2607b6007aee
- packages/conformance/src/runner.test.ts:3933d97fcdae91efeb23e24af468791d8917588b159f228433fceaa1ae317485

Read-only code, artifact and source-hash inspection. No tests, build, imports, services, installs, package operations, source edits or actual Claude launch performed. Report write used the required live-session guard; final source hashes were checked again at publication.
