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

// Package parser - ddl_test.go
// Tests for DDL statement parsing: CREATE, DROP, REFRESH for views, materialized views, tables, and indexes.

package parser

import (
	"testing"

	"github.com/aereal/sqlx/models"
	"github.com/aereal/sqlx/sql/ast"
	"github.com/aereal/sqlx/sql/keywords"
	"github.com/aereal/sqlx/sql/tokenizer"
)

func parseDDLSQL(t *testing.T, sql string) *ast.AST {
	t.Helper()

	tkz := tokenizer.GetTokenizer()
	defer tokenizer.PutTokenizer(tkz)

	tokens, err := tkz.Tokenize([]byte(sql))
	if err != nil {
		t.Fatalf("Failed to tokenize: %v", err)
	}

	parser := &Parser{}
	astObj, err := parser.ParseFromModelTokens(tokens)
	if err != nil {
		t.Fatalf("Failed to parse: %v", err)
	}

	return astObj
}

func TestParser_CreateView(t *testing.T) {
	tests := []struct {
		name      string
		sql       string
		wantName  string
		wantCols  int
		orReplace bool
		temporary bool
	}{
		{
			name:     "simple CREATE VIEW",
			sql:      "CREATE VIEW user_summary AS SELECT id, name FROM users",
			wantName: "user_summary",
		},
		{
			name:      "CREATE OR REPLACE VIEW",
			sql:       "CREATE OR REPLACE VIEW user_summary AS SELECT id, name FROM users",
			wantName:  "user_summary",
			orReplace: true,
		},
		{
			name:      "CREATE TEMPORARY VIEW",
			sql:       "CREATE TEMPORARY VIEW temp_users AS SELECT id FROM users",
			wantName:  "temp_users",
			temporary: true,
		},
		{
			name:     "CREATE VIEW with column list",
			sql:      "CREATE VIEW user_info (user_id, user_name) AS SELECT id, name FROM users",
			wantName: "user_info",
			wantCols: 2,
		},
		{
			name:     "CREATE VIEW IF NOT EXISTS",
			sql:      "CREATE VIEW IF NOT EXISTS user_view AS SELECT id FROM users",
			wantName: "user_view",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := parseDDLSQL(t, tt.sql)
			defer ast.ReleaseAST(result)

			if len(result.Statements) != 1 {
				t.Fatalf("expected 1 statement, got %d", len(result.Statements))
			}

			stmt, ok := result.Statements[0].(*ast.CreateViewStatement)
			if !ok {
				t.Fatalf("expected CreateViewStatement, got %T", result.Statements[0])
			}

			if stmt.Name != tt.wantName {
				t.Errorf("expected view name %q, got %q", tt.wantName, stmt.Name)
			}

			if stmt.OrReplace != tt.orReplace {
				t.Errorf("expected OrReplace=%v, got %v", tt.orReplace, stmt.OrReplace)
			}

			if stmt.Temporary != tt.temporary {
				t.Errorf("expected Temporary=%v, got %v", tt.temporary, stmt.Temporary)
			}

			if tt.wantCols > 0 && len(stmt.Columns) != tt.wantCols {
				t.Errorf("expected %d columns, got %d", tt.wantCols, len(stmt.Columns))
			}

			if stmt.Query == nil {
				t.Error("expected Query to be non-nil")
			}
		})
	}
}

func TestParser_CreateMaterializedView(t *testing.T) {
	tests := []struct {
		name        string
		sql         string
		wantName    string
		wantCols    int
		ifNotExists bool
		withData    *bool
	}{
		{
			name:     "simple CREATE MATERIALIZED VIEW",
			sql:      "CREATE MATERIALIZED VIEW sales_summary AS SELECT region, amount FROM sales",
			wantName: "sales_summary",
		},
		{
			name:        "CREATE MATERIALIZED VIEW IF NOT EXISTS",
			sql:         "CREATE MATERIALIZED VIEW IF NOT EXISTS mv_test AS SELECT id FROM users",
			wantName:    "mv_test",
			ifNotExists: true,
		},
		{
			name:     "CREATE MATERIALIZED VIEW with column list",
			sql:      "CREATE MATERIALIZED VIEW sales_mv (reg, total) AS SELECT region, amount FROM sales",
			wantName: "sales_mv",
			wantCols: 2,
		},
		{
			name:     "CREATE MATERIALIZED VIEW WITH DATA",
			sql:      "CREATE MATERIALIZED VIEW mv_data AS SELECT id FROM users WITH DATA",
			wantName: "mv_data",
			withData: boolPtr(true),
		},
		{
			name:     "CREATE MATERIALIZED VIEW WITH NO DATA",
			sql:      "CREATE MATERIALIZED VIEW mv_nodata AS SELECT id FROM users WITH NO DATA",
			wantName: "mv_nodata",
			withData: boolPtr(false),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := parseDDLSQL(t, tt.sql)
			defer ast.ReleaseAST(result)

			if len(result.Statements) != 1 {
				t.Fatalf("expected 1 statement, got %d", len(result.Statements))
			}

			stmt, ok := result.Statements[0].(*ast.CreateMaterializedViewStatement)
			if !ok {
				t.Fatalf("expected CreateMaterializedViewStatement, got %T", result.Statements[0])
			}

			if stmt.Name != tt.wantName {
				t.Errorf("expected view name %q, got %q", tt.wantName, stmt.Name)
			}

			if stmt.IfNotExists != tt.ifNotExists {
				t.Errorf("expected IfNotExists=%v, got %v", tt.ifNotExists, stmt.IfNotExists)
			}

			if tt.wantCols > 0 && len(stmt.Columns) != tt.wantCols {
				t.Errorf("expected %d columns, got %d", tt.wantCols, len(stmt.Columns))
			}

			if stmt.Query == nil {
				t.Error("expected Query to be non-nil")
			}

			if tt.withData != nil {
				if stmt.WithData == nil {
					t.Error("expected WithData to be non-nil")
				} else if *stmt.WithData != *tt.withData {
					t.Errorf("expected WithData=%v, got %v", *tt.withData, *stmt.WithData)
				}
			}
		})
	}
}

func TestParser_RefreshMaterializedView(t *testing.T) {
	tests := []struct {
		name         string
		sql          string
		wantName     string
		concurrently bool
		withData     *bool
	}{
		{
			name:     "simple REFRESH MATERIALIZED VIEW",
			sql:      "REFRESH MATERIALIZED VIEW sales_summary",
			wantName: "sales_summary",
		},
		{
			name:         "REFRESH MATERIALIZED VIEW CONCURRENTLY",
			sql:          "REFRESH MATERIALIZED VIEW CONCURRENTLY sales_mv",
			wantName:     "sales_mv",
			concurrently: true,
		},
		{
			name:     "REFRESH MATERIALIZED VIEW WITH DATA",
			sql:      "REFRESH MATERIALIZED VIEW mv_test WITH DATA",
			wantName: "mv_test",
			withData: boolPtr(true),
		},
		{
			name:     "REFRESH MATERIALIZED VIEW WITH NO DATA",
			sql:      "REFRESH MATERIALIZED VIEW mv_test WITH NO DATA",
			wantName: "mv_test",
			withData: boolPtr(false),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := parseDDLSQL(t, tt.sql)
			defer ast.ReleaseAST(result)

			if len(result.Statements) != 1 {
				t.Fatalf("expected 1 statement, got %d", len(result.Statements))
			}

			stmt, ok := result.Statements[0].(*ast.RefreshMaterializedViewStatement)
			if !ok {
				t.Fatalf("expected RefreshMaterializedViewStatement, got %T", result.Statements[0])
			}

			if stmt.Name != tt.wantName {
				t.Errorf("expected view name %q, got %q", tt.wantName, stmt.Name)
			}

			if stmt.Concurrently != tt.concurrently {
				t.Errorf("expected Concurrently=%v, got %v", tt.concurrently, stmt.Concurrently)
			}

			if tt.withData != nil {
				if stmt.WithData == nil {
					t.Error("expected WithData to be non-nil")
				} else if *stmt.WithData != *tt.withData {
					t.Errorf("expected WithData=%v, got %v", *tt.withData, *stmt.WithData)
				}
			}
		})
	}
}

func TestParser_DropStatement(t *testing.T) {
	tests := []struct {
		name        string
		sql         string
		objectType  string
		wantNames   []string
		ifExists    bool
		cascadeType string
	}{
		{
			name:       "DROP TABLE",
			sql:        "DROP TABLE users",
			objectType: "TABLE",
			wantNames:  []string{"users"},
		},
		{
			name:       "DROP TABLE IF EXISTS",
			sql:        "DROP TABLE IF EXISTS users",
			objectType: "TABLE",
			wantNames:  []string{"users"},
			ifExists:   true,
		},
		{
			name:        "DROP TABLE CASCADE",
			sql:         "DROP TABLE users CASCADE",
			objectType:  "TABLE",
			wantNames:   []string{"users"},
			cascadeType: "CASCADE",
		},
		{
			name:       "DROP VIEW",
			sql:        "DROP VIEW user_view",
			objectType: "VIEW",
			wantNames:  []string{"user_view"},
		},
		{
			name:       "DROP MATERIALIZED VIEW",
			sql:        "DROP MATERIALIZED VIEW sales_mv",
			objectType: "MATERIALIZED VIEW",
			wantNames:  []string{"sales_mv"},
		},
		{
			name:        "DROP MATERIALIZED VIEW IF EXISTS CASCADE",
			sql:         "DROP MATERIALIZED VIEW IF EXISTS sales_mv CASCADE",
			objectType:  "MATERIALIZED VIEW",
			wantNames:   []string{"sales_mv"},
			ifExists:    true,
			cascadeType: "CASCADE",
		},
		{
			name:       "DROP INDEX",
			sql:        "DROP INDEX idx_users_email",
			objectType: "INDEX",
			wantNames:  []string{"idx_users_email"},
		},
		{
			name:       "DROP multiple tables",
			sql:        "DROP TABLE users, orders, products",
			objectType: "TABLE",
			wantNames:  []string{"users", "orders", "products"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := parseDDLSQL(t, tt.sql)
			defer ast.ReleaseAST(result)

			if len(result.Statements) != 1 {
				t.Fatalf("expected 1 statement, got %d", len(result.Statements))
			}

			stmt, ok := result.Statements[0].(*ast.DropStatement)
			if !ok {
				t.Fatalf("expected DropStatement, got %T", result.Statements[0])
			}

			if stmt.ObjectType != tt.objectType {
				t.Errorf("expected ObjectType %q, got %q", tt.objectType, stmt.ObjectType)
			}

			if len(stmt.Names) != len(tt.wantNames) {
				t.Errorf("expected %d names, got %d", len(tt.wantNames), len(stmt.Names))
			}

			for i, name := range tt.wantNames {
				if i < len(stmt.Names) && stmt.Names[i] != name {
					t.Errorf("expected name[%d]=%q, got %q", i, name, stmt.Names[i])
				}
			}

			if stmt.IfExists != tt.ifExists {
				t.Errorf("expected IfExists=%v, got %v", tt.ifExists, stmt.IfExists)
			}

			if stmt.CascadeType != tt.cascadeType {
				t.Errorf("expected CascadeType=%q, got %q", tt.cascadeType, stmt.CascadeType)
			}
		})
	}
}

func TestParser_CreateIndex(t *testing.T) {
	tests := []struct {
		name        string
		sql         string
		indexName   string
		tableName   string
		unique      bool
		ifNotExists bool
		numColumns  int
		span        models.Span
	}{
		{
			name:       "simple CREATE INDEX",
			sql:        "CREATE INDEX idx_email ON users (email)",
			indexName:  "idx_email",
			tableName:  "users",
			numColumns: 1,
			span:       models.Span{Start: models.Location{Line: 1, Column: 1}, End: models.Location{Line: 1, Column: 40}},
		},
		{
			name:       "CREATE UNIQUE INDEX",
			sql:        "CREATE UNIQUE INDEX idx_email ON users (email)",
			indexName:  "idx_email",
			tableName:  "users",
			unique:     true,
			numColumns: 1,
			span:       models.Span{Start: models.Location{Line: 1, Column: 1}, End: models.Location{Line: 1, Column: 47}},
		},
		{
			name:        "CREATE INDEX IF NOT EXISTS",
			sql:         "CREATE INDEX IF NOT EXISTS idx_name ON users (name)",
			indexName:   "idx_name",
			tableName:   "users",
			ifNotExists: true,
			numColumns:  1,
			span:        models.Span{Start: models.Location{Line: 1, Column: 1}, End: models.Location{Line: 1, Column: 52}},
		},
		{
			name:       "CREATE INDEX with multiple columns",
			sql:        "CREATE INDEX idx_composite ON orders (user_id, order_date)",
			indexName:  "idx_composite",
			tableName:  "orders",
			numColumns: 2,
			span:       models.Span{Start: models.Location{Line: 1, Column: 1}, End: models.Location{Line: 1, Column: 59}},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := parseDDLSQL(t, tt.sql)
			defer ast.ReleaseAST(result)

			if len(result.Statements) != 1 {
				t.Fatalf("expected 1 statement, got %d", len(result.Statements))
			}

			stmt, ok := result.Statements[0].(*ast.CreateIndexStatement)
			if !ok {
				t.Fatalf("expected CreateIndexStatement, got %T", result.Statements[0])
			}

			if stmt.Name != tt.indexName {
				t.Errorf("expected index name %q, got %q", tt.indexName, stmt.Name)
			}

			if stmt.Table.Name != tt.tableName {
				t.Errorf("expected table name %q, got %q", tt.tableName, stmt.Table.Name)
			}

			if stmt.Unique != tt.unique {
				t.Errorf("expected Unique=%v, got %v", tt.unique, stmt.Unique)
			}

			if stmt.IfNotExists != tt.ifNotExists {
				t.Errorf("expected IfNotExists=%v, got %v", tt.ifNotExists, stmt.IfNotExists)
			}

			if len(stmt.Columns) != tt.numColumns {
				t.Errorf("expected %d columns, got %d", tt.numColumns, len(stmt.Columns))
			}

			assertSpanEqual(t, "CREATE INDEX", tt.span, models.Span{Start: stmt.Start, End: stmt.End})
		})
	}
}

func TestParser_CreateTableWithPartitioning(t *testing.T) {
	tests := []struct {
		name          string
		sql           string
		tableName     string
		numColumns    int
		partitionType string
		numPartitions int
	}{
		{
			name:          "CREATE TABLE with RANGE partitioning",
			sql:           "CREATE TABLE sales (sale_date DATE, amount DECIMAL) PARTITION BY RANGE (sale_date) (PARTITION p2024 VALUES LESS THAN (MAXVALUE))",
			tableName:     "sales",
			numColumns:    2,
			partitionType: "RANGE",
			numPartitions: 1,
		},
		{
			name:          "CREATE TABLE with HASH partitioning",
			sql:           "CREATE TABLE users (id INT, name VARCHAR) PARTITION BY HASH (id)",
			tableName:     "users",
			numColumns:    2,
			partitionType: "HASH",
			numPartitions: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := parseDDLSQL(t, tt.sql)
			defer ast.ReleaseAST(result)

			if len(result.Statements) != 1 {
				t.Fatalf("expected 1 statement, got %d", len(result.Statements))
			}

			stmt, ok := result.Statements[0].(*ast.CreateTableStatement)
			if !ok {
				t.Fatalf("expected CreateTableStatement, got %T", result.Statements[0])
			}

			if stmt.Table.Name != tt.tableName {
				t.Errorf("expected table name %q, got %q", tt.tableName, stmt.Table.Name)
			}

			if len(stmt.Columns) != tt.numColumns {
				t.Errorf("expected %d columns, got %d", tt.numColumns, len(stmt.Columns))
			}

			if stmt.PartitionBy == nil {
				t.Fatal("expected PartitionBy to be non-nil")
			}

			if stmt.PartitionBy.Type != tt.partitionType {
				t.Errorf("expected partition type %q, got %q", tt.partitionType, stmt.PartitionBy.Type)
			}

			if len(stmt.Partitions) != tt.numPartitions {
				t.Errorf("expected %d partitions, got %d", tt.numPartitions, len(stmt.Partitions))
			}
		})
	}
}

func TestParser_CreateIndex_postgresql(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		sql       string
		indexName string
		tableName string
		columns   []ast.IndexColumn
	}{
		{
			name:      "simple",
			sql:       "CREATE INDEX idx_email ON users (email)",
			indexName: "idx_email",
			tableName: "users",
			columns: []ast.IndexColumn{
				{Column: "email", Collate: "", Direction: "", NullsLast: false},
			},
		},
		{
			name:      "with collate",
			sql:       `CREATE INDEX users_name ON public.users USING btree (name COLLATE "ja-JP-x-icu")`,
			indexName: "users_name",
			tableName: "public.users",
			columns: []ast.IndexColumn{
				{Column: "name", Collate: "ja-JP-x-icu", Direction: "", NullsLast: false},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			d := keywords.DialectPostgreSQL
			tkz, err := tokenizer.NewWithDialect(d)
			if err != nil {
				t.Fatal(err)
			}
			tokens, err := tkz.TokenizeContext(t.Context(), []byte(tt.sql))
			if err != nil {
				t.Fatal(err)
			}
			parser := NewParser(WithDialect(string(d)))
			tree, err := parser.ParseContextFromModelTokens(t.Context(), tokens)
			if err != nil {
				t.Fatal(err)
			}

			if len(tree.Statements) != 1 {
				t.Fatalf("expected 1 statement, got %d", len(tree.Statements))
			}

			stmt, ok := tree.Statements[0].(*ast.CreateIndexStatement)
			if !ok {
				t.Fatalf("expected CreateIndexStatement, got %T", tree.Statements[0])
			}

			if stmt.Name != tt.indexName {
				t.Errorf("expected index name %q, got %q", tt.indexName, stmt.Name)
			}

			if stmt.Table.Name != tt.tableName {
				t.Errorf("expected table name %q, got %q", tt.tableName, stmt.Table.Name)
			}

			if len(stmt.Columns) != len(tt.columns) {
				t.Fatalf("len(Columns): want=%d got=%d", len(tt.columns), len(stmt.Columns))
			}

			for i := range stmt.Columns {
				got := stmt.Columns[i]
				want := tt.columns[i]
				if got.Column != want.Column {
					t.Errorf("Columns[%d]: Column: want=%s got=%s", i, want.Column, got.Column)
				}
				if got.Collate != want.Collate {
					t.Errorf("Columns[%d]: Collate: want=%s got=%s", i, want.Collate, got.Collate)
				}
				if got.Direction != want.Direction {
					t.Errorf("Columns[%d]: Direction: want=%s got=%s", i, want.Direction, got.Direction)
				}
				if got.NullsLast != want.NullsLast {
					t.Errorf("Columns[%d]: NullsLast: want=%v got=%v", i, want.NullsLast, got.NullsLast)
				}
			}
		})
	}
}

func TestParser_CreateTableSimple(t *testing.T) {
	tests := []struct {
		name        string
		sql         string
		tableName   string
		numColumns  int
		ifNotExists bool
		temporary   bool
	}{
		{
			name:       "simple CREATE TABLE",
			sql:        "CREATE TABLE users (id INT, name VARCHAR)",
			tableName:  "users",
			numColumns: 2,
		},
		{
			name:        "CREATE TABLE IF NOT EXISTS",
			sql:         "CREATE TABLE IF NOT EXISTS users (id INT, name VARCHAR)",
			tableName:   "users",
			numColumns:  2,
			ifNotExists: true,
		},
		{
			name:       "CREATE TEMPORARY TABLE",
			sql:        "CREATE TEMPORARY TABLE temp_data (id INT, value TEXT)",
			tableName:  "temp_data",
			numColumns: 2,
			temporary:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := parseDDLSQL(t, tt.sql)
			defer ast.ReleaseAST(result)

			if len(result.Statements) != 1 {
				t.Fatalf("expected 1 statement, got %d", len(result.Statements))
			}

			stmt, ok := result.Statements[0].(*ast.CreateTableStatement)
			if !ok {
				t.Fatalf("expected CreateTableStatement, got %T", result.Statements[0])
			}

			if stmt.Table.Name != tt.tableName {
				t.Errorf("expected table name %q, got %q", tt.tableName, stmt.Table.Name)
			}

			if len(stmt.Columns) != tt.numColumns {
				t.Errorf("expected %d columns, got %d", tt.numColumns, len(stmt.Columns))
			}

			if stmt.IfNotExists != tt.ifNotExists {
				t.Errorf("expected IfNotExists=%v, got %v", tt.ifNotExists, stmt.IfNotExists)
			}

			if stmt.Temporary != tt.temporary {
				t.Errorf("expected Temporary=%v, got %v", tt.temporary, stmt.Temporary)
			}
		})
	}
}

func TestParser_CreateTable_postgresql_column(t *testing.T) {
	tests := []struct {
		name           string
		sql            string
		tableName      string
		wantColumnDefs []ast.ColumnDef
		shouldErr      bool
	}{
		{
			name:      "simple CREATE TABLE",
			sql:       "CREATE TABLE users (id INT, name VARCHAR(255), age SMALLINT)",
			tableName: "users",
			shouldErr: false,
			wantColumnDefs: []ast.ColumnDef{
				{Name: "id", Type: "INT"},
				{Name: "name", Type: "VARCHAR(255)"},
				{Name: "age", Type: "SMALLINT"},
			},
		},
		{
			name:      "character varying",
			sql:       "CREATE TABLE users (id INT, name character varying(255), age SMALLINT)",
			tableName: "users",
			shouldErr: false,
			wantColumnDefs: []ast.ColumnDef{
				{Name: "id", Type: "INT"},
				{Name: "name", Type: "character varying(255)"},
				{Name: "age", Type: "SMALLINT"},
			},
		},
		{
			name:      "double precision",
			sql:       "CREATE TABLE users (id INT, rate double precision, age SMALLINT)",
			tableName: "users",
			shouldErr: false,
			wantColumnDefs: []ast.ColumnDef{
				{Name: "id", Type: "INT"},
				{Name: "rate", Type: "double precision"},
				{Name: "age", Type: "SMALLINT"},
			},
		},
		{
			name:      "timestamp with zone",
			sql:       "CREATE TABLE users (created_at timestamp with time zone not null, age SMALLINT)",
			tableName: "users",
			wantColumnDefs: []ast.ColumnDef{
				{Name: "created_at", Type: "timestamp with time zone"},
				{Name: "age", Type: "SMALLINT"},
			},
		},
		{
			name:      "timestamp with zone w/precision",
			sql:       "CREATE TABLE users (created_at timestamp(3) with time zone not null, age SMALLINT)",
			tableName: "users",
			wantColumnDefs: []ast.ColumnDef{
				{Name: "created_at", Type: "timestamp(3) with time zone"},
				{Name: "age", Type: "SMALLINT"},
			},
		},
		{
			name:      "duration with fields",
			sql:       "CREATE TABLE users (dur INTERVAL YEAR TO MONTH, age SMALLINT)",
			tableName: "users",
			wantColumnDefs: []ast.ColumnDef{
				{Name: "dur", Type: "INTERVAL YEAR TO MONTH"},
				{Name: "age", Type: "SMALLINT"},
			},
		},
		{
			name:      "qualified type",
			sql:       "CREATE TABLE users (age public.unsigned_smallint NOT NULL)",
			tableName: "users",
			wantColumnDefs: []ast.ColumnDef{
				{Name: "age", Type: "public.unsigned_smallint"},
			},
		},
		{
			name:           "ng/varying after int",
			sql:            "CREATE TABLE users (name int varying(255))",
			tableName:      "users",
			shouldErr:      true,
			wantColumnDefs: nil,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			d := keywords.DialectPostgreSQL
			tkz, err := tokenizer.NewWithDialect(d)
			if err != nil {
				t.Fatal(err)
			}

			tokens, err := tkz.Tokenize([]byte(tc.sql))
			if err != nil {
				t.Fatal(err)
			}

			parser := NewParser(WithDialect(string(d)))
			result, err := parser.ParseFromModelTokens(tokens)
			if tc.shouldErr {
				if err == nil {
					t.Fatal("expected error but got nothing")
				}
				return
			}
			if err != nil {
				t.Fatal(err)
			}

			if len(result.Statements) != 1 {
				t.Fatalf("expected 1 statement, got %d", len(result.Statements))
			}

			stmt, ok := result.Statements[0].(*ast.CreateTableStatement)
			if !ok {
				t.Fatalf("expected CreateTableStatement, got %T", result.Statements[0])
			}

			if stmt.Table.Name != tc.tableName {
				t.Errorf("expected table name %q, got %q", tc.tableName, stmt.Table.Name)
			}

			if len(stmt.Columns) != len(tc.wantColumnDefs) {
				t.Fatalf("len(Columns): want=%d got=%d", len(tc.wantColumnDefs), len(stmt.Columns))
			}

			for i := range stmt.Columns {
				got := stmt.Columns[i]
				want := tc.wantColumnDefs[i]
				if got.Name != want.Name {
					t.Errorf("Columns[%d].Name: want=%s got=%s", i, want.Name, got.Name)
				}
				if got.Type != want.Type {
					t.Errorf("Columns[%d].Type: want=%s got=%s", i, want.Type, got.Type)
				}
			}
		})
	}
}

// Helper function to create bool pointer
func boolPtr(b bool) *bool {
	return &b
}

func assertSpanEqual(t *testing.T, label string, want, got models.Span) {
	t.Helper()
	assertPosEqual(t, label+": Start", want.Start, got.Start)
	assertPosEqual(t, label+": End", want.End, got.End)
}

func assertPosEqual(t *testing.T, label string, want, got models.Location) {
	t.Helper()
	if want.Line != got.Line {
		t.Errorf("%s: Line: want=%d got=%d", label, want.Line, got.Line)
	}
	if want.Column != got.Column {
		t.Errorf("%s: Column: want=%d got=%d", label, want.Column, got.Column)
	}
}
