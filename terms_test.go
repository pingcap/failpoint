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
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
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
	"reflect"
	"testing"
)

func TestTermsString(t *testing.T) {
	tests := []struct {
		desc  string
		weval []string
	}{
		{`off`, []string{""}},
		{`2*return("abc")`, []string{"abc", "abc", ""}},
		{`0%return("abc")`, []string{"", "", ""}},
		{`100%return("abc")`, []string{"abc", "abc", "abc"}},
		{`100.0%return("abc")`, []string{"abc", "abc", "abc"}},
		{`100%2*return("abc")`, []string{"abc", "abc", ""}},
		{`2*return("abc")->1*return("def")`, []string{"abc", "abc", "def", ""}},
		{`1*return("abc")->return("def")`, []string{"abc", "def", "def"}},
	}
	for _, tt := range tests {
		ter, err := newTerms(tt.desc, nil)
		if err != nil {
			t.Fatal(err)
		}
		for _, w := range tt.weval {
			v, err := ter.eval()
			if v == nil && w == "" {
				continue
			}
			if err != nil {
				continue
			}
			if v.(string) != w {
				t.Fatalf("got %q, expected %q", v, w)
			}
		}
	}
}

func TestTermsTypes(t *testing.T) {
	tests := []struct {
		desc  string
		weval interface{}
	}{
		{`off`, nil},
		{`return("abc")`, "abc"},
		{`return(true)`, true},
		{`return(1)`, 1},
		{`return()`, struct{}{}},
	}
	for _, tt := range tests {
		ter, err := newTerms(tt.desc, nil)
		if err != nil {
			t.Fatal(err)
		}
		v, _ := ter.eval()
		if v == nil && tt.weval == nil {
			continue
		}
		if !reflect.DeepEqual(v, tt.weval) {
			t.Fatalf("got %v, expected %v", v, tt.weval)
		}
	}
}

func TestTermsRejectsMissingAction(t *testing.T) {
	const fpname = "github.com/pingcap/failpoint/missingAction"
	// A term is "<mod> <act> [ ( <val> ) ]". With the action left out there
	// is nothing to run, and the chain used to be accepted and then call a
	// nil act func on the first evaluation.
	for _, desc := range []string{`1*`, `100%`, `1*2*`, `100.0%`} {
		if _, err := newTerms(desc, nil); err == nil {
			t.Errorf("newTerms(%q) was accepted, want a parse error", desc)
		}
		if err := Enable(fpname, desc); err != nil {
			continue
		}
		func() {
			defer func() {
				if r := recover(); r != nil {
					t.Errorf("Enable(%q) was accepted and the failpoint then panicked: %v", desc, r)
				}
			}()
			_, _ = Eval(fpname)
			t.Errorf("Enable(%q) was accepted, want a parse error", desc)
		}()
		if err := Disable(fpname); err != nil {
			t.Fatal(err)
		}
	}
}
