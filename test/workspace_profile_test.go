package test

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"testing"
	"time"

	v1pb "github.com/pixb/memos-server/proto/gen/api/v1"
	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

// ================= http test ===========================
func TestGetWorkspaceProfileHttp(t *testing.T) {
	resp, err := http.Get("http://localhost:8081/api/v1/workspace/profile")
	assert.NoError(t, err)
	defer resp.Body.Close()
	content, err := io.ReadAll(resp.Body)
	assert.NoError(t, err)
	fmt.Printf("\tcontent:%s\n", string(content))
}

// ================= grpc test ===========================
func initGrpcClientWithWorkspaceProfile(t *testing.T) (*grpc.ClientConn, func()) {
	// 1. Connect to server.
	conn, err := grpc.NewClient(":8081", grpc.WithTransportCredentials(insecure.NewCredentials()))
	assert.NoError(t, err)
	return conn, func() {
		if conn != nil {
			conn.Close()
		}
	}
}

func TestGetWorkspaceProfile(t *testing.T) {
	conn, closeFunc := initGrpcClientWithWorkspaceProfile(t)
	defer closeFunc()
	// 2. Create client.
	client := v1pb.NewWorkspaceServiceClient(conn)
	// 3. Execute grpc call.
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	workspaceProfile, err := client.GetWorkspaceProfile(ctx, &v1pb.GetWorkspaceProfileRequest{})
	assert.NoError(t, err)
	fmt.Printf("\tTestGetWorkspaceProfile(),workspaceProfile:%v\n", workspaceProfile)
}
