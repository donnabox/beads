package graphstore

import (
	"context"
	"database/sql"
	"fmt"
)

// This frame is owned by exactly one operation/transaction. It is never stored
// on Store, indexed globally by Tx, reused by another request or caller-made.
// The context key is private; callback lifetime and tx identity are checked.
type nativeOperationKey struct{}
type nativeOperationContext struct {
	store   *Store
	tx      *sql.Tx
	budget  *intrinsicMetadataBudget
	capture checkedResourceCapture
	purpose txPurpose
	active  bool
}

func nativeDomainPurpose(purpose txPurpose) bool {
	switch purpose {
	case txNativeIssueCreate, txNativeIssueBatchCreate, txNativeIssueGraphCreate, txNativeIssueChildCreate, txNativeMemoryCreate, txNativeMemoryMutation, txNativeIssueMutation, txNativeLinkMutation:
		return true
	default:
		return false
	}
}

// withNativeDomainTx is called only by the enumerated domain entry points.
// Ordinary11 delegates unchanged. It does NOT lift txOrdinary or grant imports,
// setup/defaults, arbitrary authored Types, persisted custom instances or packs.
func (s *Store) withNativeDomainTx(ctx context.Context, purpose txPurpose, validateTypes bool, fn func(context.Context, *sql.Tx) error) error {
	if !nativeDomainPurpose(purpose) {
		return ErrCapabilityUnavailable
	}
	if s.LayoutVersion() != TypesFreshSchemaVersion {
		return s.withTxPolicy(ctx, txPolicy{mayWrite: true, commit: true, validateTypes: validateTypes}, func(tx *sql.Tx) error { return fn(ctx, tx) })
	}
	state := &nativeOperationContext{store: s, budget: newIntrinsicMetadataBudget(), purpose: purpose}
	defer func() { state.active = false; state.tx = nil }()
	return s.withTxPolicy(ctx, txPolicy{purpose: purpose, allowance: state.budget, native: state, mayWrite: true, commit: true, validateTypes: true}, func(tx *sql.Tx) error {
		if state.tx != tx || !state.active {
			return ErrInvalidStore
		}
		capture, err := s.captureTypeRecordsInTx(ctx, tx, state.budget)
		if err != nil {
			return err
		}
		state.capture = capture
		operationContext := context.WithValue(ctx, nativeOperationKey{}, state)
		return fn(operationContext, tx)
	})
}

// actualCheckedOperationBinding is narrowly reusable proof inside the same
// operation/Tx. No Store-global flag, cross-request cache or caller assertion
// suppresses binding/schema/policy acquisition at the transaction boundary.
func actualCheckedOperationBinding(ctx context.Context, tx *sql.Tx, o Options, budget *intrinsicMetadataBudget) bool {
	if ctx == nil {
		return false
	}
	if state, ok := ctx.Value(nativeOperationKey{}).(*nativeOperationContext); ok && state != nil && state.active && state.tx == tx && state.store != nil && state.store.options.Binding == o.Binding && state.capture.installed.catalog != nil && (budget == nil || budget == state.budget) {
		return true
	}
	if read, ok := ctx.Value(resourceReadContextKey{}).(*resourceReadContext); ok && read != nil && read.tx == tx && read.store != nil && read.store.options.Binding == o.Binding && read.installed.catalog != nil && (budget == nil || budget == read.budget) {
		return true
	}
	return false
}

func (s *Store) nativeOperationInContext(ctx context.Context, tx *sql.Tx) (*nativeOperationContext, error) {
	state, ok := ctx.Value(nativeOperationKey{}).(*nativeOperationContext)
	if !ok || state == nil || state.store != s || state.tx != tx || !state.active || !nativeDomainPurpose(state.purpose) || state.budget == nil || state.capture.installed.catalog == nil {
		return nil, fmt.Errorf("%w: native operation lacks same-transaction checked Type context", ErrInvalidStore)
	}
	return state, nil
}
