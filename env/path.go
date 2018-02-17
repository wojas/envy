package env

type Path struct {
	Changed bool
	revPath []string
}

func NewPath(paths []string) *Path {
	return &Path{
		revPath: reversePath(paths),
	}
}

func (p *Path) Add(path string) {
	p.revPath = append(p.revPath, path) // NOTE: reverse list, so append
	p.Changed = true
}

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

func (p *Path) Has(path string) bool {
	for _, x := range p.revPath {
		if x == path {
			return true
		}
	}
	return false
}

func (p *Path) Get() []string {
	return reversePath(p.revPath)
}

func reversePath(a []string) []string {
	for i := len(a)/2 - 1; i >= 0; i-- {
		opp := len(a) - 1 - i
		a[i], a[opp] = a[opp], a[i]
	}
	return a
}
