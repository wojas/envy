package checkers

import (
	"bytes"
	"io/ioutil"
	"path/filepath"

	"github.com/wojas/envy/action"
	"github.com/wojas/envy/paths"
)

// GitRootCheck checks if a path is the root of a git checkout.
type GitRootCheck struct{}

// Check implements the Checker interface.
func (c GitRootCheck) Check(path string) (actions action.List) {
	git := filepath.Join(path, ".git")
	if !paths.IsDir(git) {
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

func parseGitRef(ref []byte) string {
	if !bytes.HasPrefix(ref, []byte("ref: ")) {
		return ""
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
