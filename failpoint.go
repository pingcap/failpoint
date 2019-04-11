// Copyright 2019 PingCAP, Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// See the License for the specific language governing permissions and
// limitations under the License.

package failpoint

import (
	"context"
	"sync"
)

const failpointCtxKey string = "__failpoint_ctx_key__"

type (
	// Value represents value that retrieved from failpoint terms.
	// It can be used as following types:
	// 1. val.(int)      // GO_FAILPOINTS="failpoint-name=return(1)"
	// 2. val.(string)   // GO_FAILPOINTS="failpoint-name=return('1')"
	// 3. val.(bool)     // GO_FAILPOINTS="failpoint-name=return(true)"
	Value interface{}

	// Hook is used to filter failpoint, if the hook returns false and the
	// failpoint will not to be evaluated.
	Hook func(ctx context.Context, fpname string) bool

	failpoint struct {
		mu       sync.RWMutex
		t        *terms
		waitChan chan struct{}
	}
)

// Pause will pause until the failpoint is disabled.
func (fp *failpoint) Pause() {
	<-fp.waitChan
}

// WithHook binding a hook to a new context which base on the `ctx` parameter
func WithHook(ctx context.Context, hook Hook) context.Context {
	return context.WithValue(ctx, failpointCtxKey, hook)
}

// EvalContext evaluates a failpoint's value, will call hook if the context is
// not nil and contains hook function. It will return true and the evaluated
// value if the failpoint is active
func EvalContext(ctx context.Context, fpname string) (bool, Value) {
	if ctx != nil {
		hook := ctx.Value(failpointCtxKey)
		if hook != nil {
			h, ok := hook.(Hook)
			if ok && !h(ctx, fpname) {
				return false, nil
			}
		}
	}
	return Eval(fpname)
}

// Eval evaluates a failpoint's value, It will return true and the evaluated
// value if the failpoint is active
func Eval(fpname string) (bool, Value) {
	failpoints.mu.RLock()
	defer failpoints.mu.RUnlock()
	fp, found := failpoints.reg[fpname]
	if !found {
		return false, nil
	}

	fp.mu.RLock()
	defer fp.mu.RUnlock()
	if fp.t == nil {
		return false, nil
	}
	v := fp.t.eval()
	if v == nil {
		return false, nil
	}
	return true, v
}
