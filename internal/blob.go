package objects

import (
	"crypto/sha1"
	"encoding/hex"
	"fmt"
)

type Blob struct {
	Data []byte
	hash string // cached hash
}

func (b *Blob) Type() string {
	return ObjectTypeBlob
}

func (b *Blob) Content() []byte {
	return b.Data
}

func (b *Blob) Hash() string {
	if b.hash == "" {
		content := fmt.Sprintf("blob %d\x00%s", len(b.Data), b.Data)
		hash := sha1.Sum([]byte(content))
		b.hash = hex.EncodeToString(hash[:])
	}
	return b.hash
}
