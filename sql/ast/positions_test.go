// Copyright 2026 GoSQLX Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

// Package ast positions_test.go verifies that AST nodes carry accurate source
// position information after parsing with ParseFromModelTokensWithPositions.
//
// Issue #324: Most AST nodes didn't carry position data, so LSP/linter errors
// reported "line 0, column 0". This test suite validates the fix.
package ast_test

import (
	"testing"

	"github.com/aereal/sqlx/models"
	"github.com/aereal/sqlx/sql/ast"
	"github.com/aereal/sqlx/sql/parser"
	"github.com/aereal/sqlx/sql/tokenizer"
)

// parseWithPositions is a test helper that tokenizes and parses SQL with
// full position tracking enabled, returning the AST ready for inspection.
func parseWithPositions(t *testing.T, sql string) *ast.AST {
	t.Helper()

	tkz := tokenizer.GetTokenizer()
	defer tokenizer.PutTokenizer(tkz)

	tokens, err := tkz.Tokenize([]byte(sql))
	if err != nil {
		t.Fatalf("tokenize failed: %v", err)
	}

	p := parser.GetParser()
	defer parser.PutParser(p)

	result, err := p.ParseFromModelTokens(tokens)
	if err != nil {
		t.Fatalf("parse failed: %v", err)
	}
	return result
}

// assertPos checks that a Location is non-zero (has been populated).
func assertPos(t *testing.T, label string, loc models.Location) {
	t.Helper()
	if loc.Line == 0 && loc.Column == 0 {
		t.Errorf("%s: expected non-zero position, got line=%d col=%d", label, loc.Line, loc.Column)
	}
}

// assertPosEqual checks that a Location matches expected values.
func assertPosEqual(t *testing.T, label string, loc models.Location, wantLine, wantCol int) {
	t.Helper()
	if loc.Line != wantLine || loc.Column != wantCol {
		t.Errorf("%s: expected line=%d col=%d, got line=%d col=%d",
			label, wantLine, wantCol, loc.Line, loc.Column)
	}
}

// -----------------------------------------------------------------------------
// TestSelectStatementPosition verifies SELECT statement position
// -----------------------------------------------------------------------------

func TestSelectStatementPosition(t *testing.T) {
	tree := parseWithPositions(t, "SELECT id, name FROM users")

	if len(tree.Statements) != 1 {
		t.Fatalf("expected 1 statement, got %d", len(tree.Statements))
	}

	sel, ok := tree.Statements[0].(*ast.SelectStatement)
	if !ok {
		t.Fatalf("expected *ast.SelectStatement, got %T", tree.Statements[0])
	}

	assertPosEqual(t, "SELECT.Start", sel.Start, 1, 1)
	assertPosEqual(t, "SELECT.End", sel.End, 1, 27)
	assertPosEqual(t, "FROM.Start", sel.From[0].Start, 1, 22)
	assertPosEqual(t, "FROM.End", sel.From[0].End, 1, 27)
}

func TestSelectStatementPositionMultiLine(t *testing.T) {
	// Two statements on separate lines: verify positions differ
	sql := "SELECT 1;\nSELECT 2"
	tree := parseWithPositions(t, sql)

	if len(tree.Statements) != 2 {
		t.Fatalf("expected 2 statements, got %d", len(tree.Statements))
	}

	sel1, ok := tree.Statements[0].(*ast.SelectStatement)
	if !ok {
		t.Fatalf("expected *ast.SelectStatement for stmt 1, got %T", tree.Statements[0])
	}
	sel2, ok := tree.Statements[1].(*ast.SelectStatement)
	if !ok {
		t.Fatalf("expected *ast.SelectStatement for stmt 2, got %T", tree.Statements[1])
	}

	// Both must have non-zero positions
	assertPos(t, "SELECT1.Start", sel1.Start)
	assertPos(t, "SELECT2.Start", sel2.Start)
	assertPos(t, "SELECT1.End", sel1.End)
	assertPos(t, "SELECT2.End", sel2.End)

	// The second SELECT should be on a later line than the first
	if sel2.Start.Line <= sel1.Start.Line {
		t.Errorf("second SELECT.Pos.Line (%d) should be > first (%d)",
			sel2.Start.Line, sel1.Start.Line)
	}
}

// -----------------------------------------------------------------------------
// TestInsertStatementPosition verifies INSERT statement position
// -----------------------------------------------------------------------------

func TestInsertStatementPosition(t *testing.T) {
	tree := parseWithPositions(t, "INSERT INTO users (id, name) VALUES (1, 'Alice')")

	if len(tree.Statements) != 1 {
		t.Fatalf("expected 1 statement, got %d", len(tree.Statements))
	}

	ins, ok := tree.Statements[0].(*ast.InsertStatement)
	if !ok {
		t.Fatalf("expected *ast.InsertStatement, got %T", tree.Statements[0])
	}

	assertPosEqual(t, "INSERT.Start", ins.Start, 1, 1)
	assertPosEqual(t, "INSERT.End", ins.End, 1, 49)
	assertPosEqual(t, "INSERT table.Start", ins.Table.Start, 1, 13)
	assertPosEqual(t, "INSERT table.End", ins.Table.End, 1, 19)
}

// -----------------------------------------------------------------------------
// TestUpdateStatementPosition verifies UPDATE statement position
// -----------------------------------------------------------------------------

func TestUpdateStatementPosition(t *testing.T) {
	tree := parseWithPositions(t, "UPDATE users SET name = 'Bob' WHERE id = 1")

	if len(tree.Statements) != 1 {
		t.Fatalf("expected 1 statement, got %d", len(tree.Statements))
	}

	upd, ok := tree.Statements[0].(*ast.UpdateStatement)
	if !ok {
		t.Fatalf("expected *ast.UpdateStatement, got %T", tree.Statements[0])
	}

	assertPosEqual(t, "UPDATE.Start", upd.Start, 1, 1)
	assertPosEqual(t, "UPDATE.End", upd.End, 1, 43)
	assertPosEqual(t, "table.Start", upd.Table.Start, 1, 8)
	assertPosEqual(t, "table.End", upd.Table.End, 1, 14)
}

// -----------------------------------------------------------------------------
// TestDeleteStatementPosition verifies DELETE statement position
// -----------------------------------------------------------------------------

func TestDeleteStatementPosition(t *testing.T) {
	tree := parseWithPositions(t, "DELETE FROM users WHERE id = 1")

	if len(tree.Statements) != 1 {
		t.Fatalf("expected 1 statement, got %d", len(tree.Statements))
	}

	del, ok := tree.Statements[0].(*ast.DeleteStatement)
	if !ok {
		t.Fatalf("expected *ast.DeleteStatement, got %T", tree.Statements[0])
	}

	assertPosEqual(t, "DELETE.Start", del.Start, 1, 1)
	assertPosEqual(t, "DELETE.End", del.End, 1, 31)
}

func TestCreateTableStatementPosition(t *testing.T) {
	tree := parseWithPositions(t, "create table public.users (\n  id integer primary key,\n  name varchar(255) not null\n)")
	if len(tree.Statements) != 1 {
		t.Fatalf("expected 1 statement, got %d", len(tree.Statements))
	}
	stmt, ok := tree.Statements[0].(*ast.CreateTableStatement)
	if !ok {
		t.Fatalf("expected *ast.CreateTableStatement, got %T", tree.Statements[0])
	}
	assertPosEqual(t, "CREATE TABLE start", stmt.Start, 1, 1)
	assertPosEqual(t, "CREATE TABLE end", stmt.End, 4, 2)

	assertPosEqual(t, "CREATE TABLE column 1 start", stmt.Columns[0].Start, 2, 3)
	assertPosEqual(t, "CREATE TABLE column 1 end", stmt.Columns[0].End, 2, 25)
	assertPosEqual(t, "CREATE TABLE column 2 start", stmt.Columns[1].Start, 3, 3)
	assertPosEqual(t, "CREATE TABLE column 2 end", stmt.Columns[1].End, 4, 1)
}

func TestCreateDomainStatementPosition(t *testing.T) {
	t.Run("check", func(t *testing.T) {
		tree := parseWithPositions(t, "create domain unsigned_smallint smallint\n  constraint unsigned_smallint_check\n  check ((value >= 0))")
		if len(tree.Statements) != 1 {
			t.Fatalf("expected 1 statement, got %d", len(tree.Statements))
		}
		stmt, ok := tree.Statements[0].(*ast.CreateDomainStatement)
		if !ok {
			t.Fatalf("expected *ast.CreateDomainStatement, got %T", tree.Statements[0])
		}
		assertPosEqual(t, "CREATE DOMAIN start", stmt.Start, 1, 1)
		assertPosEqual(t, "CREATE DOMAIN end", stmt.End, 3, 23)
		checkConstraint, ok := stmt.Constraint.(*ast.DomainConstraintCheck)
		if !ok {
			t.Fatalf("expected *ast.DomainConstraintCheck, got %T", stmt.Constraint)
		}
		if stmt.Name.Name != "unsigned_smallint" {
			t.Errorf("got domain name: %q", stmt.Name.Name)
		}
		assertPosEqual(t, "constraint start", checkConstraint.Start, 2, 3)
		assertPosEqual(t, "constraint end", checkConstraint.End, 3, 23)
	})

	t.Run("not null", func(t *testing.T) {
		tree := parseWithPositions(t, "create domain public.non_empty_string as varchar(255)\n  collate \"ja_JP.UTF-8\"\n  not null")
		if len(tree.Statements) != 1 {
			t.Fatalf("expected 1 statement, got %d", len(tree.Statements))
		}
		stmt, ok := tree.Statements[0].(*ast.CreateDomainStatement)
		if !ok {
			t.Fatalf("expected *ast.CreateDomainStatement, got %T", tree.Statements[0])
		}
		assertPosEqual(t, "CREATE DOMAIN start", stmt.Start, 1, 1)
		assertPosEqual(t, "CREATE DOMAIN end", stmt.End, 3, 11)
		if stmt.Name.Name != "public.non_empty_string" {
			t.Errorf("got domain name: %q", stmt.Name.Name)
		}

		checkConstraint, ok := stmt.Constraint.(*ast.DomainConstraintNotNull)
		if !ok {
			t.Fatalf("expected *ast.DomainConstraintNotNull, got %T", stmt.Constraint)
		}
		assertPosEqual(t, "constraint start", checkConstraint.Start, 3, 3)
		assertPosEqual(t, "constraint end", checkConstraint.End, 3, 11)
	})
}

func TestCreateTypeStatementPosition(t *testing.T) {
	t.Run("enum", func(t *testing.T) {
		tree := parseWithPositions(t, "create type public.user_role as enum ('admin', 'ordinary')")
		if len(tree.Statements) != 1 {
			t.Fatalf("expected 1 statement, got %d", len(tree.Statements))
		}
		stmt, ok := tree.Statements[0].(ast.CreateTypeStatement)
		if !ok {
			t.Fatalf("expected *ast.CreateTypeStatement, got %T", tree.Statements[0])
		}
		assertPosEqual(t, "CREATE TYPE start", stmt.Span().Start, 1, 1)
		assertPosEqual(t, "CREATE TYPE end", stmt.Span().End, 1, 59)
		if stmt.TypeName().Name != "public.user_role" {
			t.Errorf("got type name: %q", stmt.TypeName().Name)
		}
	})

	t.Run("composite", func(t *testing.T) {
		tree := parseWithPositions(t, "create type loc as (x integer, y integer)")
		if len(tree.Statements) != 1 {
			t.Fatalf("expected 1 statement, got %d", len(tree.Statements))
		}
		stmt, ok := tree.Statements[0].(ast.CreateTypeStatement)
		if !ok {
			t.Fatalf("expected *ast.CreateTypeStatement, got %T", tree.Statements[0])
		}
		assertPosEqual(t, "CREATE TYPE start", stmt.Span().Start, 1, 1)
		assertPosEqual(t, "CREATE TYPE end", stmt.Span().End, 1, 42)
		if stmt.TypeName().Name != "loc" {
			t.Errorf("got type name: %q", stmt.TypeName().Name)
		}
	})
}

// -----------------------------------------------------------------------------
// TestIdentifierPosition verifies Identifier nodes carry positions
// -----------------------------------------------------------------------------

func TestIdentifierPosition(t *testing.T) {
	tree := parseWithPositions(t, "SELECT id FROM users")

	sel := tree.Statements[0].(*ast.SelectStatement)

	// The first column in SELECT should be an Identifier
	if len(sel.Columns) == 0 {
		t.Fatal("expected at least 1 column")
	}

	ident, ok := sel.Columns[0].(*ast.Identifier)
	if !ok {
		t.Fatalf("expected *ast.Identifier, got %T", sel.Columns[0])
	}

	assertPos(t, "Identifier(id).Start", ident.Start)
	// "id" starts at column 8 in "SELECT id FROM users" (SELECT is 6 chars + space = 7)
	assertPosEqual(t, "Identifier(id).Start", ident.Start, 1, 8)
	assertPosEqual(t, "Identifier(id).End", ident.End, 1, 11)
}

func TestQualifiedIdentifierPosition(t *testing.T) {
	tree := parseWithPositions(t, "SELECT u.id FROM users u")

	sel := tree.Statements[0].(*ast.SelectStatement)

	if len(sel.Columns) == 0 {
		t.Fatal("expected at least 1 column")
	}

	ident, ok := sel.Columns[0].(*ast.Identifier)
	if !ok {
		t.Fatalf("expected *ast.Identifier, got %T", sel.Columns[0])
	}

	// Qualified identifier should have position of the table qualifier
	assertPos(t, "Identifier(u.id).Start", ident.Start)
	if ident.Name != "id" || ident.Table != "u" {
		t.Errorf("expected u.id, got %s.%s", ident.Table, ident.Name)
	}
}

// -----------------------------------------------------------------------------
// TestFunctionCallPosition verifies FunctionCall nodes carry positions
// -----------------------------------------------------------------------------

func TestFunctionCallPosition(t *testing.T) {
	tree := parseWithPositions(t, "SELECT COUNT(*) FROM users")

	sel := tree.Statements[0].(*ast.SelectStatement)

	if len(sel.Columns) == 0 {
		t.Fatal("expected at least 1 column")
	}

	fn, ok := sel.Columns[0].(*ast.FunctionCall)
	if !ok {
		t.Fatalf("expected *ast.FunctionCall, got %T", sel.Columns[0])
	}

	assertPos(t, "FunctionCall(COUNT).Pos", fn.Pos)
	// COUNT starts at column 8 in "SELECT COUNT(*) FROM users"
	assertPosEqual(t, "FunctionCall(COUNT).Pos", fn.Pos, 1, 8)
}

func TestNestedFunctionCallPosition(t *testing.T) {
	tree := parseWithPositions(t, "SELECT SUM(amount) FROM orders")

	sel := tree.Statements[0].(*ast.SelectStatement)

	fn, ok := sel.Columns[0].(*ast.FunctionCall)
	if !ok {
		t.Fatalf("expected *ast.FunctionCall, got %T", sel.Columns[0])
	}

	assertPos(t, "FunctionCall(SUM).Pos", fn.Pos)
}

// -----------------------------------------------------------------------------
// TestBinaryExpressionPosition verifies BinaryExpression nodes carry positions
// -----------------------------------------------------------------------------

func TestBinaryExpressionPosition(t *testing.T) {
	tree := parseWithPositions(t, "SELECT * FROM users WHERE id = 1")

	sel := tree.Statements[0].(*ast.SelectStatement)
	if sel.Where == nil {
		t.Fatal("expected WHERE clause")
	}

	binExpr, ok := sel.Where.(*ast.BinaryExpression)
	if !ok {
		t.Fatalf("expected *ast.BinaryExpression in WHERE, got %T", sel.Where)
	}

	assertPos(t, "BinaryExpression(=).Pos", binExpr.Pos)
}

func TestANDExpressionPosition(t *testing.T) {
	tree := parseWithPositions(t, "SELECT * FROM users WHERE id = 1 AND active = true")

	sel := tree.Statements[0].(*ast.SelectStatement)
	if sel.Where == nil {
		t.Fatal("expected WHERE clause")
	}

	// The outermost expression should be AND
	andExpr, ok := sel.Where.(*ast.BinaryExpression)
	if !ok {
		t.Fatalf("expected *ast.BinaryExpression for AND, got %T", sel.Where)
	}

	if andExpr.Operator != "AND" {
		t.Errorf("expected AND operator, got %q", andExpr.Operator)
	}

	assertPos(t, "BinaryExpression(AND).Pos", andExpr.Pos)
}

// -----------------------------------------------------------------------------
// TestUnaryExpressionPosition verifies UnaryExpression nodes carry positions
// -----------------------------------------------------------------------------

func TestUnaryExpressionPosition(t *testing.T) {
	tree := parseWithPositions(t, "SELECT * FROM users WHERE NOT active")

	sel := tree.Statements[0].(*ast.SelectStatement)
	if sel.Where == nil {
		t.Fatal("expected WHERE clause")
	}

	unary, ok := sel.Where.(*ast.UnaryExpression)
	if !ok {
		t.Fatalf("expected *ast.UnaryExpression, got %T", sel.Where)
	}

	assertPos(t, "UnaryExpression(NOT).Pos", unary.Pos)
}

// -----------------------------------------------------------------------------
// TestJoinClausePosition verifies JoinClause nodes carry positions
// -----------------------------------------------------------------------------

func TestJoinClausePosition(t *testing.T) {
	tree := parseWithPositions(t, "SELECT u.id, o.total FROM users u JOIN orders o ON u.id = o.user_id")

	sel := tree.Statements[0].(*ast.SelectStatement)
	if len(sel.Joins) == 0 {
		t.Fatal("expected at least 1 JOIN")
	}

	join := sel.Joins[0]
	assertPos(t, "JoinClause.Pos", join.Pos)
}

// -----------------------------------------------------------------------------
// TestMultipleStatementPositions verifies position accuracy in batches
// -----------------------------------------------------------------------------

func TestMultipleStatementPositions(t *testing.T) {
	sql := "SELECT 1;\nINSERT INTO t (x) VALUES (2);\nDELETE FROM t WHERE x = 2"
	tree := parseWithPositions(t, sql)

	if len(tree.Statements) != 3 {
		t.Fatalf("expected 3 statements, got %d", len(tree.Statements))
	}

	// Statement 1: SELECT on line 1
	sel, ok := tree.Statements[0].(*ast.SelectStatement)
	if !ok {
		t.Fatalf("expected *ast.SelectStatement, got %T", tree.Statements[0])
	}
	assertPosEqual(t, "SELECT.Start", sel.Start, 1, 1)
	assertPosEqual(t, "SELECT.End", sel.End, 1, 9)

	// Statement 2: INSERT on line 2
	ins, ok := tree.Statements[1].(*ast.InsertStatement)
	if !ok {
		t.Fatalf("expected *ast.InsertStatement, got %T", tree.Statements[1])
	}
	assertPosEqual(t, "INSERT.Start", ins.Start, 2, 1)
	assertPosEqual(t, "INSERT.End", ins.End, 2, 29)

	// Statement 3: DELETE on line 3
	del, ok := tree.Statements[2].(*ast.DeleteStatement)
	if !ok {
		t.Fatalf("expected *ast.DeleteStatement, got %T", tree.Statements[2])
	}
	assertPosEqual(t, "DELETE.Start", del.Start, 3, 1)
	assertPosEqual(t, "DELETE.End", del.End, 3, 26)
}

// -----------------------------------------------------------------------------
// TestPositionsWithoutPositionTracking verifies that ParseFromModelTokens
// now provides accurate position information (Bug 1 fix: error positions
// were always 0,0 because the position mapping was discarded).
// -----------------------------------------------------------------------------

func TestPositionsWithoutPositionTracking(t *testing.T) {
	// Bug 1 fix: ParseFromModelTokens now uses convertModelTokensWithPositions
	// internally, so positions are always populated - even when the caller
	// does not explicitly call ParseFromModelTokensWithPositions.
	tkz := tokenizer.GetTokenizer()
	defer tokenizer.PutTokenizer(tkz)

	tokens, err := tkz.Tokenize([]byte("SELECT id FROM users"))
	if err != nil {
		t.Fatalf("tokenize failed: %v", err)
	}

	p := parser.GetParser()
	defer parser.PutParser(p)

	tree, err := p.ParseFromModelTokens(tokens)
	if err != nil {
		t.Fatalf("parse failed: %v", err)
	}

	sel := tree.Statements[0].(*ast.SelectStatement)

	// With unified token types, positions are always available from TokenWithSpan spans.
	// ParseFromModelTokens now inherently carries position info.
	if sel.Start.Line == 0 && sel.Start.Column == 0 {
		t.Errorf("expected non-zero position with unified tokens, got line=%d col=%d",
			sel.Start.Line, sel.Start.Column)
	}
}

// -----------------------------------------------------------------------------
// TestArithmeticBinaryExpressionPosition verifies positions in arithmetic
// -----------------------------------------------------------------------------

func TestArithmeticBinaryExpressionPosition(t *testing.T) {
	tree := parseWithPositions(t, "SELECT price * 1.1 FROM products")

	sel := tree.Statements[0].(*ast.SelectStatement)
	if len(sel.Columns) == 0 {
		t.Fatal("expected columns")
	}

	binExpr, ok := sel.Columns[0].(*ast.BinaryExpression)
	if !ok {
		t.Fatalf("expected *ast.BinaryExpression, got %T", sel.Columns[0])
	}

	assertPos(t, "BinaryExpression(*).Pos", binExpr.Pos)
}
