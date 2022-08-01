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

package code_test

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/pingcap/failpoint/code"
	"github.com/stretchr/testify/require"
)

func TestRestore(t *testing.T) {
	restorer := code.NewRestorer("not-exists-path")
	err := restorer.Restore()
	require.EqualError(t, err, `lstat not-exists-path: no such file or directory`)
}

func TestRestoreModification(t *testing.T) {
	var cases = []struct {
		filepath string
		original string
		modified string
		expected string
	}{
		{
			filepath: "modified-test.go",
			original: `
package rewriter_test

import (
	"fmt"

	"github.com/pingcap/failpoint"
)

func unittest() {
	failpoint.Inject("failpoint-name", func(val failpoint.Value) {
		fmt.Println("unit-test", val)
	})
}
`,
			modified: `
package rewriter_test

import (
	"fmt"

	"github.com/pingcap/failpoint"
)

func unittest() {
	if val, _err_ := failpoint.Eval(_curpkg_("failpoint-name")); _err_ == nil {
		fmt.Println("extra add line")
		fmt.Println("unit-test", val)
	}
}
`,
			expected: `
package rewriter_test

import (
	"fmt"

	"github.com/pingcap/failpoint"
)

func unittest() {
	failpoint.Inject("failpoint-name", func(val failpoint.Value) {
		fmt.Println("extra add line")
		fmt.Println("unit-test", val)
	})
}
`,
		},

		{
			filepath: "modified-test-2.go",
			original: `
package rewriter_test

import (
	"fmt"

	"github.com/pingcap/failpoint"
)

func unittest() {
	failpoint.Inject("failpoint-name", func(val failpoint.Value) {
		fmt.Println("unit-test", val)
	})
}
`,
			modified: `
package rewriter_test

import (
	"fmt"

	"github.com/pingcap/failpoint"
)

func unittest() {
	if val, _err_ := failpoint.Eval(_curpkg_("failpoint-name")); _err_ == nil {
		fmt.Println("extra add line")
		fmt.Println("unit-test", val)
	}
	fmt.Println("extra add line2")
}
`,
			expected: `
package rewriter_test

import (
	"fmt"

	"github.com/pingcap/failpoint"
)

func unittest() {
	failpoint.Inject("failpoint-name", func(val failpoint.Value) {
		fmt.Println("extra add line")
		fmt.Println("unit-test", val)
	})
	fmt.Println("extra add line2")
}
`,
		},

		{
			filepath: "modified-test-3.go",
			original: `
package rewriter_test

import (
	"fmt"

	"github.com/pingcap/failpoint"
)

func unittest() {
	failpoint.Inject("failpoint-name", func(val failpoint.Value) {
		fmt.Println("unit-test", val)
	})
}
`,
			modified: `
package rewriter_test

import (
	"fmt"

	"github.com/pingcap/failpoint"
)

func unittest() {
	if val, _err_ := failpoint.Eval(_curpkg_("failpoint-name-extra-part")); _err_ == nil {
		fmt.Println("extra add line")
		fmt.Println("unit-test", val)
	}
	fmt.Println("extra add line2")
}
`,
			expected: `
package rewriter_test

import (
	"fmt"

	"github.com/pingcap/failpoint"
)

func unittest() {
	failpoint.Inject("failpoint-name-extra-part", func(val failpoint.Value) {
		fmt.Println("extra add line")
		fmt.Println("unit-test", val)
	})
	fmt.Println("extra add line2")
}
`,
		},
	}

	// Create temp files
	err := os.MkdirAll(restorePath, 0755)
	require.NoError(t, err)
	for _, cs := range cases {
		original := filepath.Join(restorePath, cs.filepath)
		err := ioutil.WriteFile(original, []byte(cs.original), 0644)
		require.NoError(t, err)
	}

	// Clean all temp files
	defer func() {
		err := os.RemoveAll(restorePath)
		require.NoError(t, err)
	}()

	rewriter := code.NewRewriter(restorePath)
	err = rewriter.Rewrite()
	require.NoError(t, err)

	for _, cs := range cases {
		modified := filepath.Join(restorePath, cs.filepath)
		err := ioutil.WriteFile(modified, []byte(cs.modified), 0644)
		require.NoError(t, err)
	}

	// Restore workspace
	restorer := code.NewRestorer(restorePath)
	err = restorer.Restore()
	require.NoError(t, err)

	for _, cs := range cases {
		expected := filepath.Join(restorePath, cs.filepath)
		content, err := ioutil.ReadFile(expected)
		require.NoError(t, err)
		require.Equalf(t, strings.TrimSpace(cs.expected), strings.TrimSpace(string(content)), "%v", cs.filepath)
	}
}

func TestRestoreModificationBad(t *testing.T) {
	var cases = []struct {
		filepath string
		original string
		modified string
	}{
		{
			filepath: "bad-modification-test.go",
			original: `
package rewriter_test

import (
	"fmt"

	"github.com/pingcap/failpoint"
)

func unittest() {
	failpoint.Inject("failpoint-name", func(val failpoint.Value) {
		fmt.Println("unit-test", val)
	})
}
`,
			modified: `
package rewriter_test

import (
	"fmt"

	"github.com/pingcap/failpoint"
)

func unittest() {
	if val, _err_ := failpoint.EvalContext(nil, _curpkg_("failpoint-name-extra-part")); _err_ == nil {
		fmt.Println("extra add line")
		fmt.Println("unit-test", val)
	}
}
`,
		},
	}

	// Create temp files
	err := os.MkdirAll(restorePath, 0755)
	require.NoError(t, err)
	for _, cs := range cases {
		original := filepath.Join(restorePath, cs.filepath)
		err := ioutil.WriteFile(original, []byte(cs.original), 0644)
		require.NoError(t, err)
	}

	// Clean all temp files
	defer func() {
		err := os.RemoveAll(restorePath)
		require.NoError(t, err)
	}()

	rewriter := code.NewRewriter(restorePath)
	err = rewriter.Rewrite()
	require.NoError(t, err)

	for _, cs := range cases {
		modified := filepath.Join(restorePath, cs.filepath)
		err := ioutil.WriteFile(modified, []byte(cs.modified), 0644)
		require.NoError(t, err)
	}

	restorer := code.NewRestorer(restorePath)
	err = restorer.Restore()
	require.Error(t, err)
	require.Regexp(t, `cannot merge modifications back automatically.*`, err.Error())
}
