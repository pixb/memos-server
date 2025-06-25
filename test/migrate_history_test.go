package test

import (
	"context"
	"fmt"
	"testing"

	"github.com/pixb/memos-server/server/profile"
	store "github.com/pixb/memos-server/store"
	"github.com/pixb/memos-server/store/db"
	"github.com/stretchr/testify/assert"
)

func TestFindMigrationHistoryList(t *testing.T) {
	instanceProfile := &profile.Profile{
		Driver: "sqlite",
		DSN:    "../build/memos_dev.db",
	}

	ctx, cancel := context.WithCancel(context.Background())
	dbDriver, err := db.NewDBDriver(instanceProfile)
	if err != nil {
		cancel()
		assert.NoError(t, err)
	}

	migrationHistoryList, err := dbDriver.FindMigrationHistoryList(ctx, &store.FindMigrationHistory{})
	assert.NoError(t, err)
	for _, migrationHistory := range migrationHistoryList {
		fmt.Println(migrationHistory)
	}
}
