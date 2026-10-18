// One tooling invocation; imports only the derivative's own payload helper, no service.
import { readFileSync, writeFileSync, realpathSync } from "node:fs";
import { createHash } from "node:crypto";
import { createRequire } from "node:module";
import path from "node:path";
import { fileURLToPath, pathToFileURL } from "node:url";
const root = path.dirname(fileURLToPath(import.meta.url));
const binding = JSON.parse(readFileSync(path.join(root, "run/derivative-binding.json"), "utf8"));
const digest = (bytes) => createHash("sha256").update(bytes).digest("hex");
for (const [relative, hash] of Object.entries(binding.after.files))
  if (digest(readFileSync(path.join(binding.root, relative))) !== hash) throw Error("derivative bytes differ");
for (const item of binding.productionGraph) {
  const require = createRequire(item.manifest);
  for (const [name, expectedManifest] of Object.entries(item.resolvedDependencies)) {
    let directory = path.dirname(require.resolve(name));
    while (true) {
      try { const p = path.join(directory, "package.json");
        if (JSON.parse(readFileSync(p, "utf8")).name === name) {
          if (realpathSync(p) !== realpathSync(expectedManifest)) throw Error("actual Node dependency edge differs");
          break;
        }
      } catch (error) { if (error.code !== "ENOENT") throw error; }
      const parent = path.dirname(directory); if (parent === directory) throw Error("dependency root missing"); directory = parent;
    }
  }
}
const helper = path.join(path.dirname(binding.entry), "installed-payload.js");
const { installedPayloadIdentity } = await import(pathToFileURL(helper).href);
const payloadIdentity = await installedPayloadIdentity();
if (!/^[a-f0-9]{64}$/.test(payloadIdentity)) throw Error("payload identity shape");
const result = { payloadIdentity, helper: realpathSync(helper), helperSha256: digest(readFileSync(helper)),
  derivativeBindingSha256: digest(readFileSync(path.join(root, "run/derivative-binding.json"))),
  limitation: "Uses derivative's own helper plus separate byte/graph inventory; not an independently implemented payload algorithm or source attestation." };
writeFileSync(path.join(root, "run/derivative-payload.json"), JSON.stringify(result, null, 2) + "\n", { flag: "wx" });
console.log(JSON.stringify(result));
