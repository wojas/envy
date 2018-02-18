package checkers

import (
	"log"
	"os"
	"path/filepath"
	"sort"

	"github.com/subosito/gotenv" // TODO: find a better one

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
		actions = append(actions, action.Action{
			Path:        path,
			SetEnv:      k,
			SetEnvValue: v,
		})
	}
	return
}
