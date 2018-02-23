package checkers

import (
	"path/filepath"

	"github.com/wojas/envy/action"
	"github.com/wojas/envy/paths"
)

// AllCheckers is a list of all Checker interfaces to use.
var AllCheckers = []Checker{
	BinCheck{"bin"},
	BinCheck{"node_modules/.bin"},
	BinCheck{".venv/bin"},
	DotEnvCheck{".envy"},
	GoPathCheck{},
	GitRootCheck{},
}

// Checker is the interface shared by functions that check for Actions to take
// for the given path.
type Checker interface {
	Check(path string) action.List
}

// BinCheck checks for a directory with executables to add to the PATH.
type BinCheck struct {
	RelPath string
}

// Check implements the Checker interface.
func (c BinCheck) Check(path string) (actions action.List) {
	bin := filepath.Join(path, c.RelPath)
	if !paths.IsDir(bin) {
		return
	}

	actions = append(actions, action.Action{
		Path:    path,
		AddPath: bin,
	})
	return
}

// GoPathCheck checks if a path is a GOPATH.
type GoPathCheck struct{}

// Check implements the Checker interface.
func (c GoPathCheck) Check(path string) (actions action.List) {
	bin := filepath.Join(path, "bin")
	if !paths.IsDir(bin) {
		return
	}
	src := filepath.Join(path, "src")
	if !paths.IsDir(src) {
		return
	}
	pkg := filepath.Join(path, "pkg")
	if !paths.IsDir(pkg) {
		return
	}

	actions = append(actions, action.Action{
		Path:        path,
		SetEnv:      "GOPATH",
		SetEnvValue: path,
	})
	return
}
