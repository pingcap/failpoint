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

package code

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
)

const (
	failpointStashFileSuffix = "__failpoint_stash__"
	failpointBindingFileName = "binding__failpoint_binding__.go"
)

type Restorer struct {
	path string
}

// Restorer is represents a manager to restore currentFile tree which been modified by
// `failpoint-ctl enable`, e.g:
// ├── foo
// │   ├── foo.go
// │   └── foo.go__failpoint_stash__
// ├── bar
// │   ├── bar.go
// │   └── bar.go__failpoint_stash__
// └── foobar
//     ├── foobar.go
//     └── foobar.go__failpoint_stash__
// Which will be restored as below:
// ├── foo
// │   └── foo.go <- foo.go__failpoint_stash__
// ├── bar
// │   └── bar.go <- bar.go__failpoint_stash__
// └── foobar
//     └── foobar.go <- foobar.go__failpoint_stash__
func NewRestorer(path string) *Restorer {
	return &Restorer{path: path}
}

// Restore restores the currentFile tree which will delete all files generated
// by `failpoint-ctl enable` and replace it by fail point stashed currentFile
func (r Restorer) Restore() error {
	var stashFiles []string
	err := filepath.Walk(r.path, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			return nil
		}
		if strings.HasSuffix(path, failpointStashFileSuffix) ||
			strings.HasSuffix(path, failpointBindingFileName) {
			stashFiles = append(stashFiles, path)
		}
		return nil
	})
	if err != nil {
		return err
	}
	for _, filePath := range stashFiles {
		if strings.HasSuffix(filePath, failpointBindingFileName) {
			if err := os.Remove(filePath); err != nil {
				return err
			}
			continue
		}
		originFileName := filePath[:len(filePath)-len(failpointStashFileSuffix)]
		if err := os.Remove(originFileName); err != nil {
			return err
		}
		if err := os.Rename(filePath, originFileName); err != nil {
			return err
		}
	}
	return nil
}

func failpointBindingPath(path string) string {
	return filepath.Join(filepath.Dir(path), failpointBindingFileName)
}

func isBindingFileExists(path string) (bool, error) {
	bindingFile := failpointBindingPath(path)
	_, err := os.Stat(bindingFile)
	if err != nil && os.IsNotExist(err) {
		return false, nil
	}
	return true, err
}

func writeBindingFile(path, pak string) error {
	bindingFile := failpointBindingPath(path)
	bindingContent := fmt.Sprintf(`
package %s

import "reflect"

type __failpointBindingType struct {pkgpath string}
var __failpointBindingCache = &__failpointBindingType{}

func init() {
	__failpointBindingCache.pkgpath = reflect.TypeOf(__failpointBindingType{}).PkgPath()
}
func %s(name string) string {
	return  __failpointBindingCache.pkgpath + "/" + name
}
`, pak, extendPkgName)
	return ioutil.WriteFile(bindingFile, []byte(bindingContent), os.ModePerm)
}
