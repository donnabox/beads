# Administrative drain lifetime correction

The frozen successor incorrectly says existing execution deadlines bound every drain. The actual private service scheduler explicitly provides no delivery timeout. Root accepted this correction before runtime execution.

The trusted service/runtime replacement requires timeoutMs and optional signal, with timeoutMs a positive safe integer at most 2147483647. The service owns and clears its timer and abort listener. Deadline or caller cancellation abandons publication and reopens mutation admission under the old snapshot; already-admitted work continues normally. The operation checks cancellation and service lifetime immediately before synchronous owner publication. This adds no mutation deadline, wire control, receipt or rollback semantics.

The sequence witness includes non-delivered scheduler cancellation and timeout, then ordinary work completion and a successful replacement, within its same owner and zero HTTP budget. Concrete calls will be frozen with the authored ledger.
