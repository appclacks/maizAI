package aggregates

type Directory struct {
	Path      string
	Recursive bool
}

type LocalFileContext struct {
	Files       []string
	Directories []Directory
}
