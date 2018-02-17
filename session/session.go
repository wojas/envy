package session

import (
	"encoding/json"
	"log"

	"github.com/wojas/envy/paths"
)

// PathUndo describes the actions to undo for a single path
type PathUndo struct {
	Env  map[string]string // Environment vars to restore
	Path map[string]bool   // Paths to remove from PATH
}

func NewPathUndo() *PathUndo {
	return &PathUndo{
		Env:  make(map[string]string),
		Path: make(map[string]bool),
	}
}

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
func (s *Session) ToUndoFor(p string) []PathUndo {
	undo := make([]PathUndo, 0)
	delPaths := make([]string, 0)
	for path, u := range s.Undo {
		if paths.IsSubpath(p, path) {
			continue // Still active
		}
		delPaths = append(delPaths, path)
		undo = append(undo, *u)
	}
	for _, path := range delPaths {
		delete(s.Undo, path) // Remove from self
	}
	return undo
}

func New() *Session {
	return &Session{
		Undo: make(map[string]*PathUndo),
	}
}

func Dump(s *Session) string {
	blob, err := json.Marshal(s)
	if err != nil {
		log.Fatalf("Cannot marshall %#v", s)
	}
	return string(blob)
}

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
