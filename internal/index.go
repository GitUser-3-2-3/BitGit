package objects

import "time"

type IndexEntry struct {
	Path    string    `json:"path"`
	Hash    string    `json:"hash"`
	Mode    string    `json:"mode"`
	Size    int64     `json:"size"`
	ModTime time.Time `json:"mod_time"`
	Staged  bool      `json:"staged"`
}
