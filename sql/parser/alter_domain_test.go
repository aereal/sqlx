package parser_test

import (
	"testing"

	"github.com/aereal/sqlx/models"
	"github.com/aereal/sqlx/sql/ast"
	"github.com/aereal/sqlx/sql/dialect"
	"github.com/aereal/sqlx/sql/parser"
	"github.com/aereal/sqlx/sql/tokenizer"
)

func TestParse_AlterDomainStatement_owner(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name      string
		sql       string
		want      *ast.AlterDomainOwnerToStatement
		shouldErr bool
	}{
		{
			name: "current_user",
			sql:  "alter domain hoge owner to current_user",
			want: &ast.AlterDomainOwnerToStatement{
				Name:     &ast.Identifier{Name: "hoge"},
				UserName: "current_user",
				Start:    models.Location{Line: 1, Column: 1},
				End:      models.Location{Line: 1, Column: 40},
			},
			shouldErr: false,
		},
		{
			name: "specify user",
			sql:  "alter domain hoge owner to app_user",
			want: &ast.AlterDomainOwnerToStatement{
				Name:     &ast.Identifier{Name: "hoge"},
				UserName: "app_user",
				Start:    models.Location{Line: 1, Column: 1},
				End:      models.Location{Line: 1, Column: 36},
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
			got, ok := stmt.(*ast.AlterDomainOwnerToStatement)
			if !ok {
				t.Fatalf("expected *ast.AlterDomainiOwnerToStatement but got %T (%#v)", stmt, stmt)
			}
			if tc.want.Name.Name != got.Name.Name {
				t.Errorf("Name.Name: want=%q got=%q", tc.want.Name.Name, got.Name.Name)
			}
			if tc.want.UserName != got.UserName {
				t.Errorf("UserName: want=%q got=%q", tc.want.UserName, got.UserName)
			}
			assertEqualSpan(t, tc.want.Span(), got.Span())
		})
	}
}
