// Copyright 2021 PingCAP, Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package failpoint_test

import (
	"io/ioutil"
	"os"
	"testing"
	"time"

	"github.com/pingcap/errors"
	"github.com/pingcap/failpoint"
	"github.com/stretchr/testify/require"
)

func TestFailpoints(t *testing.T) {
	var fps failpoint.Failpoints

	err := fps.Enable("failpoints-test-1", "return(1)")
	require.NoError(t, err)
	val, err := fps.Eval("failpoints-test-1")
	require.NoError(t, err)
	require.Equal(t, 1, val.(int))

	err = fps.Enable("failpoints-test-2", "invalid")
	require.EqualError(t, err, `error on failpoints-test-2: failpoint: failed to parse "invalid" past "invalid"`)

	val, err = fps.Eval("failpoints-test-2")
	require.Error(t, err)
	require.Nil(t, val)

	err = fps.Disable("failpoints-test-1")
	require.NoError(t, err)

	val, err = fps.Eval("failpoints-test-1")
	require.Error(t, err)
	require.Nil(t, val)

	err = fps.Disable("failpoints-test-1")
	require.NoError(t, err)

	err = fps.Enable("failpoints-test-1", "return(1)")
	require.NoError(t, err)

	status, err := fps.Status("failpoints-test-1")
	require.NoError(t, err)
	require.Equal(t, "return(1)", status)

	err = fps.Enable("failpoints-test-3", "return(2)")
	require.NoError(t, err)

	ch := make(chan struct{})
	go func() {
		time.Sleep(time.Second)
		err := fps.Disable("gofail/testPause")
		require.NoError(t, err)
		close(ch)
	}()
	err = fps.Enable("gofail/testPause", "pause")
	require.NoError(t, err)
	start := time.Now()
	v, err := fps.Eval("gofail/testPause")
	require.NoError(t, err)
	require.Nil(t, v)

	require.True(t, time.Since(start) > 100*time.Millisecond)
	<-ch

	err = fps.Enable("failpoints-test-4", "50.0%return(5)")
	require.NoError(t, err)
	var succ int
	for i := 0; i < 1000; i++ {
		val, err = fps.Eval("failpoints-test-4")
		if err == nil && val != nil {
			succ++
			require.Equal(t, 5, val.(int))
		}
	}

	if succ < 450 || succ > 550 {
		require.FailNow(t, "prop failure: %v", succ)
	}

	err = fps.Enable("failpoints-test-5", "50*return(5)")
	require.NoError(t, err)
	for i := 0; i < 50; i++ {
		val, err = fps.Eval("failpoints-test-5")
		require.NoError(t, err)
		require.Equal(t, 5, val.(int))
	}
	val, err = fps.Eval("failpoints-test-5")
	require.Equal(t, failpoint.ErrNotAllowed, errors.Cause(err))
	require.Nil(t, val)

	points := map[string]struct{}{}
	for _, fp := range fps.List() {
		points[fp] = struct{}{}
	}
	require.Contains(t, points, "failpoints-test-1")
	require.Contains(t, points, "failpoints-test-2")
	require.Contains(t, points, "failpoints-test-3")
	require.Contains(t, points, "failpoints-test-4")
	require.Contains(t, points, "failpoints-test-5")

	err = fps.Enable("failpoints-test-6", "50*return(5)->1*return(true)->1*return(false)->10*return(20)")
	require.NoError(t, err)
	// 50*return(5)
	for i := 0; i < 50; i++ {
		val, err = fps.Eval("failpoints-test-6")
		require.NoError(t, err)
		require.Equal(t, 5, val.(int))
	}
	// 1*return(true)
	val, err = fps.Eval("failpoints-test-6")
	require.NoError(t, err)
	require.True(t, val.(bool))
	// 1*return(false)
	val, err = fps.Eval("failpoints-test-6")
	require.NoError(t, err)
	require.False(t, val.(bool))
	// 10*return(20)
	for i := 0; i < 10; i++ {
		val, err = fps.Eval("failpoints-test-6")
		require.NoError(t, err)
		require.Equal(t, 20, val.(int))
	}
	val, err = fps.Eval("failpoints-test-6")
	require.Equal(t, failpoint.ErrNotAllowed, errors.Cause(err))
	require.Nil(t, val)

	val, err = fps.Eval("failpoints-test-7")
	require.Equal(t, failpoint.ErrNotExist, errors.Cause(err))
	require.Nil(t, val)

	val, err = failpoint.Eval("failpoint-env1")
	require.NoError(t, err)
	require.Equal(t, 10, val.(int))
	val, err = failpoint.Eval("failpoint-env2")
	require.NoError(t, err)
	require.True(t, val.(bool))

	// Tests for sleep
	ch = make(chan struct{})
	go func() {
		defer close(ch)
		time.Sleep(time.Second)
		err := failpoint.Disable("gofail/test-sleep")
		require.NoError(t, err)
	}()
	err = failpoint.Enable("gofail/test-sleep", "sleep(100)")
	require.NoError(t, err)
	start = time.Now()
	v, err = failpoint.Eval("gofail/test-sleep")
	require.NoError(t, err)
	require.Nil(t, v)
	require.GreaterOrEqual(t, time.Since(start).Milliseconds(), int64(90))
	<-ch

	// Tests for sleep duration
	ch = make(chan struct{})
	go func() {
		defer close(ch)
		time.Sleep(time.Second)
		err := failpoint.Disable("gofail/test-sleep2")
		require.NoError(t, err)
	}()
	err = failpoint.Enable("gofail/test-sleep2", `sleep("100ms")`)
	require.NoError(t, err)
	start = time.Now()
	v, err = failpoint.Eval("gofail/test-sleep2")
	require.NoError(t, err)
	require.Nil(t, v)
	require.GreaterOrEqual(t, time.Since(start).Milliseconds(), int64(90))
	<-ch

	// Tests for print
	oldStdio := os.Stdout
	r, w, err := os.Pipe()
	require.NoError(t, err)
	os.Stdout = w
	err = fps.Enable("test-print", `print("hello world")`)
	require.NoError(t, err)
	val, err = fps.Eval("test-print")
	require.NoError(t, err)
	require.Nil(t, val)
	outC := make(chan string)
	// copy the output in a separate goroutine so printing can't block indefinitely
	go func() {
		defer close(outC)
		s, err := ioutil.ReadAll(r)
		require.NoError(t, err)
		outC <- string(s)
	}()
	require.NoError(t, w.Close())
	os.Stdout = oldStdio
	out := <-outC
	require.Equal(t, "failpoint print: hello world\n", out)

	// Tests for panic
	require.PanicsWithValue(t, "failpoint panic: {}", testPanic)

	err = fps.Enable("failpoints-test-7", `return`)
	require.NoError(t, err)
	val, err = fps.Eval("failpoints-test-7")
	require.NoError(t, err)
	require.Equal(t, struct{}{}, val)

	err = fps.Enable("failpoints-test-8", `return()`)
	require.NoError(t, err)
	val, err = fps.Eval("failpoints-test-8")
	require.NoError(t, err)
	require.Equal(t, struct{}{}, val)
}

func testPanic() {
	_ = failpoint.Enable("test-panic", `panic`)
	_, _ = failpoint.Eval("test-panic")
}
