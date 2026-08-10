package parser_test

import (
	"testing"

	"github.com/aereal/sqlx/sql/ast"
	"github.com/aereal/sqlx/sql/keywords"
	"github.com/aereal/sqlx/sql/parser"
)

func TestAlterSequenceStatement(t *testing.T) {
	t.Parallel()

	dialects := []keywords.SQLDialect{keywords.DialectMariaDB, keywords.DialectPostgreSQL}

	testCases := []struct {
		name            string
		sql             string
		wantRestartWith *ast.LiteralValue
		wantRestart     bool
	}{
		{
			name:            "restart with",
			sql:             "ALTER SEQUENCE seq_orders RESTART WITH 500",
			wantRestart:     false,
			wantRestartWith: &ast.LiteralValue{Type: "int", Value: "500"},
		},
		{
			name:            "restart",
			sql:             "ALTER SEQUENCE seq_orders RESTART",
			wantRestart:     true,
			wantRestartWith: nil,
		},
		{
			name:            "qualified sequence name",
			sql:             "ALTER SEQUENCE public.seq_orders RESTART",
			wantRestart:     true,
			wantRestartWith: nil,
		},
	}
	for _, tc := range testCases {
		for _, d := range dialects {
			t.Run(tc.name+"/"+string(d), func(t *testing.T) {
				t.Parallel()

				tree, err := parser.ParseWithDialect(tc.sql, d)
				if err != nil {
					t.Fatal(err)
				}
				stmt, ok := tree.Statements[0].(*ast.AlterSequenceStatement)
				if !ok {
					t.Fatalf("expected AlterSequenceStatement, got %T", tree.Statements[0])
				}
				if stmt.Options.Restart != tc.wantRestart {
					t.Errorf("Options.Restart: want=%v got=%v", tc.wantRestart, stmt.Options.Restart)
				}
				gotRestartWith := stmt.Options.RestartWith
				if tc.wantRestartWith == nil {
					if gotRestartWith != nil {
						t.Errorf("Options.RestartWith: expected no RestartWith but got: %#v", gotRestartWith)
					}
				} else { // tc.wantRestartWith != nil
					if gotRestartWith == nil {
						t.Errorf("Options.RestartWith: expected RestartWith (%#v) but got nothing", tc.wantRestartWith)
					}
				}
			})
		}
	}
}

func TestAlterSequenceStatement_ownerTo(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name          string
		sql           string
		dialect       keywords.SQLDialect
		wantOwnerName string
		shouldErr     bool
	}{
		{
			name:          "MariaDB/owner to",
			sql:           "ALTER SEQUENCE public.seq_orders OWNER TO app",
			dialect:       keywords.DialectMariaDB,
			wantOwnerName: "",
			shouldErr:     true,
		},
		{
			name:          "PostgreSQL/owner to",
			sql:           "ALTER SEQUENCE public.seq_orders OWNER TO app",
			dialect:       keywords.DialectPostgreSQL,
			wantOwnerName: "app",
			shouldErr:     false,
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			tree, err := parser.ParseWithDialect(tc.sql, tc.dialect)
			if tc.shouldErr {
				if err == nil {
					t.Errorf("expected some error but got nothing")
				}
			} else {
				if err != nil {
					t.Fatalf("expected no error but got %s", err)
				}
			}
			if err != nil {
				return
			}

			stmt, ok := tree.Statements[0].(*ast.AlterSequenceStatement)
			if !ok {
				t.Fatalf("expected AlterSequenceStatement, got %T", tree.Statements[0])
			}
			if stmt.Options.OwnerName != tc.wantOwnerName {
				t.Errorf("Options.OwnerName: want=%v got=%v", tc.wantOwnerName, stmt.Options.OwnerName)
			}
		})
	}
}
