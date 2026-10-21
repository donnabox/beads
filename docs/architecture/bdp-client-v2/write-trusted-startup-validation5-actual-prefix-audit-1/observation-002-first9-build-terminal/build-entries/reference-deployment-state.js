const states = new WeakMap();
export function registerReferenceDeployment(handle, state) {
    if (states.has(handle))
        throw new TypeError("reference deployment already registered");
    states.set(handle, Object.freeze(state));
}
export function referenceDeploymentState(handle) {
    const state = handle !== null && typeof handle === "object" ? states.get(handle) : undefined;
    if (state === undefined)
        throw new TypeError("registered reference deployment required");
    return state;
}
//# sourceMappingURL=reference-deployment-state.js.map