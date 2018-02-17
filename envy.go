package main

import (
	"flag"
	"log"
	"os"
	"path/filepath"
	"runtime/trace"
	"sort"
	"strings"
	"sync"

	"github.com/wojas/envy/action"
	"github.com/wojas/envy/checkers"
	environ "github.com/wojas/envy/env"
	"github.com/wojas/envy/paths"
	"github.com/wojas/envy/session"
	"github.com/wojas/envy/shell"
)

var traceFile = flag.String("trace", "", "Write trace to given file for use with `go tool trace`")

const (
	// TODO: use https://github.com/fatih/color
	colorBlue  = "\033[1;34m"
	colorReset = "\033[0m"
)

// getActions checks all paths for Actions using the checkers.
func getActions(paths []string) (actions action.List) {
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
	flag.Parse()

	// Setup logging: disabled timestamp and add a colored prefix
	log.SetPrefix(colorBlue + "[envy] " + colorReset)
	log.SetFlags(0)

	// If flag set, enable tracing and write to file
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

	// Options set through environment variables
	debug := os.Getenv("envy_debug") != ""

	// Get information about our environment
	home, err := paths.HomeDir()
	if err != nil {
		log.Fatalf("Could not determine home dir: %v", nil)
	}
	cwd, err := os.Getwd()
	if err != nil {
		return
	}

	// Util to shorten paths for display
	shorten := paths.Shorten{
		Home:    home,
		Current: cwd,
	}

	// Load session info from environment
	ses := session.Load(os.Getenv("_envy_session"))
	ses.Path = cwd

	// Load environment to perform magic on
	path := environ.NewPath(filepath.SplitList(os.Getenv("PATH")))
	env := environ.New()

	// Undo previous changes if the user moved to a different working directory.
	undo := ses.ToUndoFor(cwd)
	// TODO: Work from deep to shallow
	for _, u := range undo {
		for p := range u.Path {
			path.Remove(p)
			log.Printf("restore: PATH -= %s", shorten.Do(p))
		}
		for k, v := range u.Env {
			env.Set(k, v)
			log.Printf("restore: %s = %s", k, shorten.Do(v))
		}
	}

	// Perform actions for the current working directory.
	actions := getActions(paths.ToCheck(cwd, home))
	for _, a := range actions {
		if debug {
			log.Printf("action %#v", a)
		}

		if a.AddPath != "" {
			p := a.AddPath
			if !path.Has(p) {
				path.Add(p)
				u := ses.UndoFor(a.Path)
				u.Path[p] = true
				log.Printf("PATH += %s", shorten.Do(p))
			}
		}

		if a.SetEnv != "" {
			k, v := a.SetEnv, a.SetEnvValue
			if os.Getenv(k) != v {
				savedValue := env.Get(k)
				env.Set(k, v)

				// Only store current env value if we did not already store a
				// value for this path.
				// TODO: Do we also need to check higher up paths for the value?
				u := ses.UndoFor(a.Path)
				if _, exists := u.Env[k]; !exists {
					u.Env[k] = savedValue
				}
				log.Printf("%s = %s", k, shorten.Do(v))
			}
		}
	}

	// Print commands to perform environment changes
	for _, item := range env.Changes() {
		shell.SetEnv(item.Key, item.Val)
	}
	if path.Changed {
		pathenv := strings.Join(path.Get(), string(filepath.ListSeparator))
		shell.SetEnv("PATH", pathenv)
	}

	// TODO: If variables previously changed do not appear in this list, unset them

	// Set new session.
	// This one is exported too, so that if the user start a subshell,
	// envy is aware of the changes in the parent shell.
	shell.SetEnv("_envy_session", session.Dump(ses))
}
