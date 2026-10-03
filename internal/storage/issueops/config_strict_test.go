package issueops

import (
	"context"
	"errors"
	"reflect"
	"regexp"
	"strings"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/steveyegge/beads/internal/config"
	"github.com/steveyegge/beads/internal/storage/domain"
	"github.com/steveyegge/beads/internal/types"
)

func strictConfigYAML(t *testing.T) {
	t.Helper()
	t.Setenv("BEADS_TEST_IGNORE_REPO_CONFIG", "1")
	config.ResetForTesting()
	if err := config.Initialize(); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(config.ResetForTesting)
	config.Set("status.custom", []string{"yaml-active:active"})
	config.Set("types.custom", []string{"yaml-type"})
	config.Set("types.infra", []string{"yaml-infra"})
}

func TestResolveCustomConfigStrictFailures(t *testing.T) {
	strictConfigYAML(t)
	queries := []string{
		"SELECT name, category FROM custom_statuses ORDER BY name",
		"SELECT name FROM custom_types ORDER BY name",
		"SELECT `key`, value FROM config WHERE `key` IN (?,?)",
	}
	columns := [][]string{{"name", "category"}, {"name"}, {"key", "value"}}
	for stage, query := range queries {
		for _, failure := range []string{"query", "scan", "iteration"} {
			t.Run(strings.Join([]string{columns[stage][0], failure}, "/")+string(rune('0'+stage)), func(t *testing.T) {
				_, mock, tx := beginMockTx(t)
				for earlier := 0; earlier < stage; earlier++ {
					mock.ExpectQuery(regexp.QuoteMeta(queries[earlier])).WillReturnRows(sqlmock.NewRows(columns[earlier]))
				}
				sentinel := errors.New("targeted downstream config failure")
				expected := mock.ExpectQuery(regexp.QuoteMeta(query))
				if stage == 2 {
					expected.WithArgs("status.custom", "types.custom")
				}
				switch failure {
				case "query":
					expected.WillReturnError(sentinel)
				case "scan":
					rows := sqlmock.NewRows(columns[stage])
					if stage == 1 {
						rows.AddRow(nil)
					} else {
						rows.AddRow("present", nil)
					}
					expected.WillReturnRows(rows)
				case "iteration":
					rows := sqlmock.NewRows(columns[stage])
					if stage == 1 {
						rows.AddRow("present")
					} else {
						rows.AddRow("present", "active")
					}
					expected.WillReturnRows(rows.RowError(0, sentinel))
				}
				mock.ExpectRollback()
				statuses, kinds, err := ResolveCustomConfigStrictInTx(context.Background(), tx)
				if err == nil || statuses != nil || kinds != nil {
					t.Fatalf("failure must return no policy: statuses=%v types=%v err=%v", statuses, kinds, err)
				}
				if failure != "scan" && !errors.Is(err, sentinel) {
					t.Fatalf("lost underlying error: %v", err)
				}
				if failure == "scan" && !strings.Contains(err.Error(), "scan") {
					t.Fatalf("missing scan error: %v", err)
				}
				if err := tx.Rollback(); err != nil {
					t.Fatal(err)
				}
				if err := mock.ExpectationsWereMet(); err != nil {
					t.Fatal(err)
				}
			})
		}
	}
}

func TestResolveInfraTypesStrictFailures(t *testing.T) {
	strictConfigYAML(t)
	for _, failure := range []string{"query", "scan", "iteration"} {
		t.Run(failure, func(t *testing.T) {
			_, mock, tx := beginMockTx(t)
			sentinel := errors.New("infra config failed")
			expected := mock.ExpectQuery(regexp.QuoteMeta("SELECT value FROM config WHERE `key` = ?")).WithArgs("types.infra")
			switch failure {
			case "query":
				expected.WillReturnError(sentinel)
			case "scan":
				expected.WillReturnRows(sqlmock.NewRows([]string{"value"}).AddRow(nil))
			case "iteration":
				expected.WillReturnRows(sqlmock.NewRows([]string{"value"}).AddRow("infra").RowError(0, sentinel))
			}
			mock.ExpectRollback()
			got, err := ResolveInfraTypesStrictInTx(context.Background(), tx)
			if err == nil || got != nil {
				t.Fatalf("failure must not yield YAML/default policy: %v %v", got, err)
			}
			if failure != "scan" && !errors.Is(err, sentinel) {
				t.Fatalf("lost underlying error: %v", err)
			}
			if err := tx.Rollback(); err != nil {
				t.Fatal(err)
			}
			if err := mock.ExpectationsWereMet(); err != nil {
				t.Fatal(err)
			}
		})
	}
}

func TestResolveCustomConfigStrictPrecedence(t *testing.T) {
	strictConfigYAML(t)
	for _, tc := range []struct {
		name                     string
		table                    bool
		statusConfig, typeConfig string
		wantStatus, wantType     string
	}{
		{name: "normalized tables win over YAML", table: true, wantStatus: "table-active", wantType: "table-type"},
		{name: "config strings win over YAML", statusConfig: "db-active:active", typeConfig: `["db-type"]`, wantStatus: "db-active", wantType: "db-type"},
		{name: "genuinely absent config uses YAML", wantStatus: "yaml-active", wantType: "yaml-type"},
		{name: "malformed nonempty status keeps existing degraded parse policy", statusConfig: "db:invalid-category", wantType: "yaml-type"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			_, mock, tx := beginMockTx(t)
			statusRows := sqlmock.NewRows([]string{"name", "category"})
			typeRows := sqlmock.NewRows([]string{"name"})
			if tc.table {
				statusRows.AddRow("table-active", "active")
				typeRows.AddRow("table-type")
			}
			mock.ExpectQuery(regexp.QuoteMeta("SELECT name, category FROM custom_statuses ORDER BY name")).WillReturnRows(statusRows)
			mock.ExpectQuery(regexp.QuoteMeta("SELECT name FROM custom_types ORDER BY name")).WillReturnRows(typeRows)
			if !tc.table {
				rows := sqlmock.NewRows([]string{"key", "value"})
				if tc.statusConfig != "" {
					rows.AddRow("status.custom", tc.statusConfig)
				}
				if tc.typeConfig != "" {
					rows.AddRow("types.custom", tc.typeConfig)
				}
				mock.ExpectQuery(regexp.QuoteMeta("SELECT `key`, value FROM config WHERE `key` IN (?,?)")).WithArgs("status.custom", "types.custom").WillReturnRows(rows)
			}
			mock.ExpectRollback()
			statuses, kinds, err := ResolveCustomConfigStrictInTx(context.Background(), tx)
			var wantStatuses []types.CustomStatus
			if tc.wantStatus != "" {
				wantStatuses = []types.CustomStatus{{Name: tc.wantStatus, Category: types.CategoryActive}}
			}
			if err != nil || !reflect.DeepEqual(statuses, wantStatuses) || !reflect.DeepEqual(kinds, []string{tc.wantType}) {
				t.Fatalf("got statuses=%v types=%v err=%v", statuses, kinds, err)
			}
			if err := tx.Rollback(); err != nil {
				t.Fatal(err)
			}
			if err := mock.ExpectationsWereMet(); err != nil {
				t.Fatal(err)
			}
		})
	}
}

func TestResolveInfraTypesStrictPrecedence(t *testing.T) {
	strictConfigYAML(t)
	for _, tc := range []struct {
		name, value    string
		absent, noYAML bool
		want           []string
	}{
		{name: "database", value: "db-infra, other", want: []string{"db-infra", "other"}},
		{name: "absent", absent: true, want: []string{"yaml-infra"}},
		{name: "empty", want: []string{"yaml-infra"}},
		{name: "whitespace", value: " , ", want: []string{"yaml-infra"}},
		{name: "default", absent: true, noYAML: true, want: domain.DefaultInfraTypes()},
	} {
		t.Run(tc.name, func(t *testing.T) {
			config.Set("types.infra", []string{"yaml-infra"})
			if tc.noYAML {
				config.Set("types.infra", []string{})
			}
			_, mock, tx := beginMockTx(t)
			rows := sqlmock.NewRows([]string{"value"})
			if !tc.absent {
				rows.AddRow(tc.value)
			}
			mock.ExpectQuery(regexp.QuoteMeta("SELECT value FROM config WHERE `key` = ?")).WithArgs("types.infra").WillReturnRows(rows)
			mock.ExpectRollback()
			got, err := ResolveInfraTypesStrictInTx(context.Background(), tx)
			want := map[string]bool{}
			for _, name := range tc.want {
				want[name] = true
			}
			if err != nil || !reflect.DeepEqual(got, want) {
				t.Fatalf("got=%v want=%v err=%v", got, want, err)
			}
			if err := tx.Rollback(); err != nil {
				t.Fatal(err)
			}
			if err := mock.ExpectationsWereMet(); err != nil {
				t.Fatal(err)
			}
		})
	}
}

// Existing degraded consumers deliberately continue using their old behavior.
func TestResolveConfigLegacyDegradationUnchanged(t *testing.T) {
	strictConfigYAML(t)
	for _, failure := range []string{"query", "scan"} {
		t.Run(failure, func(t *testing.T) {
			_, mock, tx := beginMockTx(t)
			status := mock.ExpectQuery(regexp.QuoteMeta("SELECT name, category FROM custom_statuses ORDER BY name"))
			if failure == "query" {
				status.WillReturnError(errors.New("unavailable"))
			} else {
				status.WillReturnRows(sqlmock.NewRows([]string{"name", "category"}).AddRow("broken", nil))
			}
			kind := mock.ExpectQuery(regexp.QuoteMeta("SELECT name FROM custom_types ORDER BY name"))
			if failure == "query" {
				kind.WillReturnError(errors.New("unavailable"))
			} else {
				kind.WillReturnRows(sqlmock.NewRows([]string{"name"}).AddRow(nil))
			}
			mock.ExpectQuery(regexp.QuoteMeta("SELECT `key`, value FROM config WHERE `key` IN (?,?)")).WithArgs("status.custom", "types.custom").WillReturnError(errors.New("unavailable"))
			mock.ExpectQuery(regexp.QuoteMeta("SELECT value FROM config WHERE `key` = ?")).WithArgs("types.infra").WillReturnError(errors.New("unavailable"))
			mock.ExpectRollback()
			statuses, kinds, err := ResolveCustomConfigInTx(context.Background(), tx)
			if err != nil || !reflect.DeepEqual(statuses, []types.CustomStatus{{Name: "yaml-active", Category: types.CategoryActive}}) || !reflect.DeepEqual(kinds, []string{"yaml-type"}) {
				t.Fatalf("legacy degradation changed: %v %v %v", statuses, kinds, err)
			}
			if got := ResolveInfraTypesInTx(context.Background(), tx); !reflect.DeepEqual(got, map[string]bool{"yaml-infra": true}) {
				t.Fatalf("legacy infra changed: %v", got)
			}
			if err := tx.Rollback(); err != nil {
				t.Fatal(err)
			}
			if err := mock.ExpectationsWereMet(); err != nil {
				t.Fatal(err)
			}
		})
	}
}
