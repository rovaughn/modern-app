package main

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/neelance/graphql-go/errors"
	"io/ioutil"
	"os"
	"strings"
	"testing"
)

// Wipe the database, providing a clean slate for tests.
func prepareDB(t *testing.T) {
	t.Helper()

	if os.Getenv("environment") != "local" {
		t.Fatal("It's dangerous to wipe the database non-locally")
		return
	}

	if _, err := db.Query(`SET FOREIGN_KEY_CHECKS = 0`); err != nil {
		t.Fatal(err)
	}

	defer func() {
		if _, err := db.Query(`SET FOREIGN_KEY_CHECKS = 0`); err != nil {
			t.Fatal(err)
		}
	}()

	rows, err := db.Query(`SHOW TABLES`)
	if err != nil {
		t.Fatal(err)
	}
	defer rows.Close()
	for rows.Next() {
		var tableName string
		if err := rows.Scan(&tableName); err != nil {
			t.Fatal(err)
		}

		if _, err := db.Exec(fmt.Sprintf("DROP TABLE %s", tableName)); err != nil {
			t.Fatal(err)
		}
	}
	if err := rows.Err(); err != nil {
		t.Fatal(err)
	}

	schema, err := ioutil.ReadFile("schema.sql")
	if err != nil {
		t.Fatal(err)
	}

	queries := strings.Split(string(schema), ";")

	for _, query := range queries[:len(queries)-1] {
		if _, err := db.Exec(query); err != nil {
			t.Fatal(err)
		}
	}
}

// Run a GraphQL query against the schema.  Variables can be nil.  The JSON
// result will be unmarshaled into "v".
func query(t *testing.T, query string, variables map[string]interface{}, v interface{}) {
	t.Helper()

	result := schema.Exec(context.TODO(), query, "", variables)
	if len(result.Errors) > 0 {
		for _, e := range result.Errors {
			t.Error(e)
		}
		t.Fatal("Query returned errors")
	}
	if len(result.Extensions) > 0 {
		t.Fatal("Extensions returned")
	}

	if err := json.Unmarshal(result.Data, v); err != nil {
		t.Fatal(err)
	}
}

// Run a GraphQL query against the schema, expecting an error to occur.
func errQuery(t *testing.T, query string, variables map[string]interface{}) []*errors.QueryError {
	t.Helper()

	result := schema.Exec(context.TODO(), query, "Test", variables)
	if len(result.Errors) == 0 {
		t.Fatal("Expected at least one error")
	}
	return result.Errors
}

func TestHello(t *testing.T) {
	var result struct {
		Hello string `json:"hello"`
	}
	query(t, `
		query {
			hello(name: "John")
		}
	`, nil, &result)
	if result.Hello != "Hello John!" {
		t.Errorf("Expected %q, not %q", "Hello John!", result.Hello)
	}
}
