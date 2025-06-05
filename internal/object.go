package objects

import (
	"crypto/sha1"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"time"
)

const (
	ObjectTypeBlob   = "blob"
	ObjectTypeTree   = "tree"
	ObjectTypeCommit = "commit"
)

const (
	colorRed   = "\033[31m"
	colorReset = "\033[0m"
)

type IndexEntry struct {
	Path    string    `json:"path"`
	Hash    string    `json:"hash"`
	Mode    string    `json:"mode"`
	Size    int64     `json:"size"`
	ModTime time.Time `json:"mod_time"`
	Staged  bool      `json:"staged"`
}

type GitObject interface {
	Type() string
	Content() ([]byte, error)
	Hash() (string, error)
}

type Blob struct {
	Data []byte
	hash string // cached hash
}

func (blob *Blob) Type() string {
	return ObjectTypeBlob
}

func (blob *Blob) Content() ([]byte, error) { return blob.Data, nil }

func (blob *Blob) Hash() (string, error) {
	if blob.hash == "" {
		content := fmt.Sprintf("blob %d\x00%s", len(blob.Data), blob.Data)
		hash := sha1.Sum([]byte(content))
		blob.hash = hex.EncodeToString(hash[:])
	}
	return blob.hash, nil
}

type TreeEntry struct {
	Mode string `json:"mode"`
	Name string `json:"name"`
	Hash string `json:"hash"`
	Type string `json:"type"`
}

type Tree struct {
	Entries []TreeEntry `json:"entries"`
	hash    string
}

func (root *Tree) Type() string { return ObjectTypeTree }

func (root *Tree) Content() ([]byte, error) {
	data, err := json.Marshal(root.Entries)
	if err != nil {
		return nil, err
	}
	return data, nil
}

func (root *Tree) Hash() (string, error) {
	if root.hash == "" {
		content, err := root.Content()
		if err != nil {
			return "", err
		}
		hashStr := fmt.Sprintf("tree %d\x00%s", len(content), content)
		hash := sha1.Sum([]byte(hashStr))
		root.hash = hex.EncodeToString(hash[:])
	}
	return root.hash, nil
}

type Commit struct {
	Tree      string    `json:"tree"`
	Parent    string    `json:"parent,omitempty"`
	Author    string    `json:"author,omitempty"`
	Message   string    `json:"message"`
	Timestamp time.Time `json:"timestamp"`
	hash      string
}

func (commit *Commit) Type() string { return ObjectTypeCommit }

func (commit *Commit) Content() ([]byte, error) {
	data, err := json.Marshal(commit)
	if err != nil {
		return nil, err
	}
	return data, nil
}

func (commit *Commit) Hash() (string, error) {
	if commit.hash == "" {
		content, err := commit.Content()
		if err != nil {
			return "", err
		}
		hashStr := fmt.Sprintf("commit %d\x00%s", len(content), content)
		hash := sha1.Sum([]byte(hashStr))
		commit.hash = hex.EncodeToString(hash[:])
	}
	return commit.hash, nil
}
