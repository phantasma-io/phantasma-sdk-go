package io

import (
	"os"
	"path"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestMakeDirForFile_HappyPath(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "test")
	require.NoError(t, err)

	filePath := tempDir + "/testDir/testFile.test"
	err = MakeDirForFile(filePath, "test")
	defer removeDir(t, tempDir)
	require.NoError(t, err)

	_, errChDir := os.Create(filePath)
	require.NoError(t, errChDir)
}

func removeDir(t *testing.T, dirName string) {
	err := os.RemoveAll(dirName)
	require.NoError(t, err)
}

func TestMakeDirForFile_Negative(t *testing.T) {
	file, err := os.CreateTemp("", "test")
	require.NoError(t, err)
	require.NoError(t, file.Close())

	filePath := file.Name() + "/error"
	dir := path.Dir(filePath)
	err = MakeDirForFile(filePath, "test")
	defer removeDir(t, dir)
	require.Errorf(t, err, "could not create dir for test: mkdir %s : not a directory", filePath)
}
