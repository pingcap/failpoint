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
	ok, val := failpoint.EvalContext(ctx, "unit-test")
	c.Assert(ok, IsFalse)
	c.Assert(val, IsNil)

	ctx = failpoint.WithHook(context.Background(), func(ctx context.Context, fpname string) bool {
		return true
	})
	err := failpoint.Enable("TestWithHook-test-1", "return(1)")
	defer func() {
		err := failpoint.Disable("TestWithHook-test-1")
		c.Assert(err, IsNil)
	}()
	c.Assert(err, IsNil)
	ok, val = failpoint.EvalContext(ctx, "TestWithHook-test-1")
	c.Assert(ok, IsTrue)
	c.Assert(val.(int), Equals, 1)
}
