package code_test

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"

	"github.com/pingcap/failpoint/code"
	"github.com/stretchr/testify/require"
)

func TestRewriteFileResetRewrittenStateForNoDeclFile(t *testing.T) {
	t.Parallel()

	workDir := t.TempDir()
	failpointFile := filepath.Join(workDir, "with_failpoint.go")
	docFile := filepath.Join(workDir, "doc.go")

	require.NoError(t, os.WriteFile(failpointFile, []byte(`package sample

import "github.com/pingcap/failpoint"

func f() {
	failpoint.Inject("fp", func() {})
}
`), 0o644))
	require.NoError(t, os.WriteFile(docFile, []byte("package sample\n"), 0o644))

	rewriter := code.NewRewriter(workDir)
	rewriter.SetAllowNotChecked(true)

	var out bytes.Buffer

	rewriter.SetOutput(&out)
	require.NoError(t, rewriter.RewriteFile(failpointFile))
	require.True(t, rewriter.GetRewritten())

	out.Reset()
	rewriter.SetOutput(&out)
	require.NoError(t, rewriter.RewriteFile(docFile))
	require.False(t, rewriter.GetRewritten())
	require.Zero(t, out.Len())
}
