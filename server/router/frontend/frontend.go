package frontend

import (
	"context"
	"embed"
	"fmt"
	"io/fs"
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"

	"github.com/pixb/memos-server/internal/util"
	"github.com/pixb/memos-server/server/profile"
	"github.com/pixb/memos-server/store"
)

//go:embed dist/*
var embeddedFiles embed.FS

type FrontendService struct {
	Profile *profile.Profile
	Store   *store.Store
}

func NewFrontendService(profile *profile.Profile, store *store.Store) *FrontendService {
	return &FrontendService{
		Profile: profile,
		Store:   store,
	}
}

func (*FrontendService) Serve(_ context.Context, e *echo.Echo) {
	fmt.Println("frontend.go Serve()")
	apiSkipper := func(c echo.Context) bool {
		if util.HasPrefixes(c.Path(), "/api", "/memos.api.v1") {
			return true
		}
		// Set Cache-Control header to allow public caching with a max-age of 30 days (in seconds).
		c.Response().Header().Set(echo.HeaderCacheControl, "public, max-age=2592000")
		return false
	}

	// Route to serve the main app with HTML5 fallback for SPA behavior.
	e.Use(middleware.StaticWithConfig(middleware.StaticConfig{
		Filesystem: getFileSystem("dist"),
		HTML5:      true, // Enable fallback to index.html
		Skipper:    apiSkipper,
	}))
}

func getFileSystem(path string) http.FileSystem {
	fmt.Println("getFileSystem(), path:" + path)
	fs, err := fs.Sub(embeddedFiles, path)
	if err != nil {
		panic(err)
	}
	return http.FS(fs)
}
