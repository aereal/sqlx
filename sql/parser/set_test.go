package parser_test

import (
	"reflect"
	"testing"

	"github.com/aereal/sqlx/sql/ast"
	"github.com/aereal/sqlx/sql/parser"
	"github.com/aereal/sqlx/sql/tokenizer"
)

func TestParse_SetStatement(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name           string
		sql            string
		wantLifetime   ast.SetStatementLifetime
		wantParamName  string
		wantParamValue *ast.LiteralValue
		shouldErr      bool
	}{
		{
			name:          "SET with string literal",
			sql:           "SET client_encoding = 'UTF8'",
			wantLifetime:  ast.SetStatementLifetimeSession,
			wantParamName: "client_encoding",
			wantParamValue: &ast.LiteralValue{
				Type:  "string",
				Value: "UTF8",
			},
			shouldErr: false,
		},
		{
			name:          "SET with boolean keyword",
			sql:           "SET standard_conforming_strings = on",
			wantLifetime:  ast.SetStatementLifetimeSession,
			wantParamName: "standard_conforming_strings",
			wantParamValue: &ast.LiteralValue{
				Type:  "bool",
				Value: "on",
			},
			shouldErr: false,
		},
		{
			name:          "set local parameter",
			sql:           "SET LOCAL standard_conforming_strings = on",
			wantLifetime:  ast.SetStatementLifetimeLocal,
			wantParamName: "standard_conforming_strings",
			wantParamValue: &ast.LiteralValue{
				Type:  "bool",
				Value: "on",
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

			p := parser.NewParser()
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
			got, ok := stmt.(*ast.SetStatement)
			if !ok {
				t.Fatalf("expected *ast.SetStatement but got %T (%#v)", stmt, stmt)
			}
			if got.ParameterName != tc.wantParamName {
				t.Errorf("ParameterName: want=%q got=%q", tc.wantParamName, got.ParameterName)
			}
			if got.Lifetime != tc.wantLifetime {
				t.Errorf("Lifetime: want=%s got=%s", tc.wantLifetime, got.Lifetime)
			}
			gotLit, ok := got.ParameterValue.(*ast.LiteralValue)
			if !ok {
				t.Fatalf("ParameterValue is not a ast.LiteralValue; got %T", got.ParameterValue)
			}
			if gotLit.Type != tc.wantParamValue.Type {
				t.Errorf("ParamterValue.Type: want=%q got=%q", tc.wantParamValue.Type, gotLit.Type)
			}
			if !reflect.DeepEqual(gotLit.Value, tc.wantParamValue.Value) {
				t.Errorf("ParamterValue.Value: want=%#v got=%#v", tc.wantParamValue.Value, gotLit.Value)
			}
		})
	}
}
