package failpoint_test

import (
	"context"
	"testing"

	. "github.com/pingcap/check"
	"github.com/pingcap/failpoint"
)

func TestFailpoint(t *testing.T) {
	TestingT(t)
}

var _ = &failpointSuite{}

type failpointSuite struct{}

func (s *failpointSuite) TestWithHook(c *C) {
	ctx := failpoint.WithHook(context.Background(), func(ctx context.Context, fpname string) bool {
		return false
	})
	ok, val := failpoint.Eval("unit-test", ctx)
	c.Assert(ok, Equals, false)
	c.Assert(val, Equals, nil)

	ctx = failpoint.WithHook(context.Background(), func(ctx context.Context, fpname string) bool {
		return true
	})
	err := failpoint.Enable("TestWithHook-test-1", "return(1)")
	defer func() {
		err := failpoint.Disable("TestWithHook-test-1")
		c.Assert(err, Equals, nil)
	}()
	c.Assert(err, Equals, nil)
	ok, val = failpoint.Eval("TestWithHook-test-1", ctx)
	c.Assert(ok, Equals, true)
	c.Assert(val.(int), Equals, 1)
}