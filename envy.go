package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"os/user"
	"path/filepath"
	"regexp"
	"runtime/trace"
	"sort"
	"strings"
	"sync"

	"github.com/wojas/envy/action"
	"github.com/wojas/envy/checkers"
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

// ValidEnvVar checks if an environment variable name is valid.
// https://stackoverflow.com/questions/2821043/
var ValidEnvVar = regexp.MustCompile(`[a-zA-Z_]+[a-zA-Z0-9_]*`)

// ShellQuote quotes a value with single quotes in a way that is safe to pass
// to the shell. Since there is no way to escape ' inside single quotes, we
// exit the single quotes and quote that char with double quotes, like "'".
// The string "who's there?" will become 'Who'"'"'s there?'
func ShellQuote(s string) string {
	return "'" + strings.Replace(s, "'", `'"'"'`, -1) + "'"
}

// SetEnv prints a shell env var export command. The key is expected to be
// safe, the value is escaped.
func SetEnv(key, value string) {
	if !ValidEnvVar.MatchString(key) {
		// Should have been checked by caller
		log.Fatalf("SetEnv got an invalid env var name: %v", key)
	}
	fmt.Printf("export %s=%s\n", key, ShellQuote(value))
}

// HomeDir returns the current user's cleaned home dir path (no trailing '/').
func HomeDir() (string, error) {
	// First try HOME env var (saves ~2 ms over user API and makes testing easier)
	home := os.Getenv("HOME")
	if home != "" {
		return filepath.Clean(home), nil
	}

	// Fallback to user API
	u, err := user.Current()
	if err != nil {
		return "", err
	}
	if u.HomeDir == "" {
		return "", fmt.Errorf("home dir empty")
	}
	return filepath.Clean(u.HomeDir), nil
}

// ShortenPath shortens paths by replacing the home dir with '~' and current
// dir with '.' for display.
func ShortenPath(p, cwd, home string) string {
	if session.IsSubpath(p, cwd) {
		return "." + p[len(cwd):]
	}
	if session.IsSubpath(p, home) {
		return "~" + p[len(home):]
	}
	return p
}

var traceFile = flag.String("trace", "", "Write trace to given file for use with `go tool trace`")

// getActions checks all paths for Actions using the checkers.
func getActions(paths []string) (actions ActionList) {
	ac := make(chan action.Action)
	var wg sync.WaitGroup

	for _, p := range paths {
		for _, c := range checkers.AllCheckers {
			wg.Add(1)
			go func(path string, c checkers.Checker) {
				for _, a := range c.Check(path) {
					ac <- a
				}
				wg.Done()
			}(p, c)
		}
	}

	go func() {
		wg.Wait()
		close(ac)
	}()

	for a := range ac {
		actions = append(actions, a)
	}
	sort.Sort(actions) // shallow paths first
	return actions
}

func main() {
	// TODO: use https://github.com/fatih/color
	log.SetPrefix("\033[1;34m[envy]\033[0m ")
	log.SetFlags(0)

	flag.Parse()

	if *traceFile != "" {
		f, err := os.Create(*traceFile)
		if err != nil {
			log.Fatalf("Cannot write trace to %s: %v", *traceFile, err)
		}
		defer log.Printf("To view the trace file, run:  go tool trace %s", *traceFile)
		defer f.Close()
		trace.Start(f)
		defer trace.Stop()
	}

	home, err := HomeDir()
	if err != nil {
		log.Fatalf("Could not determine home dir: %v", nil)
	}

	cwd, err := os.Getwd()
	if err != nil {
		return
	}

	debug := os.Getenv("envy_debug") != ""

	ses := session.Load(os.Getenv("_envy_session"))
	ses.Path = cwd

	path := ReversePaths(filepath.SplitList(os.Getenv("PATH")))
	pathChanged := false

	undo := ses.ToUndoFor(cwd)
	undoneEnv := make(map[string]string)
	// TODO: Work from deep to shallow
	for _, u := range undo {
		for p := range u.Path {
			path, _ = RemoveFromPath(path, p)
			log.Printf("restore: PATH -= %s", ShortenPath(p, cwd, home))
			pathChanged = true
		}
		for k, v := range u.Env {
			SetEnv(k, v)
			undoneEnv[k] = v
			log.Printf("restore: %s = %s", k, ShortenPath(v, cwd, home))
		}
	}

	paths := PathsToCheck(cwd, home)
	actions := getActions(paths)

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
				log.Printf("PATH += %s", ShortenPath(a.AddPath, cwd, home))
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
				log.Printf("%s = %s", k, ShortenPath(v, cwd, home))
			}
		}
	}

	if pathChanged {
		pathenv := strings.Join(ReversePaths(path), string(filepath.ListSeparator))
		SetEnv("PATH", pathenv)
	}

	// TODO: If variables previously changed do not appear in this list, unset them

	// We need to export this one too, so that if the user start a subshell,
	// envy is aware of the changes in the parent shell.
	SetEnv("_envy_session", session.Dump(ses))
}
