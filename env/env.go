package env

import "os"

// Env keeps track of changes to environment variables
type Env struct {
	changed map[string]string
}

// New returns a new Env
func New() *Env {
	return &Env{
		changed: make(map[string]string),
	}
}

// Get returns the current value for an environment variable.
func (e *Env) Get(key string) string {
	if val, exists := e.changed[key]; exists {
		return val
	}
	return os.Getenv(key)
}

// Set sets an environment variable
func (e *Env) Set(key, val string) {
	e.changed[key] = val
}

type Change struct {
	Key string
	Val string
}

// Changes returns all changes to environment variables
func (e *Env) Changes() (changes []Change) {
	for k, v := range e.changed {
		changes = append(changes, Change{k, v})
	}
	return
}
