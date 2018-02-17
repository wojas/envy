package action

type Action struct {
	Path        string
	AddPath     string
	SetEnv      string
	SetEnvValue string
}

type List []Action

// Implement sort.Interface to sort from shallow path to deep path
func (a List) Len() int           { return len(a) }
func (a List) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a List) Less(i, j int) bool { return len(a[i].Path) < len(a[j].Path) }
