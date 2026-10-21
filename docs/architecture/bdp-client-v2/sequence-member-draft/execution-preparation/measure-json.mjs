// Standalone JSON/text measurement only. No product/fixture imports or test execution.
// Root must explicitly release measurement-gate.json after the four-path draft freezes.
import fs from "node:fs";
import path from "node:path";
import crypto from "node:crypto";
import { fileURLToPath } from "node:url";
const P = path.dirname(fileURLToPath(import.meta.url)), D = path.dirname(P);
const W = "/Users/dbox/repos/wt/vickie/vickie-sequence-member-evidence";
const hash = bytes => crypto.createHash("sha256").update(bytes).digest("hex");
const read = file => JSON.parse(fs.readFileSync(file, "utf8"));
const must = (ok, message) => { if (!ok) throw new Error(message); };
const gate = read(path.join(P, "measurement-gate.json"));
must(gate.status === "released-measurement-only" && gate.execution === false, "Measurement gate closed");
must(gate.session === "2026-10-08T20:26:35Z-81e5", "Session binding differs");
const acceptance = path.resolve(gate.rootAcceptance.path);
must(acceptance.startsWith(D+path.sep) && hash(fs.readFileSync(acceptance))===gate.rootAcceptance.sha256, "Root measurement acceptance changed");
const requestPath = path.join(P, "measurement-request.json");
must(hash(fs.readFileSync(requestPath)) === gate.requestSha256, "Measurement request changed");
const req = read(requestPath), bindingPath = path.join(D, "source-binding.json");
must(req.sourceBindingPath === bindingPath && req.sourceBindingSha256 === hash(fs.readFileSync(bindingPath)), "Draft binding absent");
const binding = read(bindingPath);
const names = [
 "packages/conformance/test-support/transactional/manifest-v1.json",
 "packages/conformance/test-support/transactional/coverage-v1.json",
 "packages/conformance/src/transactional-cumulative.test.ts",
 "apps/bdptest/src/transactional-matrix.test.ts"
];
must(JSON.stringify(Object.keys(binding.files).sort()) === JSON.stringify([...names].sort()), "Four-path binding required");
for (const n of names) must(hash(fs.readFileSync(path.join(W,n))) === binding.files[n].sha256, "Source changed: "+n);
function stats(bytes) {
 must(bytes.length <= 1048576, "Runtime byte bound exceeded");
 const value = JSON.parse(bytes.toString("utf8")), stack = [value];
 let nodes = 0;
 while(stack.length) { const v = stack.pop(); nodes++; if (v && typeof v === "object") stack.push(...Object.values(v)); }
 must(nodes <= 100000, "Runtime node bound exceeded");
 const normal = Buffer.from(JSON.stringify(value,null,2)+"\n"), compact = Buffer.from(JSON.stringify(value));
 return {sha256:hash(bytes),diskBytes:bytes.length,normalBytes:normal.length,normalSha256:hash(normal),compactBytes:compact.length,nodes,value};
}
const canonical = stats(fs.readFileSync(path.join(W,names[0])));
const coverage = stats(fs.readFileSync(path.join(W,names[1])));
const actions = canonical.value.scenarios.reduce((n,row)=>n+row.actions.length,0);
const assertions = canonical.value.scenarios.reduce((n,row)=>n+row.actions.reduce((a,item)=>a+item.assertions.length,0),0);
must(Array.isArray(req.specimens) && req.specimens.length === 41, "All41 changed specimen bindings required");
must(new Set(req.specimens.map(x=>x.id)).size===41, "Duplicate specimen ID");
const derivative = req.specimens.map(spec=>{
 const file = path.resolve(spec.path);
 must(file.startsWith(D+path.sep), "Specimen must be external under draft evidence");
 const bytes = fs.readFileSync(file);
 must(hash(bytes)===spec.sha256, "Specimen changed: "+spec.id);
 const measured=stats(bytes);delete measured.value;
 return {id:spec.id,path:file,...measured};
});
delete canonical.value;delete coverage.value;
const stops = {
 canonicalDisk:canonical.diskBytes<=818624,
 canonicalNormal:canonical.normalBytes<=830430,
 canonicalNodes:canonical.nodes<=21770,
 coverageDisk:coverage.diskBytes<=123139,
 coverageNormal:coverage.normalBytes<=123139,
 actions:actions===220,
 assertions:assertions<=2554,
 derivativeNormal:derivative.every(x=>x.normalBytes<=836548),
 derivativeNodes:derivative.every(x=>x.nodes<=21926)
};
const oldFences = {
 canonicalDisk:canonical.diskBytes<=742879,
 canonicalNormal:canonical.normalBytes<=754685,
 canonicalNodes:canonical.nodes<=19639,
 coverageDisk:coverage.diskBytes<=109115,
 coverageNormal:coverage.normalBytes<=109115,
 derivativeNormal:derivative.every(x=>x.normalBytes<=760803),
 derivativeNodes:derivative.every(x=>x.nodes<=19795)
};
must(req.output===path.join(P,"measurements.json"), "Output path changed");
const result={status:"MEASURED_ONLY_NO_EXECUTION_AUTHORIZATION",sourceBindingSha256:req.sourceBindingSha256,requestSha256:gate.requestSha256,canonical,coverage,actions,assertions,derivative,simultaneousPlanningStops:stops,simultaneousPlanningStopsSatisfied:Object.values(stops).every(Boolean),retainedFences:oldFences,requiresExplicitRootFenceDecision:true,testsExecuted:false,derivativeTransformationFidelity:"Not proved by this numeric tool; root/source review must bind all41 to actual app transformations and separately retain14 unchanged invocations.",parentPreservation:"Separate raw28-object and complete parent reconstruction proof required.",runtimeLimitsChanged:false};
fs.writeFileSync(req.output,JSON.stringify(result,null,2)+"\n",{flag:"wx"});
if(!result.simultaneousPlanningStopsSatisfied) process.exitCode=2;
