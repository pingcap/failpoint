package code

import (
	"os"
	"path/filepath"
	"strings"
)

const (
	failpointStashFileSuffix = "__failpoint_stash__"
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
		if strings.HasSuffix(path, failpointStashFileSuffix) {
			stashFiles = append(stashFiles, path)
		}
		return nil
	})
	if err != nil {
		return err
	}
	for _, filePath := range stashFiles {
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
