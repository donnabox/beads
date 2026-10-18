# Owned stacked draft synchronization — October 17, 2026

Publication draft merge `83f696c7cc54d9434ef78ebd1bfc4dd2eccff465` joins publication `649350ba8` with reader `b5973def8` without rebasing or force-pushing. It brings the existing reader feed-error and nested-detachment test repairs into the publication lane. The public format and CLI are not implemented here.

The focused canonical no-engine selection passed 68 test records (40 reader, 28 publication), zero failures/skips. The source/parents/hash/command receipt and raw log are adjacent. All 23 import-only commits above the original qualified base were checked against merged ancestry; none is present. The worktree was clean after the merge and no test-owned subprocess remained. No installed, engine or full-suite qualification is claimed.

Normal merge hooks were bypassed to keep ambient bd/database operations out of Janet’s exclusive engine slot. Focused formatting/source checks and pure tests ran directly; full hook, CI, exact-source installed embedded/shared-server and final reader review gates remain open. Existing inherited atomicfile durability/ACL/symlink limitations remain documented; this sync adds no publication guarantee.
