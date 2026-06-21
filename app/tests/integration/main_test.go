//go:build integration

package integrationtest

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"testing"

	_ "github.com/go-sql-driver/mysql"
	"github.com/stretchr/testify/require"
	tcmysql "github.com/testcontainers/testcontainers-go/modules/mysql"
)

var testDB *sql.DB

func TestMain(m *testing.M) {
	ctx := context.Background()

	container, err := tcmysql.Run(ctx,
		"mysql:8.0",
		tcmysql.WithDatabase("secunda"),
		tcmysql.WithUsername("root"),
		tcmysql.WithPassword("root"),
	)
	if err != nil {
		fmt.Println("start mysql container:", err)
		os.Exit(1)
	}
	defer container.Terminate(ctx) //nolint:errcheck

	dsn, err := container.ConnectionString(ctx, "parseTime=true")
	if err != nil {
		fmt.Println("connection string:", err)
		os.Exit(1)
	}

	testDB, err = sql.Open("mysql", dsn)
	if err != nil {
		fmt.Println("open db:", err)
		os.Exit(1)
	}
	defer testDB.Close()

	if err := applyMigrations(dsn); err != nil {
		fmt.Println("apply migrations:", err)
		os.Exit(1)
	}

	os.Exit(m.Run())
}

// applyMigrations uses a separate connection with multiStatements=true
// because some migration files (e.g. indexes) have multiple statements.
func applyMigrations(dsn string) error {
	db, err := sql.Open("mysql", dsn+"&multiStatements=true")
	if err != nil {
		return err
	}
	defer db.Close()

	files, err := filepath.Glob("../../migrations/*.sql")
	if err != nil {
		return err
	}
	sort.Strings(files)

	for _, f := range files {
		content, err := os.ReadFile(f)
		if err != nil {
			return fmt.Errorf("read %s: %w", f, err)
		}
		if _, err := db.Exec(string(content)); err != nil {
			return fmt.Errorf("exec %s: %w", filepath.Base(f), err)
		}
	}
	return nil
}

// cleanDB deletes all rows in dependency order.
func cleanDB(t *testing.T) {
	t.Helper()

	for _, tbl := range []string{"task_comments", "task_history", "tasks", "team_members", "teams", "users"} {
		_, err := testDB.Exec("DELETE FROM " + tbl)
		require.NoError(t, err, "clean table %s", tbl)
	}
}

func insertUser(t *testing.T, email, name string) int64 {
	t.Helper()
	res, err := testDB.Exec(
		"INSERT INTO users (email, password_hash, name) VALUES (?, ?, ?)",
		email, "$2a$10$fakehash", name,
	)
	require.NoError(t, err)
	id, _ := res.LastInsertId()
	return id
}

func insertTeam(t *testing.T, name string, createdBy int64) int64 {
	t.Helper()
	res, err := testDB.Exec(
		"INSERT INTO teams (name, created_by) VALUES (?, ?)", name, createdBy,
	)
	require.NoError(t, err)
	id, _ := res.LastInsertId()
	return id
}
