package code_test

import (
	"testing"

	. "github.com/pingcap/check"
	"github.com/pingcap/failpoint/code"
)

func TestNewRestorer(t *testing.T) {
	TestingT(t)
}

var _ = Suite(&restorerSuite{})

type restorerSuite struct{}

func (s *restorerSuite) TestRestore(c *C) {
	restorer := code.NewRestorer("not-exists-path")
	err := restorer.Restore()
	c.Assert(err, ErrorMatches, `lstat not-exists-path: no such file or directory`)
}
