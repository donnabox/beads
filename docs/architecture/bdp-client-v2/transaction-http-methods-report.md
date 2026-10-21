# T2b method-table repair

Frozen clean head:8844fe738d7457d6c9d3ea71b0b8d0fa778eecde
Parent:b1de5f342fb8c8b90c6909458018b892534863f0
Checkout:/Users/dbox/repos/wt/vickie/vickie-transaction-http-methods
Branch:codex/vickie-transaction-http-methods
Origin:https://github.com/gastownhall/bdp.git
No outward writes. Frozen b1 checkout unchanged.

Native findings: unauthenticated POST receipts/unknown incorrectly reached401;
raw TRACE operation request threw during Fetch Request construction and became500.
Normative basis:bdp.md4599–4605 defines wrong-method405 with target Allow and no
BDP Problem. Authentication-first4618 applies to actual receipt/page reads.

A single source-private methodResponse gate now classifies canonical target
shapes without consulting receipt existence/principal. Both raw listener and
handler use it. Raw listener invokes it after existing raw target/Host validation
and before Fetch Request construction or body consumption. Refused connections
close. Receipt GET/HEAD retain authentication and nondisclosure ordering.

Node routes CONNECT through its own event rather than normal request delivery;
that event now uses the same dispatch/refusal path and never establishes a tunnel.
TRACE and CONNECT do not reach Fetch Request. TRACK is rejected by Node24's HTTP
parser with native400 before dispatch (observed during test development); no raw
parser replacement or broader arbitrary-method acceptance is claimed.

Files only:transaction-http.ts, transaction-runtime.ts, transaction-http.test.ts.
No owner/receipt/admission/wire schema or public profile changes.

Verification Node24.16.0/pnpm11.20:
- frozen install; build before emitted tests PASS.
- focused-final.log:46 tests/3 files (HTTP/runtime/service), PASS.
- One rawsocket test matrix covers60 authenticated/unauthenticated wrong-method
  combinations:PUT/DELETE/OPTIONS/TRACE/CONNECT across two operation targets,
  known/unknown receipt IDs, receipt root, operation directory; plus4 POST
  receipt combinations and4 unauthenticated GET/HEAD cases. Bodyless405, Allow,
  no auth/context disclosure, preserved retained receipt checked.
- typecheck, boundaries, format, lint, installed smoke, development smoke,
  unchanged evidence verification PASS (separate logs).
- git diff --check PASS, committed worktree clean.
- Earlier raw matrix attempted TRACK and failed on Node-native400; narrowed to
  actual accepted Node methods and recorded this transport limitation above.

Root may integrate/review this successor independently. No next slice started.
