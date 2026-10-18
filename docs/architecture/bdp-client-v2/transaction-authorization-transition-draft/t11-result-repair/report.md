# T11 positive Read oracle repair

Frozen one-file successor: the two initial calls remain exactly the same. Their expected AuthorityReadResult tag is success, matching authority-read.ts:164–174 and the actual v4 result. Both complete bodies now equal the real seed receipt's Resource, with explicit success/seed guards; no weaker partial Resource assertion and no added owner/read/HTTP/callback.

Only transaction-authorization.test.ts differs from final-code-v3. Fourteen other files, including every production source, retain exact hashes. Full patch cc03010d7c9390cb3034eaf18eebd9b2ac5dd3522aa72e998e991b5d5e20e189; source binding f7cf0b51f79c56ee6b24a5be571336066a7148b97e1aa769d511a45579646cbd; correction db7ea6adf47c4aeb0c4cf233722c0b5d0731d34554aa05976f7d00dc019b4062.

The predecessor v4 remains failed: six static stages and transport2/queued1/stale1 passed; owner6 passed and T11 failed at the first incorrect kind expectation. Later T11 assertions were not reached. No complete eleven-case success is claimed. Actual v4 build can bridge unchanged production bytes; this test-only repair does not require rebuilding production. Root should typecheck newly narrowed body expressions and run only the exact affected callback under its existing bounds.

Planned11 callbacks/21HTTP/14 successful owners/11 ordinary directories+5 rejected copies/2 listeners/0 product children unchanged. No validation command, service, build, formatter, reviewer or publish executed by this repair author. Historical failures untouched; Claude packets stay closed pending targeted T11 proof.
