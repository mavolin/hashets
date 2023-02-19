package hashets

import (
	"io"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestWrapFS(t *testing.T) {
	t.Parallel()

	wrapFS, m, err := WrapFS(testdataIn, Options{})
	require.NoError(t, err)

	assert.Equal(t, expectMap, m)

	for origPath, hashedPath := range expectMap {
		expectFile, err := testdataIn.Open(origPath)
		require.NoError(t, err)
		defer expectFile.Close()

		expect, err := io.ReadAll(expectFile)
		require.NoError(t, err)

		actualFile, err := wrapFS.Open(hashedPath)
		require.NoError(t, err)
		defer actualFile.Close()

		actual, err := io.ReadAll(actualFile)
		require.NoError(t, err)

		assert.Equal(t, expect, actual)
	}
}
