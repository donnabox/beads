package graphstore

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
)

// Real canonical workspace path, otherwise complete explicit Options. This
// fixture never creates a database/engine directory and never opens a Store.
func profileAdmissionOptions(t *testing.T, version int) Options {
	t.Helper()
	workspace, err := filepath.EvalSymlinks(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	return Options{Backend: "embedded", Database: "profile_admission", Branch: "main",
		DataDir: filepath.Join(workspace, "uncreated-engine"), TypeGeneration: TypeGeneration,
		Binding: Binding{WorkspaceID: workspace, ScopeURL: "https://example.test/profile/", AuthorityID: strings.Repeat("a", 32), SchemaVersion: version}}
}

func TestOrdinaryProfileAdmissionPreserves11And14(t *testing.T) {
	for _, backend := range []string{"embedded", "server"} {
		for _, version := range []int{0, SchemaVersion, 12, 13, TypesFreshSchemaVersion, AtomicBatchSchemaVersion, 16} {
			t.Run(fmt.Sprintf("%s/%d", backend, version), func(t *testing.T) {
				o := profileAdmissionOptions(t, version)
				o.Backend = backend
				if backend == "server" {
					o.ServerHost = "127.0.0.1"
					o.ServerPort = 3306
				}
				err := validateOptions(o)
				if version == SchemaVersion || version == TypesFreshSchemaVersion {
					if err != nil {
						t.Fatalf("existing explicit layout%d rejected: %v", version, err)
					}
				} else if !errors.Is(err, ErrInvalidStore) || !strings.Contains(err.Error(), "invalid preview version") {
					t.Fatalf("ordinary admission accepted/defaulted layout%d or changed refusal: %v", version, err)
				}
				if _, err := os.Stat(o.DataDir); !errors.Is(err, os.ErrNotExist) {
					t.Fatalf("validation touched engine path: %v", err)
				}
			})
		}
	}
}

func TestExplicitProfileAdmissionRequiresExpectedVersion(t *testing.T) {
	for _, expected := range []int{SchemaVersion, TypesFreshSchemaVersion, AtomicBatchSchemaVersion} {
		for _, version := range []int{0, SchemaVersion, 12, TypesFreshSchemaVersion, AtomicBatchSchemaVersion, 16} {
			t.Run(fmt.Sprintf("expected%d/actual%d", expected, version), func(t *testing.T) {
				o := profileAdmissionOptions(t, version)
				err := validateOptionsAtVersion(o, expected)
				if version == expected {
					if err != nil {
						t.Fatal(err)
					}
				} else if !errors.Is(err, ErrInvalidStore) || !strings.Contains(err.Error(), "invalid preview version") {
					t.Fatalf("expected equality lost: expected%d actual%d: %v", expected, version, err)
				}
			})
		}
	}
}

func TestOrdinaryProfileAdmissionKeepsFullOptionsValidation(t *testing.T) {
	cases := []struct {
		name, diagnosis string
		change          func(*Options)
	}{
		{"backend", "backend must", func(o *Options) { o.Backend = "unsupported" }},
		{"database", "invalid database", func(o *Options) { o.Database = "bad-name" }},
		{"branch", "unsupported branch", func(o *Options) { o.Branch = "other" }},
		{"authority", "authority ID", func(o *Options) { o.Binding.AuthorityID = "invalid" }},
		{"generation", "Type generation", func(o *Options) { o.TypeGeneration = TypeGeneration + 1 }},
		{"scope", "invalid Scope", func(o *Options) { o.Binding.ScopeURL = "not-a-scope" }},
		{"workspace", "absolute canonical path", func(o *Options) { o.Binding.WorkspaceID = "relative" }},
		{"data-dir", "embedded data path", func(o *Options) { o.DataDir = "relative" }},
		{"server", "explicit server address", func(o *Options) { o.Backend = "server" }},
	}
	for _, version := range []int{SchemaVersion, TypesFreshSchemaVersion} {
		for _, tc := range cases {
			t.Run(fmt.Sprintf("%d/%s", version, tc.name), func(t *testing.T) {
				o := profileAdmissionOptions(t, version)
				tc.change(&o)
				err := validateOptions(o)
				if !errors.Is(err, ErrInvalidStore) || !strings.Contains(err.Error(), tc.diagnosis) {
					t.Fatalf("full option guard lost: %v", err)
				}
			})
		}
	}
}

func TestProfileEntryRefusalsPrecedeEngineProvisioning(t *testing.T) {
	// Poisoned relative DataDir is an independent pre-backend fence if a schema
	// predicate regresses: then the wrong diagnostic fails WITHOUT launching
	// an engine. Positive initialization/reopen belongs to existing cgo tests.
	for _, version := range []int{0, 12, AtomicBatchSchemaVersion, 16} {
		t.Run(fmt.Sprintf("ordinary/%d", version), func(t *testing.T) {
			o := profileAdmissionOptions(t, version)
			o.DataDir = "must-not-reach-backend"
			if err := Init(context.Background(), o); !errors.Is(err, ErrInvalidStore) || !strings.Contains(err.Error(), "invalid preview version") {
				t.Fatalf("Init wrong pre-engine refusal: %v", err)
			}
			s, err := OpenExisting(context.Background(), o)
			if s != nil || !errors.Is(err, ErrInvalidStore) || !strings.Contains(err.Error(), "invalid preview version") {
				t.Fatalf("OpenExisting wrong pre-engine refusal: %v %v", s, err)
			}
		})
	}
	for _, version := range []int{0, SchemaVersion, 12, TypesFreshSchemaVersion, 16} {
		t.Run(fmt.Sprintf("atomic/%d", version), func(t *testing.T) {
			o := profileAdmissionOptions(t, version)
			o.DataDir = "must-not-reach-backend"
			if err := InitAtomicBatch(context.Background(), o); !errors.Is(err, ErrInvalidStore) || !strings.Contains(err.Error(), "invalid preview version") {
				t.Fatalf("InitAtomicBatch wrong pre-engine refusal: %v", err)
			}
			s, err := OpenAtomicBatch(context.Background(), o)
			if s != nil || !errors.Is(err, ErrInvalidStore) || !strings.Contains(err.Error(), "invalid preview version") {
				t.Fatalf("OpenAtomicBatch wrong pre-engine refusal: %v %v", s, err)
			}
		})
	}
}

func TestOrdinaryAllocatedIDAbsenceNeverReadsBatchTable(t *testing.T) {
	const catalogQuery = `SELECT 1 FROM graph_preview_catalog
        WHERE (
            (backing='issue' AND backing_key=?) OR
            (backing IN ('generic','dependency','informational','intrinsic','intrinsic-link','type-definition','installed-type','installed-type-link') AND COALESCE(public_id,path)=?)
        ) LIMIT 1`
	const versionQuery = `SELECT schema_version FROM graph_preview_scope WHERE singleton=1`
	for _, version := range []int{SchemaVersion, TypesFreshSchemaVersion} {
		for _, mode := range []string{"absent", "catalog-taken", "catalog-error", "scope-error", "scope-missing", "scope-conversion"} {
			t.Run(fmt.Sprintf("%d/%s", version, mode), func(t *testing.T) {
				tx, mock := intrinsicProfileTransaction(t)
				fault := errors.New("allocation fixture fault")
				q := mock.ExpectQuery(catalogQuery).WithArgs("new-id", "new-id")
				switch mode {
				case "catalog-taken":
					q.WillReturnRows(sqlmock.NewRows([]string{"exists"}).AddRow(1))
				case "catalog-error":
					q.WillReturnError(fault)
				default:
					q.WillReturnError(sql.ErrNoRows)
				}
				if mode != "catalog-taken" && mode != "catalog-error" {
					q := mock.ExpectQuery(versionQuery)
					switch mode {
					case "scope-error":
						q.WillReturnError(fault)
					case "scope-missing":
						q.WillReturnError(sql.ErrNoRows)
					case "scope-conversion":
						q.WillReturnRows(sqlmock.NewRows([]string{"schema_version"}).AddRow("not-a-version"))
					default:
						q.WillReturnRows(sqlmock.NewRows([]string{"schema_version"}).AddRow(version))
					}
				}
				owner, err := allocatedResourceIDOwnerInTx(context.Background(), tx, "new-id")
				switch mode {
				case "absent":
					if err != nil || owner != "" {
						t.Fatalf("ordinary absence refused: %q %v", owner, err)
					}
				case "catalog-taken":
					if err != nil || owner != "a live or retained Resource allocation" {
						t.Fatalf("existing allocation lost: %q %v", owner, err)
					}
				default:
					if err == nil || owner != "" {
						t.Fatalf("partial owner on failed read: %q %v", owner, err)
					}
					if (mode == "catalog-error" || mode == "scope-error") && !errors.Is(err, fault) {
						t.Fatalf("lost driver cause: %v", err)
					}
					if mode == "scope-missing" && !errors.Is(err, sql.ErrNoRows) {
						t.Fatalf("lost missing-binding cause: %v", err)
					}
					if mode == "scope-conversion" && !strings.Contains(err.Error(), "Scan error") {
						t.Fatalf("lost binding conversion refusal: %v", err)
					}
				}
				// Strict QueryMatcherEqual and no batch-table expectation mean
				// querying the15 table on11/14 fails, never matches a wildcard.
			})
		}
	}
}
