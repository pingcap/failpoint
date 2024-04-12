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

package main

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/pingcap/errors"
	"github.com/pingcap/failpoint/code"
	"golang.org/x/mod/modfile"
)

var logger = log.New(os.Stderr, "[failpoint-toolexec]", log.LstdFlags)

func main() {
	if len(os.Args) < 2 {
		return
	}
	goCmd, buildArgs := os.Args[1], os.Args[2:]
	goCmdBase := filepath.Base(goCmd)
	if runtime.GOOS == "windows" {
		goCmdBase = strings.TrimSuffix(goCmd, ".exe")
	}

	if strings.ToLower(goCmdBase) == "compile" {
		if err := injectFailpoint(&buildArgs); err != nil {
			logger.Println("failed to inject failpoint", err)
		}
	}

	cmd := exec.Command(goCmd, buildArgs...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		logger.Println("failed to run command", err)
	}
}

func injectFailpoint(argsP *[]string) error {
	callersModule, err := findCallersModule()
	if err != nil {
		return err
	}

	// ref https://pkg.go.dev/cmd/compile#hdr-Command_Line
	var module string
	args := *argsP
	for i, arg := range args {
		if arg == "-p" {
			if i+1 < len(args) {
				module = args[i+1]
			}
			break
		}
	}
	if !strings.HasPrefix(module, callersModule) && module != "main" {
		return nil
	}

	fileIndices := make([]int, 0, len(args))
	for i, arg := range args {
		// find the golang source files of the caller's package
		if strings.HasSuffix(arg, ".go") && !inSDKOrMod(arg) {
			fileIndices = append(fileIndices, i)
		}
	}

	needExtraFile := false
	writer := &code.Rewriter{}
	writer.SetAllowNotChecked(true)
	for _, idx := range fileIndices {
		needExtraFile = injectFailpointForFile(writer, &args[idx], module) || needExtraFile
	}
	if needExtraFile {
		newFile := filepath.Join(tmpFolder, module, "failpoint_toolexec_extra.go")
		if err := writeExtraFile(newFile, writer.GetCurrentFile().Name.Name, module); err != nil {
			return err
		}
		*argsP = append(args, newFile)
	}
	return nil
}

// ref https://github.com/golang/go/blob/bdd27c4debfb51fe42df0c0532c1c747777b7a32/src/cmd/go/internal/modload/init.go#L1511
func findCallersModule() (string, error) {
	cwd, err := os.Getwd()
	if err != nil {
		return "", err
	}
	dir := filepath.Clean(cwd)

	// Look for enclosing go.mod.
	for {
		goModPath := filepath.Join(dir, "go.mod")
		if fi, err := os.Stat(goModPath); err == nil && !fi.IsDir() {
			data, err := os.ReadFile(goModPath)
			if err != nil {
				return "", err
			}
			f, err := modfile.ParseLax(goModPath, data, nil)
			if err != nil {
				return "", err
			}
			return f.Module.Mod.Path, err
		}
		d := filepath.Dir(dir)
		if d == dir {
			break
		}
		dir = d
	}
	return "", errors.New("go.mod file not found")
}

var goModCache = os.Getenv("GOMODCACHE")
var goRoot = runtime.GOROOT()

func inSDKOrMod(path string) bool {
	absPath, err := filepath.Abs(path)
	if err != nil {
		logger.Println("failed to get absolute path", err)
		return false
	}

	if goModCache != "" && strings.HasPrefix(absPath, goModCache) {
		return true
	}
	if strings.HasPrefix(absPath, goRoot) {
		return true
	}
	return false
}

var tmpFolder = filepath.Join(os.TempDir(), "failpoint-toolexec")

func injectFailpointForFile(w *code.Rewriter, file *string, module string) bool {
	newFile := filepath.Join(tmpFolder, module, filepath.Base(*file))
	newFileDir := filepath.Dir(newFile)
	if err := os.MkdirAll(newFileDir, 0700); err != nil {
		logger.Println("failed to create temp folder", err)
		return false
	}
	f, err := os.OpenFile(newFile, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		logger.Println("failed to open temp file", err)
		return false
	}
	defer f.Close()
	w.SetOutput(f)

	if err := w.RewriteFile(*file); err != nil {
		logger.Println("failed to rewrite file", err)
		return false
	}
	if !w.GetRewritten() {
		return false
	}
	*file = newFile
	return true
}

func writeExtraFile(filePath, packageName, module string) error {
	bindingContent := fmt.Sprintf(`
package %s

func %s(name string) string {
	return "%s/" + name
}
`, packageName, code.ExtendPkgName, module)
	return os.WriteFile(filePath, []byte(bindingContent), 0644)
}
