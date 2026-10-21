/** Private identity registry. Never exported by a package entry. */
import type { PreparedReferenceDeployment } from "./reference-deployment.js";
import type { captureInstalledReferenceConfiguration } from "./installed-reference-configuration.js";
import type { captureInstalledTransactionConfiguration } from "./transaction-installed-configuration.js";
import type { captureTransactionRuntimeConfiguration } from "./transaction-runtime-configuration.js";
export type ReferenceDeploymentState = {
    readonly profile: "read-update";
    readonly projectedInput: Readonly<Record<string, unknown>>;
    readonly installed: ReturnType<typeof captureInstalledReferenceConfiguration>;
} | {
    readonly profile: "transactional";
    readonly projectedInput: Readonly<Record<string, unknown>>;
    readonly installed: ReturnType<typeof captureInstalledTransactionConfiguration>;
    readonly preflight: ReturnType<typeof captureTransactionRuntimeConfiguration>;
};
export declare function registerReferenceDeployment(handle: PreparedReferenceDeployment, state: ReferenceDeploymentState): void;
export declare function referenceDeploymentState(handle: unknown): ReferenceDeploymentState;
//# sourceMappingURL=reference-deployment-state.d.ts.map