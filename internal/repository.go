package objects

import (
	"compress/zlib"
	"encoding/json"
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
	hash, err := obj.Hash()
	if err != nil {
		return err
	}
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
	bytes, err := obj.Content()
	if err != nil {
		return err
	}
	content := fmt.Sprintf("%s %d\x00%s", obj.Type(), len(bytes), bytes)
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

// ReadIndex Staging Area
func (repo *Repository) ReadIndex() ([]IndexEntry, error) {
	indexPath := filepath.Join(repo.GitDir, "index")
	data, err := os.ReadFile(indexPath)
	if err != nil {
		return nil, err
	}
	var entries []IndexEntry
	if err = json.Unmarshal(data, &entries); err != nil {
		return nil, err
	}
	return entries, nil
}

// WriteIndex Staging Area
func (repo *Repository) WriteIndex(entries []IndexEntry) error {
	indexPath := filepath.Join(repo.GitDir, "index")

	data, err := json.Marshal(entries)
	if err != nil {
		return err
	}
	return os.WriteFile(indexPath, data, 0644)
}

// Add Adds a file to staging area with proper staus tracking
func (repo *Repository) Add(filePath string) error {
	fullPath := filepath.Join(repo.WorkDir, filePath)

	info, err := os.Stat(fullPath)
	if err != nil {
		return fmt.Errorf("%sError::%s file not found: %s", colorRed,
			colorReset, filePath)
	}
	data, err := os.ReadFile(fullPath)
	if err != nil {
		return err
	}
	blob := &Blob{Data: data}
	if err = repo.StoreObject(blob); err != nil {
		return err
	}
	// Determine file mode
	var mode = "100644"
	if info.Mode()&0111 != 0 {
		mode = "100755"
	}
	// update index
	entries, err := repo.ReadIndex()
	if err != nil {
		return err
	}
	for i, entry := range entries {
		if entry.Path == filePath {
			entries = append(entries[:i], entries[i+1:]...)
			break
		}
	}
	hash, err := blob.Hash()
	if err != nil {
		return err
	}
	entries = append(entries, IndexEntry{
		Path:    filePath,
		Hash:    hash,
		Mode:    mode,
		Size:    info.Size(),
		ModTime: info.ModTime(),
		Staged:  true,
	})
	return repo.WriteIndex(entries)
}
