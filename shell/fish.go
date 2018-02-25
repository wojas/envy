package shell

import (
	"bytes"
	"fmt"
	"log"
	"strings"

	"github.com/wojas/envy/env"
)

// Fish returns a Shell interface for the fish shell.
func Fish() Shell {
	return fish{}
}

type fish struct{}

// Quote quotes a value with single quotes in a way that is safe to pass
// to the shell.
func (sh fish) Quote(s string) string {
	// "The only backslash escape accepted within single quotes is \', which
	// escapes a single quote and \\, which escapes the backslash symbol."
	// https://fishshell.com/docs/current/index.html
	return "'" + strings.Replace(strings.Replace(s, `\`, `\\`, -1), "'", `\'`, -1) + "'"
}

// SetEnv prints a shell env var export command. The key is expected to be
// safe, the value is escaped.
func (sh fish) SetEnv(key, value string) {
	if !ValidEnvVar(key) {
		// Should have been checked by caller
		log.Printf("SetEnv got an invalid env var name: %v", key)
		return
	}

	buf := bytes.NewBuffer(nil)
	buf.WriteString("set -xg ")
	buf.WriteString(key)
	buf.WriteByte(' ')
	buf.WriteString(sh.Quote(value))
	buf.WriteByte(';')
	fmt.Println(buf.String())
}

// SetPath prints the shell command to set a new path.
func (sh fish) SetPath(path *env.Path) {
	pathlist := path.Get()
	if len(pathlist) == 0 {
		log.Printf("Refusing to set an empty PATH")
		return
	}

	buf := bytes.NewBuffer(nil)
	buf.WriteString("set -xg PATH")
	for _, p := range pathlist {
		buf.WriteByte(' ')
		buf.WriteString(sh.Quote(p))
	}
	buf.WriteByte(';')
	fmt.Println(buf.String())
}
