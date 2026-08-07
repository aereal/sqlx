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

package main

import (
	"fmt"
	"log"
	"strings"

	"github.com/aereal/sqlx/gosqlx"
	"github.com/aereal/sqlx/sql/ast"
)

func main() {
	// Example 1: Basic multi-row INSERT
	sql1 := `INSERT INTO users (name, email) VALUES
		('John', 'john@example.com'),
		('Jane', 'jane@example.com'),
		('Bob', 'bob@example.com')`

	fmt.Println("Example 1: Basic Multi-Row INSERT")
	fmt.Println("SQL:", sql1)
	fmt.Println()

	astResult, err := gosqlx.Parse(sql1)
	if err != nil {
		log.Fatalf("Failed to parse: %v", err)
	}

	insertStmt, _ := astResult.Statements[0].(*ast.InsertStatement)
	fmt.Printf("Table: %s\n", insertStmt.Table.Name)
	fmt.Printf("Columns: %d\n", len(insertStmt.Columns))
	fmt.Printf("Rows: %d\n", len(insertStmt.Values))
	fmt.Println()

	// Example 2: Multi-row INSERT with ON CONFLICT (upsert)
	sql2 := `INSERT INTO products (id, name, price)
		VALUES (1, 'Widget', 9.99), (2, 'Gadget', 19.99), (3, 'Thing', 29.99)
		ON CONFLICT (id) DO UPDATE SET price = EXCLUDED.price`

	fmt.Println("Example 2: Multi-Row INSERT with ON CONFLICT")
	fmt.Println("SQL:", sql2)
	fmt.Println()

	astResult2, err := gosqlx.Parse(sql2)
	if err != nil {
		log.Fatalf("Failed to parse: %v", err)
	}

	insertStmt2, _ := astResult2.Statements[0].(*ast.InsertStatement)
	fmt.Printf("Table: %s\n", insertStmt2.Table.Name)
	fmt.Printf("Rows: %d\n", len(insertStmt2.Values))
	fmt.Printf("Has ON CONFLICT: %t\n", insertStmt2.OnConflict != nil)
	fmt.Println()

	// Example 3: Multi-row INSERT with RETURNING
	sql3 := `INSERT INTO events (name, timestamp)
		VALUES ('Login', NOW()), ('Logout', NOW()), ('PageView', NOW())
		RETURNING id, created_at`

	fmt.Println("Example 3: Multi-Row INSERT with RETURNING")
	fmt.Println("SQL:", sql3)
	fmt.Println()

	astResult3, err := gosqlx.Parse(sql3)
	if err != nil {
		log.Fatalf("Failed to parse: %v", err)
	}

	insertStmt3, _ := astResult3.Statements[0].(*ast.InsertStatement)
	fmt.Printf("Table: %s\n", insertStmt3.Table.Name)
	fmt.Printf("Rows: %d\n", len(insertStmt3.Values))
	fmt.Printf("RETURNING columns: %d\n", len(insertStmt3.Returning))
	fmt.Println()

	// Example 4: Large batch insert
	sql4 := "INSERT INTO bulk_data (id, value) VALUES "
	var sql4Sb89 strings.Builder
	for i := 1; i <= 10; i++ {
		if i > 1 {
			sql4Sb89.WriteString(", ")
		}
		sql4Sb89.WriteString(fmt.Sprintf("(%d, 'value%d')", i, i))
	}
	sql4 += sql4Sb89.String()

	fmt.Println("Example 4: Large Batch INSERT (10 rows)")
	fmt.Println()

	astResult4, err := gosqlx.Parse(sql4)
	if err != nil {
		log.Fatalf("Failed to parse: %v", err)
	}

	insertStmt4, _ := astResult4.Statements[0].(*ast.InsertStatement)
	fmt.Printf("Table: %s\n", insertStmt4.Table.Name)
	fmt.Printf("Rows: %d\n", len(insertStmt4.Values))
	fmt.Printf("Values per row: %d\n", len(insertStmt4.Values[0]))
	fmt.Println()

	fmt.Println("✓ All multi-row INSERT examples parsed successfully!")
}
