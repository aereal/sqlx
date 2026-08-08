package parser_test

import (
	"testing"

	"github.com/aereal/sqlx/models"
	"github.com/aereal/sqlx/sql/ast"
	"github.com/aereal/sqlx/sql/dialect"
	"github.com/aereal/sqlx/sql/parser"
	"github.com/aereal/sqlx/sql/tokenizer"
)

func TestParse_AlterTypeStatement_owner(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name      string
		sql       string
		want      *ast.AlterTypeOwnerToStatement
		shouldErr bool
	}{
		{
			name: "current_user",
			sql:  "alter type hoge owner to current_user",
			want: &ast.AlterTypeOwnerToStatement{
				Name:     &ast.Identifier{Name: "hoge"},
				UserName: "current_user",
				Start:    models.Location{Line: 1, Column: 1},
				End:      models.Location{Line: 1, Column: 38},
			},
			shouldErr: false,
		},
		{
			name: "specify user",
			sql:  "alter type hoge owner to app_user",
			want: &ast.AlterTypeOwnerToStatement{
				Name:     &ast.Identifier{Name: "hoge"},
				UserName: "app_user",
				Start:    models.Location{Line: 1, Column: 1},
				End:      models.Location{Line: 1, Column: 34},
			},
			shouldErr: false,
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

			if tc.shouldErr {
				if err == nil {
					t.Error("Expected error, got nil")
				}
				return
			}

			if err != nil {
				t.Fatalf("Expected no error, got: %v", err)
			}
			if len(tree.Statements) != 1 {
				t.Fatalf("expected only 1 statement but got %d statement(s)", len(tree.Statements))
			}
			stmt := tree.Statements[0]
			got, ok := stmt.(*ast.AlterTypeOwnerToStatement)
			if !ok {
				t.Fatalf("expected *ast.SetStatement but got %T (%#v)", stmt, stmt)
			}
			if tc.want.TypeName().Name != got.TypeName().Name {
				t.Errorf("TypeName.Name: want=%q got=%q", tc.want.TypeName().Name, got.TypeName().Name)
			}
			if tc.want.UserName != got.UserName {
				t.Errorf("UserName: want=%q got=%q", tc.want.UserName, got.UserName)
			}
			assertEqualSpan(t, tc.want.Span(), got.Span())
		})
	}
}

func assertEqualSpan(t *testing.T, want, got models.Span) {
	t.Helper()

	assertEqualLocation(t, "Start", want.Start, got.Start)
	assertEqualLocation(t, "End", want.End, got.End)
}

func assertEqualLocation(t *testing.T, label string, want, got models.Location) {
	t.Helper()

	if want.Line != got.Line {
		t.Errorf("%s/Line: want=%d got=%d", label, want.Line, got.Line)
	}
	if want.Column != got.Column {
		t.Errorf("%s/Column: want=%d got=%d", label, want.Column, got.Column)
	}
}
