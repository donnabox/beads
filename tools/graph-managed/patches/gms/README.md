# Pinned GMS terminal-session source bundle

This offline bundle reproduces four reviewed GMS source changes: terminal connection cleanup uses an existing session without creating one, while ordinary context creation and reset behavior remain intact. It includes eight component controls and four protocol-fixture controls. Materialization produces source and a receipt; it does not build or run that source.

The upstream module is `github.com/dolthub/go-mysql-server` at `v0.20.1-0.20260805191915-e5eafe0da809` (origin `e5eafe0da809714c3b88be7a54b9e2163f2550bf`). Its 44,505,336-byte ZIP has SHA256 `9d494d82a0a9d2a6fc5239a5d1dd95c5dfda372377fa68cdcfab26cfabd9a9be` and Go checksum `h1:c+S3O08LgGS+czU/CuN+/pOKMri5F6kxajoWrTyrE1s=`. The complete 1,688 original and 1,690 final path/size/SHA256 inventories are in `manifest.json`.

The final source identity is local commit `02101eba443c807888abfd649551f8754400d7f0`, tree `8ee0dc18641ccfe8599b74a2d5237ff6708e367a`. That commit belongs to a synthetic archive checkout; it is not an upstream fetchable module revision. Exactly these GMS paths change:

- `server/context.go`
- `server/handler.go`
- `server/close_existing_session_test.go` (new)
- `server/terminal_connection_integration_test.go` (new)

The canonical plain unified patch is 36,756 bytes, SHA256 `93a84536debc560ff3a3510fa78f5f4695147dcfc31bf5f808ab48489bbdd720`. The compact manifest is 462,003 bytes, SHA256 `a3d7e07a020431fa4a15d03a10bb1bcbd07263d6c2a0656929c5b6c58cd884fc`. The historical Git-format provenance patch is a separate artifact and is not accepted by this parser. Matching hashes establish consistency with the reviewed checkout, not a signature or independent remote attestation.

## Offline use

Supply the existing ZIP and exact Dolt consumer module file explicitly. The tool never discovers a module cache, downloads an input, or chooses a version. The consumer is `github.com/dolthub/dolt/go` at `v0.40.5-0.20260806213044-796d07497741`; its complete `go.mod` must match the manifest byte-for-byte. The exact GMS requirement was verified in that pinned source file; the tool does not re-parse it. Even an unrelated change to that file refuses admission.

```sh
python3 -I -B tools/graph-managed/materialize-gms.py \
  --archive /absolute/path/to/pinned-gms.zip \
  --consumer-go-mod /absolute/path/to/pinned-dolt/go.mod \
  --output-parent /absolute/canonical/private/parent \
  --name gms-terminal-session
```

The existing output parent must be owned by the effective user with no group/world permissions. Its path must be absolute and contain no symlink, empty, dot or dot-dot component. The fresh output name must be one ASCII component of at most 64 characters. POSIX descriptor-relative no-follow APIs are required; unsupported hosts refuse.

Success creates `gms-terminal-session/source/` containing precisely the 1,690 verified ordinary files, plus receipt files outside that source tree. A fully written, closed and verified `receipt.pending` is published as `receipt.json` using an atomic exclusive hard link in the same owned container. That link is the success point; the pending name remains nonauthoritative. These two receipt names share one inode; source files are never links. Directories are mode0700; source files mode0644. The receipt identifies the materializer, manifest, patch, archive, consumer and complete output map, and reports the exact local source commit. Each source file is created exclusively once. There is no extract-and-overwrite step.

An existing destination always refuses, even if empty. Failure before container creation has bounded stderr only. A later failure retains its owned partial container and reports its absolute path, with `receipt.failure.json` when safely writable within the remaining deadline. Expiry can leave that best-effort diagnostic missing or incomplete; the full owned location and primary failure are reported separately, without starting a new deadline. Only a valid `receipt.json` signifies success; a pending or failure receipt never does. A missing or invalid success receipt means the tree is unusable. After publication, descriptor cleanup errors are reported separately as `postpublication_cleanup`; output-reporting failure exits2 and states that source was committed. Neither reverses the published receipt. This is a local source-publication boundary, not a durability or whole-host crash guarantee. Retry with a fresh name. An operator can separately inspect and remove a verified owned partial container; this tool has no recursive deletion, overwrite or cleanup command. A malicious process with the same user's write authority is outside this local filesystem guarantee.

Fixed administrative limits are manifest2MiB, patch1MiB, archive64MiB, 2,048 archive entries, 8MiB per file, 128MiB aggregate input/output payload, 256UTF-8 bytes and 16 components per relative path. Exact pinned identities impose tighter bounds. The 120-second cooperative deadline checks work between I/O operations; use an outer process deadline as well. These limits do not promise storage throughput or interrupt a blocked kernel I/O call. The exact archive hash is checked before ZIP metadata parsing; this is not a general arbitrary-ZIP memory guarantee.

## Verification and provenance boundaries

The default synthetic controls use only standard-library fixtures in test-owned temporary directories. These Python controls are operator-invoked and are not currently part of repository CI. They do not search for a real archive:

```sh
python3 -I -B -m unittest discover -s tools/graph-managed -p test_materialize_gms.py
```

The separate real-archive gate requires explicit inputs and a private test parent:

```sh
python3 -I -B tools/graph-managed/test_materialize_gms.py \
  --real-archive /absolute/path/to/pinned-gms.zip \
  --consumer-go-mod /absolute/path/to/pinned-dolt/go.mod \
  --output-parent /absolute/canonical/private/test-parent
```

Use an outer 120-second bound for synthetic controls and 180 seconds for the real gate. Missing inputs fail rather than skip or download. The real gate verifies the complete source map, requires an existing-destination refusal without receipt changes, and also requires no post-publication cleanup diagnostics. Both synthetic and real controls clean only their own temporary children on every exit, including failure; they do not leave a persistent installed source tree. These commands are proposed validation gates, not a claim that this bundle's tools have already been executed.

Previous GMS component, terminal, server, race and vet observations belong to their exact recorded source/binary/toolchain inputs. The final local source adds the two explicit production modification comments and removes an unsupported copyright-holder line from the newly authored component test. Those three changes are comments only; 1,687 other files, including the terminal test, are byte-identical to the executed predecessor. The manifest records the before/after linkage and preserved receipt history. Old receipts are not relabeled as runs of the final comment successor or this materializer.

This bundle does not patch an installed Dolt executable. Beads' ordinary launcher resolves an external `dolt`; a Go replacement in Beads cannot change that binary. A later Dolt pairing must reconcile its actual module/initializer graph, compiler inputs and binary identity, and separately validate the selected client. The current standalone GMS/Dolt client requirement is mysql1.9.3, while the independently patched Beads client candidate is1.10.0. No version change or pairing is selected here. Managed activation, authentication, credentials/profile, listener, host, drain and route-restore qualification remain separate gates.

## Attribution

Original GMS production headers and the upstream Apache2.0 license are preserved. The two modified production files carry `Modified 2026-09-16: avoid creating sessions during terminal connection cleanup.` The two new locally authored tests retain Apache boilerplate without an invented copyright holder. The original archive has one root `LICENSE` and no `NOTICE` member. Every other source/license byte remains covered by the complete file inventory.

The new Python tools and documentation follow the Beads repository contribution policy. Embedded GMS patch context and the following license retain their upstream Apache attribution; they are not silently relicensed by the surrounding repository. This source inventory is not a legal authorship ruling or a license review of an eventual Dolt/native dependency binary.

The following appendix is the complete pinned upstream `LICENSE`, SHA256 `c71d239df91726fc519c6eb72d318ec65820627232b2f796219e87dcf35d0ab4`.

## Upstream LICENSE (verbatim)

```text
                                 Apache License
                           Version 2.0, January 2004
                        http://www.apache.org/licenses/

   TERMS AND CONDITIONS FOR USE, REPRODUCTION, AND DISTRIBUTION

   1. Definitions.

      "License" shall mean the terms and conditions for use, reproduction,
      and distribution as defined by Sections 1 through 9 of this document.

      "Licensor" shall mean the copyright owner or entity authorized by
      the copyright owner that is granting the License.

      "Legal Entity" shall mean the union of the acting entity and all
      other entities that control, are controlled by, or are under common
      control with that entity. For the purposes of this definition,
      "control" means (i) the power, direct or indirect, to cause the
      direction or management of such entity, whether by contract or
      otherwise, or (ii) ownership of fifty percent (50%) or more of the
      outstanding shares, or (iii) beneficial ownership of such entity.

      "You" (or "Your") shall mean an individual or Legal Entity
      exercising permissions granted by this License.

      "Source" form shall mean the preferred form for making modifications,
      including but not limited to software source code, documentation
      source, and configuration files.

      "Object" form shall mean any form resulting from mechanical
      transformation or translation of a Source form, including but
      not limited to compiled object code, generated documentation,
      and conversions to other media types.

      "Work" shall mean the work of authorship, whether in Source or
      Object form, made available under the License, as indicated by a
      copyright notice that is included in or attached to the work
      (an example is provided in the Appendix below).

      "Derivative Works" shall mean any work, whether in Source or Object
      form, that is based on (or derived from) the Work and for which the
      editorial revisions, annotations, elaborations, or other modifications
      represent, as a whole, an original work of authorship. For the purposes
      of this License, Derivative Works shall not include works that remain
      separable from, or merely link (or bind by name) to the interfaces of,
      the Work and Derivative Works thereof.

      "Contribution" shall mean any work of authorship, including
      the original version of the Work and any modifications or additions
      to that Work or Derivative Works thereof, that is intentionally
      submitted to Licensor for inclusion in the Work by the copyright owner
      or by an individual or Legal Entity authorized to submit on behalf of
      the copyright owner. For the purposes of this definition, "submitted"
      means any form of electronic, verbal, or written communication sent
      to the Licensor or its representatives, including but not limited to
      communication on electronic mailing lists, source code control systems,
      and issue tracking systems that are managed by, or on behalf of, the
      Licensor for the purpose of discussing and improving the Work, but
      excluding communication that is conspicuously marked or otherwise
      designated in writing by the copyright owner as "Not a Contribution."

      "Contributor" shall mean Licensor and any individual or Legal Entity
      on behalf of whom a Contribution has been received by Licensor and
      subsequently incorporated within the Work.

   2. Grant of Copyright License. Subject to the terms and conditions of
      this License, each Contributor hereby grants to You a perpetual,
      worldwide, non-exclusive, no-charge, royalty-free, irrevocable
      copyright license to reproduce, prepare Derivative Works of,
      publicly display, publicly perform, sublicense, and distribute the
      Work and such Derivative Works in Source or Object form.

   3. Grant of Patent License. Subject to the terms and conditions of
      this License, each Contributor hereby grants to You a perpetual,
      worldwide, non-exclusive, no-charge, royalty-free, irrevocable
      (except as stated in this section) patent license to make, have made,
      use, offer to sell, sell, import, and otherwise transfer the Work,
      where such license applies only to those patent claims licensable
      by such Contributor that are necessarily infringed by their
      Contribution(s) alone or by combination of their Contribution(s)
      with the Work to which such Contribution(s) was submitted. If You
      institute patent litigation against any entity (including a
      cross-claim or counterclaim in a lawsuit) alleging that the Work
      or a Contribution incorporated within the Work constitutes direct
      or contributory patent infringement, then any patent licenses
      granted to You under this License for that Work shall terminate
      as of the date such litigation is filed.

   4. Redistribution. You may reproduce and distribute copies of the
      Work or Derivative Works thereof in any medium, with or without
      modifications, and in Source or Object form, provided that You
      meet the following conditions:

      (a) You must give any other recipients of the Work or
          Derivative Works a copy of this License; and

      (b) You must cause any modified files to carry prominent notices
          stating that You changed the files; and

      (c) You must retain, in the Source form of any Derivative Works
          that You distribute, all copyright, patent, trademark, and
          attribution notices from the Source form of the Work,
          excluding those notices that do not pertain to any part of
          the Derivative Works; and

      (d) If the Work includes a "NOTICE" text file as part of its
          distribution, then any Derivative Works that You distribute must
          include a readable copy of the attribution notices contained
          within such NOTICE file, excluding those notices that do not
          pertain to any part of the Derivative Works, in at least one
          of the following places: within a NOTICE text file distributed
          as part of the Derivative Works; within the Source form or
          documentation, if provided along with the Derivative Works; or,
          within a display generated by the Derivative Works, if and
          wherever such third-party notices normally appear. The contents
          of the NOTICE file are for informational purposes only and
          do not modify the License. You may add Your own attribution
          notices within Derivative Works that You distribute, alongside
          or as an addendum to the NOTICE text from the Work, provided
          that such additional attribution notices cannot be construed
          as modifying the License.

      You may add Your own copyright statement to Your modifications and
      may provide additional or different license terms and conditions
      for use, reproduction, or distribution of Your modifications, or
      for any such Derivative Works as a whole, provided Your use,
      reproduction, and distribution of the Work otherwise complies with
      the conditions stated in this License.

   5. Submission of Contributions. Unless You explicitly state otherwise,
      any Contribution intentionally submitted for inclusion in the Work
      by You to the Licensor shall be under the terms and conditions of
      this License, without any additional terms or conditions.
      Notwithstanding the above, nothing herein shall supersede or modify
      the terms of any separate license agreement you may have executed
      with Licensor regarding such Contributions.

   6. Trademarks. This License does not grant permission to use the trade
      names, trademarks, service marks, or product names of the Licensor,
      except as required for reasonable and customary use in describing the
      origin of the Work and reproducing the content of the NOTICE file.

   7. Disclaimer of Warranty. Unless required by applicable law or
      agreed to in writing, Licensor provides the Work (and each
      Contributor provides its Contributions) on an "AS IS" BASIS,
      WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or
      implied, including, without limitation, any warranties or conditions
      of TITLE, NON-INFRINGEMENT, MERCHANTABILITY, or FITNESS FOR A
      PARTICULAR PURPOSE. You are solely responsible for determining the
      appropriateness of using or redistributing the Work and assume any
      risks associated with Your exercise of permissions under this License.

   8. Limitation of Liability. In no event and under no legal theory,
      whether in tort (including negligence), contract, or otherwise,
      unless required by applicable law (such as deliberate and grossly
      negligent acts) or agreed to in writing, shall any Contributor be
      liable to You for damages, including any direct, indirect, special,
      incidental, or consequential damages of any character arising as a
      result of this License or out of the use or inability to use the
      Work (including but not limited to damages for loss of goodwill,
      work stoppage, computer failure or malfunction, or any and all
      other commercial damages or losses), even if such Contributor
      has been advised of the possibility of such damages.

   9. Accepting Warranty or Additional Liability. While redistributing
      the Work or Derivative Works thereof, You may choose to offer,
      and charge a fee for, acceptance of support, warranty, indemnity,
      or other liability obligations and/or rights consistent with this
      License. However, in accepting such obligations, You may act only
      on Your own behalf and on Your sole responsibility, not on behalf
      of any other Contributor, and only if You agree to indemnify,
      defend, and hold each Contributor harmless for any liability
      incurred by, or claims asserted against, such Contributor by reason
      of your accepting any such warranty or additional liability.

   END OF TERMS AND CONDITIONS

   APPENDIX: How to apply the Apache License to your work.

      To apply the Apache License to your work, attach the following
      boilerplate notice, with the fields enclosed by brackets "[]"
      replaced with your own identifying information. (Don't include
      the brackets!)  The text should be enclosed in the appropriate
      comment syntax for the file format. We also recommend that a
      file or class name and description of purpose be included on the
      same "printed page" as the copyright notice for easier
      identification within third-party archives.

   Copyright [yyyy] [name of copyright owner]

   Licensed under the Apache License, Version 2.0 (the "License");
   you may not use this file except in compliance with the License.
   You may obtain a copy of the License at

       http://www.apache.org/licenses/LICENSE-2.0

   Unless required by applicable law or agreed to in writing, software
   distributed under the License is distributed on an "AS IS" BASIS,
   WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
   See the License for the specific language governing permissions and
   limitations under the License.

```
