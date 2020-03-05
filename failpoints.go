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

// Copyright 2016 CoreOS, Inc.
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

package failpoint

import (
	"context"
	"fmt"
	"os"
	"sort"
	"strings"
	"sync"
)

var (
	// ErrNoExist represents can not found a failpoint by specified name
	ErrNoExist = fmt.Errorf("failpoint: failpoint does not exist")
	// ErrDisabled represents a failpoint is be disabled
	ErrDisabled = fmt.Errorf("failpoint: failpoint is disabled")
)

func init() {
	failpoints.reg = make(map[string]*Failpoint)
	if s := os.Getenv("GO_FAILPOINTS"); len(s) > 0 {
		// format is <FAILPOINT>=<TERMS>[;<FAILPOINT>=<TERMS>;...]
		for _, fp := range strings.Split(s, ";") {
			fpTerms := strings.Split(fp, "=")
			if len(fpTerms) != 2 {
				fmt.Printf("bad failpoint %q\n", fp)
				os.Exit(1)
			}
			err := Enable(fpTerms[0], fpTerms[1])
			if err != nil {
				fmt.Printf("bad failpoint %s\n", err)
				os.Exit(1)
			}
		}
	}
	if s := os.Getenv("GO_FAILPOINTS_HTTP"); len(s) > 0 {
		if err := serve(s); err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
	}
}

// Failpoints manages multiple failpoints
type Failpoints struct {
	mu  sync.RWMutex
	reg map[string]*Failpoint
}

// Enable a failpoint on failpath
func (fps *Failpoints) Enable(failpath, inTerms string) error {
	fps.mu.Lock()
	fps.mu.Unlock()

	fp := fps.reg[failpath]
	if fp == nil {
		fp = &Failpoint{}
		fps.reg[failpath] = fp
	}
	return fp.Enable(inTerms)
}

// Disable a failpoint on failpath
func (fps *Failpoints) Disable(failpath string) error {
	fps.mu.Lock()
	fps.mu.Unlock()

	fp := fps.reg[failpath]
	if fp == nil {
		return ErrDisabled
	}
	return fp.Disable()
}

// Status gives the current setting for the failpoint
func (fps *Failpoints) Status(failpath string) (string, error) {
	fps.mu.RLock()
	fp := fps.reg[failpath]
	fps.mu.RUnlock()
	if fp == nil {
		return "", ErrNoExist
	}
	fp.mu.RLock()
	t := fp.t
	fp.mu.RUnlock()
	if t == nil {
		return "", ErrDisabled
	}
	return t.desc, nil
}

// List returns all the failpoints information
func (fps *Failpoints) List() []string {
	fps.mu.RLock()
	ret := make([]string, 0, len(failpoints.reg))
	for fp := range fps.reg {
		ret = append(ret, fp)
	}
	fps.mu.RUnlock()
	sort.Strings(ret)
	return ret
}

// EvalContext evaluates a failpoint's value, and calls hook if the context is
// not nil and contains hook function. It will return the evaluated value and
// true if the failpoint is active. Always returns false if ctx is nil
// or context does not contains a hook function
func (fps *Failpoints) EvalContext(ctx context.Context, failpath string) (Value, bool) {
	if ctx == nil {
		return nil, false
	}
	hook, ok := ctx.Value(failpointCtxKey).(Hook)
	if !ok {
		return nil, false
	}
	if !hook(ctx, failpath) {
		return nil, false
	}
	return fps.Eval(failpath)
}

// Eval evaluates a failpoint's value, It will return the evaluated value and
// true if the failpoint is active
func (fps *Failpoints) Eval(failpath string) (Value, bool) {
	fps.mu.RLock()
	fp, found := failpoints.reg[failpath]
	fps.mu.RUnlock()
	if !found {
		return nil, false
	}

	return fp.Eval()
}

// failpoints is the default
var failpoints Failpoints

// Enable sets a failpoint to a given failpoint description.
func Enable(failpath, inTerms string) error {
	return failpoints.Enable(failpath, inTerms)
}

// Disable stops a failpoint from firing.
func Disable(failpath string) error {
	return failpoints.Disable(failpath)
}

// Status gives the current setting for the failpoint
func Status(failpath string) (string, error) {
	return failpoints.Status(failpath)
}

// List returns all the failpoints information
func List() []string {
	return failpoints.List()
}

// WithHook binds a hook to a new context which is based on the `ctx` parameter
func WithHook(ctx context.Context, hook Hook) context.Context {
	return context.WithValue(ctx, failpointCtxKey, hook)
}

// EvalContext evaluates a failpoint's value, and calls hook if the context is
// not nil and contains hook function. It will return the evaluated value and
// true if the failpoint is active. Always returns false if ctx is nil
// or context does not contains hook function
func EvalContext(ctx context.Context, failpath string) (Value, bool) {
	return failpoints.EvalContext(ctx, failpath)
}

// Eval evaluates a failpoint's value, It will return the evaluated value and
// true if the failpoint is active
func Eval(failpath string) (Value, bool) {
	return failpoints.Eval(failpath)
}
