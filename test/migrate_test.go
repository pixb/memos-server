package test

import (
	"embed"
	"fmt"
	"io/fs"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"testing"

	"github.com/pixb/memos-server/server/profile"
	"github.com/pixb/memos-server/server/version"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
)

const (
	// MigrateFileNameSplit is the split character between the patch version and the description in the migration file name.
	// For example, "1__create_table.sql".
	MigrateFileNameSplit = "__"
	// LatestSchemaFileName is the name of the latest schema file.
	// This file is used to apply the latest schema when no migration history is found.
	LatestSchemaFileName = "LATEST.sql"
)

//go:embed migration
var migrationFS embed.FS

var p = &profile.Profile{
	Mode:   "dev",
	Driver: "sqlite",
}

func TestGetCurrentSchemaVersion(t *testing.T) {
	currentSchemaVersion, err := GetCurrentSchemaVersion()
	fmt.Printf("\tCurrent schema version: %s\n", currentSchemaVersion)
	assert.NoError(t, err)
}

func TestFilePath(t *testing.T) {
	filePaths, err := fs.Glob(migrationFS, fmt.Sprintf("%s*/*.sql", getMigrationBasePath()))
	assert.NoError(t, err)
	sort.Strings(filePaths)
	fmt.Println(filePaths)
}

func GetCurrentSchemaVersion() (string, error) {
	currentVersion := version.GetCurrentVersion(p.Mode)
	minorVersion := version.GetMinorVersion(currentVersion)
	fmt.Printf("\tGetCurrentSchemaVersion(), Current version: %s\n", currentVersion)
	fmt.Printf("\tGetCurrentSchemaVersion(), Minor version: %s\n", minorVersion)
	filePaths, err := fs.Glob(migrationFS, fmt.Sprintf("%s%s/*.sql", getMigrationBasePath(), minorVersion))
	if err != nil {
		return "", errors.Wrap(err, "failed to read migration files")
	}
	sort.Strings(filePaths)
	fmt.Printf("\tGetCurrentSchemaVersion(), filePaths.lenth: %d\n", len(filePaths))
	if len(filePaths) == 0 {
		return fmt.Sprintf("%s.0", minorVersion), nil
	}
	return getSchemaVersionOfMigrateScript(filePaths[len(filePaths)-1])
}

func getMigrationBasePath() string {
	return fmt.Sprintf("migration/%s/", p.Driver)
}

func getSchemaVersionOfMigrateScript(filePath string) (string, error) {
	fmt.Printf("\tgetSchemaVersionOfMigrateScript(), filePath: %s\n", filePath)
	// If the file is the latest schema file, return the current schema version.
	if strings.HasSuffix(filePath, LatestSchemaFileName) {
		return GetCurrentSchemaVersion()
	}

	normalizedPath := filepath.ToSlash(filePath)
	elements := strings.Split(normalizedPath, "/")
	if len(elements) < 2 {
		return "", errors.Errorf("invalid file path: %s", filePath)
	}
	minorVersion := elements[len(elements)-2]
	rawPatchVersion := strings.Split(elements[len(elements)-1], MigrateFileNameSplit)[0]
	patchVersion, err := strconv.Atoi(rawPatchVersion)
	if err != nil {
		return "", errors.Wrapf(err, "failed to convert patch version to int: %s", rawPatchVersion)
	}
	return fmt.Sprintf("%s.%d", minorVersion, patchVersion+1), nil
}
