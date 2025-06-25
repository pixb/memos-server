package test

import (
	"context"
	"database/sql"
	"fmt"
	"testing"

	// Import the SQLite driver.
	_ "modernc.org/sqlite"
)

func TestQueryMigrationHistory(t *testing.T) {
	dbPath := "../build/memos_dev.db"
	db, err := sql.Open("sqlite", dbPath+"?_pragma=foreign_keys(0)&_pragma=busy_timeout(10000)&_pragma=journal_mode(WAL)")
	if err != nil {
		fmt.Println(err)
	}
	ctx, cancel := context.WithCancel(context.Background())
	query := "SELECT `version`, `created_ts` FROM `migration_history` ORDER BY `created_ts` DESC"
	rows, err := db.QueryContext(ctx, query)
	if err != nil {
		fmt.Println(err)
		cancel()
		return
	}
	defer rows.Close()
	for rows.Next() {
		var Version string
		var CreatedTs int64
		if err := rows.Scan(
			&Version,
			&CreatedTs,
		); err != nil {
			fmt.Println(err)
			cancel()
			return
		}
		fmt.Printf("Version: %s,CreatedTs: %d\n", Version, CreatedTs)
	}
	if err := rows.Err(); err != nil {
		fmt.Println(err)
		cancel()
		return
	}
	cancel()
}
