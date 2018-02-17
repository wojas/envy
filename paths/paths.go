package paths

import (
	"fmt"
	"os"
	"os/user"
	"path/filepath"
	"strings"
)

// IsSubpath checks if path `p` is a subpath of `parent`
func IsSubpath(p, parent string) bool {
	if !strings.HasPrefix(p, parent) {
		return false
	}
	if len(p) > len(parent) {
		return p[len(parent)] == filepath.Separator
	}
	return true
}

// ToCheck returns all paths to check using the checkers.
// Currently only paths under the home directory are returned for security,
// but this will become configurable in a future version.
func ToCheck(cwd, home string) (paths []string) {
	// First check if we are within the user's home dir
	if !IsSubpath(cwd, home) {
		return nil
	}

	p := cwd
	for strings.HasPrefix(p, home) {
		paths = append(paths, p)
		p = filepath.Dir(p)
	}
	return paths
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

// Shorten shortens paths by replacing the home dir with '~' and current
// dir with '.' for display.
type Shorten struct {
	Home    string
	Current string
}

// Do performs the actual shortening of a path.
func (s Shorten) Do(p string) string {
	if IsSubpath(p, s.Current) {
		return "." + p[len(s.Current):]
	}
	if IsSubpath(p, s.Home) {
		return "~" + p[len(s.Home):]
	}
	return p
}

// IsDir checks if a path is a directory.
func IsDir(path string) bool {
	fi, err := os.Stat(path)
	if err != nil {
		return false
	}
	return fi.IsDir()
}
