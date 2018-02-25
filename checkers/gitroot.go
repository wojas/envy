package checkers

import (
	"bytes"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/wojas/envy/action"
	"github.com/wojas/envy/paths"
)

// GitRootCheck checks if a path is the root of a git checkout.
type GitRootCheck struct{}

// Check implements the Checker interface.
func (c GitRootCheck) Check(path string) (actions action.List) {
	git := filepath.Join(path, ".git")

	fi, err := os.Stat(git)
	if err != nil {
		return
	}

	if fi.Mode().IsRegular() {
		// Plain file with "gitdir: ..." (worktrees and subrepos)
		contents, err := ioutil.ReadFile(git)
		if err != nil {
			return
		}
		git = parseGitRedirect(path, contents)
		if git == "" {
			return
		}
	} else if !fi.IsDir() {
		return
	}

	actions = append(actions, action.Action{
		Path:        path,
		SetEnv:      "_ENVY_GITROOT",
		SetEnvValue: path,
	})

	ref, err := ioutil.ReadFile(filepath.Join(git, "HEAD"))
	if err != nil {
		return
	}
	branch := parseGitRef(ref)
	if branch == "" {
		return
	}

	actions = append(actions, action.Action{
		Path:        path,
		SetEnv:      "_ENVY_BRANCH",
		SetEnvValue: branch,
	})

	return
}

func parseGitRedirect(path string, contents []byte) string {
	if !bytes.HasPrefix(contents, []byte("gitdir: ")) {
		return ""
	}
	contents = contents[8:]
	if idx := bytes.IndexByte(contents, '\n'); idx >= 0 {
		contents = contents[:idx]
	}
	git := string(contents)
	if git == "" {
		return ""
	}
	if git[0] != '/' {
		git = filepath.Join(path, git)
	}
	if !paths.IsDir(git) {
		return ""
	}
	return git
}

func parseGitRef(ref []byte) string {
	if !bytes.HasPrefix(ref, []byte("ref: ")) {
		return string(ref[:7]) // detached head, commit hash
	}
	if idx := bytes.IndexByte(ref, '\n'); idx >= 0 {
		ref = ref[:idx]
	}
	idx := bytes.LastIndexByte(ref, '/')
	if idx < 0 {
		return ""
	}
	branch := string(ref[idx+1:])
	return branch
}
