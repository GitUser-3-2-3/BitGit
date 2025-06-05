package objects

import (
	"os"
	"path/filepath"
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
	return nil, nil
}
