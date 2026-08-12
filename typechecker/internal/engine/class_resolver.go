package engine

type ClassResolver interface {
	Resolve(name string) (*ClassObject, error)
}
