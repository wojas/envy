package env

// Path contains operations for paths in a PATH env var.
type Path struct {
	Changed bool
	Added   map[string]bool
	Removed map[string]bool
	revPath []string
}

// NewPath returns a new Path
func NewPath(paths []string) *Path {
	return &Path{
		Added:   make(map[string]bool),
		Removed: make(map[string]bool),
		revPath: ReversePaths(paths),
	}
}

// Add adds a path to the list of paths.
func (p *Path) Add(path string) {
	p.revPath = append(p.revPath, path) // NOTE: reverse list, so append
	p.Changed = true
	p.Added[path] = true
}

// Remove removes a path from the list of paths.
func (p *Path) Remove(path string) {
	for i, x := range p.revPath {
		if x == path {
			p.revPath = append(p.revPath[:i], p.revPath[i+1:]...)
			p.Changed = true
			p.Removed[path] = true
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
	return ReversePaths(p.revPath)
}

// GetReversed returns the list of paths reversed.
func (p *Path) GetReversed() []string {
	return p.revPath
}

// ReversePaths reverses a list of paths and returns a new slice.
func ReversePaths(a []string) []string {
	res := make([]string, len(a))
	n := len(a)
	for i, p := range a {
		res[n-i-1] = p
	}
	return res
}
