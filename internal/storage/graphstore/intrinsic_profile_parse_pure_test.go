package graphstore

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/dolthub/vitess/go/vt/sqlparser"
)

// These are source syntax controls, not actual SHOW fixtures or native
// canonical equivalence evidence. The enclosing profile remains unsupported.
func TestIntrinsicSyntaxCandidates(t *testing.T) {
	for _, text := range []string{
		"CREATE TABLE demo (id BIGINT NOT NULL PRIMARY KEY, label VARCHAR(32), INDEX ix_label (label));",
		"CREATE TABLE `demo` (`id` BIGINT PRIMARY KEY)",
	} {
		got, err := parseIntrinsicSyntax(context.Background(), text, "demo", "BASE TABLE")
		if err != nil || len(got.columns) == 0 || len(got.tokens) == 0 {
			t.Fatalf("source syntax: %v", err)
		}
	}
	if _, err := parseIntrinsicSyntax(context.Background(), "CREATE VIEW demo AS SELECT id FROM issues", "demo", "VIEW"); err != nil {
		t.Fatal(err)
	}
	// A dot within a quoted literal name is not a database qualifier.
	got, err := parseIntrinsicSyntax(context.Background(), "CREATE TABLE `other.demo` (id INT)", "other.demo", "BASE TABLE")
	if err != nil || got.name != "other.demo" {
		t.Fatalf("literal name changed: %q, %v", got.name, err)
	}
}

func TestIntrinsicSyntaxRefusals(t *testing.T) {
	for _, text := range []string{
		"CREATE TABLE wrong (id INT)",
		"CREATE TABLE demo (id INT); SELECT 1",
		"CREATE TABLE demo (id INT);;",
		"CREATE TABLE demo (id INT) /* comment */",
		"/*! CREATE TABLE demo (id INT) */",
		"CREATE TABLE demo (id INT DEFAULT ?)",
		"CREATE TEMPORARY TABLE demo (id INT)",
		"CREATE TABLE demo AS SELECT 1",
		"CREATE TABLE demo LIKE other",
		"CREATE TABLE demo (id INT, id INT)",
		"CREATE TABLE demo (id INT, v INT GENERATED ALWAYS AS (id+1))",
		"CREATE TABLE demo (id INT, FULLTEXT INDEX ix (id))",
		"CREATE TABLE demo (id INT) this is not a valid clause",
	} {
		t.Run(text, func(t *testing.T) {
			got, err := parseIntrinsicSyntax(context.Background(), text, "demo", "BASE TABLE")
			if err == nil || got.name != "" || len(got.tokens) != 0 {
				t.Fatalf("unsupported syntax accepted: %v", err)
			}
		})
	}
	for _, text := range []string{"CREATE VIEW demo(alias) AS SELECT id FROM issues", "CREATE VIEW wrong AS SELECT id FROM issues"} {
		if _, err := parseIntrinsicSyntax(context.Background(), text, "demo", "VIEW"); err == nil {
			t.Fatal(text)
		}
	}
	// These controls require the pinned parser to produce a qualified MySQL
	// name before checking our refusal, rather than passing on a parse error.
	for _, tc := range []struct {
		name string
		text string
		kind string
	}{
		{"database-qualified-table", "CREATE TABLE other.demo (id INT)", "BASE TABLE"},
		{"database-qualified-view", "CREATE VIEW other.demo AS SELECT id FROM issues", "VIEW"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			stmt, _, err := sqlparser.ParseOneWithOptions(context.Background(), tc.text, sqlparser.ParserOptions{})
			if err != nil {
				t.Fatalf("qualified MySQL source did not parse: %v", err)
			}
			ddl, ok := stmt.(*sqlparser.DDL)
			if !ok {
				t.Fatalf("qualified MySQL source is %T", stmt)
			}
			name := ddl.Table
			if tc.kind == "VIEW" {
				if ddl.ViewSpec == nil {
					t.Fatal("missing qualified view")
				}
				name = ddl.ViewSpec.ViewName
			}
			if name.Name.String() != "demo" || name.DbQualifier.String() != "other" {
				t.Fatalf("missing parsed qualifier: %#v", name)
			}
			got, err := parseIntrinsicSyntax(context.Background(), tc.text, "demo", tc.kind)
			if !errors.Is(err, ErrInvalidStore) || got.name != "" || got.kind != "" || len(got.tokens) != 0 || len(got.columns) != 0 {
				t.Fatalf("qualified name did not refuse with zero result: %#v, %v", got, err)
			}
		})
	}
	// SchemaQualifier is unused by the pinned MySQL grammar. These are AST-only
	// field-law controls, not claims that native SHOW or MySQL parsing emits it.
	for _, tc := range []struct {
		name string
		db   string
	}{
		{"schema-qualifier-AST-only", ""},
		{"database-and-schema-AST-only", "otherdb"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			name := sqlparser.TableName{Name: sqlparser.NewTableIdent("demo"), DbQualifier: sqlparser.NewTableIdent(tc.db), SchemaQualifier: sqlparser.NewTableIdent("other")}
			if intrinsicUnqualifiedSQLName(name, "demo") {
				t.Fatal("schema qualifier accepted")
			}
			name.DbQualifier = sqlparser.TableIdent{}
			name.SchemaQualifier = sqlparser.TableIdent{}
			if !intrinsicUnqualifiedSQLName(name, "demo") {
				t.Fatal("unqualified positive control refused")
			}
		})
	}
}

func TestIntrinsicSyntaxBoundsAndCancellation(t *testing.T) {
	for _, text := range []string{
		strings.Repeat(" ", intrinsicStatementLimit+1),
		"SELECT " + strings.Repeat("(", intrinsicNestingLimit+1) + "1" + strings.Repeat(")", intrinsicNestingLimit+1),
		strings.Repeat("x ", intrinsicTokenLimit+1),
	} {
		if _, err := scanIntrinsicSQL(context.Background(), text); !errors.Is(err, ErrLimitExceeded) {
			t.Fatalf("input bound: %v", err)
		}
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if _, err := parseIntrinsicSyntax(ctx, "CREATE TABLE demo (id INT)", "demo", "BASE TABLE"); !errors.Is(err, context.Canceled) {
		t.Fatal(err)
	}
	var deep any = 1
	for i := 0; i < intrinsicNestingLimit+1; i++ {
		deep = []any{deep}
	}
	if err := boundIntrinsicAST(context.Background(), deep); !errors.Is(err, ErrLimitExceeded) {
		t.Fatal(err)
	}
}

func TestIntrinsicSyntaxRetainsIndexDirection(t *testing.T) {
	a, err := parseIntrinsicSyntax(context.Background(), "CREATE TABLE demo (id INT, INDEX ix (id ASC))", "demo", "BASE TABLE")
	if err != nil {
		t.Fatal(err)
	}
	b, err := parseIntrinsicSyntax(context.Background(), "CREATE TABLE demo (id INT, INDEX ix (id DESC))", "demo", "BASE TABLE")
	if err != nil {
		t.Fatal(err)
	}
	if len(a.tokens) != len(b.tokens) {
		t.Fatal("unexpected token control")
	}
	different := false
	for i := range a.tokens {
		if a.tokens[i] != b.tokens[i] {
			different = true
		}
	}
	if !different {
		t.Fatal("direction erased by parser formatter")
	}
}

// Captured from real fresh production initialization, still non-admitting until
// source correspondence is qualified; no engine is opened by this parser check.
func TestIntrinsicNativeProjectionParser(t *testing.T) {
	m, err := loadIntrinsicExpected()
	if err != nil {
		t.Fatal(err)
	}
	if m.Native == nil || len(m.Native.Tables) != 42 {
		t.Fatal("missing native observation")
	}
	for name, ddl := range m.Native.Tables {
		t.Run(name, func(t *testing.T) {
			if _, err := parseIntrinsicSyntax(context.Background(), ddl, name, "BASE TABLE"); err != nil {
				t.Error(err)
			}
		})
	}
	for name, v := range m.Native.Views {
		t.Run(name, func(t *testing.T) {
			if _, err := parseIntrinsicSyntax(context.Background(), v.Fragment, name, "VIEW"); err != nil {
				t.Error(err)
			}
		})
	}
}
