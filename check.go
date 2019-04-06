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

import "context"

const failpointCtxKey = "__failpoint_ctx_key__"

type Hook func(ctx context.Context, fpname string) bool

func WithHook(ctx context.Context, hook Hook) context.Context {
	return context.WithValue(ctx, failpointCtxKey, hook)
}

type Arg struct {
}

var activeFPs = map[string]*Arg{}

func IsActive(ctx context.Context, fpname string) (bool, *Arg) {
	if ctx != nil {
		hook := ctx.Value(failpointCtxKey)
		if hook != nil {
			h, ok := hook.(Hook)
			if ok && !h(ctx, fpname) {
				return false, nil
			}
		}
	}

	// TODO: implement check fail point activity algorithm
	arg, found := activeFPs[fpname]
	return found, arg
}

func (a *Arg) Int() int {
	return 0
}

func (a *Arg) String() string {
	return ""
}
