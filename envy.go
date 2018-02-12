package main

import (
	"fmt"
	"log"
	"os"
	"os/user"
	"path/filepath"
	"sort"
	"strings"
	"sync"

	"github.com/wojas/envy/action"
	"github.com/wojas/envy/env"
	"github.com/wojas/envy/session"
)

func PathsToCheck(cwd, home string) (paths []string) {
	// First check if we are within the user's home dir
	if !session.IsSubpath(cwd, home) {
		return nil
	}

	p := cwd
	for strings.HasPrefix(p, home) {
		paths = append(paths, p)
		p = filepath.Dir(p)
	}
	return paths
}

type ActionList []action.Action

// Implement sort.Interface to sort from shallow to deep
func (a ActionList) Len() int           { return len(a) }
func (a ActionList) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a ActionList) Less(i, j int) bool { return len(a[i].Path) < len(a[j].Path) }

func AddToPath(path []string, p string) ([]string, bool) {
	for _, x := range path {
		if x == p {
			return path, false
		}
	}
	path = append(path, p) // NOTE: reverse list
	return path, true
}

func RemoveFromPath(path []string, p string) ([]string, bool) {
	idx := -1
	for i, x := range path {
		if x == p {
			idx = i
		}
	}
	if idx < 0 {
		return path, false
	}
	path = append(path[:idx], path[idx+1:]...)
	return path, true
}

func ReversePaths(a []string) []string {
	for i := len(a)/2 - 1; i >= 0; i-- {
		opp := len(a) - 1 - i
		a[i], a[opp] = a[opp], a[i]
	}
	return a
}

func SetEnv(key, value string) {
	// TODO: escape value
	fmt.Printf("export %s='%s'\n", key, value)
}

func undoOld(cwd string, ses *session.Session) {
}

func main() {
	log.SetPrefix("\033[1;34m[envy]\033[0m ")
	log.SetFlags(0)

	u, err := user.Current()
	if err != nil || u.HomeDir == "" {
		log.Fatalf("Could not determine home dir.")
	}

	cwd, err := os.Getwd()
	if err != nil {
		return
	}

	debug := os.Getenv("envy_debug") != ""

	ses := session.Load(os.Getenv("_envy_session"))
	ses.Path = cwd
	undoOld(cwd, ses)

	path := ReversePaths(filepath.SplitList(os.Getenv("PATH")))
	pathChanged := false

	undo := ses.ToUndoFor(cwd)
	undoneEnv := make(map[string]string)
	// TODO: Work from deep to shallow
	for _, u := range undo {
		if u.Path != nil {
			for p := range u.Path {
				path, _ = RemoveFromPath(path, p)
				log.Printf("undo: PATH -= %s", p)
				pathChanged = true
			}
		}
		if u.Env != nil {
			for k, v := range u.Env {
				SetEnv(k, v)
				undoneEnv[k] = v
				log.Printf("undo: %s = %s", k, v)
			}
		}
	}

	paths := PathsToCheck(cwd, u.HomeDir)
	//log.Printf("Checking: %v", paths)

	ac := make(chan action.Action)
	var wg sync.WaitGroup

	for _, p := range paths {
		e := env.Env{
			Path:     p,
			ActionsC: ac,
			WG:       &wg,
		}
		for _, c := range env.AllCheckers {
			wg.Add(1)
			go c.Check(e)
		}
	}

	go func() {
		wg.Wait()
		close(ac)
	}()

	var actions ActionList
	for a := range ac {
		actions = append(actions, a)
	}
	sort.Sort(actions) // shallow paths first

	for _, a := range actions {
		if debug {
			log.Printf("action %#v", a)
		}

		if a.AddPath != "" {
			var changed bool
			path, changed = AddToPath(path, a.AddPath)
			pathChanged = pathChanged || changed
			if changed {
				u := ses.UndoFor(a.Path)
				u.Path[a.AddPath] = true
				log.Printf("PATH += %s", a.AddPath)
			}
		}

		if a.SetEnv != "" {
			k, v := a.SetEnv, a.SetEnvValue
			if os.Getenv(k) != v {
				SetEnv(k, v)
				u := ses.UndoFor(a.Path)

				savedValue := os.Getenv(k)
				if uv, exists := undoneEnv[k]; exists {
					// If we just restored this env var, it will not be visible
					// to getenv yet.
					savedValue = uv
				}
				// Only store current env value if we did not already store a
				// value for this path.
				// TODO: Do we also need to check higher up paths for the value?
				if _, exists := u.Env[k]; !exists {
					u.Env[k] = savedValue
				}
				log.Printf("%s = %s", k, v)
			}
		}
	}

	if pathChanged {
		pathenv := strings.Join(ReversePaths(path), string(filepath.ListSeparator))
		SetEnv("PATH", pathenv)
	}

	// We need to export this one too, so that if the user start a subshell,
	// envy is aware of the changes in the parent shell.
	SetEnv("_envy_session", session.Dump(ses))
}
