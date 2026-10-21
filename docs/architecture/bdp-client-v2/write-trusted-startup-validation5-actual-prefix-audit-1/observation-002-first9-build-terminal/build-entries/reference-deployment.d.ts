export interface ReferenceDeploymentReadLimits {
    readonly page: {
        readonly defaultItems: number;
        readonly maximumItems: number;
    };
    readonly selector: {
        readonly bytes: number;
        readonly depth: number;
        readonly nodes: number;
    };
    readonly cursorTtlMilliseconds: number;
}
export interface ReferenceDeploymentInput {
    readonly deploymentFile: string;
    readonly scope: {
        readonly url: string;
    };
    readonly server: {
        readonly host: string;
        readonly port: number;
        readonly advertisedProfile?: string;
        readonly limits: ReferenceDeploymentReadLimits;
        readonly internalFaultResource?: string;
    };
    readonly bdptest?: {
        readonly fixture?: string;
        readonly deploymentFile?: string;
    };
}
interface PreparedSummary {
    readonly format: "bdptest-reference-deployment-1";
    readonly canonicalScope: string;
    readonly listener: {
        readonly host: string;
        readonly port: number;
    };
    /** Code/capture comparison only; persisted owner readback remains mandatory. */
    readonly preflightReadLimits: ReferenceDeploymentReadLimits;
    readonly artifactIdentity: string;
    readonly fileDigest: string;
}
export interface PreparedReadUpdateDeployment extends PreparedSummary {
    readonly profile: "read-update";
    readonly configurationIdentity: string;
    readonly policyIdentity: string;
    readonly authenticationIdentity: string;
}
export interface PreparedTransactionalDeployment extends PreparedSummary {
    readonly profile: "transactional";
}
export type PreparedReferenceDeployment = PreparedReadUpdateDeployment | PreparedTransactionalDeployment;
export type ReferenceDeploymentErrorCode = "configuration" | "file" | "bounds" | "syntax" | "contract" | "internal";
export declare class ReferenceDeploymentError extends Error {
    readonly code: ReferenceDeploymentErrorCode;
    constructor(code: ReferenceDeploymentErrorCode);
}
/** The byte ceiling is the sole predecode bound. The strict decoder allocates
 * its lossless tree before the iterative depth/node checks below. Trusted local
 * input accepts that bounded-byte allocation/work risk. */
export declare function loadReferenceDeployment(input: ReferenceDeploymentInput): Promise<PreparedReferenceDeployment>;
export {};
//# sourceMappingURL=reference-deployment.d.ts.map