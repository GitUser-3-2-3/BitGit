package objects

import (
	"compress/zlib"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"slices"
	"sort"
	"strings"
	"time"
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
	case ObjectTypeTree:
		var entries []TreeEntry

		if err = json.Unmarshal(content, &entries); err != nil {
			return nil, err
		}
		return &Tree{Entries: entries, hash: hash}, nil
	case ObjectTypeCommit:
		var commit Commit

		if err = json.Unmarshal(content, &commit); err != nil {
			return nil, err
		}
		commit.hash = hash
		return &commit, nil
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

type DirNode struct {
	Name     string
	Children map[string]*DirNode
	Files    []IndexEntry
}

func (repo *Repository) CreateTreeFromIndex() (*Tree, error) {
	entries, err := repo.ReadIndex()
	if err != nil {
		return nil, err
	}

	// initialize directory tree structure
	root := &DirNode{Name: "",
		Children: make(map[string]*DirNode),
		Files:    []IndexEntry{},
	}

	// organize files into directory structure
	for _, entry := range entries {
		parts := strings.Split(filepath.ToSlash(entry.Path), "/")
		var current = root

		// navigate directory structure
		for _, part := range parts[:len(parts)-1] {
			if current.Children[part] == nil {
				current.Children[part] = &DirNode{
					Name:     part,
					Children: make(map[string]*DirNode),
					Files:    []IndexEntry{},
				}
			}
			current = current.Children[part]
		}
		current.Files = append(current.Files, entry)
	}
	return repo.createTreeFromDirNode(root)
}

func (repo *Repository) createTreeFromDirNode(root *DirNode) (*Tree, error) {
	var treeEntries []TreeEntry

	for _, file := range root.Files {
		treeEntries = append(treeEntries, TreeEntry{
			Mode: file.Mode,
			Name: filepath.Base(file.Path),
			Hash: file.Hash,
			Type: ObjectTypeBlob,
		})
	}

	// recursively process subdirectories
	var dirNames []string
	for name := range root.Children {
		dirNames = append(dirNames, name)
	}
	sort.Strings(dirNames)

	for _, name := range dirNames {
		child := root.Children[name]

		childTree, err := repo.createTreeFromDirNode(child)
		if err != nil {
			return nil, err
		}
		// store the child tree
		if err := repo.StoreObject(childTree); err != nil {
			return nil, err
		}
		childTreeHash, err := childTree.Hash()
		if err != nil {
			return nil, err
		}
		treeEntries = append(treeEntries, TreeEntry{
			Mode: "040000",
			Name: name,
			Hash: childTreeHash,
			Type: ObjectTypeTree,
		})
	}
	// sort all entries by name (Git practice)
	slices.SortFunc(treeEntries, func(a, b TreeEntry) int {
		return strings.Compare(a.Name, b.Name)
	})
	tree := &Tree{Entries: treeEntries}
	return tree, nil
}

func (repo *Repository) GetHEAD() (string, error) {
	headPath := filepath.Join(repo.GitDir, "HEAD")

	data, err := os.ReadFile(headPath)
	if err != nil {
		return "", err
	}
	head := strings.TrimSpace(string(data))

	if !strings.HasPrefix(head, "ref: ") {
		return head, nil
	}
	refPath := filepath.Join(repo.GitDir, head[5:])

	if _, err := os.Stat(refPath); os.IsNotExist(err) {
		return "", err // no commits yet
	}
	refData, err := os.ReadFile(refPath)
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(refData)), nil
}

func (repo *Repository) UpdateRef(branch, hash string) error {
	refPath := filepath.Join(repo.GitDir, "refs", "heads", branch)
	return os.WriteFile(refPath, []byte(hash+"\n"), 0644)
}

func (repo *Repository) Commit(message, author string) (string, error) {
	tree, err := repo.CreateTreeFromIndex()
	if err != nil {
		return "", err
	}
	parentHash, err := repo.GetHEAD()
	if err != nil {
		return "", err
	}
	treeHash, err := tree.Hash()
	if err != nil {
		return "", err
	}
	commit := &Commit{Tree: treeHash,
		Parent:    parentHash,
		Author:    author,
		Message:   message,
		Timestamp: time.Now(),
	}
	if err = repo.StoreObject(commit); err != nil {
		return "", err
	}
	commitHash, err := commit.Hash()
	if err != nil {
		return "", err
	}
	if err = repo.UpdateRef("main", commitHash); err != nil {
		return "", err
	}
	return commit.Hash()
}
