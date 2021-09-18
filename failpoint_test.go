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
	"context"
	"sync"
	"testing"
	"time"

	"github.com/pingcap/failpoint"
	"github.com/stretchr/testify/require"
)

func TestWithHook(t *testing.T) {
	err := failpoint.Enable("TestWithHook-test-0", "return(1)")
	require.NoError(t, err)

	val, err := failpoint.EvalContext(context.Background(), "TestWithHook-test-0")
	require.Nil(t, val)
	require.Error(t, err)

	val, err = failpoint.EvalContext(nil, "TestWithHook-test-0")
	require.Nil(t, val)
	require.Error(t, err)

	ctx := failpoint.WithHook(context.Background(), func(ctx context.Context, fpname string) bool {
		return false
	})
	val, err = failpoint.EvalContext(ctx, "unit-test")
	require.Error(t, err)
	require.Nil(t, val)

	ctx = failpoint.WithHook(context.Background(), func(ctx context.Context, fpname string) bool {
		return true
	})
	err = failpoint.Enable("TestWithHook-test-1", "return(1)")
	require.NoError(t, err)
	defer func() {
		err := failpoint.Disable("TestWithHook-test-1")
		require.NoError(t, err)
	}()
	val, err = failpoint.EvalContext(ctx, "TestWithHook-test-1")
	require.NoError(t, err)
	require.Equal(t, 1, val.(int))
}

func TestConcurrent(t *testing.T) {
	err := failpoint.Enable("TestWithHook-test-2", "pause")
	require.NoError(t, err)
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		ctx := failpoint.WithHook(context.Background(), func(ctx context.Context, fpname string) bool {
			return true
		})
		val, _ := failpoint.EvalContext(ctx, "TestWithHook-test-2")
		require.Nil(t, val)
		wg.Done()
	}()
	time.Sleep(1 * time.Second)
	err = failpoint.Enable("TestWithHook-test-3", "return(1)")
	require.NoError(t, err)
	err = failpoint.Disable("TestWithHook-test-3")
	require.NoError(t, err)
	err = failpoint.Disable("TestWithHook-test-2")
	require.NoError(t, err)
	wg.Wait()
}
