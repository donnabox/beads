package graphops

import (
	"context"
	"errors"
	"reflect"
	"sync/atomic"
	"time"

	graph "github.com/steveyegge/beads/graphops"
	"github.com/steveyegge/beads/internal/storage/authority"
)

type readAttemptBinding struct{ installationKey, database, branch string }

// loadReadAttemptFacts is passive intake, not an authority verdict. Construction
// binds Manager and installationKey; an absent Snapshot cannot verify that bond.
func loadReadAttemptFacts(ctx context.Context, manager *authority.Manager, binding readAttemptBinding) (expectedReadFacts, error) {
	if err := observationContext(ctx); err != nil {
		return expectedReadFacts{}, err
	}
	if manager == nil {
		return expectedReadFacts{}, graph.ErrNotAuthority
	}
	want := expectedReadFacts{installationKey: binding.installationKey, database: binding.database, branch: binding.branch}
	if err := checkReadOperands(want); err != nil {
		return expectedReadFacts{}, err
	}
	snapshot, err := manager.Load(ctx)
	if err != nil {
		return expectedReadFacts{}, err
	}
	want.pending = snapshot.Pending()
	want.witnessInstallationKey = snapshot.InstallationKey()
	if witness, ok := snapshot.Witness(); ok {
		want.witness = &witness
	}
	return want, nil
}

// readAttemptResource groups one snapshot and its matching connection's cleanup.
// Its cooperative methods are not a physical-terminal receipt or an issuer.
// There is no production implementation or acquisition callback in this slice.
type readAttemptResource interface {
	queryer
	rollbackRead(context.Context) error
	closeReadConnection(context.Context) error
}

type readAttemptOwner struct {
	self     *readAttemptOwner
	resource readAttemptResource
	consumed atomic.Bool
}

// ownReadAttempt transfers cleanup ownership without validation or I/O. Failed
// acquisition may transfer its remaining resource with runReadAttempt's entryErr.
func ownReadAttempt(resource readAttemptResource) *readAttemptOwner {
	owner := &readAttemptOwner{resource: resource}
	owner.self = owner
	return owner
}

func (owner *readAttemptOwner) take() (readAttemptResource, error) {
	if owner == nil || owner.self != owner || !owner.consumed.CompareAndSwap(false, true) {
		return nil, graph.ErrNotAuthority
	}
	if owner.resource == nil {
		return nil, graph.ErrNotAuthority
	}
	value := reflect.ValueOf(owner.resource)
	switch value.Kind() {
	case reflect.Chan, reflect.Func, reflect.Interface, reflect.Map, reflect.Pointer, reflect.Slice:
		if value.IsNil() {
			return nil, graph.ErrNotAuthority
		}
	}
	return owner.resource, nil
}

type readAttemptStage uint8

const (
	readAttemptEntry readAttemptStage = iota
	readAttemptObserve
	readAttemptCompare
	readAttemptBody
)

type readAttemptCompletion struct {
	stage        readAttemptStage
	diagnostic   error // Private retained cause; completion has no Error or Unwrap.
	cleanupOK    bool
	stateChanged bool
	workLive     bool // Observed before a future entrypoint releases its claim.
}

func (c readAttemptCompletion) retryEligible() bool {
	return c.stage == readAttemptCompare && c.stateChanged && c.cleanupOK && c.workLive
}

const readAttemptCleanupTimeout = 5 * time.Second

// cleanupReadCall catches a panic only so the other cleanup obligation can run.
func cleanupReadCall(ctx context.Context, call func(context.Context) error) (err error, panicValue any, panicked bool) {
	panicked = true
	defer func() {
		if panicked {
			panicValue = recover()
		}
	}()
	err = call(ctx)
	panicked = false
	return
}

func closeReadAttempt(ctx context.Context, resource readAttemptResource) (err error, panicValue any, panicked bool) {
	parent := context.Background()
	if ctx != nil {
		parent = context.WithoutCancel(ctx)
	}
	cleanupCtx, cancel := context.WithTimeout(parent, readAttemptCleanupTimeout)
	defer cancel()
	rollbackErr, rollbackPanic, rollbackPanicked := cleanupReadCall(cleanupCtx, resource.rollbackRead)
	connectionErr, connectionPanic, connectionPanicked := cleanupReadCall(cleanupCtx, resource.closeReadConnection)
	err = errors.Join(rollbackErr, connectionErr, cleanupCtx.Err())
	if rollbackPanicked {
		return err, rollbackPanic, true
	}
	return err, connectionPanic, connectionPanicked
}

// runReadAttempt owns observe -> compare -> body -> checked cleanup. This is a
// single claimless mechanism; it is neither a protected Reader nor a provider.
// The caller must not retain/use the resource after transfer. Work keeps exactly
// ctx; cleanup has a separate bounded context. An adapter must independently prove
// terminal cleanup: neither a deadline nor sql.ErrTxDone establishes that fact.
func runReadAttempt[T any](ctx context.Context, owner *readAttemptOwner, want expectedReadFacts, entryErr error, body func(context.Context, queryer) (T, error)) (result T, completion readAttemptCompletion, err error) {
	resource, err := owner.take()
	if err != nil {
		completion.diagnostic = errors.Join(err, entryErr)
		return result, completion, err
	}
	normalReturn := false
	defer func() {
		var workPanic any
		if !normalReturn {
			workPanic = recover()
		}
		completion.diagnostic = err
		cleanupErr, cleanupPanic, cleanupPanicked := closeReadAttempt(ctx, resource)
		completion.cleanupOK = cleanupErr == nil && !cleanupPanicked
		var contextErr error
		if ctx != nil {
			contextErr = ctx.Err()
		}
		completion.workLive = ctx != nil && contextErr == nil
		if cleanupErr != nil || contextErr != nil {
			// Do not expose discarded result-bearing stage causes through Is or As.
			err = errors.Join(cleanupErr, contextErr)
		}
		if err != nil || !normalReturn || cleanupPanicked {
			var zero T
			result = zero
		}
		// Under qualified default Go1.26, real panic(nil) is non-nil on recover.
		// A nil recovery here is Goexit; allow its unwinding to continue.
		if !normalReturn && workPanic != nil {
			panic(workPanic)
		}
		if cleanupPanicked {
			panic(cleanupPanic)
		}
	}()
	result, completion.stage, completion.stateChanged, err = executeReadAttempt(ctx, resource, want, entryErr, body)
	normalReturn = true
	return
}

func executeReadAttempt[T any](ctx context.Context, resource readAttemptResource, want expectedReadFacts, entryErr error, body func(context.Context, queryer) (T, error)) (T, readAttemptStage, bool, error) {
	var zero T
	if entryErr != nil {
		return zero, readAttemptEntry, false, entryErr
	}
	if err := observationContext(ctx); err != nil {
		return zero, readAttemptEntry, false, err
	}
	if body == nil {
		return zero, readAttemptEntry, false, graph.ErrNotAuthority
	}
	// Do not pre-judge pending, absent, unverified, binding or Scope states here.
	// Observers own structural query safety; checkReadFacts owns verdict ordering.
	var recordedSeq *uint64
	if want.witness != nil {
		witness := *want.witness
		want.witness = &witness
		seq := witness.Fields().LedgerSeq
		recordedSeq = &seq
	}
	observed, err := observePreconditionsInTx(ctx, resource, recordedSeq)
	if err != nil {
		return zero, readAttemptObserve, false, err
	}
	if err := checkReadFacts(ctx, want, observed); err != nil {
		return zero, readAttemptCompare, err == graph.ErrStateChanged, err
	}
	result, err := body(ctx, resource)
	return result, readAttemptBody, false, err
}
