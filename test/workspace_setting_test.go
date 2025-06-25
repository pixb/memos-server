package test

import (
	"context"
	"fmt"
	"log/slog"
	"testing"

	storepb "github.com/pixb/memos-server/proto/gen/store"
	"github.com/pixb/memos-server/server/profile"
	"github.com/pixb/memos-server/store"
	"github.com/pixb/memos-server/store/db"
	"github.com/stretchr/testify/assert"
	"google.golang.org/protobuf/encoding/protojson"
)

func TestWorkspaceSettingKey(t *testing.T) {
	fmt.Println(storepb.WorkspaceSettingKey_WORKSPACE_SETTING_KEY_UNSPECIFIED.String())
	fmt.Println(storepb.WorkspaceSettingKey_BASIC.String())
	fmt.Println(storepb.WorkspaceSettingKey_GENERAL.String())
	fmt.Println(storepb.WorkspaceSettingKey_STORAGE.String())
	fmt.Println(storepb.WorkspaceSettingKey_MEMO_RELATED.String())
}

func TestWorkspaceSettingValue(t *testing.T) {
	fmt.Println(storepb.WorkspaceSettingKey_value["WORKSPACE_SETTING_KEY_UNSPECIFIED"])
	fmt.Println(storepb.WorkspaceSettingKey_value["BASIC"])
	fmt.Println(storepb.WorkspaceSettingKey_value["GENERAL"])
	fmt.Println(storepb.WorkspaceSettingKey_value["STORAGE"])
	fmt.Println(storepb.WorkspaceSettingKey_value["MEMO_RELATED"])
}

func TestProtoToJson(t *testing.T) {
	workspaceSetting := &storepb.WorkspaceSetting{
		Key: storepb.WorkspaceSettingKey_GENERAL,
		Value: &storepb.WorkspaceSetting_GeneralSetting{
			GeneralSetting: &storepb.WorkspaceGeneralSetting{
				AdditionalScript: "",
			},
		},
	}
	workspaceSettingJson, err := protojson.Marshal(workspaceSetting)
	assert.NoError(t, err)
	fmt.Println("\tworkSpaceSettingJson:" + string(workspaceSettingJson))
}

func TestStoreGetWorkspaceSetting(t *testing.T) {
	instanceProfile := &profile.Profile{
		Mode:   "dev",
		Driver: "sqlite",
		DSN:    "../build/memos_dev.db",
	}

	ctx, cancel := context.WithCancel(context.Background())
	dbDriver, err := db.NewDBDriver(instanceProfile)
	if err != nil {
		cancel()
		assert.NoError(t, err)
	}
	storeInstance := store.New(dbDriver, instanceProfile)
	if err := storeInstance.Migrate(ctx); err != nil {
		cancel()
		slog.Error("failed to migrate", "error", err)
		return
	}
	GetWorkspaceSetting(t, ctx, storeInstance)
	UpsertGeneralWorkspaceSetting(t, ctx, storeInstance)
	storeInstance.Close()
	cancel()
}

func GetWorkspaceSetting(t *testing.T, ctx context.Context, ts *store.Store) {
	fmt.Println("\t=== GetWorkspaceSetting() ===")
	setting, err := ts.GetWorkspaceSetting(ctx, &store.FindWorkspaceSetting{
		Name: storepb.WorkspaceSettingKey_BASIC.String(),
	})
	assert.NoError(t, err)
	fmt.Printf("\tGetWorkspaceSetting(),BASIC setting:%+v\n", setting)
}

func UpsertGeneralWorkspaceSetting(t *testing.T, ctx context.Context, ts *store.Store) {
	fmt.Println("\t=== UpsertGeneralWorkspaceSetting ===")
	workspaceSetting, err := ts.UpsertWorkspaceSetting(ctx, &storepb.WorkspaceSetting{
		Key: storepb.WorkspaceSettingKey_GENERAL,
		Value: &storepb.WorkspaceSetting_GeneralSetting{
			GeneralSetting: &storepb.WorkspaceGeneralSetting{
				AdditionalScript: "",
			},
		},
	})
	assert.NoError(t, err)
	assert.NotNil(t, workspaceSetting)
}
