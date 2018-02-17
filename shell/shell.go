package shell

import (
	"fmt"
	"log"
	"regexp"
	"strings"
)

var valid = regexp.MustCompile(`[a-zA-Z_]+[a-zA-Z0-9_]*`)

// ValidEnvVar checks if an environment variable name is valid.
// https://stackoverflow.com/questions/2821043/
func ValidEnvVar(key string) bool {
	return valid.MatchString(key)
}

// Quote quotes a value with single quotes in a way that is safe to pass
// to the shell. Since there is no way to escape ' inside single quotes, we
// exit the single quotes and quote that char with double quotes, like "'".
// The string "who's there?" will become 'Who'"'"'s there?'
func Quote(s string) string {
	return "'" + strings.Replace(s, "'", `'"'"'`, -1) + "'"
}

// SetEnv prints a shell env var export command. The key is expected to be
// safe, the value is escaped.
func SetEnv(key, value string) {
	if !ValidEnvVar(key) {
		// Should have been checked by caller
		// TODO: Move fatal things to main
		log.Fatalf("SetEnv got an invalid env var name: %v", key)
	}
	fmt.Printf("export %s=%s\n", key, Quote(value))
}
