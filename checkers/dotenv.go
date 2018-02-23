package checkers

import (
	"log"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/subosito/gotenv" // TODO: find a better one
	"github.com/wojas/envy/env"

	"github.com/wojas/envy/action"
	"github.com/wojas/envy/paths"
)

// DotEnvCheck checks for a .env file and loads the variables defined there
type DotEnvCheck struct {
	RelPath string
}

// Check implements the Checker interface.
func (c DotEnvCheck) Check(path string) (actions action.List) {
	p := filepath.Join(path, c.RelPath)
	if !paths.IsFile(p) {
		return
	}

	f, err := os.Open(p)
	if err != nil {
		log.Printf("Warning: could not open %s: %v", p, err)
		return
	}
	defer f.Close()

	// TODO: better parser that does not stop on error
	// TODO: ignore PATH and ENVY_* vars
	dotenv := gotenv.Parse(f)
	keys := make([]string, 0, len(dotenv))
	for k := range dotenv {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	for _, k := range keys {
		v := dotenv[k]
		if strings.HasPrefix(k, "ENVY_") {
			actions = handleEnvyVar(actions, path, k, v)
		} else {
			actions = append(actions, action.Action{
				Path:        path,
				Priority:    -1,
				SetEnv:      k,
				SetEnvValue: v,
			})
		}
	}
	return
}

func handleEnvyVar(actions action.List, path, k, v string) action.List {
	// TODO: make paths absolute
	switch k {
	case "ENVY_EXTEND_PATH":
		for _, p := range env.ReversePaths(filepath.SplitList(v)) {
			actions = append(actions, action.Action{
				Path:     path,
				Priority: -1,
				AddPath:  p,
			})
		}
	case "ENVY_GOROOT":
		actions = append(actions, action.Action{
			Path:     path,
			Priority: -1,
			AddPath:  filepath.Join(v, "bin"),
		}, action.Action{
			Path:        path,
			Priority:    -1,
			SetEnv:      "GOROOT",
			SetEnvValue: v,
		})
	case "ENVY_PYTHONROOT":
		actions = append(actions, action.Action{
			Path:     path,
			Priority: -1,
			AddPath:  filepath.Join(v, "bin"),
		})
	default:
		log.Printf("%s not supported in env files", k)
	}
	return actions
}
