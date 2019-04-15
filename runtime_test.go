package failpoint_test

import (
	"testing"
	"time"

	. "github.com/pingcap/check"
	"github.com/pingcap/failpoint"
)

func TestNewRestorer(t *testing.T) {
	TestingT(t)
}

var _ = Suite(&runtimeSuite{})

type runtimeSuite struct{}

func (s *runtimeSuite) TestRuntime(c *C) {
	err := failpoint.Enable("runtime-test-1", "return(1)")
	c.Assert(err, IsNil)
	ok, val := failpoint.Eval("runtime-test-1")
	c.Assert(ok, IsTrue)
	c.Assert(val.(int), Equals, 1)

	err = failpoint.Enable("runtime-test-2", "invalid")
	c.Assert(err, ErrorMatches, `failpoint: could not parse terms`)

	ok, val = failpoint.Eval("runtime-test-2")
	c.Assert(ok, IsFalse)

	err = failpoint.Disable("runtime-test-1")
	c.Assert(err, IsNil)

	ok, val = failpoint.Eval("runtime-test-1")
	c.Assert(ok, IsFalse)

	err = failpoint.Disable("runtime-test-1")
	c.Assert(err, ErrorMatches, `failpoint: failpoint is disabled`)

	err = failpoint.Enable("runtime-test-1", "return(1)")
	c.Assert(err, IsNil)

	status, err := failpoint.Status("runtime-test-1")
	c.Assert(err, IsNil)
	c.Assert(status, Equals, "return(1)")

	err = failpoint.Enable("runtime-test-3", "return(2)")
	c.Assert(err, IsNil)

	ch := make(chan struct{})
	go func() {
		time.Sleep(time.Second)
		err := failpoint.Disable("gofail/testPause")
		c.Assert(err, IsNil)
		close(ch)
	}()
	err = failpoint.Enable("gofail/testPause", "pause")
	c.Assert(err, IsNil)
	start := time.Now()
	ok, v := failpoint.Eval("gofail/testPause")
	c.Assert(ok, IsFalse)
	c.Assert(v, IsNil)
	c.Assert(time.Since(start), GreaterEqual, 100*time.Millisecond, Commentf("not paused"))
	<-ch

	err = failpoint.Enable("runtime-test-4", "50.0%return(5)")
	c.Assert(err, IsNil)
	var succ int
	for i := 0; i < 1000; i++ {
		ok, val = failpoint.Eval("runtime-test-4")
		if ok {
			succ++
			c.Assert(val.(int), Equals, 5)
		}
	}
	if succ < 450 || succ > 550 {
		c.Fatalf("prop failure: %v", succ)
	}

	err = failpoint.Enable("runtime-test-5", "50*return(5)")
	c.Assert(err, IsNil)
	for i := 0; i < 50; i++ {
		ok, val = failpoint.Eval("runtime-test-5")
		c.Assert(ok, Equals, true)
		c.Assert(val.(int), Equals, 5)
	}
	ok, val = failpoint.Eval("runtime-test-5")
	c.Assert(ok, IsFalse)

	fps := map[string]struct{}{}
	for _, fp := range failpoint.List() {
		fps[fp] = struct{}{}
	}
	c.Assert(fps, HasKey, "runtime-test-1")
	c.Assert(fps, HasKey, "runtime-test-2")
	c.Assert(fps, HasKey, "runtime-test-3")
	c.Assert(fps, HasKey, "runtime-test-4")
	c.Assert(fps, HasKey, "runtime-test-5")

	err = failpoint.Enable("runtime-test-6", "50*return(5)->1*return(true)->1*return(false)->10*return(20)")
	c.Assert(err, IsNil)
	// 50*return(5)
	for i := 0; i < 50; i++ {
		ok, val = failpoint.Eval("runtime-test-6")
		c.Assert(ok, IsTrue)
		c.Assert(val.(int), Equals, 5)
	}
	// 1*return(true)
	ok, val = failpoint.Eval("runtime-test-6")
	c.Assert(ok, IsTrue)
	c.Assert(val.(bool), Equals, true)
	// 1*return(false)
	ok, val = failpoint.Eval("runtime-test-6")
	c.Assert(ok, IsTrue)
	c.Assert(val.(bool), Equals, false)
	// 10*return(20)
	for i := 0; i < 10; i++ {
		ok, val = failpoint.Eval("runtime-test-6")
		c.Assert(ok, IsTrue)
		c.Assert(val.(int), Equals, 20)
	}
	ok, val = failpoint.Eval("runtime-test-6")
	c.Assert(ok, IsFalse)

	ok, val = failpoint.Eval("failpoint-env1")
	c.Assert(ok, IsTrue)
	c.Assert(val.(int), Equals, 10)
	ok, val = failpoint.Eval("failpoint-env2")
	c.Assert(ok, IsTrue)
	c.Assert(val.(bool), Equals, true)
}
