package test

import (
	"fmt"
	"testing"

	storepb "github.com/pixb/memos-server/proto/gen/store"
)

func TestPB(t *testing.T) {
	name := storepb.WorkspaceSettingKey_BASIC.String()
	fmt.Printf("TestPB(), name: %s\n", name)
}
