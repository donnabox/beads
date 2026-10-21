# Alias integration root source verification

Independently verified clean7171251, direct docs-only successor offc980. fc980 is exact automatic merge3959+8d238, no source resolutions. Six non-document alias files equal3959; four evidence files equal8d238; all remaining tracked files equal main0432. Only exception is the scoped historical documentation sentence, reproduced below. Fresh evidence and independent integration closure remain separate gates; no profile or release claim.

```diff
diff --git a/docs/design/transaction-implementation-progress.md b/docs/design/transaction-implementation-progress.md
index c7e7bd4..cd8766c 100644
--- a/docs/design/transaction-implementation-progress.md
+++ b/docs/design/transaction-implementation-progress.md
@@ -149,8 +149,9 @@ persisted witnesses are owner faults, not ordinary missing aliases or key
 conflicts. These are internal metadata, not new receipt or wire fields.
 
 Alias operations remain forbidden in batches, as required by the normative
-contract. There are no HTTP alias routes, redirects, replica alias projections
-or new profile claims.
+contract. These alias singletons add no HTTP alias routes, redirects, replica
+alias projections or new profile claims; current alias read routing is described
+under the private HTTP receiver and current alias minimums below.
 
 ## First private sequence adapter
```
