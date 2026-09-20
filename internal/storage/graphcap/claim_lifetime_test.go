package graphcap

import (
	"context"
	"errors"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"reflect"
	"sort"
	"strconv"
	"strings"
	"sync"
	"testing"
	"time"

	graph "github.com/steveyegge/beads/graphops"
)

// These are synthetic lifetime tests, not a SQL issuer or authority proof.
type testClock struct {
	mu sync.Mutex
	t  time.Time
}

func (c *testClock) now() time.Time      { c.mu.Lock(); defer c.mu.Unlock(); return c.t }
func (c *testClock) add(d time.Duration) { c.mu.Lock(); defer c.mu.Unlock(); c.t = c.t.Add(d) }

func syntheticOwner(t *testing.T) (*Owner, *testClock) {
	t.Helper()
	o, err := newOwner(ownerBinding{"/synthetic/.beads", "installation", "workspace", "main"}, 30*time.Second, 10*time.Second)
	if err != nil {
		t.Fatal(err)
	}
	c := &testClock{t: time.Now()}
	o.now = c.now
	t.Cleanup(func() {
		ctx, cancel := context.WithTimeout(context.Background(), time.Second)
		defer cancel()
		if err := o.Close(ctx); err != nil {
			t.Errorf("synthetic owner cleanup: %v", err)
		}
	})
	return o, c
}

func syntheticEra() eraBinding {
	return eraBinding{"https://graph.example/", "authority", "2026-09-20T00:00:00Z", 2}
}

func activateSynthetic(t *testing.T, o *Owner) {
	t.Helper()
	if err := o.activate(o.capture(), syntheticEra(), 30*time.Second); err != nil {
		t.Fatal(err)
	}
}

func parentContext(t *testing.T) context.Context {
	t.Helper()
	ctx, cancel := context.WithTimeout(t.Context(), time.Minute)
	t.Cleanup(cancel)
	return ctx
}

func admitted(t *testing.T, o *Owner, budget time.Duration) (LeaseClaim, context.Context, func()) {
	t.Helper()
	c, ctx, release, err := o.Begin(parentContext(t), budget)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(release)
	return c, ctx, release
}

func requireRefused(t *testing.T, err error) {
	t.Helper()
	if !errors.Is(err, graph.ErrNotAuthority) {
		t.Fatalf("want ErrNotAuthority, got %v", err)
	}
}

func TestZeroAndInactiveRefuse(t *testing.T) {
	var claim LeaseClaim
	requireRefused(t, claim.Check(parentContext(t)))
	if claim.InstallationKey() != "" || claim.Database() != "" || claim.Branch() != "" || claim.Epoch() != 0 {
		t.Fatal("zero claim exposed identity")
	}
	o, _ := syntheticOwner(t)
	for _, candidate := range []*Owner{nil, {}, o} {
		got, ctx, release, err := candidate.Begin(parentContext(t), time.Second)
		requireRefused(t, err)
		if got.admission != nil || ctx != nil || release != nil {
			t.Fatal("partial admission on refusal")
		}
	}
}

func TestBindingOperands(t *testing.T) {
	for _, name := range []string{"workspace", "installation", "database", "branch", "lease", "cap", "cap exceeds lease"} {
		t.Run(name, func(t *testing.T) {
			b := ownerBinding{"workspace", "installation", "database", "main"}
			lease, cap := 30*time.Second, 10*time.Second
			switch name {
			case "workspace":
				b.workspace = ""
			case "installation":
				b.installationKey = ""
			case "database":
				b.database = ""
			case "branch":
				b.branch = ""
			case "lease":
				lease = 0
			case "cap":
				cap = 0
			case "cap exceeds lease":
				cap = 31 * time.Second
			}
			o, err := newOwner(b, lease, cap)
			requireRefused(t, err)
			if o != nil {
				t.Fatal("invalid owner returned")
			}
		})
	}
	for _, name := range []string{"scope", "authority", "grant", "epoch"} {
		t.Run(name, func(t *testing.T) {
			o, _ := syntheticOwner(t)
			b := syntheticEra()
			switch name {
			case "scope":
				b.scopeURL = ""
			case "authority":
				b.authorityID = ""
			case "grant":
				b.grantedAt = ""
			case "epoch":
				b.epoch = 0
			}
			requireRefused(t, o.activate(o.capture(), b, time.Second))
		})
	}
}

func TestAdmissionContextAndIdentity(t *testing.T) {
	o, _ := syntheticOwner(t)
	activateSynthetic(t, o)
	c, ctx, release := admitted(t, o, time.Second)
	if err := c.Check(ctx); err != nil {
		t.Fatal(err)
	}
	if c.InstallationKey() != "installation" || c.Database() != "workspace" || c.Branch() != "main" || c.Epoch() != 2 {
		t.Fatal("binding accessors changed")
	}
	other, _ := syntheticOwner(t) // identical text, distinct provider owner
	activateSynthetic(t, other)
	foreign, foreignCtx, _ := admitted(t, other, time.Second)
	requireRefused(t, c.Check(foreignCtx))
	requireRefused(t, foreign.Check(ctx))
	_, secondCtx, _ := admitted(t, o, time.Second)
	descendant, cancel := context.WithCancel(ctx)
	defer cancel()
	shorter, stop := context.WithTimeout(ctx, 100*time.Millisecond)
	defer stop()
	for _, candidate := range []context.Context{nil, context.Background(), parentContext(t), secondCtx,
		context.WithoutCancel(ctx), descendant, shorter} {
		requireRefused(t, c.Check(candidate))
	}
	copy := c
	release()
	release()
	requireRefused(t, c.Check(ctx))
	requireRefused(t, copy.Check(ctx))
	if !errors.Is(ctx.Err(), context.Canceled) {
		t.Fatal("release did not cancel operation")
	}
}

func TestAnchorAndIntervalRefusals(t *testing.T) {
	for _, name := range []string{"zero", "other owner", "pre revoke", "elapsed", "equal elapsed", "nonpositive", "over max"} {
		t.Run(name, func(t *testing.T) {
			o, clock := syntheticOwner(t)
			a, remaining := o.capture(), time.Second
			switch name {
			case "zero":
				a = anchor{}
			case "other owner":
				other, _ := syntheticOwner(t)
				a = other.capture()
			case "pre revoke":
				o.Revoke()
			case "elapsed":
				clock.add(2 * time.Second)
			case "equal elapsed":
				clock.add(time.Second)
			case "nonpositive":
				remaining = 0
			case "over max":
				remaining = 31 * time.Second
			}
			requireRefused(t, o.activate(a, syntheticEra(), remaining))
		})
	}
	o, clock := syntheticOwner(t)
	a := o.capture()
	clock.add(3 * time.Second)
	if err := o.activate(a, syntheticEra(), 5*time.Second); err != nil {
		t.Fatal(err)
	}
	// The proof work's three seconds cannot be added back by a later timestamp.
	_, _, _, err := o.Begin(parentContext(t), 2*time.Second)
	requireRefused(t, err)
	admitted(t, o, time.Second)
}

func TestStrictBudgetsAndParentCancellation(t *testing.T) {
	o, clock := syntheticOwner(t)
	activateSynthetic(t, o)
	for _, budget := range []time.Duration{0, -1, 10 * time.Second, 11 * time.Second} {
		_, _, _, err := o.Begin(parentContext(t), budget)
		requireRefused(t, err)
	}
	_, _, _, err := o.Begin(context.Background(), time.Second)
	requireRefused(t, err)
	cancelled, cancel := context.WithCancel(parentContext(t))
	cancel()
	_, _, _, err = o.Begin(cancelled, time.Second)
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("cancelled parent: %v", err)
	}
	clock.add(25 * time.Second)
	_, _, _, err = o.Begin(parentContext(t), 5*time.Second)
	requireRefused(t, err)
	clock.add(5 * time.Second)
	_, _, _, err = o.Begin(parentContext(t), time.Nanosecond)
	requireRefused(t, err)

	o2, _ := syntheticOwner(t)
	activateSynthetic(t, o2)
	parent, cancelParent := context.WithCancel(parentContext(t))
	c, ctx, release, err := o2.Begin(parent, time.Second)
	if err != nil {
		t.Fatal(err)
	}
	defer release()
	cancelParent()
	if !errors.Is(c.Check(ctx), context.Canceled) {
		t.Fatal("parent cancellation lost")
	}
}

func TestExtensionPreservesOneAdmission(t *testing.T) {
	o, clock := syntheticOwner(t)
	activateSynthetic(t, o)
	c, ctx, release := admitted(t, o, 5*time.Second)
	originalDeadline, _ := ctx.Deadline()
	clock.add(time.Second)
	if err := o.extend(o.capture(), 30*time.Second); err != nil {
		t.Fatal(err)
	}
	deadline, _ := ctx.Deadline()
	if deadline != originalDeadline || ctx.Err() != nil {
		t.Fatal("routine renewal changed/cancelled admission")
	}
	if err := c.Check(ctx); err != nil {
		t.Fatal(err)
	}
	_, fresh, _ := admitted(t, o, 5*time.Second)
	requireRefused(t, c.Check(fresh)) // old token cannot authorize another Begin
	clock.add(4 * time.Second)
	if !errors.Is(c.Check(ctx), context.DeadlineExceeded) {
		t.Fatal("extension prolonged old admission")
	}
	release()
	requireRefused(t, c.Check(ctx))
	clock.add(24 * time.Second) // t0+29: original has 1s, renewed interval has 2s
	_, _, _, err := o.Begin(parentContext(t), 2*time.Second)
	requireRefused(t, err)
	admitted(t, o, 1500*time.Millisecond) // requires extension to have taken effect
	// Renewal does not replace an active era or shorten its usable interval.
	requireRefused(t, o.activate(o.capture(), syntheticEra(), 30*time.Second))
	requireRefused(t, o.extend(o.capture(), time.Second))
	clock.add(30 * time.Second)
	requireRefused(t, o.extend(o.capture(), 30*time.Second))
}

func TestRevokeThenSyntheticActivationNeverRevivesCopies(t *testing.T) {
	o, _ := syntheticOwner(t)
	activateSynthetic(t, o)
	c, ctx, _ := admitted(t, o, time.Second)
	copy := c
	staleAnchor := o.capture()
	o.Revoke()
	o.Revoke()
	if !errors.Is(context.Cause(ctx), graph.ErrNotAuthority) {
		t.Fatal("revocation did not cancel")
	}
	requireRefused(t, o.extend(staleAnchor, 30*time.Second))
	requireRefused(t, o.activate(staleAnchor, syntheticEra(), 30*time.Second))
	activateSynthetic(t, o) // synthetic inactive activation, not production re-arm
	_, fresh, _ := admitted(t, o, time.Second)
	requireRefused(t, c.Check(ctx))
	requireRefused(t, copy.Check(ctx))
	requireRefused(t, c.Check(fresh))
}

func TestCloseWaitsForReleaseAndIsTerminal(t *testing.T) {
	o, _ := syntheticOwner(t)
	activateSynthetic(t, o)
	c, ctx, release := admitted(t, o, time.Second)
	closing, cancel := context.WithTimeout(context.Background(), 20*time.Millisecond)
	defer cancel()
	// A borrower ignoring cancellation is not reported as cleaned up.
	if err := o.Close(closing); !errors.Is(err, context.DeadlineExceeded) {
		t.Fatalf("close: %v", err)
	}
	requireRefused(t, c.Check(ctx))
	_, _, _, err := o.Begin(parentContext(t), time.Second)
	requireRefused(t, err)
	requireRefused(t, o.activate(o.capture(), syntheticEra(), time.Second))
	requireRefused(t, o.extend(o.capture(), time.Second))
	release()
	if err := o.Close(parentContext(t)); err != nil {
		t.Fatal(err)
	}
	if err := o.Close(parentContext(t)); err != nil {
		t.Fatal(err)
	}
}

func TestCloseAndReleaseConcurrent(t *testing.T) {
	o, _ := syntheticOwner(t)
	activateSynthetic(t, o)
	_, ctx, release := admitted(t, o, time.Second)
	result := make(chan error, 1)
	closing := parentContext(t)
	go func() { result <- o.Close(closing) }()
	select {
	case <-ctx.Done():
	case <-time.After(time.Second):
		t.Fatal("Close did not cancel borrower")
	}
	release()
	select {
	case err := <-result:
		if err != nil {
			t.Fatal(err)
		}
	case <-time.After(time.Second):
		t.Fatal("Close did not join released borrower")
	}
}

func TestBeginVersusRevoke(t *testing.T) {
	o, _ := syntheticOwner(t)
	activateSynthetic(t, o)
	parent := parentContext(t)
	start, done := make(chan struct{}), make(chan error, 1)
	go func() {
		<-start
		c, ctx, release, err := o.Begin(parent, time.Second)
		if err != nil {
			done <- err
			return
		}
		defer release()
		<-ctx.Done()
		done <- c.Check(ctx)
	}()
	close(start)
	o.Revoke()
	select {
	case err := <-done:
		requireRefused(t, err)
	case <-time.After(time.Second):
		t.Fatal("admission escaped revoke")
	}
}

func TestCopiedOwnerRefusesBeforeSharedState(t *testing.T) {
	o, _ := syntheticOwner(t)
	activateSynthetic(t, o)
	c, ctx, release := admitted(t, o, time.Second)
	// Reflection reproduces an external value copy without adding an intentional
	// copylocks diagnostic to the repository. No concurrent source copy occurs.
	value := reflect.New(reflect.TypeOf(Owner{}))
	value.Elem().Set(reflect.ValueOf(o).Elem())
	copy := value.Interface().(*Owner)
	got, _, cleanup, err := copy.Begin(parentContext(t), time.Second)
	if cleanup != nil {
		cleanup()
	}
	if !errors.Is(err, graph.ErrNotAuthority) || got.admission != nil {
		t.Errorf("copied owner admitted: %v", err)
	}
	if copy.capture() != (anchor{}) {
		t.Error("copied owner captured an anchor")
	}
	if !errors.Is(copy.activate(o.capture(), syntheticEra(), time.Second), graph.ErrNotAuthority) {
		t.Error("copied owner activated")
	}
	if !errors.Is(copy.extend(o.capture(), 30*time.Second), graph.ErrNotAuthority) {
		t.Error("copied owner extended")
	}
	copy.Revoke()
	if err := c.Check(ctx); err != nil {
		t.Errorf("copied revoke affected original: %v", err)
	}
	closing, stop := context.WithTimeout(context.Background(), 20*time.Millisecond)
	defer stop()
	if !errors.Is(copy.Close(closing), graph.ErrNotAuthority) {
		t.Error("copied owner did not refuse Close")
	}
	if err := c.Check(ctx); err != nil {
		t.Errorf("copied Close affected original: %v", err)
	}
	release()
	admitted(t, o, time.Second)
}

type checkedContext struct {
	context.Context
	once    sync.Once
	checked chan struct{}
}

func (c *checkedContext) Err() error {
	err := c.Context.Err()
	c.once.Do(func() { close(c.checked) })
	return err
}

func TestCanceledWaiterIsNotAdmitted(t *testing.T) {
	o, _ := syntheticOwner(t)
	activateSynthetic(t, o)
	parent, cancel := context.WithCancel(parentContext(t))
	defer cancel()
	ctx := &checkedContext{Context: parent, checked: make(chan struct{})}
	result := make(chan error, 1)
	o.mu.Lock()
	go func() {
		_, _, release, err := o.Begin(ctx, time.Second)
		if release != nil {
			release()
		}
		result <- err
	}()
	select {
	case <-ctx.checked:
	case <-time.After(time.Second):
		o.mu.Unlock()
		t.Fatal("Begin did not perform initial context check")
	}
	cancel()
	o.mu.Unlock()
	select {
	case err := <-result:
		if !errors.Is(err, context.Canceled) {
			t.Fatalf("canceled waiter was admitted: %v", err)
		}
	case <-time.After(time.Second):
		t.Fatal("Begin did not return after owner lock released")
	}
}

func TestNilZeroAndCanceledClose(t *testing.T) {
	for _, o := range []*Owner{nil, {}} {
		o.Revoke()
		requireRefused(t, o.Close(parentContext(t)))
		if o.capture() != (anchor{}) {
			t.Fatal("invalid owner captured anchor")
		}
		requireRefused(t, o.activate(anchor{}, syntheticEra(), time.Second))
		requireRefused(t, o.extend(anchor{}, time.Second))
	}
	o, _ := syntheticOwner(t)
	activateSynthetic(t, o)
	canceled, cancel := context.WithCancel(parentContext(t))
	cancel()
	if err := o.Close(canceled); !errors.Is(err, context.Canceled) {
		t.Fatalf("canceled idle Close: %v", err)
	}
	requireRefused(t, o.activate(o.capture(), syntheticEra(), time.Second))
	if err := o.Close(parentContext(t)); err != nil {
		t.Fatal(err)
	}
	o2, _ := syntheticOwner(t)
	activateSynthetic(t, o2)
	requireRefused(t, o2.Close(context.Background()))
	requireRefused(t, o2.activate(o2.capture(), syntheticEra(), time.Second))
}

func TestJustInsideBudgetAndExtensionBoundaries(t *testing.T) {
	o, clock := syntheticOwner(t)
	preActivation := o.capture()
	activateSynthetic(t, o)
	admitted(t, o, 10*time.Second-time.Nanosecond)
	clock.add(25 * time.Second)
	admitted(t, o, 5*time.Second-time.Nanosecond)
	requireRefused(t, o.extend(o.capture(), 5*time.Second)) // equal usable-until
	requireRefused(t, o.extend(preActivation, 30*time.Second))
}

func TestSurfaceHasNoPositiveConstructorOrSQL(t *testing.T) {
	packages, err := parser.ParseDir(token.NewFileSet(), ".", func(info os.FileInfo) bool { return !strings.HasSuffix(info.Name(), "_test.go") }, 0)
	if err != nil {
		t.Fatal(err)
	}
	var exported []string
	allowed := map[string]bool{"context": true, "sync": true, "time": true, "github.com/steveyegge/beads/graphops": true}
	for _, f := range packages["graphcap"].Files {
		for _, imp := range f.Imports {
			name, err := strconv.Unquote(imp.Path.Value)
			if err != nil || !allowed[name] {
				t.Fatalf("unreviewed dependency: %s", imp.Path.Value)
			}
		}
		for _, decl := range f.Decls {
			switch d := decl.(type) {
			case *ast.FuncDecl:
				if d.Recv != nil && receiverName(d.Recv.List[0].Type) == "Owner" {
					if _, ok := d.Recv.List[0].Type.(*ast.StarExpr); !ok {
						t.Fatal("Owner method copies receiver")
					}
				}
				if d.Name.IsExported() {
					name := d.Name.Name
					if d.Recv != nil {
						name = receiverName(d.Recv.List[0].Type) + "." + name
					}
					exported = append(exported, name)
				}
			case *ast.GenDecl:
				for _, spec := range d.Specs {
					if value, ok := spec.(*ast.ValueSpec); ok {
						for _, name := range value.Names {
							if name.IsExported() {
								t.Fatalf("exported value: %s", name)
							}
						}
					}
					if typ, ok := spec.(*ast.TypeSpec); ok && typ.Name.IsExported() {
						exported = append(exported, typ.Name.Name)
						if st, ok := typ.Type.(*ast.StructType); ok {
							for _, field := range st.Fields.List {
								if len(field.Names) == 0 {
									t.Fatalf("embedded field in exported type %s", typ.Name.Name)
								}
								for _, name := range field.Names {
									if name.IsExported() {
										t.Fatalf("exported field: %s", name)
									}
								}
							}
						}
					}
				}
			}
		}
	}
	want := []string{"LeaseClaim", "LeaseClaim.Branch", "LeaseClaim.Check", "LeaseClaim.Database", "LeaseClaim.Epoch", "LeaseClaim.InstallationKey", "Owner", "Owner.Begin", "Owner.Close", "Owner.Revoke"}
	sort.Strings(exported)
	if !reflect.DeepEqual(exported, want) {
		t.Fatalf("exported surface: %v", exported)
	}
}

func receiverName(expr ast.Expr) string {
	if pointer, ok := expr.(*ast.StarExpr); ok {
		expr = pointer.X
	}
	return expr.(*ast.Ident).Name
}
