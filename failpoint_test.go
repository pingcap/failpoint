package failpoint_test

import (
	"context"
	"sync"
	"testing"
	"time"

	. "github.com/pingcap/check"
	"github.com/pingcap/failpoint"
)

func TestT(t *testing.T) {
	TestingT(t)
}

var _ = Suite(&failpointSuite{})

type failpointSuite struct{}

func (s *failpointSuite) TestWithHook(c *C) {
	err := failpoint.Enable("TestWithHook-test-0", "return(1)")
	c.Assert(err, IsNil)

	val, err := failpoint.EvalContext(context.Background(), "TestWithHook-test-0")
	c.Assert(val, IsNil)
	c.Assert(err, NotNil)

	val, err = failpoint.EvalContext(nil, "TestWithHook-test-0")
	c.Assert(val, IsNil)
	c.Assert(err, NotNil)

	ctx := failpoint.WithHook(context.Background(), func(ctx context.Context, fpname string) bool {
		return false
	})
	val, err = failpoint.EvalContext(ctx, "unit-test")
	c.Assert(err, NotNil)
	c.Assert(val, IsNil)

	ctx = failpoint.WithHook(context.Background(), func(ctx context.Context, fpname string) bool {
		return true
	})
	err = failpoint.Enable("TestWithHook-test-1", "return(1)")
	c.Assert(err, IsNil)
	defer func() {
		err := failpoint.Disable("TestWithHook-test-1")
		c.Assert(err, IsNil)
	}()
	val, err = failpoint.EvalContext(ctx, "TestWithHook-test-1")
	c.Assert(err, IsNil)
	c.Assert(val.(int), Equals, 1)
}

func (s *failpointSuite) TestConcurrent(c *C) {
	err := failpoint.Enable("TestWithHook-test-2", "pause")
	c.Assert(err, IsNil)
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		ctx := failpoint.WithHook(context.Background(), func(ctx context.Context, fpname string) bool {
			return true
		})
		val, _ := failpoint.EvalContext(ctx, "TestWithHook-test-2")
		c.Assert(val, IsNil)
		wg.Done()
	}()
	time.Sleep(1 * time.Second)
	err = failpoint.Enable("TestWithHook-test-3", "return(1)")
	c.Assert(err, IsNil)
	err = failpoint.Disable("TestWithHook-test-3")
	c.Assert(err, IsNil)
	err = failpoint.Disable("TestWithHook-test-2")
	c.Assert(err, IsNil)
	wg.Wait()
}
