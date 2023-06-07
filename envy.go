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
	"github.com/wojas/envy/config"
	environ "github.com/wojas/envy/env"
	"github.com/wojas/envy/paths"
	"github.com/wojas/envy/session"
	"github.com/wojas/envy/shell"
)

var traceFile = flag.String("trace", "", "Write trace to given file for use with `go tool trace`")
var fish = flag.Bool("fish", false, "Output fish shell commands instead of the default bash/zsh commands.")

const ConfigFile = ".envy.yml"

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

	// Get configuration
	home, err := paths.HomeDir()
	if err != nil {
		log.Fatalf("Could not determine home dir: %v", nil)
	}
	conf := config.Default()
	err = conf.LoadYAMLFile(filepath.Join(home, ConfigFile))
	if err != nil && !os.IsNotExist(err) {
		log.Printf("Could not open ~/%s config file: %v", ConfigFile, err)
	}
	if err := conf.Check(); err != nil {
		log.Fatalf("Error in ~/%s config: %v", ConfigFile, err)
	}
	if debug {
		log.Printf("Effective config:\n%s", conf)
	}

	// Get information about our environment
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

	// Step 1: Undo previous changes if the user moved to a different working directory.
	undo := ses.ToUndoFor(cwd)
	for _, u := range undo {
		for p := range u.Path {
			path.Remove(p)
		}
		for k, v := range u.Env {
			env.Restore(k, v)
		}
	}

	// Step 2: Perform actions for the current working directory.
	toCheck := paths.ToCheck(cwd, conf.TrustedPaths)
	if conf.AlwaysLoadHome {
		// TODO: handle home with trailing /
		if len(toCheck) == 0 || toCheck[len(toCheck)-1] != home {
			toCheck = append(toCheck, home)
		}
	}
	if debug {
		log.Printf("Paths to check: %v", toCheck)
	}
	actions := getActions(toCheck)
	seenEnvs := make(map[string]bool)
	seenPaths := make(map[string]bool)
	for _, a := range actions {
		if debug {
			log.Printf("action %#v", a)
		}

		if a.AddPath != "" {
			p := a.AddPath
			seenPaths[p] = true
			if !path.Has(p) {
				path.Add(p)
				u := ses.UndoFor(a.Path)
				u.Path[p] = true
			}
		}

		if a.SetEnv != "" {
			k, v := a.SetEnv, a.SetEnvValue
			seenEnvs[k] = true
			if prevValue := env.Get(k); prevValue != v {
				env.Set(k, v)

				// Only store current env value if we did not already store a
				// value for this path.
				// TODO: Do we also need to check higher up paths for the value?
				u := ses.UndoFor(a.Path)
				if _, exists := u.Env[k]; !exists {
					u.Env[k] = prevValue
				}
			}
		}
	}

	// Step 3: Undo changes that no longer appear in the current list of actions
	// For this we check which env vars occur in the undo state that were not
	// reported by the checkers and changed in the env.
	// Note that after Step 1 the session undo data only contains active paths.

	// TODO: this will not be triggered the first time if the var was already
	//       restored in Step 1, in which case it will also be changed. Maybe
	//       merge Step 1 into this one, because we do not need to restore
	//       vars that are changed by checkers anyway? But then we need to be
	//       careful about which undo value we store in the session.

	// These two are defined outside of the loop to also apply all removes from
	// shallow paths to deeper ones.
	// This relies on the undo items being sorted from shallow to deep paths.
	removeEnvs := make([]string, 0)
	removePaths := make([]string, 0)
	for _, u := range ses.PathUndoList() {
		// For environment variables
		for k, v := range u.Env {
			if !seenEnvs[k] {
				env.Restore(k, v)
				removeEnvs = append(removeEnvs, k)
				seenEnvs[k] = true // Prevent triggering again
			}
		}
		for _, k := range removeEnvs {
			delete(u.Env, k) // Remove from session, no longer relevant
		}

		// For PATH elements
		// TODO: If a shallower path reappears later, it will be added to the front.
		//       This is undesirable. Avoiding this would require storing the
		//       original PATH before we do any changes in a session.
		for p := range u.Path {
			if !seenPaths[p] {
				path.Remove(p)
				removePaths = append(removePaths, p)
				seenPaths[p] = true // Prevent triggering again
			}
		}
		for _, p := range removePaths {
			delete(u.Path, p)
		}
	}

	// Print commands to perform environment changes for different shells
	var sh shell.Shell
	if *fish {
		sh = shell.Fish()
	} else {
		sh = shell.Bash() // Also used for zsh
	}

	// Environment changes
	for _, item := range env.Changes() {
		if os.Getenv(item.Key) != item.Val {
			sh.SetEnv(item.Key, item.Val)
			if strings.HasPrefix(item.Key, "_ENVY_") {
				continue // Do not log gitroot env changes
			}

			if item.Restored {
				log.Printf("restore: %s = %s", item.Key, shorten.Do(item.Val))
			} else {
				log.Printf("%s = %s", item.Key, shorten.Do(item.Val))
			}

			// Easiest to implement when this changes
			if item.Key == "ENVY_COLOR" {
				color := item.Val
				if color == "" {
					color = "RESET"
				}
				s, exists := conf.Colors[color]
				if !exists {
					log.Printf("WARNING: Color %q not defined in ~/.envy.yml", color)
				}
				if debug {
					log.Printf("Color: %s = %q", color, s)
				}
				_, _ = os.Stderr.WriteString(s) // no newline
			}
		}
	}

	// PATH changes
	if path.Changed {
		sh.SetPath(path)

		// Print removed paths
		var removed []string
		for p := range path.Removed {
			removed = append(removed, p)
		}
		sort.Strings(removed)
		for _, p := range removed {
			if !path.Added[p] {
				log.Printf("restore: PATH -= %s", shorten.Do(p))
			}
		}

		// Print added paths
		for _, p := range path.GetReversed() {
			if path.Added[p] && !path.Removed[p] {
				log.Printf("PATH += %s", shorten.Do(p))
			}
		}
	}

	// Set new session.
	// This one is exported too, so that if the user start a subshell,
	// envy is aware of the changes in the parent shell.
	sh.SetEnv("_envy_session", session.Dump(ses))
}
