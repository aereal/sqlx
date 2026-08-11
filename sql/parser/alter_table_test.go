package parser_test

import (
	"testing"

	"github.com/aereal/sqlx/sql/ast"
	"github.com/aereal/sqlx/sql/dialect"
	"github.com/aereal/sqlx/sql/keywords"
	"github.com/aereal/sqlx/sql/parser"
	"github.com/aereal/sqlx/sql/tokenizer"
)

func TestParse_AlterTableStatement_owner(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name string
		sql  string
		want *ast.AlterStatement
	}{
		{
			name: "current_user",
			sql:  "ALTER TABLE public.users OWNER TO current_user",
			want: &ast.AlterStatement{
				Name: "public.users",
				Type: ast.AlterTypeTable,
				Operation: &ast.AlterTableOperation{
					Type:     ast.OwnerTo,
					UserName: "current_user",
				},
			},
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			tkz := tokenizer.GetTokenizer()
			t.Cleanup(func() { tokenizer.PutTokenizer(tkz) })

			tokens, err := tkz.Tokenize([]byte(tc.sql))
			if err != nil {
				t.Fatalf("tokenization failed: %v", err)
			}

			p := parser.NewParser(parser.WithDialect(dialect.PostgreSQL.String()))
			t.Cleanup(func() { p.Release() })

			tree, err := p.ParseContextFromModelTokens(t.Context(), tokens)

			if err != nil {
				t.Fatalf("Expected no error, got: %v", err)
			}
			if len(tree.Statements) != 1 {
				t.Fatalf("expected only 1 statement but got %d statement(s)", len(tree.Statements))
			}
			stmt := tree.Statements[0]
			alterStmt, ok := tree.Statements[0].(*ast.AlterStatement)
			if !ok {
				t.Fatalf("expected *ast.SetStatement but got %T (%#v)", stmt, stmt)
			}
			if alterStmt.Name != tc.want.Name {
				t.Errorf("Name: want=%q got=%q", tc.want.Name, alterStmt.Name)
			}
			if alterStmt.Type != tc.want.Type {
				t.Errorf("Type: want=%d got=%d", tc.want.Type, alterStmt.Type)
			}
			gotOp, ok := alterStmt.Operation.(*ast.AlterTableOperation)
			if !ok {
				t.Fatalf("expected ast.AlterTableOperation but got %T", alterStmt.Operation)
			}
			wantOp, ok := tc.want.Operation.(*ast.AlterTableOperation)
			if !ok {
				t.Fatalf("expected ast.AlterTableOperation but got %T", tc.want.Operation)
			}
			if gotOp.Type != wantOp.Type {
				t.Errorf("Operation.Type: want=%d got=%d", wantOp.Type, gotOp.Type)
			}
			if gotOp.UserName != wantOp.UserName {
				t.Errorf("Operation.UserName: want=%s got=%s", wantOp.UserName, gotOp.UserName)
			}
		})
	}
}

func TestParse_AlterTableStatement_postgresql(t *testing.T) {
	t.Parallel()

	d := keywords.DialectPostgreSQL

	tkz, err := tokenizer.NewWithDialect(d)
	if err != nil {
		t.Fatal(err)
	}
	tokens, err := tkz.TokenizeContext(t.Context(), []byte("ALTER TABLE ONLY public.users ALTER COLUMN id SET NULL"))
	if err != nil {
		t.Fatal(err)
	}
	tree, err := parser.NewParser(parser.WithDialect(string(d))).ParseContextFromModelTokens(t.Context(), tokens)
	if err != nil {
		t.Fatal(err)
	}
	if len(tree.Statements) != 1 {
		t.Fatalf("expected 1 statement but got %d statement(s)", len(tree.Statements))
	}
	stmt, ok := tree.Statements[0].(*ast.AlterStatement)
	if !ok {
		t.Fatalf("expected *ast.AlterStatement but got %T", tree.Statements[0])
	}
	if stmt.Type != ast.AlterTypeTable {
		t.Errorf("expected AlterTable but got %s", stmt.Type)
	}
	if _, ok := stmt.Operation.(*ast.AlterTableOperation); !ok {
		t.Errorf("Operation: expected *ast.AlterTableOperation but got %T", stmt.Operation)
	}
}

func TestParse_AlterTableStatement_postgresql_default(t *testing.T) {
	t.Parallel()

	d := keywords.DialectPostgreSQL

	tkz, err := tokenizer.NewWithDialect(d)
	if err != nil {
		t.Fatal(err)
	}
	tokens, err := tkz.TokenizeContext(t.Context(), []byte("ALTER TABLE ONLY public.users ALTER COLUMN id SET DEFAULT nextval('public.users_id_seq'::regclass)"))
	if err != nil {
		t.Fatal(err)
	}
	tree, err := parser.NewParser(parser.WithDialect(string(d))).ParseContextFromModelTokens(t.Context(), tokens)
	if err != nil {
		t.Fatal(err)
	}
	if len(tree.Statements) != 1 {
		t.Fatalf("expected 1 statement but got %d statement(s)", len(tree.Statements))
	}
	stmt, ok := tree.Statements[0].(*ast.AlterStatement)
	if !ok {
		t.Fatalf("expected *ast.AlterStatement but got %T", tree.Statements[0])
	}
	if stmt.Type != ast.AlterTypeTable {
		t.Errorf("expected AlterTable but got %s", stmt.Type)
	}
	if _, ok := stmt.Operation.(*ast.AlterTableOperation); !ok {
		t.Errorf("Operation: expected *ast.AlterTableOperation but got %T", stmt.Operation)
	}
}
