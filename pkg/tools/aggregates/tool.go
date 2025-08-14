package aggregates

type Tools interface {
	Execute(jsonInput string) (string, error)
	Spec() Tool
}

type Tool struct {
	Name        string
	Description string
	Properties  map[string]any
}
