// Copyright 2024 PingCAP, Inc.
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

package injectcall

import (
	"context"
	"testing"

	"github.com/pingcap/failpoint"
	"github.com/stretchr/testify/require"
)

func TestFoo(t *testing.T) {
	ctx := context.WithValue(context.Background(), "key", "ctx-value")
	ctx, cancel := context.WithCancel(ctx)
	var (
		capturedCtxVal   string
		capturedArgCount int
	)
	require.NoError(t, failpoint.EnableCall("github.com/pingcap/failpoint/examples/injectcall/test",
		func(ctx context.Context, i, count int) {
			if i == 5 {
				cancel()
				capturedCtxVal = ctx.Value("key").(string)
				capturedArgCount = count
			}
		},
	))
	t.Cleanup(func() {
		require.NoError(t, failpoint.Disable("github.com/pingcap/failpoint/examples/injectcall/test"))
	})

	loopCount := foo(ctx, 123)
	require.EqualValues(t, "ctx-value", capturedCtxVal)
	require.EqualValues(t, 5, loopCount)
	require.EqualValues(t, 123, capturedArgCount)
}
