# Root-executable source binding, unexecuted

`bind-source.py` is a data-only binder. It has not run; no final-source.json, input-binding.json, root-release.json or execution-gate.json was created. The15 prepared execution files and prior cache-correction history are unchanged.

Required invocation inputs: exact full sequence author commit, combined merge commit and tree, plus root's accepted source-only authorization JSON and its SHA256. The root authorization must be inside the combined-validation directory and contain `accepted:true`, the exact `head`, and `sourceOnly:true`. `root-authorization.example.json` is closed.

The script verifies the session guard; all18 pinned map/council/review/audit inputs; the current20-path map6180b009 and final seven-file sequence bindingb937a736; minimum5ca/tree274766; clean actual SW HEAD/tree and ordered parents[sequence-author,minimum5ca]; sequence author as a direct df80 child with exactlyseven frozen file blobs; every combined20 mapped blob/length/mode; all outside-union df80 objects/modes/types; all tracked working bytes against actual Git blobs; all15 prepared hashes; Node24.16.0/pnpm11.20.0/Vitest/Biome binaries and entry realpaths; source lock. It reads only supplied branch objects, never main/remotes. No fetch, merge, Git write or product import occurs.

Only after every check passes does it exclusively write final-source.json, input-binding.json, binding-verification.json and root-release.json. It writes execution-gate.json last, with only this exact source prefix released:

`typecheck → fresh-build → format-check → lint → boundaries → read-seal → preparation-positive → sequence-positive → http-positive → minimum-http → build-inventory`

This is23 selected callbacks/193 modeled sourceHTTP. Control, package, installed and adjudication phase flags stayfalse; consumer lock/provenance remainnull. The script does not launch even the released source commands. Root's invocation authorization is recorded explicitly; mechanical checks are not misrepresented as a new independent review.

Root invocation (replace every placeholder with actual supplied values; run only after review/merge):

```sh
python3 /Users/dbox/repos/wt/vickie/bdp-integration-logs/guarded-action.py python3 /Users/dbox/repos/wt/vickie/bdp-integration-logs/sequence-minimum-combined-validation/binding-preparation/bind-source.py --sequence-author SEQUENCE_AUTHOR_40HEX --combined-head COMBINED_HEAD_40HEX --combined-tree COMBINED_TREE_40HEX --root-authorization /Users/dbox/repos/wt/vickie/bdp-integration-logs/sequence-minimum-combined-validation/root-source-authorization.json --root-authorization-sha256 AUTHORIZATION_64HEX
```

Checks performed during preparation: stdlib AST parse of binder text; rehash all18 frozen input artifacts; read-only base tree inspection confirms only regular100644/100755 blobs, matching the binder's fail-closed mode constraint. No binder execution or simulation, test, build, source mutation, or gate release. Any changed final-source composition, tool hash or council input stops rather than guessing an identity or overriding a failure.
