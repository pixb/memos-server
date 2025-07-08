package test

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"testing"
	"time"

	apiv1 "github.com/pixb/memos-server/proto/gen/api/v1"
	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

// get auth status http
func TestGetAuthStatusHttp(t *testing.T) {
	getAuthStatusUrl := "http://localhost:8081/api/v1/auth/status"
	contentType := "application/json;charset=utf-8"
	resp, err := http.Post(getAuthStatusUrl, contentType, nil)
	assert.NoError(t, err)
	defer resp.Body.Close()
	content, err := io.ReadAll(resp.Body)
	assert.NoError(t, err)
	fmt.Printf("\tTestGetAuthStatusHttp(), content: %s\n", string(content))
}

// init grpc client
func initGrpcClient(t *testing.T) (*grpc.ClientConn, func()) {
	conn, err := grpc.NewClient(":8081", grpc.WithTransportCredentials(insecure.NewCredentials()))
	assert.NoError(t, err)
	assert.NotNil(t, conn)
	return conn, func() {
		conn.Close()
	}
}

// get auth status grpc
func TestGetAuthStatusGrpc(t *testing.T) {
	conn, closeFunc := initGrpcClient(t)
	defer closeFunc()
	client := apiv1.NewAuthServiceClient(conn)
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	user, err := client.GetAuthStatus(ctx, &apiv1.GetAuthStatusRequest{})
	assert.NoError(t, err)
	fmt.Printf("\tTestGetAuthStatusGrpc(), body: %v\n", user)
}
