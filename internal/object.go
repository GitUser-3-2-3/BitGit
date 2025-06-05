package objects

const (
	ObjectTypeBlob   = "blob"
	ObjectTypeTree   = "tree"
	ObjectTypeCommit = "commit"
)

type GitObject interface {
	Type() string
	Content() []byte
	Hash() string
}
