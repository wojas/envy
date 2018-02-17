package env

// Path contains operations for paths in a PATH env var.
type Path struct {
	Changed bool
	revPath []string
}

// NewPath returns a new Path
func NewPath(paths []string) *Path {
	return &Path{
		revPath: reversePath(paths),
	}
}

// Add adds a path to the list of paths.
func (p *Path) Add(path string) {
	p.revPath = append(p.revPath, path) // NOTE: reverse list, so append
	p.Changed = true
}

// Remove removes a path from the list of paths.
func (p *Path) Remove(path string) {
	for i, x := range p.revPath {
		if x == path {
			p.revPath = append(p.revPath[:i], p.revPath[i+1:]...)
			p.Changed = true
			return
		}
	}
	return
}

// Has checks if a path is already included in the list of paths.
func (p *Path) Has(path string) bool {
	for _, x := range p.revPath {
		if x == path {
			return true
		}
	}
	return false
}

// Get returns the list of paths.
func (p *Path) Get() []string {
	return reversePath(p.revPath)
}

// reversePath reverses a list of paths
func reversePath(a []string) []string {
	for i := len(a)/2 - 1; i >= 0; i-- {
		opp := len(a) - 1 - i
		a[i], a[opp] = a[opp], a[i]
	}
	return a
}
