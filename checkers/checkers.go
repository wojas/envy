package checkers

import (
	"os"
	"path/filepath"

	"github.com/wojas/envy/action"
)

var AllCheckers = []Checker{
	BinCheck{"bin"},
	BinCheck{"node_modules/.bin"},
	BinCheck{".venv/bin"},
	GoPathCheck{},
}

type Checker interface {
	Check(path string) []action.Action
}

func IsDir(path string) bool {
	fi, err := os.Stat(path)
	if err != nil {
		return false
	}
	return fi.IsDir()
}

type BinCheck struct {
	RelPath string
}

func (c BinCheck) Check(path string) (actions []action.Action) {
	bin := filepath.Join(path, c.RelPath)
	if !IsDir(bin) {
		return
	}

	actions = append(actions, action.Action{
		Path:    path,
		AddPath: bin,
	})
	return
}

type GoPathCheck struct{}

func (c GoPathCheck) Check(path string) (actions []action.Action) {
	bin := filepath.Join(path, "bin")
	if !IsDir(bin) {
		return
	}
	src := filepath.Join(path, "src")
	if !IsDir(src) {
		return
	}
	pkg := filepath.Join(path, "pkg")
	if !IsDir(pkg) {
		return
	}

	actions = append(actions, action.Action{
		Path:        path,
		SetEnv:      "GOPATH",
		SetEnvValue: path,
	})
	return
}
