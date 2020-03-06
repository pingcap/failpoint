package failpoint_test

import (
	"io/ioutil"
	"os"
	"time"

	. "github.com/pingcap/check"
	"github.com/pingcap/failpoint"
)

var _ = Suite(&failpointsSuite{})

type failpointsSuite struct{}

func (s *failpointsSuite) TestFailpoints(c *C) {
	var fps failpoint.Failpoints

	err := fps.Enable("failpoints-test-1", "return(1)")
	c.Assert(err, IsNil)
	val, err := fps.Eval("failpoints-test-1")
	c.Assert(err, IsNil)
	c.Assert(val.(int), Equals, 1)

	err = fps.Enable("failpoints-test-2", "invalid")
	c.Assert(err, ErrorMatches, `failpoint: failed to parse \"invalid\" past \"invalid\"`)

	val, err = fps.Eval("failpoints-test-2")
	c.Assert(err, NotNil)
	c.Assert(val, IsNil)

	err = fps.Disable("failpoints-test-1")
	c.Assert(err, IsNil)

	val, err = fps.Eval("failpoints-test-1")
	c.Assert(err, NotNil)
	c.Assert(val, IsNil)

	err = fps.Disable("failpoints-test-1")
	c.Assert(err, ErrorMatches, `failpoint: failpoint is disabled`)

	err = fps.Enable("failpoints-test-1", "return(1)")
	c.Assert(err, IsNil)

	status, err := fps.Status("failpoints-test-1")
	c.Assert(err, IsNil)
	c.Assert(status, Equals, "return(1)")

	err = fps.Enable("failpoints-test-3", "return(2)")
	c.Assert(err, IsNil)

	ch := make(chan struct{})
	go func() {
		time.Sleep(time.Second)
		err := fps.Disable("gofail/testPause")
		c.Assert(err, IsNil)
		close(ch)
	}()
	err = fps.Enable("gofail/testPause", "pause")
	c.Assert(err, IsNil)
	start := time.Now()
	v, err := fps.Eval("gofail/testPause")
	c.Assert(err, IsNil)
	c.Assert(v, IsNil)
	c.Assert(time.Since(start), GreaterEqual, 100*time.Millisecond, Commentf("not paused"))
	<-ch

	err = fps.Enable("failpoints-test-4", "50.0%return(5)")
	c.Assert(err, IsNil)
	var succ int
	for i := 0; i < 1000; i++ {
		val, err = fps.Eval("failpoints-test-4")
		if err == nil {
			succ++
			c.Assert(val.(int), Equals, 5)
		}
	}
	if succ < 450 || succ > 550 {
		c.Fatalf("prop failure: %v", succ)
	}

	err = fps.Enable("failpoints-test-5", "50*return(5)")
	c.Assert(err, IsNil)
	for i := 0; i < 50; i++ {
		val, err = fps.Eval("failpoints-test-5")
		c.Assert(err, IsNil)
		c.Assert(val.(int), Equals, 5)
	}
	val, err = fps.Eval("failpoints-test-5")
	c.Assert(err, NotNil)
	c.Assert(val, IsNil)

	points := map[string]struct{}{}
	for _, fp := range fps.List() {
		points[fp] = struct{}{}
	}
	c.Assert(points, HasKey, "failpoints-test-1")
	c.Assert(points, HasKey, "failpoints-test-2")
	c.Assert(points, HasKey, "failpoints-test-3")
	c.Assert(points, HasKey, "failpoints-test-4")
	c.Assert(points, HasKey, "failpoints-test-5")

	err = fps.Enable("failpoints-test-6", "50*return(5)->1*return(true)->1*return(false)->10*return(20)")
	c.Assert(err, IsNil)
	// 50*return(5)
	for i := 0; i < 50; i++ {
		val, err = fps.Eval("failpoints-test-6")
		c.Assert(err, IsNil)
		c.Assert(val.(int), Equals, 5)
	}
	// 1*return(true)
	val, err = fps.Eval("failpoints-test-6")
	c.Assert(err, IsNil)
	c.Assert(val.(bool), IsTrue)
	// 1*return(false)
	val, err = fps.Eval("failpoints-test-6")
	c.Assert(err, IsNil)
	c.Assert(val.(bool), IsFalse)
	// 10*return(20)
	for i := 0; i < 10; i++ {
		val, err = fps.Eval("failpoints-test-6")
		c.Assert(err, IsNil)
		c.Assert(val.(int), Equals, 20)
	}
	val, err = fps.Eval("failpoints-test-6")
	c.Assert(err, NotNil)
	c.Assert(val, IsNil)

	val, err = failpoint.Eval("failpoint-env1")
	c.Assert(err, IsNil)
	c.Assert(val.(int), Equals, 10)
	val, err = failpoint.Eval("failpoint-env2")
	c.Assert(err, IsNil)
	c.Assert(val.(bool), IsTrue)

	// Tests for sleep
	ch = make(chan struct{})
	go func() {
		defer close(ch)
		time.Sleep(time.Second)
		err := failpoint.Disable("gofail/test-sleep")
		c.Assert(err, IsNil)
	}()
	err = failpoint.Enable("gofail/test-sleep", "sleep(100)")
	c.Assert(err, IsNil)
	start = time.Now()
	v, err = failpoint.Eval("gofail/test-sleep")
	c.Assert(err, IsNil)
	c.Assert(v, IsNil)
	c.Assert(time.Since(start), GreaterEqual, 90*time.Millisecond, Commentf("not sleep"))
	<-ch

	// Tests for sleep duration
	ch = make(chan struct{})
	go func() {
		defer close(ch)
		time.Sleep(time.Second)
		err := failpoint.Disable("gofail/test-sleep2")
		c.Assert(err, IsNil)
	}()
	err = failpoint.Enable("gofail/test-sleep2", `sleep("100ms")`)
	c.Assert(err, IsNil)
	start = time.Now()
	v, err = failpoint.Eval("gofail/test-sleep2")
	c.Assert(err, IsNil)
	c.Assert(v, IsNil)
	c.Assert(time.Since(start), GreaterEqual, 90*time.Millisecond, Commentf("not sleep"))
	<-ch

	// Tests for print
	oldStdio := os.Stdout
	r, w, err := os.Pipe()
	c.Assert(err, IsNil)
	os.Stdout = w
	err = fps.Enable("test-print", `print("hello world")`)
	c.Assert(err, IsNil)
	val, err = fps.Eval("test-print")
	c.Assert(err, IsNil)
	c.Assert(val, IsNil)
	outC := make(chan string)
	// copy the output in a separate goroutine so printing can't block indefinitely
	go func() {
		defer close(outC)
		s, err := ioutil.ReadAll(r)
		c.Assert(err, IsNil)
		outC <- string(s)
	}()
	w.Close()
	os.Stdout = oldStdio
	out := <-outC
	c.Assert(out, Equals, "failpoint print: hello world\n")

	// Tests for panic
	c.Assert(testPanic, PanicMatches, "failpoint panic.*")

	err = fps.Enable("failpoints-test-7", `return`)
	c.Assert(err, IsNil)
	val, err = fps.Eval("failpoints-test-7")
	c.Assert(err, IsNil)
	c.Assert(val, Equals, struct{}{})

	err = fps.Enable("failpoints-test-8", `return()`)
	c.Assert(err, IsNil)
	val, err = fps.Eval("failpoints-test-8")
	c.Assert(err, IsNil)
	c.Assert(val, Equals, struct{}{})
}

func testPanic() {
	_ = failpoint.Enable("test-panic", `panic`)
	failpoint.Eval("test-panic")
}
