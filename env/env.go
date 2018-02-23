package env

import (
	"os"
	"sort"
)

// Env keeps track of changes to environment variables
type Env struct {
	changed  map[string]string
	restored map[string]bool
}

// New returns a new Env
func New() *Env {
	return &Env{
		changed:  make(map[string]string),
		restored: make(map[string]bool),
	}
}

// Get returns the current value for an environment variable.
func (e *Env) Get(key string) string {
	if val, exists := e.changed[key]; exists {
		return val
	}
	return os.Getenv(key)
}

// Set sets an environment variable to a new value.
func (e *Env) Set(key, val string) {
	e.changed[key] = val
	delete(e.restored, key)
}

// Restore sets an environment variable to a previous value and marks it as restored.
func (e *Env) Restore(key, val string) {
	e.changed[key] = val
	e.restored[key] = true
}

// Changes returns all changes to environment variables
func (e *Env) Changes() (changes ChangeList) {
	for k, v := range e.changed {
		changes = append(changes, Change{k, v, e.restored[k]})
	}
	sort.Sort(changes)
	return
}

// Change describes a single environment variable change.
type Change struct {
	Key      string
	Val      string
	Restored bool
}

// ChangeList is a slice of changes, with added sort interface.
// The sort.Interface lists restored changes first and sorts by Key.
type ChangeList []Change

func (cl ChangeList) Len() int      { return len(cl) }
func (cl ChangeList) Swap(i, j int) { cl[i], cl[j] = cl[j], cl[i] }
func (cl ChangeList) Less(i, j int) bool {
	if cl[i].Restored != cl[j].Restored {
		return cl[i].Restored // Restored items go first
	}
	return cl[i].Key < cl[j].Key
}
