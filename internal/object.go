package objects

const (
	ObjectTypeBlob   = "blob"
	ObjectTypeTree   = "tree"
	ObjectTypeCommit = "commit"
)

const (
	colorRed   = "\033[31m"
	colorReset = "\033[0m"
)

type GitObject interface {
	Type() string
	Content() ([]byte, error)
	Hash() (string, error)
}
