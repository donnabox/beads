# Installed CONNECT test review dispositions

Native root and external review agree the prior destroy-after-end regression gap and two wording nits are closed at94c863b, integrated inbf462023. The external review ran no new probes; author mutant evidence and root combined execution remain separately attributed.

External LOW1 (socket errno hidden by timeout) is rejected as a false positive: Node events.once(emitter, otherEvent) rejects on error even when another error listener collects it. Root ran a Node24.16.0 counterexample with an additional collecting listener: once(readable) rejected the identical ECONNRESET object immediately. The test uses once for both connect/readable. Adding another error race would duplicate the existing behavior.

INFO2 is a precondition check, correctly useful: writableEnded verifies the test client never called end; it is not intended to prove that the server sent no FIN. The runtime property is independently tested by real child close and retained-store reopen. No claim of wire FIN inspection.

INFO3 is optional budget headroom; the test's specific shutdown bound already fires at two seconds and cleanup is bounded independently. No test change needed. INFO4 cross-platform behavior remains pending actual Linux CI; do not infer it from Darwin. Import organization and Markdown line wrapping are nonblocking style observations outside required format/lint gates. No runtime/code change is needed from this review.
