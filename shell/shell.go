package shell

import (
	"regexp"

	"github.com/wojas/envy/env"
)

var valid = regexp.MustCompile(`[a-zA-Z_]+[a-zA-Z0-9_]*`)

// ValidEnvVar checks if an environment variable name is valid.
// https://stackoverflow.com/questions/2821043/
func ValidEnvVar(key string) bool {
	return valid.MatchString(key)
}

// Shell defines the interface that shell support modules must implement.
type Shell interface {
	Quote(s string) string
	SetEnv(key, value string)
	SetPath(path *env.Path)
}
