package objects

import (
	"compress/zlib"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

type Repository struct {
	WorkDir string
	GitDir  string
}

// InitRepo initializes a git repository
func InitRepo(path string) (*Repository, error) {
	gitDir := filepath.Join(path, ".git")

	// create directory structure
	dirs := []string{gitDir,
		filepath.Join(gitDir, "objects"),
		filepath.Join(gitDir, "refs"),
		filepath.Join(gitDir, "refs", "heads"),
	}
	for _, dir := range dirs {
		if err := os.Mkdir(dir, 0755); err != nil {
			return nil, err
		}
	}
	headPath := filepath.Join(gitDir, "HEAD")
	err := os.WriteFile(headPath, []byte("ref: refs/heads/main\n"), 0644)
	if err != nil {
		return nil, err
	}
	indexPath := filepath.Join(gitDir, "index")
	err = os.WriteFile(indexPath, []byte("[]"), 0644)
	if err != nil {
		return nil, err
	}
	return &Repository{WorkDir: path, GitDir: gitDir}, nil
}

func LoadRepo(path string) (*Repository, error) {
	gitDir := filepath.Join(path, ".git")
	if _, err := os.Stat(gitDir); os.IsNotExist(err) {
		return nil, fmt.Errorf("not a git repository")
	}
	return &Repository{WorkDir: path,
		GitDir: gitDir,
	}, nil
}

func (repo *Repository) StoreObject(obj GitObject) error {
	hash := obj.Hash()
	objDir := filepath.Join(repo.GitDir, "objects", hash[:2])
	objPath := filepath.Join(objDir, hash[2:])

	if _, err := os.Stat(objPath); nil == err {
		return nil // object already exists
	}
	if err := os.MkdirAll(objDir, 0755); err != nil {
		return err
	}

	file, err := os.Create(objPath)
	if err != nil {
		return err
	}
	defer func(file *os.File) {
		err = file.Close()
	}(file)
	if err != nil {
		return err
	}

	writer := zlib.NewWriter(file)
	defer func(writer *zlib.Writer) {
		err = writer.Close()
	}(writer)
	if err != nil {
		return err
	}

	content := fmt.Sprintf("%s %d\x00%s", obj.Type(), len(obj.Content()), obj.Content())
	_, err = writer.Write([]byte(content))
	return err
}

func (repo *Repository) LoadObject(hash string) (GitObject, error) {
	objPath := filepath.Join(repo.GitDir, "objects", hash[:2], hash[2:])

	file, err := os.Open(objPath)
	if err != nil {
		return nil, err
	}
	defer func(file *os.File) {
		err = file.Close()
	}(file)
	if err != nil {
		return nil, err
	}
	reader, err := zlib.NewReader(file)
	if err != nil {
		return nil, err
	}
	defer func(reader io.ReadCloser) {
		err = reader.Close()
	}(reader)
	if err != nil {
		return nil, err
	}

	data, err := io.ReadAll(reader)
	if err != nil {
		return nil, err
	}

	// parse object header
	nullIndex := strings.Index(string(data), "\x00")
	if nullIndex == -1 {
		return nil, fmt.Errorf("invalid object format")
	}
	header := string(data[:nullIndex])
	content := data[nullIndex+1:]

	parts := strings.Split(header, " ")
	if len(parts) != 2 {
		return nil, fmt.Errorf("invalid object header")
	}

	objType := parts[0]

	switch objType {
	case ObjectTypeBlob:
		return &Blob{Data: content, hash: hash}, nil
	default:
		return nil, fmt.Errorf("what even is that type? %s", objType)
	}
}
