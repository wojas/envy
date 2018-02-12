package env

import (
	"os"
	"path/filepath"
	"sync"

	"github.com/wojas/envy/action"
)

var AllCheckers = []Checker{
	BinCheck{},
	NodeModulesCheck{},
	PythonVEnvCheck{},
	GoPathCheck{},
}

type Env struct {
	Path     string
	ActionsC chan action.Action
	WG       *sync.WaitGroup
}

func (e *Env) Add(a action.Action) {
	a.Path = e.Path
	e.ActionsC <- a
}

func (e *Env) Done() {
	e.WG.Done()
}

type Checker interface {
	Check(e Env)
}

func IsDir(path string) bool {
	fi, err := os.Stat(path)
	if err != nil {
		return false
	}
	return fi.IsDir()
}

type BinCheck struct{}

func (c BinCheck) Check(e Env) {
	defer e.Done()

	bin := filepath.Join(e.Path, "bin")
	if !IsDir(bin) {
		return
	}

	e.Add(action.Action{
		AddPath: bin,
	})
}

type NodeModulesCheck struct{}

func (c NodeModulesCheck) Check(e Env) {
	defer e.Done()

	p := filepath.Join(e.Path, "node_modules", ".bin")
	if !IsDir(p) {
		return
	}

	e.Add(action.Action{
		AddPath: p,
	})
}

type PythonVEnvCheck struct{}

func (c PythonVEnvCheck) Check(e Env) {
	defer e.Done()

	bin := filepath.Join(e.Path, ".venv", "bin")
	if !IsDir(bin) {
		return
	}

	e.Add(action.Action{
		AddPath: bin,
	})
}

type GoPathCheck struct{}

func (c GoPathCheck) Check(e Env) {
	defer e.Done()

	bin := filepath.Join(e.Path, "bin")
	if !IsDir(bin) {
		return
	}
	src := filepath.Join(e.Path, "src")
	if !IsDir(src) {
		return
	}
	pkg := filepath.Join(e.Path, "pkg")
	if !IsDir(pkg) {
		return
	}

	e.Add(action.Action{
		SetEnv:      "GOPATH",
		SetEnvValue: e.Path,
	})
}
