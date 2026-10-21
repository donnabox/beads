/** Handle-only private lifecycle. Neither exported admission nor a public startup path. */
import type { PreparedReferenceDeployment } from "./reference-deployment.js";
import { referenceDeploymentState } from "./reference-deployment-state.js";
import { startPreparedDevelopmentReferenceRuntime } from "./development-runtime.js";
import { startPreparedTransactionRuntime } from "./transaction-runtime.js";

export async function startPreparedReferenceDeployment(
  handle: PreparedReferenceDeployment,
  signal?: AbortSignal,
) {
  const state = referenceDeploymentState(handle);
  signal?.throwIfAborted();
  const runtime =
    state.profile === "read-update"
      ? await startPreparedDevelopmentReferenceRuntime(handle, signal)
      : await startPreparedTransactionRuntime(handle, signal);
  if (
    !("completion" in runtime) ||
    !("ownerFacts" in runtime) ||
    !("history" in runtime) ||
    !("listenUrl" in runtime) ||
    !(runtime.completion instanceof Promise) ||
    typeof runtime.scope !== "string" ||
    typeof runtime.listenUrl !== "string" ||
    typeof runtime.history !== "boolean"
  ) {
    const primary = new Error("prepared lifecycle unavailable");
    try {
      await runtime.close();
    } catch (cleanup) {
      throw new AggregateError([primary, cleanup], "prepared lifecycle cleanup failed", {
        cause: primary,
      });
    }
    throw primary;
  }
  // No captures, token, target map, file/authentication digest or capability escapes.
  return Object.freeze({
    scope: runtime.scope,
    listenUrl: runtime.listenUrl,
    profile: state.profile,
    history: runtime.history,
    claimEligible: false as const,
    ownerFacts: runtime.ownerFacts,
    close: runtime.close,
    completion: runtime.completion,
  });
}
