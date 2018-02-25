package shell

import (
	"fmt"
	"log"
	"path/filepath"
	"strings"

	"github.com/wojas/envy/env"
)

// Bash returns a Shell interface for the bash and zsh shells.
func Bash() Shell {
	return bash{}
}

type bash struct{}

// Quote quotes a value with single quotes in a way that is safe to pass
// to the shell. Since there is no way to escape ' inside single quotes, we
// exit the single quotes and quote that char with double quotes, like "'".
// The string "who's there?" will become 'Who'"'"'s there?'
func (sh bash) Quote(s string) string {
	return "'" + strings.Replace(s, "'", `'"'"'`, -1) + "'"
}

// SetEnv prints a shell env var export command. The key is expected to be
// safe, the value is escaped.
func (sh bash) SetEnv(key, value string) {
	if !ValidEnvVar(key) {
		// Should have been checked by caller
		log.Printf("SetEnv got an invalid env var name: %v", key)
		return
	}
	fmt.Printf("export %s=%s\n", key, sh.Quote(value))
}

// SetPath prints the shell command to set a new path.
func (sh bash) SetPath(path *env.Path) {
	pathenv := strings.Join(path.Get(), string(filepath.ListSeparator))
	sh.SetEnv("PATH", pathenv)
}
