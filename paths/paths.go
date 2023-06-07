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

// IsSubpathOfAny checks if path `p` is a subpath of any given parent
func IsSubpathOfAny(p string, parents []string) bool {
	for _, candidate := range parents {
		if IsSubpath(p, candidate) {
			return true
		}
	}
	return false
}

// ToCheck returns all paths to check using the checkers.
// It only includes trusted paths.
func ToCheck(cwd string, trusted []string) (paths []string) {
	p := cwd
	for IsSubpathOfAny(p, trusted) {
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

// IsFile checks if a path is a regular file
func IsFile(path string) bool {
	fi, err := os.Stat(path)
	if err != nil {
		return false
	}
	return fi.Mode().IsRegular()
}
