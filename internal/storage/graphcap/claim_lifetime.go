package graphcap

import (
	"context"
	"sync"
	"time"

	graph "github.com/steveyegge/beads/graphops"
)

type ownerBinding struct{ workspace, installationKey, database, branch string }
type eraBinding struct {
	scopeURL, authorityID, grantedAt string
	epoch                            uint64
}

// generation has nonzero size: distinct live allocations have distinct identity.
type generation struct{ marker byte }
type era struct {
	binding eraBinding
	until   time.Time
}
type anchor struct {
	owner      *Owner
	generation *generation
	era        *era
	at         time.Time
}

// Owner is one provider's process-local admission lifetime. There is no exported
// constructor and its zero value refuses. Copies refuse before touching shared
// state; use the original pointer, not a copied Owner value.
type Owner struct {
	self               *Owner // immutable identity; checked before locking
	mu                 sync.Mutex
	binding            ownerBinding
	maxLease, perOpCap time.Duration
	now                func() time.Time
	generation         *generation
	current            *era
	closed             bool
	admissions         map[*admission]struct{}
	idle               chan struct{}
}

type admission struct {
	owner    *Owner
	era      *era
	ctx      context.Context
	cancel   func(error)
	until    time.Time
	released bool // protected by owner.mu
}

// LeaseClaim is an opaque token for exactly one admitted operation. Copies share
// release and revocation. Its identity accessors return data, never permission.
type LeaseClaim struct{ admission *admission }

func newOwner(binding ownerBinding, maxLease, perOpCap time.Duration) (*Owner, error) {
	if binding.workspace == "" || binding.installationKey == "" || binding.database == "" || binding.branch == "" ||
		maxLease <= 0 || perOpCap <= 0 || perOpCap > maxLease {
		return nil, graph.ErrNotAuthority
	}
	idle := make(chan struct{})
	close(idle)
	o := &Owner{binding: binding, maxLease: maxLease, perOpCap: perOpCap, now: time.Now,
		generation: &generation{}, admissions: make(map[*admission]struct{}), idle: idle}
	o.self = o
	return o, nil
}

func (o *Owner) original() bool { return o != nil && o.self == o }

// capture precedes future issuer work. Its private clock supplies provenance;
// there is no timestamp argument and no inspection of time.Time internals.
func (o *Owner) capture() anchor {
	if !o.original() {
		return anchor{}
	}
	o.mu.Lock()
	defer o.mu.Unlock()
	if o.closed || o.now == nil {
		return anchor{}
	}
	return anchor{owner: o, generation: o.generation, era: o.current, at: o.now()}
}

// interval is called under mu. Remaining must already be conservative under the
// future issuer's qualified DB observation. Anchoring it before that work cannot
// extend permission by the time spent completing the proof.
func (o *Owner) interval(a anchor, remaining time.Duration) (time.Time, error) {
	if o.closed || o.now == nil || a.owner != o || a.generation != o.generation || a.era != o.current ||
		a.at.IsZero() || remaining <= 0 || remaining > o.maxLease {
		return time.Time{}, graph.ErrNotAuthority
	}
	until := a.at.Add(remaining)
	if !o.now().Before(until) {
		return time.Time{}, graph.ErrNotAuthority
	}
	return until, nil
}

// activate has only synthetic test callers. An eventual same-package issuer must
// establish real authority before calling it. It cannot replace an active era.
func (o *Owner) activate(a anchor, binding eraBinding, remaining time.Duration) error {
	if !o.original() {
		return graph.ErrNotAuthority
	}
	o.mu.Lock()
	defer o.mu.Unlock()
	if o.current != nil || binding.scopeURL == "" || binding.authorityID == "" || binding.grantedAt == "" || binding.epoch == 0 {
		return graph.ErrNotAuthority
	}
	until, err := o.interval(a, remaining)
	if err != nil {
		return err
	}
	o.current = &era{binding: binding, until: until}
	return nil
}

// extend is routine same-era renewal only. It does not alter or re-admit any
// existing operation. An expired or revoked era cannot be resurrected here.
func (o *Owner) extend(a anchor, remaining time.Duration) error {
	if !o.original() {
		return graph.ErrNotAuthority
	}
	o.mu.Lock()
	defer o.mu.Unlock()
	until, err := o.interval(a, remaining)
	if err != nil {
		return err
	}
	if o.current == nil || !o.now().Before(o.current.until) || !until.After(o.current.until) {
		return graph.ErrNotAuthority
	}
	o.current.until = until
	return nil
}

// Begin atomically admits one operation. Budget must be strictly below both the
// configured cap and remaining lease interval; it is never silently capped.
// The parent must have a deadline and may cancel/shorten the operation normally.
// Use exactly the returned context with Check, and release after checked cleanup.
// On error all other outputs are nil/zero; check the error before deferring release.
func (o *Owner) Begin(ctx context.Context, budget time.Duration) (LeaseClaim, context.Context, func(), error) {
	if err := finiteContext(ctx); err != nil {
		return LeaseClaim{}, nil, nil, err
	}
	if !o.original() {
		return LeaseClaim{}, nil, nil, graph.ErrNotAuthority
	}
	o.mu.Lock()
	defer o.mu.Unlock()
	// Waiting for the owner lock must not admit an already-canceled parent.
	if err := finiteContext(ctx); err != nil {
		return LeaseClaim{}, nil, nil, err
	}
	if o.closed || o.current == nil || budget <= 0 || budget >= o.perOpCap {
		return LeaseClaim{}, nil, nil, graph.ErrNotAuthority
	}
	now := o.now()
	if budget >= o.current.until.Sub(now) {
		return LeaseClaim{}, nil, nil, graph.ErrNotAuthority
	}
	until := now.Add(budget)
	bounded, stop := context.WithDeadline(ctx, until)
	operation, cancel := context.WithCancelCause(bounded)
	a := &admission{owner: o, era: o.current, ctx: operation, until: until,
		cancel: func(err error) { cancel(err); stop() }}
	if len(o.admissions) == 0 {
		o.idle = make(chan struct{})
	}
	o.admissions[a] = struct{}{}
	return LeaseClaim{admission: a}, operation, func() { o.release(a) }, nil
}

func finiteContext(ctx context.Context) error {
	if ctx == nil {
		return graph.ErrNotAuthority
	}
	if err := ctx.Err(); err != nil {
		return err
	}
	if _, ok := ctx.Deadline(); !ok {
		return graph.ErrNotAuthority
	}
	return nil
}

// Check validates this admission and its exact returned context. Ordinary
// descendants also refuse; detached caller deadlines cannot extend the token.
func (c LeaseClaim) Check(ctx context.Context) error {
	a := c.admission
	if a == nil || ctx != a.ctx {
		return graph.ErrNotAuthority
	}
	o := a.owner
	o.mu.Lock()
	defer o.mu.Unlock()
	if a.released || o.closed || o.current != a.era {
		return graph.ErrNotAuthority
	}
	if err := a.ctx.Err(); err != nil {
		return err
	}
	if !o.now().Before(a.until) {
		return context.DeadlineExceeded
	}
	return nil
}

func (o *Owner) release(a *admission) {
	if !o.original() {
		return
	}
	o.mu.Lock()
	defer o.mu.Unlock()
	if a.released {
		return
	}
	a.released = true
	a.cancel(context.Canceled)
	delete(o.admissions, a)
	if len(o.admissions) == 0 {
		close(o.idle)
	}
}

// Revoke explicitly cancels all admitted work and invalidates every old token.
// It does not assert database rollback or wait for borrowers to release.
func (o *Owner) Revoke() {
	if !o.original() {
		return
	}
	o.mu.Lock()
	defer o.mu.Unlock()
	o.revoke()
}

func (o *Owner) revoke() {
	o.current = nil
	o.generation = &generation{}
	for a := range o.admissions {
		a.cancel(graph.ErrNotAuthority)
	}
}

// Close terminally revokes, then waits for explicit releases under a finite
// caller deadline. Cancellation alone is not successful borrower cleanup.
// Even an invalid/canceled Close context leaves the owner terminal.
// A canceled context reports its error even if already idle; retrying Close with
// a live finite context may confirm drainage. Nil, zero and copied owners refuse.
func (o *Owner) Close(ctx context.Context) error {
	if !o.original() {
		return graph.ErrNotAuthority
	}
	o.mu.Lock()
	o.closed = true
	o.revoke()
	idle := o.idle
	o.mu.Unlock()
	if err := finiteContext(ctx); err != nil {
		return err
	}
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-idle:
		return ctx.Err()
	}
}

// InstallationKey returns the admission's copied installation identity.
func (c LeaseClaim) InstallationKey() string {
	if c.admission == nil {
		return ""
	}
	return c.admission.owner.binding.installationKey
}

// Database returns the admission's bound workspace database.
func (c LeaseClaim) Database() string {
	if c.admission == nil {
		return ""
	}
	return c.admission.owner.binding.database
}

// Branch returns the admission's configured default branch.
func (c LeaseClaim) Branch() string {
	if c.admission == nil {
		return ""
	}
	return c.admission.owner.binding.branch
}

// Epoch returns the admission's era epoch; this value alone grants nothing.
func (c LeaseClaim) Epoch() uint64 {
	if c.admission == nil {
		return 0
	}
	return c.admission.era.binding.epoch
}
