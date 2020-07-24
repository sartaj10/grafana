package fs

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCopyFile(t *testing.T) {
	src, err := ioutil.TempFile("", "")
	require.NoError(t, err)
	defer os.RemoveAll(src.Name())
	err = ioutil.WriteFile(src.Name(), []byte("Contents"), 0600)
	require.NoError(t, err)

	dst, err := ioutil.TempFile("", "")
	require.NoError(t, err)
	defer os.RemoveAll(dst.Name())

	err = CopyFile(src.Name(), dst.Name())
	require.NoError(t, err)
}

func TestCopyFile_Permissions(t *testing.T) {
	const perms = os.FileMode(0700)

	src, err := ioutil.TempFile("", "")
	require.NoError(t, err)
	defer os.RemoveAll(src.Name())
	err = ioutil.WriteFile(src.Name(), []byte("Contents"), 0600)
	require.NoError(t, err)
	err = os.Chmod(src.Name(), perms)
	require.NoError(t, err)

	dst, err := ioutil.TempFile("", "")
	require.NoError(t, err)
	defer os.RemoveAll(dst.Name())

	err = CopyFile(src.Name(), dst.Name())
	require.NoError(t, err)

	fi, err := os.Stat(dst.Name())
	require.NoError(t, err)

	assert.Equal(t, perms, fi.Mode()&os.ModePerm)
}

// Test case where destination directory doesn't exist.
func TestCopyFile_NonExistentDestDir(t *testing.T) {
	src, err := ioutil.TempFile("", "")
	require.NoError(t, err)
	defer os.RemoveAll(src.Name())

	err = CopyFile(src.Name(), "non-existent/dest")
	require.EqualError(t, err, "destination directory doesn't exist: \"non-existent\"")
}

func TestCopyRecursive_NonExistentDest(t *testing.T) {
	src, err := ioutil.TempDir("", "")
	require.NoError(t, err)
	defer os.RemoveAll(src)

	err = os.MkdirAll(filepath.Join(src, "data"), 0755)
	require.NoError(t, err)
	// nolint:gosec
	err = ioutil.WriteFile(filepath.Join(src, "data", "file.txt"), []byte("Test"), 0644)
	require.NoError(t, err)

	dstParent, err := ioutil.TempDir("", "")
	require.NoError(t, err)
	defer os.RemoveAll(dstParent)

	dst := filepath.Join(dstParent, "dest")

	err = CopyRecursive(src, dst)
	require.NoError(t, err)

	compareDirs(t, src, dst)
}

func TestCopyRecursive_ExistentDest(t *testing.T) {
	src, err := ioutil.TempDir("", "")
	require.NoError(t, err)
	defer os.RemoveAll(src)

	err = os.MkdirAll(filepath.Join(src, "data"), 0755)
	require.NoError(t, err)
	// nolint:gosec
	err = ioutil.WriteFile(filepath.Join(src, "data", "file.txt"), []byte("Test"), 0644)
	require.NoError(t, err)

	dst, err := ioutil.TempDir("", "")
	require.NoError(t, err)
	defer os.RemoveAll(dst)

	err = CopyRecursive(src, dst)
	require.NoError(t, err)

	compareDirs(t, src, dst)
}

func compareDirs(t *testing.T, src, dst string) {
	sfi, err := os.Stat(src)
	require.NoError(t, err)
	dfi, err := os.Stat(dst)
	require.NoError(t, err)

	require.Equal(t, sfi.Mode(), dfi.Mode())

	err = filepath.Walk(src, func(srcPath string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		relPath := strings.TrimPrefix(srcPath, src)
		dstPath := filepath.Join(dst, relPath)
		sfi, err := os.Stat(srcPath)
		require.NoError(t, err)

		dfi, err := os.Stat(dstPath)
		require.NoError(t, err)
		require.Equal(t, sfi.Mode(), dfi.Mode())

		if sfi.IsDir() {
			return nil
		}

		srcData, err := ioutil.ReadFile(srcPath)
		require.NoError(t, err)
		dstData, err := ioutil.ReadFile(dstPath)
		require.NoError(t, err)

		require.Equal(t, srcData, dstData)

		return nil
	})
	require.NoError(t, err)
}
