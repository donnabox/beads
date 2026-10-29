package graphstore

import (
	"context"
	"fmt"
	"reflect"
	"strings"
	"unicode/utf8"

	"github.com/dolthub/vitess/go/vt/sqlparser"
)

const (
	intrinsicStatementLimit = 128 << 10
	intrinsicTokenLimit     = 32768
	intrinsicNestingLimit   = 64
)

// intrinsicSyntax is deliberately a syntax acquisition, NOT a qualified native
// descriptor. Tokens retain index direction and other spelling the selected AST
// formatter loses. No syntax result can construct checkedStoreProfile.
type intrinsicSyntax struct {
	name    string
	kind    string
	tokens  []intrinsicSQLToken
	columns []string
}

type intrinsicSQLToken struct {
	kind  int
	value string
}

// scanIntrinsicSQL bounds input before the non-preemptible selected parser.
// Context checks do not promise to interrupt a Scan or native parser call.
func scanIntrinsicSQL(ctx context.Context, text string) ([]intrinsicSQLToken, error) {
	if len(text) > intrinsicStatementLimit {
		return nil, fmt.Errorf("%w: intrinsic DDL bytes", ErrLimitExceeded)
	}
	if !utf8.ValidString(text) {
		return nil, fmt.Errorf("%w: DDL UTF-8", ErrInvalidStore)
	}
	tokenizer := sqlparser.NewStringTokenizer(text)
	tokenizer.SkipSpecialComments = true
	tokenizer.AllowComments = true
	var out []intrinsicSQLToken
	depth, ended := 0, false
	for {
		if err := ctx.Err(); err != nil {
			return nil, err
		}
		kind, value := tokenizer.Scan()
		if tokenizer.LastError != nil {
			return nil, fmt.Errorf("%w: DDL lexer: %v", ErrInvalidStore, tokenizer.LastError)
		}
		if kind == 0 {
			break
		}
		if ended {
			return nil, fmt.Errorf("%w: extra DDL statement", ErrInvalidStore)
		}
		if kind == sqlparser.LEX_ERROR || kind == sqlparser.COMMENT || kind == sqlparser.VALUE_ARG || kind == sqlparser.LIST_ARG {
			return nil, fmt.Errorf("%w: DDL comment or bind", ErrInvalidStore)
		}
		switch kind {
		case '(':
			depth++
		case ')':
			depth--
		case ';':
			ended = true
		}
		if depth < 0 || depth > intrinsicNestingLimit {
			return nil, fmt.Errorf("%w: DDL nesting", ErrLimitExceeded)
		}
		if len(out) == intrinsicTokenLimit {
			return nil, fmt.Errorf("%w: DDL tokens", ErrLimitExceeded)
		}
		out = append(out, intrinsicSQLToken{kind: kind, value: string(value)})
	}
	if depth != 0 || len(out) == 0 {
		return nil, fmt.Errorf("%w: incomplete DDL", ErrInvalidStore)
	}
	return out, nil
}

func parseIntrinsicSyntax(ctx context.Context, text, name, kind string) (intrinsicSyntax, error) {
	tokens, err := scanIntrinsicSQL(ctx, text)
	if err != nil {
		return intrinsicSyntax{}, err
	}
	stmt, consumed, err := sqlparser.ParseOneWithOptions(ctx, text, sqlparser.ParserOptions{})
	if err != nil {
		return intrinsicSyntax{}, fmt.Errorf("%w: DDL parser: %v", ErrInvalidStore, err)
	}
	if err = ctx.Err(); err != nil {
		return intrinsicSyntax{}, err
	}
	if consumed < 0 || consumed > len(text) || strings.TrimSpace(text[consumed:]) != "" {
		return intrinsicSyntax{}, fmt.Errorf("%w: unconsumed DDL", ErrInvalidStore)
	}
	if err = boundIntrinsicAST(ctx, stmt); err != nil {
		return intrinsicSyntax{}, err
	}
	ddl, ok := stmt.(*sqlparser.DDL)
	if !ok || ddl.Action != sqlparser.CreateStr || ddl.Temporary || ddl.SpecialCommentMode ||
		ddl.OptSelect != nil || ddl.OptLike != nil || ddl.PartitionSpec != nil ||
		ddl.TriggerSpec != nil || ddl.ProcedureSpec != nil || ddl.EventSpec != nil ||
		ddl.Authentication != nil || ddl.AccountLimits != nil || ddl.AutoIncSpec != nil ||
		ddl.IndexSpec != nil || ddl.DefaultSpec != nil || ddl.ColumnOrder != nil ||
		ddl.AlterCollationSpec != nil || ddl.AlterCommentSpec != nil || ddl.NotNullSpec != nil ||
		ddl.ColumnTypeSpec != nil || ddl.IfExists || ddl.ConstraintIfExists || ddl.ColumnAction != "" || (ddl.ConstraintAction != "" && !(kind == "BASE TABLE" && ddl.ConstraintAction == sqlparser.AddStr && ddl.TableSpec != nil && len(ddl.TableSpec.Constraints) > 0)) {
		return intrinsicSyntax{}, fmt.Errorf("%w: %s %s unsupported DDL shape", ErrInvalidStore, name, kind)
	}
	out := intrinsicSyntax{name: name, kind: kind, tokens: tokens}
	switch kind {
	case "BASE TABLE":
		if !intrinsicUnqualifiedSQLName(ddl.Table, name) || ddl.ViewSpec != nil || ddl.OrReplace ||
			ddl.TableSpec == nil || ddl.TableSpec.PartitionOpt != nil || len(ddl.TableSpec.Columns) == 0 ||
			len(ddl.TableSpec.Columns) > 128 || len(ddl.TableSpec.Indexes) > 256 || len(ddl.TableSpec.Constraints) > 256 {
			return intrinsicSyntax{}, fmt.Errorf("%w: table DDL shape", ErrInvalidStore)
		}
		seen := make(map[string]bool)
		for _, col := range ddl.TableSpec.Columns {
			if col == nil || col.Name.IsEmpty() || seen[col.Name.String()] || col.Type.GeneratedExpr != nil || col.Type.SRID != nil || col.Type.ResolvedType != nil {
				return intrinsicSyntax{}, fmt.Errorf("%w: table column", ErrInvalidStore)
			}
			seen[col.Name.String()] = true
			out.columns = append(out.columns, col.Name.String())
		}
		keys := make(map[string]bool)
		for _, index := range ddl.TableSpec.Indexes {
			if index == nil || index.Info == nil || index.Info.Spatial || index.Info.Fulltext || index.Info.Vector || len(index.Options) != 0 || len(index.Fields) == 0 {
				return intrinsicSyntax{}, fmt.Errorf("%w: table index", ErrInvalidStore)
			}
			key := index.Info.Name.String()
			if index.Info.Primary {
				key = "PRIMARY"
			}
			if key == "" || keys[key] {
				return intrinsicSyntax{}, fmt.Errorf("%w: duplicate/unnamed index", ErrInvalidStore)
			}
			keys[key] = true
			for _, field := range index.Fields {
				if field == nil || field.Expression != nil || !seen[field.Column.String()] {
					return intrinsicSyntax{}, fmt.Errorf("%w: functional/unknown index column", ErrInvalidStore)
				}
			}
		}
	case "VIEW":
		v := ddl.ViewSpec
		if v == nil || ddl.TableSpec != nil || !intrinsicUnqualifiedSQLName(v.ViewName, name) || v.ViewExpr == nil || len(v.Columns) != 0 || v.Definer != "" || v.Algorithm != "" || v.Security != "" || v.CheckOption != sqlparser.ViewCheckOptionUnspecified {
			return intrinsicSyntax{}, fmt.Errorf("%w: view DDL shape", ErrInvalidStore)
		}
	default:
		return intrinsicSyntax{}, fmt.Errorf("%w: DDL kind", ErrInvalidStore)
	}
	// No general AST equivalence/formatter normalization is inferred here. Full
	// native descriptor correspondence remains unqualified in the fixed manifest.
	return out, nil
}

// Check both qualification fields directly: the pinned MySQL formatter and
// TableName.Walk omit SchemaQualifier. Neither field may qualify a local name.
func intrinsicUnqualifiedSQLName(table sqlparser.TableName, name string) bool {
	return table.Name.String() == name && table.DbQualifier.IsEmpty() && table.SchemaQualifier.IsEmpty()
}

func intrinsicSQLModes(raw string) error {
	want := map[string]bool{"NO_ENGINE_SUBSTITUTION": false, "ONLY_FULL_GROUP_BY": false, "STRICT_TRANS_TABLES": false}
	for _, part := range strings.Split(raw, ",") {
		part = strings.ToUpper(strings.TrimSpace(part))
		seen, ok := want[part]
		if !ok || seen {
			return fmt.Errorf("%w: unsupported/duplicate SQL mode", ErrInvalidStore)
		}
		want[part] = true
	}
	for _, seen := range want {
		if !seen {
			return fmt.Errorf("%w: missing SQL mode", ErrInvalidStore)
		}
	}
	return nil
}

func intrinsicIgnoreBoolean(raw any) (bool, error) {
	switch v := raw.(type) {
	case bool:
		return v, nil
	case int8, int64:
		// The pinned native Boolean is SQL Int8; database/sql mocks supply
		// Int64. Accept only exact 0/1, never general truthiness or floats.
		n, err := intrinsicUnsigned(v)
		if err == nil && n <= 1 {
			return n == 1, nil
		}
	case []byte:
		if string(v) == "0" || string(v) == "1" {
			return string(v) == "1", nil
		}
	case string:
		if v == "0" || v == "1" {
			return v == "1", nil
		}
	}
	return false, fmt.Errorf("%w: ignore boolean encoding", ErrInvalidStore)
}

// This traversal is iterative and bounded AFTER native parser allocation. It
// counts structural values, including fields omitted by SQLNode.Walk; it is not
// a semantic comparator or a claim about parser allocation/CPU preemption.
func boundIntrinsicAST(ctx context.Context, root any) error {
	type frame struct {
		value reflect.Value
		depth int
	}
	stack := []frame{{value: reflect.ValueOf(root)}}
	count := 0
	for len(stack) != 0 {
		if err := ctx.Err(); err != nil {
			return err
		}
		f := stack[len(stack)-1]
		stack = stack[:len(stack)-1]
		if !f.value.IsValid() {
			continue
		}
		count++
		if count > intrinsicTokenLimit || f.depth > intrinsicNestingLimit {
			return fmt.Errorf("%w: DDL AST traversal", ErrLimitExceeded)
		}
		push := func(v reflect.Value) { stack = append(stack, frame{value: v, depth: f.depth + 1}) }
		switch f.value.Kind() {
		case reflect.Interface, reflect.Pointer:
			if !f.value.IsNil() {
				push(f.value.Elem())
			}
		case reflect.Struct:
			for i := f.value.NumField() - 1; i >= 0; i-- {
				push(f.value.Field(i))
			}
		case reflect.Slice, reflect.Array:
			if f.value.Len() > intrinsicTokenLimit-count {
				return fmt.Errorf("%w: DDL AST sequence", ErrLimitExceeded)
			}
			for i := f.value.Len() - 1; i >= 0; i-- {
				push(f.value.Index(i))
			}
		case reflect.Map:
			// No selected source DDL needs dynamic Auth.Extra or other maps.
			if f.value.Len() != 0 {
				return fmt.Errorf("%w: DDL AST map", ErrInvalidStore)
			}
		case reflect.Func, reflect.Chan, reflect.UnsafePointer:
			return fmt.Errorf("%w: DDL AST value", ErrInvalidStore)
		}
	}
	return nil
}
