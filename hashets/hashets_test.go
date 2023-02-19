package hashets

import (
	"errors"
	"io"
	"io/fs"
	"os"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestHash(t *testing.T) {
	t.Parallel()

	t.Run("no options", func(t *testing.T) {
		t.Parallel()

		actual, err := Hash(testdataIn, Options{})
		require.NoError(t, err)

		assert.Equal(t, expectMap, actual)
	})

	t.Run("with ignore func", func(t *testing.T) {
		t.Parallel()

		actual, err := Hash(testdataIn, Options{
			Ignore: IgnorePrefix("folder"),
		})
		require.NoError(t, err)

		expect := make(Map, len(expectMap)-1)
		for k, v := range expectMap {
			if !strings.HasPrefix(k, "folder") {
				expect[k] = v
			}
		}

		assert.Equal(t, expect, actual)
	})
}

func TestHashToDir(t *testing.T) {
	t.Parallel()

	var (
		testdataIn     = os.DirFS("../testdata/in")
		testdataExpect = os.DirFS("../testdata/expect")
	)

	dir := t.TempDir()

	actual, err := HashToDir(testdataIn, dir, Options{})
	require.NoError(t, err)

	assert.Equal(t, expectMap, actual)
	dirsEqual(t, testdataExpect, os.DirFS(dir))
}

func TestHashToTempDir(t *testing.T) {
	t.Parallel()

	var (
		testdataIn     = os.DirFS("../testdata/in")
		testdataExpect = os.DirFS("../testdata/expect")
	)

	actualFS, actualMap, cleanup, err := HashToTempDir(testdataIn, Options{})
	require.NoError(t, err)
	t.Cleanup(func() {
		require.NoError(t, cleanup())
	})

	assert.Equal(t, expectMap, actualMap)
	dirsEqual(t, testdataExpect, actualFS)
}

func dirsEqual(t *testing.T, a, b fs.FS) {
	_ = fs.WalkDir(a, ".", func(path string, d fs.DirEntry, err error) error {
		var statA, statB fs.FileInfo

		if path != "." {
			statA, err = fs.Stat(a, path)
			require.NoError(t, err)

			statB, err = fs.Stat(b, path)
			if err != nil {
				if errors.Is(err, fs.ErrNotExist) {
					assert.Failf(t, "a: %s: not found", path)
					return nil
				} else {
					require.NoError(t, err)
				}
			}

			if statA.IsDir() != statB.IsDir() {
				assert.Fail(t, "%s: IsDir() mismatch", path)
			}
		}

		if path == "." || statA.IsDir() {
			dirA, err := fs.ReadDir(a, path)
			require.NoError(t, err)

			dirB, err := fs.ReadDir(b, path)
			require.NoError(t, err)

			assert.Equalf(t, len(dirA), len(dirB), "%s: len mismatch", path)
		} else {
			fileA, err := a.Open(path)
			require.NoError(t, err)

			fileB, err := b.Open(path)
			require.NoError(t, err)

			dataA, err := io.ReadAll(fileA)
			if err != nil {
				return err
			}

			dataB, err := io.ReadAll(fileB)
			if err != nil {
				return err
			}

			assert.Equal(t, dataA, dataB)
		}

		return nil
	})
}
