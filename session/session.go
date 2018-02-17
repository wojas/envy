package session

import (
	"encoding/json"
	"log"
	"sort"

	"github.com/wojas/envy/paths"
)

// PathUndo describes the actions to undo for a single path
type PathUndo struct {
	Env  map[string]string // Environment vars to restore
	Path map[string]bool   // Paths to remove from PATH
}

// NewPathUndo created a new PathUndo
func NewPathUndo() *PathUndo {
	return &PathUndo{
		Env:  make(map[string]string),
		Path: make(map[string]bool),
	}
}

// PathUndoList is a slice of PathUndo
type PathUndoList []*PathUndo

// Session describes an envy session
type Session struct {
	Path string
	Undo map[string]*PathUndo
}

// UndoFor returns the PathUndo for a directory
func (s *Session) UndoFor(p string) *PathUndo {
	u, ok := s.Undo[p]
	if !ok {
		u = NewPathUndo()
		s.Undo[p] = u
	}
	return u
}

// ToUndoFor returns a list of action to undo for a new working dir, and removes
// the items from the session.
func (s *Session) ToUndoFor(p string) PathUndoList {
	undo := make(PathUndoList, 0)
	delPaths := make([]string, 0)
	for path := range s.Undo {
		if paths.IsSubpath(p, path) {
			continue // Still active
		}
		delPaths = append(delPaths, path)
	}

	// We need reverse sorting (from deep to shallow path) for undo. Since
	// the sort package does not provide it, we use normal ascending sort and
	// then iterate backwards.
	sort.Strings(delPaths)
	for i := len(delPaths) - 1; i >= 0; i-- {
		path := delPaths[i]
		undo = append(undo, s.Undo[path])
		delete(s.Undo, path) // Remove from self
	}
	return undo
}

// PathUndoList returns a list of all the current PathUndo instances, sorted
// by path
// ToUndoFor returns a list of action to undo for a new working dir, and removes
// the items from the session.
func (s *Session) PathUndoList() PathUndoList {
	undo := make(PathUndoList, 0)
	pathList := make([]string, 0, len(s.Undo))
	for path := range s.Undo {
		pathList = append(pathList, path)
	}

	// We need these entries sorted by path from shallow to deep for the
	// stage where we prune env vars that are no longer set by checkers.
	sort.Strings(pathList)
	for _, path := range pathList {
		undo = append(undo, s.Undo[path])
	}
	return undo
}

// New creates a Session object.
func New() *Session {
	return &Session{
		Undo: make(map[string]*PathUndo),
	}
}

// Dump marshals the session as JSON.
func Dump(s *Session) string {
	blob, err := json.Marshal(s)
	if err != nil {
		log.Fatalf("Cannot marshall %#v", s)
	}
	return string(blob)
}

// Load loads a previously marshaled session.
func Load(data string) *Session {
	s := New()
	if data == "" {
		return s
	}
	err := json.Unmarshal([]byte(data), s)
	if err != nil {
		log.Printf("WARNING: Cannot unmarshal session")
	}
	return s
}
